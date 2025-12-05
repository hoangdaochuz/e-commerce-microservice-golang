package apigateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/configs"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/circuitbreaker"
	custom_nats "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/custom-nats"
	di "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/dependency-injection"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/logging"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/metric"
	ratelimiter "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/rate_limiter"
	redis_pkg "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/redis"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

type APIGateway struct {
	natsConn *nats.Conn
	timeout  time.Duration
	server   *http.Server
	mux      *http.ServeMux
	ctx      context.Context
	// ctx      context.Context
}

func NewAPIGateway(natsConn *nats.Conn, server *http.Server, mux *http.ServeMux, ctx context.Context) *APIGateway {
	gateway := &APIGateway{
		natsConn: natsConn,
		timeout:  30 * time.Second,
		server:   server,
		mux:      mux,
		ctx:      ctx,
	}
	return gateway
}

func (gw *APIGateway) sendErrorResponse(w http.ResponseWriter, err string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errResponse := map[string]string{
		"error": err,
	}
	json.NewEncoder(w).Encode(errResponse)
}

func (gw *APIGateway) writeResponse(w http.ResponseWriter, response custom_nats.Response) {
	// Set default content type

	// Copy headers from response but skip Content-Length
	for key, items := range response.Headers {
		// Skip headers that should be handled by Go HTTP server
		if key == "Content-Length" || key == "Transfer-Encoding" {
			continue
		}
		for _, v := range items {
			w.Header().Add(key, v)
		}
	}
	w.WriteHeader(response.StatusCode)
	_, err := w.Write(response.Body)
	if err != nil {
		logging.GetSugaredLogger().Errorf("fail to write response body: %v", err)
	}
}

func (gw *APIGateway) ServeHTTP() func(http.ResponseWriter, *http.Request) {
	// Here is entry point for api gateway
	return func(w http.ResponseWriter, r *http.Request) {

		natsReq, err := custom_nats.HttpRequestToNatsRequest(*r)
		if err != nil {
			gw.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Add necessary infomation to header for updating to context and use it if we need
		natsReq.AddHeader("X-User-Id", "test@1234")
		// continue add if we want

		serviceName := natsReq.GetServiceName()
		circuitBreakerConfigService := configs.LoadNatsCircuitBreakerConfigByServiceName(serviceName)
		cbRegistry := circuitbreaker.GetRegistry[*nats.Msg]()
		breaker, err := cbRegistry.GetOrCreateBreaker(serviceName, circuitbreaker.ToCircuitBreakerConfig(serviceName, circuitBreakerConfigService))
		if err != nil {
			gw.sendErrorResponse(w, fmt.Errorf("fail to get or create a circuit breaker: %w", err).Error(), http.StatusInternalServerError)
			return
		}

		natsConnWithCircuitBreakerWrapper := custom_nats.NewNatsConnWithCircuitBreaker(gw.natsConn, breaker)

		// natsSubject := natsReq.Subject
		natsReqByte, err := json.Marshal(*natsReq)
		if err != nil {
			gw.sendErrorResponse(w, fmt.Errorf("fail to marshal nats request: %w", err).Error(), http.StatusInternalServerError)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), gw.timeout)
		defer cancel()
		natsSendRequest := &custom_nats.NatsSendRequest{
			Subject: natsReq.Subject,
			Content: natsReqByte,
		}
		msgResponse, err := natsConnWithCircuitBreakerWrapper.SendRequest(ctx, natsSendRequest)
		if err != nil {
			gw.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var natsResponse custom_nats.Response
		err = json.Unmarshal(msgResponse.Data, &natsResponse)
		if err != nil {
			gw.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if strings.Contains(r.URL.Path, "Logout") {
			// clear cookie
			http.SetCookie(w, &http.Cookie{
				Name:     viper.GetString("zitadel_configs.cookie_name"),
				Path:     "/",
				Domain:   "",
				MaxAge:   -1,
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteNoneMode,
			})
		}

		gw.writeResponse(w, natsResponse)
	}
}

func useMiddleware(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	return MiddlewareChain(handler, middlewares...)
}

func (gw *APIGateway) Start() error {
	logging.GetSugaredLogger().Infof("Starting API Gateway in port %s", gw.server.Addr)
	_, err := configs.Load()
	if err != nil {
		logging.GetSugaredLogger().Fatalf("failed to load configuration: %v", err)
		return err
	}

	var redisClient *redis.Client
	di.Resolve(func(redisPkg *redis_pkg.Redis) {
		redisClient = redisPkg.GetClient()
	})
	defer redisClient.Close()
	rateLimiter := ratelimiter.NewRateLimiter(redisClient, 50, 1*time.Minute, gw.ctx)

	registryWrapper := metric.NewMetricWrapper()
	registryWrapper.RegisterCollectorDefault()
	registry := registryWrapper.GetRegistry()

	rootHandler := http.HandlerFunc(gw.ServeHTTP())
	_ = useMiddleware(rootHandler, CorsMiddleware, ContentTypeMiddleware, RateLimitMiddleware(rateLimiter), LoggingMiddleware)
	protectResourceHandler := useMiddleware(rootHandler, CorsMiddleware, ContentTypeMiddleware, RateLimitMiddleware(rateLimiter), MetricMiddleware(registry), AuthMiddleware, LoggingMiddleware)
	healthCheckHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
		})
	})
	healthResourceHanlder := useMiddleware(healthCheckHandler, CorsMiddleware, ContentTypeMiddleware, MetricMiddleware(registry), LoggingMiddleware)
	gw.mux.Handle("/", protectResourceHandler)
	gw.mux.Handle("/health", healthResourceHanlder)
	gw.mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	errChan := make(chan error, 1)

	go func() {
		defer gw.server.Close()
		if err := gw.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.GetSugaredLogger().Errorf("API Gateway listen and serve fail: %v", err)
			errChan <- err
		}
	}()

	err = <-errChan
	gw.Stop()
	logging.GetSugaredLogger().Errorf("api gateway listen and serve fail: %v", err)
	return err
}

func (gw *APIGateway) Stop() error {
	err := gw.server.Close()
	if err != nil {
		return err
	}
	return gw.natsConn.Drain()
}

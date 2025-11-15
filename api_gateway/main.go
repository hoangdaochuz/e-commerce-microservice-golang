package apigateway

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/configs"
	custom_nats "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/custom-nats"
	di "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/dependency-injection"
	ratelimiter "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/rate_limiter"
	redis_pkg "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/redis"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

type APIGateway struct {
	natsConn *custom_nats.NatsConnWithCircuitBreaker
	timeout  time.Duration
	server   *http.Server
	mux      *http.ServeMux
	ctx      context.Context
	// ctx      context.Context
}

func NewAPIGateway(natsConn *custom_nats.NatsConnWithCircuitBreaker, server *http.Server, mux *http.ServeMux, ctx context.Context) *APIGateway {
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
		fmt.Println("fail to write response body: ", err)
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

		// natsSubject := natsReq.Subject
		natsReqByte, err := json.Marshal(*natsReq)
		if err != nil {
			gw.sendErrorResponse(w, fmt.Errorf("fail to marshal nats request: %w", err).Error(), http.StatusInternalServerError)
			return
		}
		ctx, _ := context.WithTimeout(context.Background(), gw.timeout)
		natsSendRequest := &custom_nats.NatsSendRequest{
			Subject: natsReq.Subject,
			Content: natsReqByte,
		}
		msgResponse, err := gw.natsConn.SendRequest(ctx, natsSendRequest)
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
	fmt.Printf("Starting API Gateway in port %s\n", gw.server.Addr)
	_, err := configs.Load()
	if err != nil {
		log.Fatal("failed to load configuration: %w", err)
		return err
	}

	var redisClient *redis.Client
	di.Resolve(func(redisPkg *redis_pkg.Redis) {
		redisClient = redisPkg.GetClient()
	})
	defer redisClient.Close()
	rateLimiter := ratelimiter.NewRateLimiter(redisClient, 50, 1*time.Minute, gw.ctx)
	rootHandler := http.HandlerFunc(gw.ServeHTTP())
	_ = useMiddleware(rootHandler, CorsMiddleware, ContentTypeMiddleware, RateLimitMiddleware(rateLimiter), LoggingMiddleware)
	protectResourceHandler := useMiddleware(rootHandler, CorsMiddleware, ContentTypeMiddleware, RateLimitMiddleware(rateLimiter), AuthMiddleware, LoggingMiddleware)
	healthCheckHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
		})
	})
	healthResourceHanlder := useMiddleware(healthCheckHandler, CorsMiddleware, ContentTypeMiddleware, LoggingMiddleware)
	gw.mux.Handle("/", protectResourceHandler)
	gw.mux.Handle("/health", healthResourceHanlder)
	errChan := make(chan error, 1)

	go func() {
		defer gw.server.Close()
		if err := gw.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("API Gateway listen and serve fail")
			errChan <- err
		}
	}()

	err = <-errChan
	gw.Stop()
	return fmt.Errorf("api gateway listen and serve fail: %w", err)
}

func (gw *APIGateway) Stop() error {
	err := gw.server.Close()
	if err != nil {
		return err
	}
	return gw.natsConn.GetNatsConn().Drain()
}

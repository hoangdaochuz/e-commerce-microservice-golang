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
	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/tracing"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
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
	if encodeErr := json.NewEncoder(w).Encode(errResponse); encodeErr != nil {
		logging.GetSugaredLogger().Errorf("failed to encode error response: %v", encodeErr)
	}
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

func (gw *APIGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Here is entry point for api gateway
	timeoutCtx, cancel := context.WithTimeout(r.Context(), gw.timeout)
	defer cancel()
	ctx, span := tracing.SpanContext(timeoutCtx, r.Header, fmt.Sprintf("outbound request: %s", r.URL.Path))
	defer span.End()
	// r, err := http.NewRequestWithContext(ctx, r.Method, r.URL.Path, r.Body)
	// if err != nil {
	// 	logging.GetSugaredLogger().Panicf("fail to create a http req: %w", err)
	// }
	r = r.WithContext(ctx)
	tracing.InjectTraceIntoHttpReq(ctx, r)

	natsReq, err := custom_nats.HttpRequestToNatsRequest(*r)
	if err != nil {
		gw.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Add necessary information to header for updating to context and use it if we need
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
	natsSendRequest := &custom_nats.NatsSendRequest{
		Subject: natsReq.Subject,
		Content: natsReqByte,
	}
	// logging.GetSugaredLogger().Infof("Sending request ")
	start := time.Now()
	msgResponse, err := natsConnWithCircuitBreakerWrapper.SendRequest(timeoutCtx, natsSendRequest)

	if err != nil {
		// set span attribute error
		tracing.SetSpanError(span, err)
		gw.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		logging.GetSugaredLogger().Errorf("%s %s %v statusCode: %v traceId: %s", r.Method, r.URL.Path, time.Since(start), http.StatusInternalServerError, span.SpanContext().TraceID().String())
		return
	}

	var natsResponse custom_nats.Response
	err = json.Unmarshal(msgResponse.Data, &natsResponse)
	if err != nil {
		tracing.SetSpanError(span, err)
		gw.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		logging.GetSugaredLogger().Errorf("%s %s %v statusCode: %v traceId: %s", r.Method, r.URL.Path, time.Since(start), http.StatusInternalServerError, span.SpanContext().TraceID().String())
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
	span.SetAttributes(
		semconv.HTTPMethod(r.Method),
		semconv.HTTPRoute(r.URL.Path),
		semconv.HTTPScheme(r.URL.Scheme),
		semconv.HTTPStatusCode(natsResponse.StatusCode),
	)
	logging.GetSugaredLogger().Infof("%s %s %v statusCode: %v traceId: %s", r.Method, r.URL.Path, time.Since(start), natsResponse.StatusCode, span.SpanContext().TraceID().String())
	gw.writeResponse(w, natsResponse)
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
	otlpEndpoint := viper.GetString("general_config.otlp_endpoint")
	var redisClient *redis.Client
	_ = di.Resolve(func(redisPkg *redis_pkg.Redis) {
		redisClient = redisPkg.GetClient()
	})
	defer redisClient.Close()
	rateLimiter := ratelimiter.NewRateLimiter(redisClient, 50, 1*time.Minute, gw.ctx)

	registryWrapper := metric.NewMetricWrapper()
	registryWrapper.RegisterCollectorDefault()
	registry := registryWrapper.GetRegistry()

	shutdownTracing, err := tracing.InitializeTraceRegistry(&tracing.TracingConfig{
		ServiceName: "api_gateway",
		// Attributes:,
		SamplingRate: 1,
		BatchTimeout: 5 * time.Second,
		BatchMaxSize: 512,
		OtelEndpoint: otlpEndpoint,
	})
	if err != nil {
		logging.GetSugaredLogger().Error(err)
		return err
	}
	defer shutdownTracing()

	rootHandler := otelhttp.NewHandler(gw, "", otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
		return r.URL.Path
	}))
	_ = useMiddleware(rootHandler, CorsMiddleware, ContentTypeMiddleware, RateLimitMiddleware(rateLimiter))
	protectResourceHandler := useMiddleware(rootHandler, CorsMiddleware, ContentTypeMiddleware, RateLimitMiddleware(rateLimiter), MetricMiddleware(registry), AuthMiddleware)
	healthCheckHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
		})
	})
	healthResourceHanlder := useMiddleware(healthCheckHandler, CorsMiddleware, ContentTypeMiddleware, MetricMiddleware(registry))
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
	_ = gw.Stop()
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

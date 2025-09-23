package apigateway

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/configs"
	custom_nats "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/custom-nats"
	ratelimiter "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/rate_limiter"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
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
		msgResponse, err := gw.natsConn.Request(natsReq.Subject, natsReqByte, gw.timeout)
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

		gw.writeResponse(w, natsResponse)
	}
}

func useMiddleware(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	return MiddlewareChain(handler, middlewares...)
}

func (gw *APIGateway) Start() error {
	fmt.Printf("Starting API Gateway in port %s\n", gw.server.Addr)
	config, err := configs.Load()
	if err != nil {
		log.Fatal("failed to load configuration: %w", err)
		return err
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: config.Redis.Address + ":" + config.Redis.Port,
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
	return gw.natsConn.Drain()
}

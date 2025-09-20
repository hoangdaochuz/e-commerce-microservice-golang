package apigateway

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/configs"
	custom_nats "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/custom-nats"
	di "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/dependency-injection"
	"github.com/nats-io/nats.go"
)

type APIGateway struct {
	natsConn *nats.Conn
	timeout  time.Duration
	server   *http.Server
}

func NewAPIGateway(natsConn *nats.Conn, serviceRegistryReqTimeout time.Duration, server *http.Server) *APIGateway {
	gateway := &APIGateway{
		natsConn: natsConn,
		timeout:  30 * time.Second,
		server:   server,
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

func ServeHTTP(gw *APIGateway) func(http.ResponseWriter, *http.Request) {
	// Here is entry point for api gateway
	return func(w http.ResponseWriter, r *http.Request) {

		natsReq, err := custom_nats.HttpRequestToNatsRequest(*r)
		if err != nil {
			gw.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}

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

func Start(port string) (*APIGateway, error) {
	fmt.Printf("Starting API Gateway in port %s\n", port)
	config, err := configs.Load()
	if err != nil {
		log.Fatal("failed to load configuration: %w", err)
	}
	natsConn, err := nats.Connect(config.NatsAuth.NATSUrl, nats.UserInfo(config.NatsAuth.NATSApps[0].Username, config.NatsAuth.NATSApps[0].Password))
	if err != nil {
		log.Fatal("Failed to connect to nats: ", err)
	}
	// defer natsConn.Drain()
	log.Println("Connected to nats successfully")
	serviceRegistryReqTimout := config.ServiceRegistry.RequestTimeout
	mux := http.NewServeMux()
	apigatewayServer := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	gateway := NewAPIGateway(natsConn, serviceRegistryReqTimout, apigatewayServer)
	di.Make(func() *APIGateway {
		return gateway
	})

	// mux.HandleFunc("/", ServeHTTP(gateway))
	rootHandler := http.HandlerFunc(ServeHTTP(gateway))
	_ = useMiddleware(rootHandler, CorsMiddleware, ContentTypeMiddleware, RateLimitMiddleware, LoggingMiddleware)
	protectResourceHandler := useMiddleware(rootHandler, CorsMiddleware, ContentTypeMiddleware, RateLimitMiddleware, AuthMiddleware, LoggingMiddleware)

	healthCheckHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
		})
	})
	healthResourceHanlder := useMiddleware(healthCheckHandler, CorsMiddleware, ContentTypeMiddleware, LoggingMiddleware)

	mux.Handle("/", protectResourceHandler)
	mux.Handle("/health", healthResourceHanlder)

	errChan := make(chan error, 1)

	go func() {
		defer apigatewayServer.Close()
		if err := apigatewayServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("API Gateway listen and serve fail")
			errChan <- err
		}
	}()
	select {
	case e := <-errChan:
		gateway.Stop()
		return nil, e
	default:
		return gateway, nil
	}
}

func (gw *APIGateway) Stop() error {
	err := gw.server.Close()
	if err != nil {
		return err
	}
	return gw.natsConn.Drain()
}

package apigateway

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"time"

	"github.com/gorilla/mux"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/configs"
	di "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/dependency-injection"
	serviceregistry "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/service-registry"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type APIGateway struct {
	router          *mux.Router
	serviceRegistry *serviceregistry.ServiceRegistry
	serviceRoutes   map[string]ServiceRoute
	natsConn        *nats.Conn
}

type ServiceRoute struct {
	ServiceName string
	BasePath    string
	Methods     map[string]MethodRoute
}

type MethodRoute struct {
	HTTPMethod   string
	Path         string
	MethodName   string
	RequestType  reflect.Type
	ResponseType reflect.Type
}

func (gw *APIGateway) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func (gw *APIGateway) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (gw *APIGateway) contentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func (gw *APIGateway) setupMiddleware() {
	gw.router.Use(gw.loggingMiddleware, gw.corsMiddleware, gw.contentTypeMiddleware)
}

func NewAPIGateway(natsConn *nats.Conn, serviceRegistryReqTimeout time.Duration) *APIGateway {
	gateway := &APIGateway{
		natsConn:        natsConn,
		router:          mux.NewRouter(),
		serviceRegistry: serviceregistry.NewServiceRegistry(natsConn, serviceRegistryReqTimeout),
		serviceRoutes:   map[string]ServiceRoute{},
	}
	//Setup middleware for gateway
	gateway.setupMiddleware()
	gateway.GetServiceAppsAndRegisterRouteMethod()
	return gateway
}

func (gw *APIGateway) RegisterServiceWithAutoRoute(serviceName, basePath string, serviceImpl interface{}) error {

	// register service's method to nats
	if err := gw.serviceRegistry.RegistryOperationOfServices(serviceName, serviceImpl); err != nil {
		return fmt.Errorf("failed to register service %s: %w", serviceName, err)
	}

	// auto config route of each method of service
	routeConfig := serviceregistry.NewAutoRouteConfig(basePath)
	routeConfig.DiscoveryRoutesFromService(serviceImpl, serviceName)

	// register http route
	return gw.RegisterServiceRouteWithConfig(serviceName, routeConfig)
}

func (gw *APIGateway) registerMethodRoute(serviceRoute *ServiceRoute, methodRoute MethodRoute) {
	fullPath := serviceRoute.BasePath + methodRoute.Path

	handler := gw.createMethodHandler(serviceRoute.ServiceName, methodRoute)
	gw.router.HandleFunc(fullPath, handler).Methods(methodRoute.HTTPMethod)
	log.Printf("Registered route: %s %s -> %s.%s\n",
		methodRoute.HTTPMethod, fullPath, serviceRoute.ServiceName, methodRoute.MethodName)
}

func (gw *APIGateway) createMethodHandler(serviceName string, methodRoute MethodRoute) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// convert http request to proto.Message
		protoReq, err := gw.buildRequestMessage(r, methodRoute)
		if err != nil {
			gw.sendErrorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}
		// call target service for this url
		protoRes, errRes := gw.serviceRegistry.CallService(serviceName, methodRoute.MethodName, protoReq)
		if errRes != nil {
			gw.sendErrorResponse(w, errRes.Err, errRes.StatusCode)
			return
		}

		if protoRes == nil {
			gw.sendErrorResponse(w, "response not found", http.StatusNotFound)
		}

		// Convert protoRes (proto.Message) to HTTP Response json
		gw.sendSuccessResponse(w, protoRes)
	}
}

func (gw *APIGateway) sendErrorResponse(w http.ResponseWriter, err string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errResponse := map[string]string{
		"error": err,
	}
	json.NewEncoder(w).Encode(errResponse)
}

func (gw *APIGateway) sendSuccessResponse(w http.ResponseWriter, protoResponse proto.Message) {
	w.Header().Set("Content-Type", "application/json")
	// parse proto message to json
	protoDataByte, err := protojson.Marshal(protoResponse)
	if err != nil {
		gw.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(protoDataByte)
	w.WriteHeader(http.StatusOK)
}

func (gw *APIGateway) buildRequestMessage(r *http.Request, methodRoute MethodRoute) (proto.Message, error) {
	reqPtr := reflect.New(methodRoute.RequestType.Elem())
	req := reqPtr.Interface().(proto.Message)

	// always POST method
	return gw.parseBodyRequest(r, req)
}

func (gw *APIGateway) parseBodyRequest(r *http.Request, protoReq proto.Message) (proto.Message, error) {

	vars := mux.Vars(r)

	// parse r.Body to json
	var jsonBody map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&jsonBody); err != nil {
		return nil, fmt.Errorf("fail to decode json body: %w", err)
	}
	// merge path variable to body json
	for key, val := range vars {
		jsonBody[key] = val
	}
	// convert json body data to proto.Message
	jsonByteData, err := json.Marshal(jsonBody)
	if err != nil {
		return nil, fmt.Errorf("faild to marshal json body")
	}
	err = protojson.Unmarshal(jsonByteData, protoReq)
	if err != nil {
		return nil, fmt.Errorf("faild to marshal to proto Message")
	}
	return protoReq, nil
}

func (gw *APIGateway) RegisterServiceRouteWithConfig(serviceName string, routeConfig serviceregistry.ServiceRouteConfig) error {
	serviceRoute := ServiceRoute{
		ServiceName: serviceName,
		BasePath:    routeConfig.GetBasePath(),
		Methods:     make(map[string]MethodRoute),
	}

	for methodName, method := range routeConfig.GetRoutes() {
		methodRoute := MethodRoute{
			HTTPMethod:   string(method.HTTPMethod),
			Path:         method.PathPattern,
			MethodName:   methodName,
			RequestType:  method.RequestType,
			ResponseType: method.ResponseType,
		}
		serviceRoute.Methods[methodName] = methodRoute
		gw.registerMethodRoute(&serviceRoute, methodRoute)
	}
	gw.serviceRoutes[serviceName] = serviceRoute
	return nil
}

// GetServiceRoutes returns the service routes for code generation
func (gw *APIGateway) GetServiceRoutes() map[string]ServiceRoute {
	return gw.serviceRoutes
}

func Start(port string) error {
	fmt.Printf("Starting API Gateway in port %s\n", port)
	config, err := configs.Load()
	if err != nil {
		log.Fatal("failed to load configuration: %w", err)
	}
	natsUrl := fmt.Sprintf("nats://%s:%s@localhost:4222", config.ServiceRegistry.NATSUser, config.ServiceRegistry.NATSPassword)
	natsConn, err := nats.Connect(natsUrl)
	if err != nil {
		log.Fatal("Failed to connect to nats")
	}
	log.Println("Connected to nats successfully")
	serviceRegistryReqTimout := config.ServiceRegistry.RequestTimeout
	gateway := NewAPIGateway(natsConn, serviceRegistryReqTimout)
	di.Make(func() *APIGateway {
		return gateway
	})
	return http.ListenAndServe(":"+port, gateway.router)
}

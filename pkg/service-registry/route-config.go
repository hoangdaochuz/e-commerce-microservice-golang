package serviceregistry

import (
	"fmt"
	"reflect"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/service-registry/utils"
)

type HTTPMethod string

const (
	POST HTTPMethod = "POST"
)

// RouteConfig: Define route config for a service method
type RouteConfig struct {
	HTTPMethod   HTTPMethod   // POST,PUT,GET,DELETE,...
	PathPattern  string       // /users/{id}
	RequestType  reflect.Type // proto.Message
	ResponseType reflect.Type //proto.Message
}

// a service will have many method, each methods will have a route config
type ServiceRouteConfig interface {
	GetBasePath() string
	GetRoutes() map[string]RouteConfig
}

// auto route config for each service
type AutoRouteConfig struct {
	BasePath string
	Methods  map[string]RouteConfig
}

func NewAutoRouteConfig(basePath string) *AutoRouteConfig {
	return &AutoRouteConfig{
		BasePath: basePath,
		Methods:  make(map[string]RouteConfig),
	}
}
func (arc *AutoRouteConfig) GetBasePath() string {
	return arc.BasePath
}
func (arc *AutoRouteConfig) GetRoutes() map[string]RouteConfig {
	return arc.Methods
}

func (arc *AutoRouteConfig) DiscoveryRoutesFromService(serviceImpl interface{}) {
	serviceType := reflect.TypeOf(serviceImpl)
	serviceName := serviceType.Name()
	for i := 0; i < serviceType.NumMethod(); i++ {
		method := serviceType.Method(i)
		if utils.IsValidServiceMethod(method) {
			routeConfig := arc.generateRouteFromMethod(method, serviceName)
			arc.Methods[method.Name] = routeConfig
		}
	}
}

func (arc *AutoRouteConfig) generateRouteFromMethod(method reflect.Method, serviceName string) RouteConfig {
	reqType := method.Type.In(2)
	resType := method.Type.Out(0)

	return RouteConfig{
		HTTPMethod:   POST,
		PathPattern:  fmt.Sprintf("/%s/%s", serviceName, method.Name),
		RequestType:  reqType,
		ResponseType: resType,
	}
}

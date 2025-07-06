package apigateway

import (
	"github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/handler/order"
	di "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/dependency-injection"
)

func (gw *APIGateway) GetServiceAppsAndRegisterRouteMethod() {
	di.Resolve(func(order *order.OrderServiceApp) {
		gw.RegisterServiceWithAutoRoute("order", "/api/v1", order)
	})
}

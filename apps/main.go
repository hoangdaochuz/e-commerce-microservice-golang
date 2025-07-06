package apps

import (
	// apigateway "github.com/hoangdaochuz/ecommerce-microservice-golang/api_gateway"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/handler/order"
	"go.uber.org/dig"
)

func GetServiceAppsAndRegisterRouteMethod() {
	c := dig.New()
	c.Invoke(func(order *order.OrderServiceApp) {
		// gw.RegisterServiceWithAutoRoute("order", "/api/v1", order)
	})
}

package order

import (
	"context"

	custom_nats "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/custom-nats"
)

const (
	NATS_SUBJECT          = "/api/v1/order"
	ORDER_CREATE_ORDER    = NATS_SUBJECT + "/CreateOrder"
	ORDER_GET_ORDER_BY_ID = NATS_SUBJECT + "/GetOrderById"
)

type OrderService interface {
	CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error)
	GetOrderById(ctx context.Context, req *GetOrderByIdRequest) (*OrderResponse, error)
}

type OrderServiceProxy struct {
	service OrderService
}

func NewOrderServiceProxy(service OrderService) *OrderServiceProxy {
	return &OrderServiceProxy{
		service: service,
	}
}

func (o *OrderServiceProxy) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	return o.service.CreateOrder(ctx, req)
}

func (o *OrderServiceProxy) GetOrderById(ctx context.Context, req *GetOrderByIdRequest) (*OrderResponse, error) {
	return o.service.GetOrderById(ctx, req)
}

type OrderServiceRouter struct {
	proxy *OrderServiceProxy
}

func NewOrderServiceRouter(proxy *OrderServiceProxy) *OrderServiceRouter {
	return &OrderServiceRouter{
		proxy: proxy,
	}
}

func (o *OrderServiceRouter) Register(natsRouter custom_nats.Router) {
	natsRouter.RegisterRoute("POST", ORDER_CREATE_ORDER, o.proxy.CreateOrder)
	natsRouter.RegisterRoute("POST", ORDER_GET_ORDER_BY_ID, o.proxy.GetOrderById)
}

package order

import (
	"context"
	"fmt"
	"time"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/api/order"
	di "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/dependency-injection"
)

type OrderServiceApp struct {
	order.UnimplementedOrderServiceServer
	// order_service_layer here
	// other service
}

func NewOrderServiceApp() *OrderServiceApp {
	return &OrderServiceApp{}
}

var _ = di.Make(NewOrderServiceApp)

func (o *OrderServiceApp) CreateOrder(ctx context.Context, req *order.CreateOrderRequest) (*order.CreateOrderResponse, error) {

	customerId := req.GetCustomerId()

	if customerId == "" {
		return nil, fmt.Errorf("customer_id is required")
	}

	orderId := fmt.Sprintf("order_%s_%d", customerId, time.Now().Unix())

	fmt.Printf("Order created successfully for customer: %s, order_id: %s\n", customerId, orderId)

	return &order.CreateOrderResponse{
		OrderId: orderId,
	}, nil
}

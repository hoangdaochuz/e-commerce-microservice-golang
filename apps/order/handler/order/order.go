package order

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/api/order"
	order_repository "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/repository"
	order_service "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/services/order"
	di "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/dependency-injection"
)

type OrderServiceApp struct {
	order.UnimplementedOrderServiceServer
	service *order_service.OrderService
	// order_service_layer here
	// other service
}

func NewOrderServiceApp(orderService *order_service.OrderService) *OrderServiceApp {
	return &OrderServiceApp{
		service: orderService,
	}
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

func (o *OrderServiceApp) toOrder(item order_repository.Order) *order.OrderResponse {
	return &order.OrderResponse{
		Id:   item.ID.String(),
		Name: item.Name,
	}
}

func (o *OrderServiceApp) GetOrderById(ctx context.Context, req *order.GetOrderByIdRequest) (*order.OrderResponse, error) {
	order, err := o.service.GetOrderById(ctx, &order_service.GetOrderByIdRequest{
		Id: uuid.MustParse(req.Id),
	})
	fmt.Println("hello ", order)
	fmt.Println("err: ", err)
	if err != nil {
		return nil, err
	}
	if order == nil {
		fmt.Println("kakaka")
		return nil, fmt.Errorf("order not found")
	}
	fmt.Println("What the fuck")
	return o.toOrder(*order), nil
}

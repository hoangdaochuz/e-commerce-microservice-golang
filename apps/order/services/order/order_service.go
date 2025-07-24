package order_service

import (
	"context"

	"github.com/google/uuid"
	order_repository "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/repository"
	di "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/dependency-injection"
)

type OrderService struct {
	OrderRepo *order_repository.OrderRepository
}

var OrderServiceMod = di.Make(NewOrderService)

func NewOrderService(repo *order_repository.OrderRepository) *OrderService {
	return &OrderService{
		OrderRepo: repo,
	}
}

type GetOrderByIdRequest struct {
	Id uuid.UUID
}

func (o *OrderService) GetOrderById(ctx context.Context, req *GetOrderByIdRequest) (*order_repository.Order, error) {
	entity, err := o.OrderRepo.FindOrderById(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if entity == nil {
		return nil, nil
	}
	return entity, nil
}

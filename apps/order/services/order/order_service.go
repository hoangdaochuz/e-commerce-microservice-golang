package order_service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	order_repository "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/repository"
	di "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/dependency-injection"
)

// OrderServiceInterface defines the contract for order service operations
type OrderServiceInterface interface {
	GetOrderById(ctx context.Context, req *GetOrderByIdRequest) (*order_repository.Order, error)
}

type OrderService struct {
	OrderRepo order_repository.OrderRepositoryInterface
}

// Ensure OrderService implements OrderServiceInterface
var _ OrderServiceInterface = (*OrderService)(nil)

var OrderServiceMod = di.Make[*OrderService](NewOrderService)

func NewOrderService(repo *order_repository.OrderRepository) *OrderService {
	return &OrderService{
		OrderRepo: repo,
	}
}

// NewOrderServiceWithRepo creates an OrderService with a custom repository (for testing)
func NewOrderServiceWithRepo(repo order_repository.OrderRepositoryInterface) *OrderService {
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
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return entity, nil
}

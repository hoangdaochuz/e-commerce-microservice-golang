package order_repository

import (
	"context"

	"github.com/google/uuid"
	order_configs "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/configs"
	di "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/dependency-injection"
	repo_pkg "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/repo"
	postgres "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/repo/postgres_sqlx"
)

type OrderRepository struct {
	repo repo_pkg.Repo[*Order]
}

var OrderRepositoryMod = di.Make(NewOrderRepository)

func NewOrderRepository() *OrderRepository {
	orderDb := order_configs.NewOrderDatabase()

	dbClient := postgres.NewPostgresDBClient(orderDb.Conn)
	return &OrderRepository{
		repo: repo_pkg.NewRepo[*Order](dbClient),
	}
}

func (o *OrderRepository) FindOrderById(ctx context.Context, id uuid.UUID) (*Order, error) {
	entity := &Order{}
	err := o.repo.FindOne(ctx, entity, `SELECT * FROM "order" WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (o *OrderRepository) CreateOrderWithTransaction(ctx context.Context, data Order, other ...interface{}) (interface{}, error) {
	handler := func(ctx context.Context, others ...interface{}) (interface{}, error) {
		// query :=
		// handler create order here
		return nil, nil
	}

	return o.repo.WithTransaction(ctx, handler, other...)
}

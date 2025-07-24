package order_repository

import (
	"context"

	"github.com/google/uuid"
	order_configs "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/configs"
	di "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/dependency-injection"
	repo_pkg "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/repo"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/repo/postgres"
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

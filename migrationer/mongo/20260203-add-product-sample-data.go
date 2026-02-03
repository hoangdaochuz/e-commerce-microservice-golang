package migrationer

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	product_repository "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/product/repository"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/repo"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/repo/mongo"
)

type AddProductSampleData struct {
}

func (p *AddProductSampleData) Name() string {
	return "AddProductSampleData"
}

func (p *AddProductSampleData) Up(ctx context.Context, conn repo.IDBConnection) error {
	mongoConn, ok := conn.(*mongo.MongoConnection)
	if !ok {
		return fmt.Errorf("Connection is invalid for mongodb")
	}

	mongoClient := mongo.NewMongoDBClient(mongoConn, "product_service", "products")
	id := uuid.New()
	sampleProduct := product_repository.Product{
		Name:  "Clear Man",
		Price: 50000,
		ID:    &id,
	}
	err := mongoClient.Insert(ctx, sampleProduct)
	if err != nil {
		return err
	}
	return nil
}

func (p *AddProductSampleData) Down(ctx context.Context, conn repo.IDBConnection) error {
	return nil
}

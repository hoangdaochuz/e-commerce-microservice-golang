package main

import (
	"context"
	"fmt"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/configs"
	migrationer "github.com/hoangdaochuz/ecommerce-microservice-golang/migrationer/mongo"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/logging"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/migration"
	mongo_repo "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/repo/mongo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
)

func main() {
	ctx := context.Background()
	config, err := configs.Load()
	if err != nil {
		panic("Cannot load config value")
	}

	mongoDSN := fmt.Sprintf("mongodb://%s:%s@%s:%s/", config.MongoDatabase.User, config.MongoDatabase.Password, config.MongoDatabase.Host, config.MongoDatabase.Port)
	client, err := mongo.Connect(ctx,
		options.Client().ApplyURI(mongoDSN),
		options.Client().SetMonitor(otelmongo.NewMonitor()),
		options.Client().SetRegistry(mongo_repo.NewRegistryWithUUID()),
	)
	if err != nil {
		panic("Cannot connect to MongoDB: " + err.Error())
	}
	mongoDBConn := &mongo_repo.MongoConnection{
		Client: client,
	}

	mongodb := mongo_repo.NewMongoDBClient(mongoDBConn, "e-commerce-migration", "migration")

	mongoMigrationer := migration.NewMigrationRunner(ctx, mongodb)
	err = mongoMigrationer.Register(&migrationer.AddProductSampleData{})
	if err != nil {
		panic(err)
	}
	err = mongoMigrationer.Run(migration.BOTH)
	if err != nil {
		panic(err)
	}
	logging.GetSugaredLogger().Infof("Migrate all migration successfully")
}

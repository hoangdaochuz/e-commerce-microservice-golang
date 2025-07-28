package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
)

type MongoConnection struct {
	client *mongo.Client
}

func (m *MongoConnection) Connect(connectionString string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx,
		options.Client().ApplyURI(connectionString),
		options.Client().SetMonitor(otelmongo.NewMonitor()),
		options.Client().SetRegistry(bson.NewRegistry()),
	)

	if err != nil {
		return err
	}
	m.client = client
	return nil
}

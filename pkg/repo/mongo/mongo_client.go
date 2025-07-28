package mongo

import (
	"context"
	"fmt"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/repo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBClient struct {
	conn           *MongoConnection
	databaseName   string
	collectionName string
	db             *mongo.Database
}

var (
	Field_ID = "_id"
)

// BulkCreate implements repo.IDBClient.
func (m *MongoDBClient) BulkCreate(ctx context.Context, query interface{}, data []interface{}, out interface{}, others ...interface{}) error {
	result, err := m.db.Collection(m.collectionName).InsertMany(ctx, data, options.InsertMany().SetOrdered(true))
	if err != nil {
		return err
	}
	insertIds := result.InsertedIDs

	if out != nil {
		filter := bson.M{
			Field_ID: bson.M{
				"$in": insertIds,
			},
		}
		cursor, err := m.db.Collection(m.collectionName).Find(ctx, filter)
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)
		return cursor.All(ctx, out)
	}
	return nil
}

// Count implements repo.IDBClient.
func (m *MongoDBClient) Count(ctx context.Context, query interface{}, others ...interface{}) (int, error) {
	result, err := m.db.Collection(m.collectionName).CountDocuments(ctx, query)

	if err != nil {
		return -1, err
	}
	return int(result), nil
}

// Create implements repo.IDBClient.
func (m *MongoDBClient) Create(ctx context.Context, data interface{}, out repo.BaseModel, others ...interface{}) error {
	inserResult, err := m.db.Collection(m.collectionName).InsertOne(ctx, data)
	if err != nil {
		return err
	}
	id := inserResult.InsertedID
	if out != nil {
		filter := bson.M{
			Field_ID: id,
		}
		return m.FindOne(ctx, out, filter)
	}
	return nil
}

// Delete implements repo.IDBClient.
func (m *MongoDBClient) Delete(ctx context.Context, query interface{}, others ...interface{}) error {
	if query == nil {
		return fmt.Errorf("query must not be nil")
	}
	_, err := m.db.Collection(m.collectionName).DeleteMany(ctx, query)
	if err != nil {
		return err
	}
	return nil
}

// FindAll implements repo.IDBClient.
func (m *MongoDBClient) FindAll(ctx context.Context, out, query interface{}, others ...interface{}) error {
	if query == nil {
		return fmt.Errorf("query must not be nil")
	}
	cursor, err := m.db.Collection(m.collectionName).Find(ctx, query)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, out)
	return err
}

// FindOne implements repo.IDBClient.
func (m *MongoDBClient) FindOne(ctx context.Context, out repo.BaseModel, query interface{}, others ...interface{}) error {
	if query == nil {
		return fmt.Errorf("query must not be nil")
	}
	err := m.db.Collection(m.collectionName).FindOne(ctx, query).Decode(out)
	if err != nil {
		return err
	}
	return nil
}

// // Paginate implements repo.IDBClient.
// func (m *MongoDBClient) Paginate(ctx context.Context, destColl string, paginationParams repo.PaginationRequest, others ...interface{}) (*repo.Pagination, error) {
// 	panic("unimplemented")
// }

// PaginateV2 implements repo.IDBClient.
func (m *MongoDBClient) Paginate(ctx context.Context, query, out interface{}, paginationParams repo.PaginationRequest, others ...interface{}) (*repo.Pagination, error) {
	if query == nil {
		return nil, fmt.Errorf("query must not be nil")
	}
	cursor, err := m.db.Collection(m.collectionName).Find(
		ctx,
		query,
		options.Find().SetLimit(int64(paginationParams.Limit)).SetSkip(int64((paginationParams.Page-1)*paginationParams.Limit)),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, out)
	if err != nil {
		return nil, err
	}
	total, err := m.Count(ctx, query, others...)
	if err != nil {
		return nil, err
	}
	return &repo.Pagination{
		Total: total,
		Limit: paginationParams.Limit,
	}, nil
}

// UpdateMany implements repo.IDBClient.
func (m *MongoDBClient) UpdateMany(ctx context.Context, filter, update interface{}, out interface{}, others ...interface{}) error {
	if filter == nil || update == nil {
		return fmt.Errorf("filter and update must not be nil")
	}
	result, err := m.db.Collection(m.collectionName).UpdateMany(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("no documents matched the filter")
	}
	if out != nil {
		err = m.FindAll(ctx, out, filter)
		if err != nil {
			return err
		}
	}
	return nil
}

// UpdateOneAndReturn implements repo.IDBClient.
func (m *MongoDBClient) UpdateOneAndReturn(ctx context.Context, query, update interface{}, out repo.BaseModel, others ...interface{}) error {
	if query == nil || update == nil {
		return fmt.Errorf("query and update must not be nil")
	}
	err := m.db.Collection(m.collectionName).FindOneAndUpdate(ctx, query, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(out)
	return err
}

// Upsert implements repo.IDBClient.
func (m *MongoDBClient) Upsert(ctx context.Context, filter, update interface{}, out repo.BaseModel, others ...interface{}) error {
	if filter == nil || update == nil {
		return fmt.Errorf("filter and update must not be nil")
	}
	_, err := m.db.Collection(m.collectionName).UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}
	if out != nil {
		err = m.FindOne(ctx, out, filter)
		if err != nil {
			return err
		}
	}
	return nil
}

// WithTransaction implements repo.IDBClient.
func (m *MongoDBClient) WithTransaction(ctx context.Context, fn func(ctx context.Context, others ...interface{}) (interface{}, error), others ...interface{}) (interface{}, error) {
	session, err := m.conn.client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)
	err = session.StartTransaction()
	if err != nil {
		return nil, err
	}
	fun := func(sessionContext mongo.SessionContext) (interface{}, error) {
		return fn(sessionContext, others...)
	}
	result, err := session.WithTransaction(ctx, fun)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func NewMongoDBClient(conn *MongoConnection, databaseName, collectionName string) repo.IDBClient {
	return &MongoDBClient{
		conn:           conn,
		databaseName:   databaseName,
		collectionName: collectionName,
		db:             conn.client.Database(databaseName),
	}
}

package repo

import (
	"context"
)

type IDBClient interface {
	FindAll(ctx context.Context, out []BaseModel, query interface{}, others ...interface{}) error
	FindOne(ctx context.Context, out BaseModel, query interface{}, others ...interface{}) error
	// comming soon
}

type Repo[model BaseModel] struct {
	IDBClient IDBClient
}

func NewRepo[model BaseModel](idbClient IDBClient) Repo[model] {
	return Repo[model]{
		IDBClient: idbClient,
	}
}

func (r *Repo[model]) FindAll(ctx context.Context, out []BaseModel, query interface{}, others ...interface{}) error {
	return r.IDBClient.FindAll(ctx, out, query, others...)
}

func (r *Repo[model]) FindOne(ctx context.Context, out BaseModel, query interface{}, others ...interface{}) error {
	return r.IDBClient.FindOne(ctx, out, query, others...)
}

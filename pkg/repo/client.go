package repo

import (
	"context"
)

type IDBClient interface {
	Create(ctx context.Context, query interface{}, data BaseModel, others ...interface{}) error
	BulkCreate(ctx context.Context, query interface{}, data []interface{}, out interface{}, others ...interface{}) error
	Upsert(ctx context.Context, filter, update interface{}, out BaseModel, others ...interface{}) error
	UpdateOneAndReturn(ctx context.Context, query, update interface{}, out BaseModel, others ...interface{}) error
	UpdateMany(ctx context.Context, filter, update, out interface{}, others ...interface{}) error
	FindAll(ctx context.Context, out, query interface{}, others ...interface{}) error
	FindOne(ctx context.Context, out BaseModel, query interface{}, others ...interface{}) error
	Delete(ctx context.Context, query interface{}, others ...interface{}) error
	Count(ctx context.Context, query interface{}, others ...interface{}) (int, error)
	Paginate(ctx context.Context, query, out interface{}, paginationParams PaginationRequest, others ...interface{}) (*Pagination, error)
	WithTransaction(ctx context.Context, fn func(ctx context.Context, others ...interface{}) (interface{}, error), others ...interface{}) (interface{}, error)
	// coming soon
}

type Repo[model BaseModel] struct {
	IDBClient IDBClient
}

type Pagination struct {
	Total int
	Limit int
	// Page  int
	Data []BaseModel
}

type QueryParams struct {
	Key      string
	Value    interface{}
	Operator string
}

type PaginationRequest struct {
	Limit int
	Page  int
	Query []QueryParams
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

func (r *Repo[model]) Create(ctx context.Context, query interface{}, data BaseModel, others ...interface{}) error {
	return r.IDBClient.Create(ctx, query, data, others...)
}

func (r *Repo[model]) BulkCreate(ctx context.Context, query interface{}, data []interface{}, out interface{}, others ...interface{}) error {
	return r.IDBClient.BulkCreate(ctx, query, data, out, others...)
}

func (r *Repo[model]) Upsert(ctx context.Context, filter, update interface{}, out BaseModel, others ...interface{}) error {
	return r.IDBClient.Upsert(ctx, filter, update, out, others...)
}

func (r *Repo[model]) UpdateOneAndReturn(ctx context.Context, query interface{}, data, out BaseModel, others ...interface{}) error {
	return r.IDBClient.UpdateOneAndReturn(ctx, query, data, out, others...)
}

func (r *Repo[model]) UpdateMany(ctx context.Context, filter, update, out interface{}, others ...interface{}) error {
	return r.IDBClient.UpdateMany(ctx, filter, update, out, others...)
}

func (r *Repo[model]) Delete(ctx context.Context, query interface{}, others ...interface{}) error {
	return r.IDBClient.Delete(ctx, query, others...)
}

func (r *Repo[model]) Count(ctx context.Context, query interface{}, others ...interface{}) (int, error) {
	return r.IDBClient.Count(ctx, query, others...)
}

func (r *Repo[model]) WithTransaction(ctx context.Context, fn func(ctx context.Context, others ...interface{}) (interface{}, error), others ...interface{}) (interface{}, error) {
	return r.IDBClient.WithTransaction(ctx, fn, others...)
}

// func (r *Repo[model]) Paginate(ctx context.Context, destColl string, paginationParams PaginationRequest, others ...interface{}) (*Pagination, error) {
// 	return r.IDBClient.Paginate(ctx, destColl, paginationParams, others...)
// }

func (r *Repo[model]) Paginate(ctx context.Context, query interface{}, out []BaseModel, paginationParams PaginationRequest, others ...interface{}) (*Pagination, error) {
	return r.IDBClient.Paginate(ctx, query, out, paginationParams, others...)
}

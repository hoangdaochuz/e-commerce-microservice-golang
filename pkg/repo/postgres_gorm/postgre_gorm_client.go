package postgres_gorm

import (
	"context"
	"fmt"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/repo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PostgresGormClient struct {
	conn *PostgresGormConnection
}

//NOTE: Query will a struct Model to DB

// BulkCreate implements repo.IDBClient.
func (p *PostgresGormClient) BulkCreate(ctx context.Context, query interface{}, data []interface{}, out interface{}, others ...interface{}) error {
	return p.conn.Db.WithContext(ctx).Create(data).Error
}

// Count implements repo.IDBClient.
func (p *PostgresGormClient) Count(ctx context.Context, query interface{}, others ...interface{}) (int, error) {
	if query == nil {
		return -1, fmt.Errorf("query must not be nil")
	}

	var total int64
	err := p.conn.Db.WithContext(ctx).Where(query).Count(&total).Error
	if err != nil {
		return -1, err
	}
	return int(total), nil
}

// Create implements repo.IDBClient.
func (p *PostgresGormClient) Create(ctx context.Context, query interface{}, data repo.BaseModel, others ...interface{}) error {
	return p.conn.Db.WithContext(ctx).Create(data).Error
}

// Delete implements repo.IDBClient.
func (p *PostgresGormClient) Delete(ctx context.Context, query interface{}, others ...interface{}) error {

	if query == nil {
		return fmt.Errorf("query must not be nil")
	}
	return p.conn.Db.WithContext(ctx).Delete(query).Error
}

// FindAll implements repo.IDBClient.
func (p *PostgresGormClient) FindAll(ctx context.Context, out interface{}, query interface{}, others ...interface{}) error {
	if query == nil {
		return fmt.Errorf("query must not be nil")
	}
	return p.conn.Db.WithContext(ctx).Where(query).Find(out).Error
}

// FindOne implements repo.IDBClient.
func (p *PostgresGormClient) FindOne(ctx context.Context, out repo.BaseModel, query interface{}, others ...interface{}) error {

	if query == nil {
		return fmt.Errorf("query must not be nil")
	}
	return p.conn.Db.WithContext(ctx).Where(query).First(out).Error

}

// Paginate implements repo.IDBClient.
func (p *PostgresGormClient) Paginate(ctx context.Context, query interface{}, out interface{}, paginationParams repo.PaginationRequest, others ...interface{}) (*repo.Pagination, error) {

	total, err := p.Count(ctx, query, others...)
	if err != nil {
		return nil, err
	}

	// query must be a struct model to DB, gorm will use Model struct to create a WHERE query on UPDATE query
	err = p.conn.Db.WithContext(ctx).Scopes(func(db *gorm.DB) *gorm.DB {
		return db.Offset((paginationParams.Page - 1) * paginationParams.Limit).Limit(paginationParams.Limit)
	}).Where(query).Find(out).Error

	if err != nil {
		return nil, err
	}
	return &repo.Pagination{
		Total: total,
		Limit: paginationParams.Limit,
	}, nil

}

// UpdateMany implements repo.IDBClient.
func (p *PostgresGormClient) UpdateMany(ctx context.Context, filter interface{}, update interface{}, out interface{}, others ...interface{}) error {
	panic("unimplemented")
}

// UpdateOneAndReturn implements repo.IDBClient.
func (p *PostgresGormClient) UpdateOneAndReturn(ctx context.Context, query interface{}, update interface{}, out repo.BaseModel, others ...interface{}) error {
	// query must be a struct model to DB, gorm will use Model struct to create a WHERE query on UPDATE query
	err := p.conn.Db.WithContext(ctx).Model(query).Updates(update).Error
	if err != nil {
		return err
	}
	return p.FindOne(ctx, out, query, others...)
}

// Upsert implements repo.IDBClient.
func (p *PostgresGormClient) Upsert(ctx context.Context, filter interface{}, update interface{}, out repo.BaseModel, others ...interface{}) error {
	err := p.conn.Db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(update).Error
	if err != nil {
		return err
	}
	return p.FindOne(ctx, out, update)
}

// WithTransaction implements repo.IDBClient.
func (p *PostgresGormClient) WithTransaction(ctx context.Context, fn func(ctx context.Context, others ...interface{}) (interface{}, error), others ...interface{}) (interface{}, error) {
	p.conn.Db.Begin()

	result, err := fn(ctx, others...)
	if err != nil {
		p.conn.Db.Rollback()
		return nil, err
	}
	p.conn.Db.Commit()
	return result, nil
}

func NewPostgresGormClient(conn *PostgresGormConnection) repo.IDBClient {
	return &PostgresGormClient{
		conn: conn,
	}
}

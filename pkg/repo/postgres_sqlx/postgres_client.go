package postgres

import (
	"context"
	"fmt"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/repo"
)

type PostgresDBClient struct {
	conn *PostgresConnection
}

func NewPostgresDBClient(conn *PostgresConnection) repo.IDBClient {
	return &PostgresDBClient{
		conn: conn,
	}
}

func (p *PostgresDBClient) FindAll(ctx context.Context, out []repo.BaseModel, query interface{}, others ...interface{}) error {
	_query, ok := query.(string)
	if !ok {
		return fmt.Errorf("query must be string")
	}
	return p.conn.DB.SelectContext(ctx, out, _query, others...)
}

func (p *PostgresDBClient) FindOne(ctx context.Context, out repo.BaseModel, query interface{}, others ...interface{}) error {
	_query, ok := query.(string)
	if !ok {
		return fmt.Errorf("query must be string")
	}
	err := p.conn.DB.GetContext(ctx, out, _query, others...)
	return err
}

// BulkCreate implements repo.IDBClient.
func (p *PostgresDBClient) BulkCreate(ctx context.Context, query interface{}, data []repo.BaseModel, others ...interface{}) error {
	_query, ok := query.(string)
	if !ok {
		return fmt.Errorf("query must be string")
	}
	_, err := p.conn.DB.NamedExecContext(ctx, _query, data)
	return err
}

// Create implements repo.IDBClient.
func (p *PostgresDBClient) Create(ctx context.Context, query interface{}, data repo.BaseModel, others ...interface{}) error {
	_query, ok := query.(string)
	if !ok {
		return fmt.Errorf("query must be string")
	}

	_, err := p.conn.DB.NamedExecContext(ctx, _query, data)
	return err

}

// Count implements repo.IDBClient.
func (p *PostgresDBClient) Count(ctx context.Context, query interface{}, others ...interface{}) (int, error) {
	_query, ok := query.(string)
	if !ok {
		return -1, fmt.Errorf("query muste be string")
	}
	var result int
	err := p.conn.DB.GetContext(ctx, &result, _query, others...)
	if err != nil {
		return -1, err
	}
	return result, nil
}

// Delete implements repo.IDBClient.
func (p *PostgresDBClient) Delete(ctx context.Context, query interface{}, others ...interface{}) error {
	if _query, ok := query.(string); !ok {
		return fmt.Errorf("query must be string")
	} else {
		_, err := p.conn.DB.NamedExecContext(ctx, _query, others)
		return err
	}
}

func (p *PostgresDBClient) buildWhereConditionSQL(queries []repo.QueryParams) string {
	whereQueryString := "WHERE "
	for _, query := range queries {
		whereQueryString += fmt.Sprintf("%s = %s %s ", query.Key, query.Value, query.Operator)
	}
	return whereQueryString
}

// PaginateV2 implements repo.IDBClient.
func (p *PostgresDBClient) Paginate(ctx context.Context, table string, paginationParams repo.PaginationRequest, others ...interface{}) (*repo.Pagination, error) {
	whereQueryString := p.buildWhereConditionSQL(paginationParams.Query)
	queryString := fmt.Sprintf("SELECT * FROM %s %s LIMIT=%d OFFSET=%d", table, whereQueryString, paginationParams.Limit, (paginationParams.Page-1)*paginationParams.Limit)
	var result []repo.BaseModel
	err := p.conn.DB.SelectContext(ctx, &result, queryString, others...)
	if err != nil {
		return nil, err
	}
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", table, whereQueryString)
	err = p.conn.DB.GetContext(ctx, &total, countQuery, others...)
	if err != nil {
		return nil, err
	}

	return &repo.Pagination{
		Total: total,
		Limit: paginationParams.Limit,
		Data:  result,
	}, nil
}

// PaginateV2 implements repo.IDBClient.
func (p *PostgresDBClient) PaginateV2(ctx context.Context, query interface{}, out []repo.BaseModel, paginationParams repo.PaginationRequest, others ...interface{}) (*repo.Pagination, error) {

	total, err := p.Count(ctx, query, others...)
	if err != nil {
		return nil, err
	}

	_query, _ := query.(string)
	_query = fmt.Sprintf("%s LIMIT=%d OFFSET=%d", _query, paginationParams.Limit, (paginationParams.Page-1)*paginationParams.Limit)

	// var result []repo.BaseModel
	err = p.conn.DB.SelectContext(ctx, out, _query, others...)
	if err != nil {
		return nil, err
	}
	return &repo.Pagination{
		Limit: paginationParams.Limit,
		Total: total,
		Data:  out,
	}, nil
}

// UpdateMany implements repo.IDBClient.
func (p *PostgresDBClient) UpdateMany(ctx context.Context, update repo.BaseModel, out repo.BaseModel, others ...interface{}) error {
	panic("unimplemented")
}

// UpdateOne implements repo.IDBClient.
func (p *PostgresDBClient) UpdateOneAndReturn(ctx context.Context, query interface{}, data, out repo.BaseModel, others ...interface{}) error {
	_query, ok := query.(string)
	if !ok {
		return fmt.Errorf("query must be string")
	}
	// var result repo.BaseModel
	return p.conn.DB.GetContext(ctx, out, _query, data)
}

// Upsert implements repo.IDBClient.
func (p *PostgresDBClient) Upsert(ctx context.Context, data repo.BaseModel, others ...interface{}) error {
	panic("unimplemented")
}

func (p *PostgresDBClient) WithTransaction(ctx context.Context, fn func(ctx context.Context, others ...interface{}) error, others ...interface{}) error {
	transaction := p.conn.DB.MustBegin()
	err := fn(ctx, others...)
	if err != nil {
		err = transaction.Rollback()
		return err
	}
	return transaction.Commit()
}

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

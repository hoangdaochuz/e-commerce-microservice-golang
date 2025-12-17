package postgres

import (
	"fmt"

	"github.com/XSAM/otelsql"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/logging"
	"github.com/jmoiron/sqlx"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

type PostgresConnection struct {
	DB *sqlx.DB
}

func (p *PostgresConnection) Connect(dsn string) error {
	db, err := otelsql.Open("postgres", dsn, otelsql.WithAttributes(semconv.DBSystemPostgreSQL))
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	err = otelsql.RegisterDBStatsMetrics(db, otelsql.WithAttributes(semconv.DBSystemPostgreSQL))
	if err != nil {
		return fmt.Errorf("fail to register db stats metric tracing: %w", err)
	}
	p.DB = sqlx.NewDb(db, "postgres")
	logging.GetSugaredLogger().Infof("Connect to Postgres successfully ✅")
	return nil
}

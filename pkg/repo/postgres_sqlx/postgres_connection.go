package postgres

import (
	"fmt"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/logging"
	"github.com/jmoiron/sqlx"
)

type PostgresConnection struct {
	DB *sqlx.DB
}

func (p *PostgresConnection) Connect(dsn string) error {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	p.DB = db
	logging.GetSugaredLogger().Infof("Connect to Postgres successfully ✅")
	return nil
}

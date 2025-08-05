package postgres_gorm

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresGormConnection struct {
	Db *gorm.DB
}

func (p *PostgresGormConnection) Connect(dsn string) error {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("fail to connect to postgres gorm")
	}
	p.Db = db
	return nil
}

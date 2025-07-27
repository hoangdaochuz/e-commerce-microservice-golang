package order_configs

import (
	"fmt"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/configs"
	postgres "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/repo/postgres_sqlx"
	_ "github.com/lib/pq"
)

type OrderDatabase struct {
	Conn *postgres.PostgresConnection
}

func NewOrderDatabase() *OrderDatabase {
	config, err := configs.Load()
	if err != nil {
		panic("Fail to load config file")
	}
	dsn := fmt.Sprintf("host=%s port =%s user=%s password=%s dbname=%s sslmode=disable", config.OrderDatabase.Host, config.OrderDatabase.Port, config.OrderDatabase.User, config.OrderDatabase.Password, config.OrderDatabase.DBname)

	conn := &postgres.PostgresConnection{}
	err = conn.Connect(dsn)
	if err != nil {
		panic("Fail to connect to order database: " + err.Error())
	}
	return &OrderDatabase{
		Conn: conn,
	}
}

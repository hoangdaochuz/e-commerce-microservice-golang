package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	order_api "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/api/order"
	order_configs "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/configs"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/handler/order"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/configs"
	custom_nats "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/custom-nats"
	di "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/dependency-injection"
	"github.com/nats-io/nats.go"
)

func main() {
	config, err := configs.Load()
	if err != nil {
		log.Fatal("fail to get config data")
	}
	natsConn, err := nats.Connect(config.NatsAuth.NATSUrl, nats.UserInfo(config.NatsAuth.NATSApps[0].Username, config.NatsAuth.NATSApps[0].Password))
	if err != nil {
		log.Fatal("fail to connect to nats")
	}

	chi := chi.NewRouter()
	router := custom_nats.NewRouter(chi)
	var orderApp *order.OrderServiceApp
	di.Resolve(func(orderImplement *order.OrderServiceApp) {
		orderApp = orderImplement
	})
	orderAppProxy := order_api.NewOrderServiceProxy(orderApp)

	orderRouterClient := order_api.NewOrderServiceRouter(orderAppProxy)
	order_configs.NewOrderDatabase()

	server := custom_nats.NewServer(natsConn, router, order_api.NATS_SUBJECT, orderRouterClient)
	err = server.Start()
	if err != nil {
		log.Fatal("fail to start order server")
		server.Stop()
	}
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGTERM, syscall.SIGINT)
	<-shutdownChan
	fmt.Println("Shutting down server peacefully")
	err = server.Stop()
	if err != nil {
		log.Fatal("fail to stop server")
	}
}

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	apigateway "github.com/hoangdaochuz/ecommerce-microservice-golang/api_gateway"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/configs"
	di "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/dependency-injection"
	"github.com/nats-io/nats.go"
)

func main() {
	di.InitDIContainer()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	config, err := configs.Load()
	if err != nil {
		log.Fatal("failed to load configuration: %w", err)
	}
	natsConn, err := nats.Connect(config.NatsAuth.NATSUrl, nats.UserInfo(config.NatsAuth.NATSApps[0].Username, config.NatsAuth.NATSApps[0].Password))
	if err != nil {
		log.Fatal("Failed to connect to nats: ", err)
	}
	log.Println("Connected to nats successfully")

	mux := http.NewServeMux()
	apigatewayServer := &http.Server{
		Addr:    ":" + config.Apigateway.Port,
		Handler: mux,
	}

	gateway := apigateway.NewAPIGateway(natsConn, apigatewayServer, mux, ctx)
	di.Make[*apigateway.APIGateway](func() *apigateway.APIGateway {
		return gateway
	})
	err = gateway.Start()
	fmt.Println("hello, what is this")
	if err != nil {
		err = gateway.Stop()
		if err != nil {
			log.Fatal("fail to shut down api gateway...")
			return
		}
		log.Fatal("fail to start api gateway: ", err)
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutting down api gateway...")
	err = gateway.Stop()
	if err != nil {
		log.Fatal("fail to shut down api gateway...")
		return
	}
	fmt.Println("shut down api gateway successfully")
}

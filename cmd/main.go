package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	apigateway "github.com/hoangdaochuz/ecommerce-microservice-golang/api_gateway"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/configs"
	di "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/dependency-injection"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/logging"
	"github.com/nats-io/nats.go"
)

func main() {
	di.InitDIContainer()
	config, err := configs.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := logging.GetSugaredLogger()
	natsConn, err := nats.Connect(config.NatsAuth.NATSUrl, nats.UserInfo(config.NatsAuth.NATSApps[0].Username, config.NatsAuth.NATSApps[0].Password))
	if err != nil {
		logger.Fatalf("Failed to connect to nats: %v", err)
	}
	// log.Println("Connected to nats successfully")
	logger.Info("Connected to nats successfully")

	mux := http.NewServeMux()
	apigatewayServer := &http.Server{
		Addr:              ":" + config.Apigateway.Port,
		Handler:           mux,
		ReadHeaderTimeout: 30 * time.Second,
	}

	gateway := apigateway.NewAPIGateway(natsConn, apigatewayServer, mux, ctx)
	_ = di.Make[*apigateway.APIGateway](func() *apigateway.APIGateway {
		return gateway
	})
	err = gateway.Start()
	if err != nil {
		err = gateway.Stop()
		if err != nil {
			logging.GetSugaredLogger().Fatalf("fail to shut down api gateway: %v", err)
			return
		}
		logging.GetSugaredLogger().Fatalf("fail to start api gateway: %v", err)
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logging.GetSugaredLogger().Infof("Shutting down api gateway...")
	err = gateway.Stop()
	if err != nil {
		logging.GetSugaredLogger().Fatalf("fail to shut down api gateway: %v", err)
		return
	}
	logging.GetSugaredLogger().Infof("shut down api gateway successfully")
}

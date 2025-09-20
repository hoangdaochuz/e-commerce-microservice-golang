package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	apigateway "github.com/hoangdaochuz/ecommerce-microservice-golang/api_gateway"
	di "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/dependency-injection"
)

func main() {
	di.InitDIContainer()
	gateway, err := apigateway.Start("8080")
	if err != nil {
		log.Fatal(err)
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

package main

import (
	"log"

	apigateway "github.com/hoangdaochuz/ecommerce-microservice-golang/api_gateway"
	order_cmd "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/cmd"
	di "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/dependency-injection"
)

func main() {
	di.InitDIContainer()
	order_cmd.Start()
	err := apigateway.Start("8080")
	if err != nil {
		log.Fatal(err)
	}
}

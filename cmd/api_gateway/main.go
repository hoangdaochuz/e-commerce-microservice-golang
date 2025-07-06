package main

import (
	"log"

	apigateway "github.com/hoangdaochuz/ecommerce-microservice-golang/api_gateway"
	di "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/dependency-injection"
)

func main() {
	di.InitDIContainer()
	err := apigateway.Start("8080")
	if err != nil {
		log.Fatal(err)
	}
}

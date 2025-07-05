package main

import (
	"log"

	apigateway "github.com/hoangdaochuz/ecommerce-microservice-golang/api_gateway"
)

func main() {
	err := apigateway.Start("8080")
	if err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	apigateway "github.com/hoangdaochuz/ecommerce-microservice-golang/api_gateway"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/configs"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/codegen/codegen-frontend"
	di "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/dependency-injection"
	"github.com/nats-io/nats.go"
)

// func getServiceRoutesFromGateway(serviceRoutes *map[string]apigateway.ServiceRoute) error {
// 	err := di.Resolve(func(gw *apigateway.APIGateway) error {

// 		// *serviceRoutes = gw.GetServiceRoutes()
// 		return nil
// 	})
// 	fmt.Println(err)
// 	return err
// }

func setupAPIGateway() error {
	configs, err := configs.Load()
	if err != nil {
		fmt.Errorf("failed to load config value %w", err)
	}
	natsUrl := fmt.Sprintf("nats://%s:%s@localhost:4222", configs.NatsAuth.NATSApps[0].Username, configs.NatsAuth.NATSApps[0].Password)
	natsConn, err := nats.Connect(natsUrl)
	if err != nil {
		return fmt.Errorf("failed to connect to nats")
	}
	gw := apigateway.NewAPIGateway(natsConn, configs.ServiceRegistry.RequestTimeout, &http.Server{})
	di.Make(func() *apigateway.APIGateway {
		return gw
	})
	return nil
}

func main() {
	var (
		ourDir      = flag.String("outdir", "./frontend/apis", "Output directory for generated frontend code")
		baseURL     = flag.String("baseurl", "http://localhost:8080", "Base URL for API endpoints")
		serviceName = flag.String("service", "", "service name for generated code, if this field has not been set, it will generated frontend code for all service")
		help        = flag.Bool("help", false, "Show help message")
	)

	flag.Parse() // this help cli read the argument
	if *help {
		fmt.Println("Frontend Code Generator")
		fmt.Println("Useage: go run pkg/codegen/cli/main.go [options]")
		fmt.Println("\nOptions:\n")
		flag.PrintDefaults()
		os.Exit(0)
	}
	di.InitDIContainer()
	err := setupAPIGateway()
	if err != nil {
		log.Fatal("failed to setup api gateway for get service routes")
	}
	// err = getServiceRoutesFromGateway(&serviceRoutes)
	// if err != nil {
	// 	log.Fatal("failed to get service routes from api gateway")
	// }
	// var _serviceName string
	// if serviceName == nil {
	// 	_serviceName = ""
	// }else{
	// 	_serviceName = *serviceName
	// }
	frontendCodeGenerator := codegen.NewFrontendGenerator(*ourDir, *baseURL, *serviceName)
	err = frontendCodeGenerator.GenerateAllFECode()
	if err != nil {
		log.Fatal("failed while generate code frontend : %w", err)
	}
	fmt.Println("🌈✅ Generated frontend code successfully")
}

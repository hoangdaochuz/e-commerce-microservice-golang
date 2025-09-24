package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/codegen/codegen-frontend"
	proto_to_dgo "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/codegen/proto-to-dgo"
	di "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/dependency-injection"
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
	// For now, just return nil as we don't need actual API gateway setup for proto-to-dgo generation
	return nil
}

func main() {
	var (
		ourDir      = flag.String("outdir", "./frontend/apis", "Output directory for generated frontend code")
		baseURL     = flag.String("baseurl", "http://localhost:8080", "Base URL for API endpoints")
		serviceName = flag.String("service", "", "service name for generated code, if this field has not been set, it will generated frontend code for all service")
		protoFile   = flag.String("proto", "", "Path to .proto file for generating .d.go file")
		dgoOutput   = flag.String("dgo-output", "", "Output path for generated .d.go file")
		help        = flag.Bool("help", false, "Show help message")
	)

	flag.Parse() // this help cli read the argument
	if *help {
		fmt.Println("Code Generator")
		fmt.Println("Usage: go run pkg/codegen/cli/main.go [options]")
		fmt.Println("\nOptions:\n")
		flag.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  Generate .d.go from .proto:")
		fmt.Println("    go run pkg/codegen/cli/main.go -proto=apps/order/proto/order.proto -dgo-output=apps/order/api/order/order.d.go")
		fmt.Println("  Generate frontend code:")
		fmt.Println("    go run pkg/codegen/cli/main.go -service=order -outdir=./frontend/apis")
		os.Exit(0)
	}

	// Check if proto-to-dgo generation is requested
	if *protoFile != "" {
		if *dgoOutput == "" {
			log.Fatal("dgo-output flag is required when proto flag is specified")
		}

		generator := proto_to_dgo.NewDGoGenerator()
		err := generator.GenerateFromProto(*protoFile, *dgoOutput)
		if err != nil {
			log.Fatalf("failed to generate .d.go file: %v", err)
		}
		fmt.Println("🌈✅ Generated .d.go file successfully")
		return
	}

	// Continue with frontend code generation if no proto file specified
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

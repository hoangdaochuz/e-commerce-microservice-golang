package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/codegen/codegen-frontend"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/codegen/proto2dgo"
)

// func getServiceRoutesFromGateway(serviceRoutes *map[string]apigateway.ServiceRoute) error {
// 	err := di.Resolve(func(gw *apigateway.APIGateway) error {

// 		// *serviceRoutes = gw.GetServiceRoutes()
// 		return nil
// 	})
// 	fmt.Println(err)
// 	return err
// }

func handleGenGoContractFile(protoFile, outputFile string) error {
	generater := proto2dgo.NewProto2dgoGenerater()
	return generater.GenerateProto2Dgo(protoFile, outputFile)
}

func handleGenFECode(ourDir, baseURL, serviceName string) error {
	frontendCodeGenerator := codegen.NewFrontendGenerator(ourDir, baseURL, serviceName)
	return frontendCodeGenerator.GenerateAllFECode()

}

func main() {
	var (
		// ourDir        = flag.String("outdir", "./frontend/apis", "Output directory for generated frontend code")
		// baseURL       = flag.String("baseurl", "http://localhost:8080", "Base URL for API endpoints")
		// serviceName   = flag.String("service", "", "service name for generated code, if this field has not been set, it will generated frontend code for all service")
		help          = flag.Bool("help", false, "Show help message")
		genType       = flag.String("type", "", "Code gen type (ex. backend-contract, fronend)")
		dgoOutput     = flag.String("dgoOutput", "", "Generated file has name and place at")
		protofilePath = flag.String("protofilePath", "", "Path to proto file need for generating contract go code")
	)

	flag.Parse() // this help cli read the argument
	if *help {
		fmt.Println("Go Contract Generator")
		// fmt.Println("Useage: go run pkg/codegen/cli/main.go [options]")
		fmt.Println("\nOptions:\n")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *genType == "backend-contract" {
		err := handleGenGoContractFile(*protofilePath, *dgoOutput)
		if err != nil {
			log.Fatal("failed while generating contract go code from proto file")
		}
	} else {
		// Implement later
		// err := handleGenFECode(*ourDir, *baseURL, *serviceName)
		// if err != nil {
		// 	log.Fatal("failed while generating fe code")
		// }
	}
	fmt.Println("🌈✅ Generated frontend code successfully")
}

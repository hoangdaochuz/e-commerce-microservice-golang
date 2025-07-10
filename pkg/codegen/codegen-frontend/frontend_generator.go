package codegen

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"

	apigateway "github.com/hoangdaochuz/ecommerce-microservice-golang/api_gateway"
)

// FrontendGenerator generates TypeScript/React Query code
type FrontendGenerator struct {
	outDir        string
	baseURL       string
	serviceName   string
	serviceRoutes map[string]apigateway.ServiceRoute
}

type CodeGenConstant struct {
	Timestamp string
	Methods   []MethodConstantData
}

type MethodConstantData struct {
	FullPath           string
	MethodConstantName string
}

func NewFrontendGenerator(outDir, baseURL, serviceName string, serviceRoutes map[string]apigateway.ServiceRoute) *FrontendGenerator {
	return &FrontendGenerator{
		outDir:        outDir,
		baseURL:       baseURL,
		serviceName:   serviceName,
		serviceRoutes: serviceRoutes,
	}
}

func (fg *FrontendGenerator) AddServiceRoutes(serviceName string, serviceRoutes apigateway.ServiceRoute) {
	fg.serviceRoutes[serviceName] = serviceRoutes
}

func (fg *FrontendGenerator) GenerateAllFECode() error {
	// check if the directory is already exist
	_, err := os.Stat(fg.outDir)
	isNotExist := os.IsNotExist(err)
	if isNotExist {
		err := os.MkdirAll(fg.outDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create a directory for store codegen, %w", err)
		}
	}

	if fg.serviceName == "" {
		// gen all fe code for all service
		for serviceName, _ := range fg.serviceRoutes {
			if err := fg.generateCodeForSpecificService(serviceName); err != nil {
				return fmt.Errorf("failed to generated frontend code for service %s", serviceName)
			}
		}
	} else {
		if err := fg.generateCodeForSpecificService(fg.serviceName); err != nil {
			return fmt.Errorf("failed to generated frontend code for service %s", fg.serviceName)
		}
	}
	fmt.Println("Generated frontend code successfully👏☘️☘️")
	return nil
}

func (fg *FrontendGenerator) generateCodeForSpecificService(serviceName string) error {
	if err := fg.generateConstantFile(serviceName); err != nil {
		return fmt.Errorf("failed to generate a constant file %w", err)
	}

	// generate type file
	if err := fg.generateTypeFile(serviceName); err != nil {
		return fmt.Errorf("failed to generate a type file %w", err)
	}

	// generate service client file
	if err := fg.generateServiceClientFile(serviceName); err != nil {
		return fmt.Errorf("failed to generate a service client file %w", err)
	}

	// generate index file
	if err := fg.generateIndexFile(serviceName); err != nil {
		return fmt.Errorf("failed to generate index file %w", err)
	}
	return nil
}

func (fg *FrontendGenerator) generateConstantFile(serviceName string) error {
	templateConstant := `
// This is codegen - DO NOT EDIT
// Code generated at: {{.Timestamp}}
{{range .Methods}}
export const {{.MethodConstantName}} = "{{.FullPath}}";
{{end}}
`

	data := CodeGenConstant{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Methods:   fg.buildMethodConstantData(fg.serviceRoutes[serviceName]),
	}
	return fg.writeTemplateToFile("constant.ts", templateConstant, data, serviceName)
}

func (fg *FrontendGenerator) writeTemplateToFile(filename, templateString string, data interface{}, serviceName string) error {
	filePath := filepath.Join(fg.outDir, serviceName, filename)

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory")
	}
	templ, err := template.New(filename).Parse(templateString)
	if err != nil {
		return fmt.Errorf("failed to parse template string %w", err)
	}
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create a file %w", err)
	}
	defer file.Close()
	err = templ.Execute(file, data)
	return err
}

func (fg *FrontendGenerator) buildMethodConstantData(serviceRoutes apigateway.ServiceRoute) []MethodConstantData {
	var methods []MethodConstantData
	for methodName, methodInfo := range serviceRoutes.Methods {
		methods = append(methods, MethodConstantData{
			FullPath:           serviceRoutes.BasePath + methodInfo.Path,
			MethodConstantName: fg.buildMethodConstantName(serviceRoutes.ServiceName, methodName),
		})
	}
	return methods
}

func (fg *FrontendGenerator) buildMethodConstantName(serviceName, methodName string) string {
	return fmt.Sprintf("%s_%s_URL", serviceName, methodName)
}

func (fg *FrontendGenerator) generateTypeFile(serviceName string) error {
	return nil
}

func (fg *FrontendGenerator) generateServiceClientFile(serviceName string) error {
	return nil
}
func (fg *FrontendGenerator) generateIndexFile(serviceName string) error {
	return nil
}

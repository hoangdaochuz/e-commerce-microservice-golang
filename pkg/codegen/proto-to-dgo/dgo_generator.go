package proto_to_dgo

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// DGoGenerator generates .d.go files from proto definitions
type DGoGenerator struct {
	parser *ProtoParser
}

// NewDGoGenerator creates a new DGO generator
func NewDGoGenerator() *DGoGenerator {
	return &DGoGenerator{
		parser: NewProtoParser(),
	}
}

// DGoTemplateData holds the data for generating .d.go files
type DGoTemplateData struct {
	Package     string
	ServiceName string
	ImportPath  string
	NatsSubject string
	Methods     []MethodData
}

// MethodData holds method information for template
type MethodData struct {
	Name         string
	RequestType  string
	ResponseType string
	ConstantName string
}

// GenerateFromProto generates a .d.go file from a .proto file
func (g *DGoGenerator) GenerateFromProto(protoPath, outputPath string) error {
	// Parse the proto file
	protoFile, err := g.parser.ParseFile(protoPath)
	if err != nil {
		return fmt.Errorf("failed to parse proto file: %w", err)
	}

	if len(protoFile.Services) == 0 {
		return fmt.Errorf("no services found in proto file")
	}

	// For now, generate for the first service
	service := protoFile.Services[0]

	// Prepare template data
	templateData := g.prepareTemplateData(protoFile, service)

	// Generate the .d.go file
	return g.generateDGoFile(templateData, outputPath)
}

func (g *DGoGenerator) prepareTemplateData(protoFile *ProtoFile, service ProtoService) *DGoTemplateData {
	// Extract package name from go_package or use proto package
	packageName := protoFile.Package
	if protoFile.GoPackage != "" {
		// Extract the last part of the go_package path
		parts := strings.Split(protoFile.GoPackage, "/")
		if len(parts) > 0 {
			packageName = parts[len(parts)-1]
		}
	}

	// Generate NATS subject
	natsSubject := fmt.Sprintf("/api/v1/%s", strings.ToLower(protoFile.Package))

	// Prepare methods data
	var methods []MethodData
	for _, method := range service.Methods {
		constantName := fmt.Sprintf("%s_%s",
			strings.ToUpper(protoFile.Package),
			strings.ToUpper(convertCamelToSnake(method.Name)))

		methods = append(methods, MethodData{
			Name:         method.Name,
			RequestType:  method.RequestType,
			ResponseType: method.ResponseType,
			ConstantName: constantName,
		})
	}

	return &DGoTemplateData{
		Package:     packageName,
		ServiceName: service.Name,
		ImportPath:  "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/custom-nats",
		NatsSubject: natsSubject,
		Methods:     methods,
	}
}

func (g *DGoGenerator) generateDGoFile(data *DGoTemplateData, outputPath string) error {
	// Create output directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Load template from embedded filesystem
	tmpl, err := template.ParseFS(templateFS, "templates/generated.d.tmpl")
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Execute template
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// convertCamelToSnake converts CamelCase to snake_case
func convertCamelToSnake(input string) string {
	var result strings.Builder
	for i, r := range input {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToUpper(result.String())
}

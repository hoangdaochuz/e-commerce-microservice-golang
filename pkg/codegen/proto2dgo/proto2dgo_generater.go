package proto2dgo

import (
	"embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

type Proto2dgoGenerater struct {
	parser *ProtoParser
}

type Proto2dgoTemplate struct {
	ImportPath  []string
	NatsSubject string
	ProtoModel  ProtoModel
	GoPackage   string
}

func NewProto2dgoGenerater() *Proto2dgoGenerater {
	return &Proto2dgoGenerater{
		parser: NewProtoParser(),
	}
}

func (g *Proto2dgoGenerater) GenerateProto2Dgo(protoPath, outputPath string) error {
	protoModel, err := g.parser.ParseProtoFile(protoPath)
	if err != nil {
		return err
	}
	if protoModel == nil {
		return fmt.Errorf("model proto is nil")
	}
	template := g.prepareProto2dgoTemplate(protoModel)

	return g.generateDgo(template, outputPath)
}

func (g *Proto2dgoGenerater) generateDgo(data *Proto2dgoTemplate, outputPath string) error {
	// create directory if it isn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("fail to create dir: %w", err)
	}
	tmpl, err := template.ParseFS(templateFS, "templates/generated.d.tmpl")
	if err != nil {
		return fmt.Errorf("fail to parse template from FS: %w", err)
	}

	// create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("fail to create a output file: %w", err)
	}
	defer file.Close()

	return tmpl.Execute(file, data)
}

func (g *Proto2dgoGenerater) prepareProto2dgoTemplate(protoModel *ProtoModel) *Proto2dgoTemplate {
	importPath := []string{}
	for _, path := range protoModel.ImportPaths {
		importPath = append(importPath, path.Path)
	}
	goPackage := protoModel.GoPackage
	splits := strings.Split(goPackage, "/")

	natsSubject := fmt.Sprintf("/api/v1/%s", splits[len(splits)-1])
	return &Proto2dgoTemplate{
		NatsSubject: natsSubject,
		ImportPath:  importPath,
		ProtoModel:  *protoModel,
		GoPackage:   splits[len(splits)-1],
	}
}

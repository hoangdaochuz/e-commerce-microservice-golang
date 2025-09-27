package proto2dgo

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type ProtoParser struct {
}

func NewProtoParser() *ProtoParser {
	return &ProtoParser{}
}

func (p *ProtoParser) ParseProtoFile(path string) (*ProtoModel, error) {
	protoFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("fail to open .proto file: %w", err)
	}
	defer protoFile.Close()

	protoModel := &ProtoModel{}
	//
	scanner := bufio.NewScanner(protoFile)
	protoImports := []ImportModel{}

	enumModels := []EnumModel{}
	enumItem := EnumModel{}
	enumField := EnumField{}
	inEnumBody := false
	enumOrderValue := 0

	inService := false
	serviceModels := []ServiceModel{}
	serviceModelItem := ServiceModel{}
	serviceMethods := []MethodModel{}

	messageModels := []MessageModel{}
	messageItem := MessageModel{}
	inMessage := false
	messageFields := []FieldModel{}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		if strings.HasPrefix(line, "syntax") {
			syntax := p.extractSyntax(line)
			protoModel.Syntax = syntax
		}

		if strings.HasPrefix(line, "package") {
			protoPackage := p.extractProtoPackage(line)
			protoModel.ProtoPackage = protoPackage
		}

		if strings.Contains(line, "go_package") {
			goPackage := p.extractGoPackage(line)
			protoModel.GoPackage = goPackage
		}

		if strings.HasPrefix(line, "import") {
			importPath := p.extractImportPath(line)
			protoImports = append(protoImports, *importPath)
		}

		if strings.HasPrefix(line, "enum") {
			enumName := p.extractEnumName(line)
			enumItem.Name = enumName
			inEnumBody = true
			continue
		}
		if inEnumBody {
			if !strings.Contains(line, "}") {
				fieldEnum := p.extractEnumField(line)
				enumField.Key = fieldEnum
				enumField.Value = enumOrderValue
				enumOrderValue++
				enumItem.EnumFields = append(enumItem.EnumFields, enumField)
			} else {
				inEnumBody = false
				enumOrderValue = 0
				enumModels = append(enumModels, enumItem)
				continue
			}
		}

		if strings.HasPrefix(line, "service") {
			inService = true
			serviceName := p.extractServiceName(line)
			serviceModelItem.Name = serviceName
			continue
		}
		if inService {
			if !strings.Contains(line, "}") {
				if strings.Contains(line, "rpc") {
					method := p.extractMethodService(line, protoModel.GoPackage)
					if method != nil {
						serviceMethods = append(serviceMethods, *method)
					}
				}
			} else {
				serviceModelItem.Methods = serviceMethods
				serviceModels = append(serviceModels, serviceModelItem)
				inService = false
				serviceMethods = []MethodModel{}
				continue
			}
		}

		if strings.HasPrefix(line, "message") {
			inMessage = true
			messageName := p.extractMessageName(line)
			messageItem.MessageName = messageName
			continue
		}
		if inMessage {
			if !strings.Contains(line, "}") {
				field, err := p.getFieldOfMessage(line)
				if err != nil {
					return nil, err
				}
				messageFields = append(messageFields, *field)
			} else {
				inMessage = false
				messageItem.Fields = messageFields
				messageFields = []FieldModel{}
				messageModels = append(messageModels, messageItem)
			}
		}

	}
	protoModel.ImportPaths = protoImports
	protoModel.Enums = enumModels
	protoModel.Services = serviceModels
	protoModel.Messages = messageModels

	return protoModel, nil

}

func (p *ProtoParser) getFieldOfMessage(line string) (*FieldModel, error) {
	pattern := `(?:(repeated|optional)\s+)?(\w+(?:\.\w+)*)\s+(\w+)\s*=\s*(\d+)\s*;?`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(line)
	var fieldInfo *FieldModel

	if len(matches) >= 5 {
		isRepeated := false
		isOptional := false
		if matches[1] != "" {
			if strings.TrimSpace(matches[1]) == "repeated" {
				isRepeated = true
			} else {
				isOptional = true
			}
		}
		order, err := strconv.Atoi(matches[4])
		if err != nil {
			return nil, err
		}
		fieldInfo = &FieldModel{
			IsRepeat:   isRepeated,
			IsOptional: isOptional,
			Type:       matches[2],
			Name:       matches[3],
			Order:      order,
		}
		return fieldInfo, nil
	}
	return nil, nil
}

func (p *ProtoParser) extractMessageName(line string) string {
	regex := `message\s+(\w+)\s*\{`
	return p.extractFirstSubstringCaptureMatch(line, regex)
}

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

func (p *ProtoParser) extractMethodService(line string, goPackage string) *MethodModel {
	re := regexp.MustCompile(`rpc\s+(\w+)\s*\(\s*(\w+)\s*\)\s*returns\s*\(\s*(\w+)\s*\)`)
	matches := re.FindStringSubmatch(line)
	method := &MethodModel{}
	goPkgSlits := strings.Split(goPackage, "/")
	methodConstantPrefix := goPkgSlits[len(goPkgSlits)-1]
	if len(matches) >= 4 {
		method.Name = matches[1]
		method.RequestType = matches[2]
		method.ResponseType = matches[3]
		method.ConstantName = strings.ToUpper(methodConstantPrefix) + "_" + strings.ToUpper(convertCamelToSnake(matches[1]))
		return method
	}
	return nil
}

func (p *ProtoParser) extractFirstSubstringCaptureMatch(line, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func (p *ProtoParser) extractServiceName(line string) string {
	regexPattern := `service\s+(\w+)`
	return p.extractFirstSubstringCaptureMatch(line, regexPattern)
}

func (p *ProtoParser) extractSyntax(line string) string {
	regexPattern := `syntax\s*=\s*"([^"]+)"`
	return p.extractFirstSubstringCaptureMatch(line, regexPattern)
}

func (p *ProtoParser) extractProtoPackage(line string) string {
	regexPattern := `package\s+(\w+)\s*;`
	return p.extractFirstSubstringCaptureMatch(line, regexPattern)
}

func (p *ProtoParser) extractGoPackage(line string) string {
	regexPattern := `option\s*go_package\s*=\s*"([^"]+)"`
	return p.extractFirstSubstringCaptureMatch(line, regexPattern)
}

func (p *ProtoParser) extractImportPath(line string) *ImportModel {
	regexPattern := `import\s*(public\s+|weak\s+)?"([^"]+)"`
	re := regexp.MustCompile(regexPattern)
	matches := re.FindStringSubmatch(line)
	if len(matches) >= 3 {
		var mode ImportMode
		if matches[1] != "" {
			mode = ImportMode(strings.TrimSpace(matches[1]))
		} else {
			mode = NORMAL
		}
		return &ImportModel{
			Path: matches[2],
			Mode: mode,
		}
	}
	return nil
}

func (p *ProtoParser) extractEnumName(line string) string {
	regexPattern := `enum\s+(\w+)\s*\{`
	return p.extractFirstSubstringCaptureMatch(line, regexPattern)
}

func (p *ProtoParser) extractEnumField(line string) string {
	regexPattern := `(\w+)\s*=\s*(\d+)\s*;?`
	return p.extractFirstSubstringCaptureMatch(line, regexPattern)
}

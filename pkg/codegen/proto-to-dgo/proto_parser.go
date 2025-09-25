package proto_to_dgo

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// TODO:
// - [ ] About repeated fields for array
// - [ ] About enum fields
// - [ ] About imports
// - [ ] About empty response/ noresponse rpc method
// - [ ] About optional fields
// - [ ] About require fields

// ---------------------------
// - [ ] About oneof fields -- low priority
// - [ ] About map fields
// - [ ] About default values
// - [ ] About reserved fields -- low priority
// - [ ] About extensions

// ProtoService represents a gRPC service definition
type ProtoService struct {
	Name    string
	Methods []ProtoMethod
}

// ProtoMethod represents a gRPC method definition
type ProtoMethod struct {
	Name         string
	RequestType  string
	ResponseType string
}

// ProtoMessage represents a protobuf message definition
type ProtoMessage struct {
	Name   string
	Fields []ProtoField
}

// ProtoField represents a field in a protobuf message
type ProtoField struct {
	Type string
	Name string
	Tag  int
}

// ProtoFile represents the parsed content of a .proto file
type ProtoFile struct {
	Package   string
	GoPackage string
	Services  []ProtoService
	Messages  []ProtoMessage
}

// ProtoParser parses .proto files
type ProtoParser struct{}

// NewProtoParser creates a new proto parser
func NewProtoParser() *ProtoParser {
	return &ProtoParser{}
}

// ParseFile parses a .proto file and returns the parsed content
func (p *ProtoParser) ParseFile(filePath string) (*ProtoFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open proto file: %w", err)
	}
	defer file.Close()

	protoFile := &ProtoFile{}
	scanner := bufio.NewScanner(file)

	var currentService *ProtoService
	var currentMessage *ProtoMessage
	var inService, inMessage bool

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Parse package
		if strings.HasPrefix(line, "package ") {
			protoFile.Package = p.extractPackage(line)
		}

		// Parse go_package option
		if strings.Contains(line, "go_package") {
			protoFile.GoPackage = p.extractGoPackage(line)
		}

		// Parse service definition
		if strings.HasPrefix(line, "service ") {
			serviceName := p.extractServiceName(line)
			currentService = &ProtoService{Name: serviceName}
			inService = true
			inMessage = false
		}

		// Parse message definition
		if strings.HasPrefix(line, "message ") {
			messageName := p.extractMessageName(line)
			currentMessage = &ProtoMessage{Name: messageName}
			inMessage = true
			inService = false
		}

		// Parse RPC methods
		if inService && strings.Contains(line, "rpc ") {
			method := p.extractMethod(line)
			if method != nil {
				currentService.Methods = append(currentService.Methods, *method)
			}
		}

		// Parse message fields
		if inMessage && !strings.HasPrefix(line, "message ") && line != "}" {
			field := p.extractField(line)
			if field != nil {
				currentMessage.Fields = append(currentMessage.Fields, *field)
			}
		}

		// End of service or message
		if line == "}" {
			if inService && currentService != nil {
				protoFile.Services = append(protoFile.Services, *currentService)
				currentService = nil
				inService = false
			}
			if inMessage && currentMessage != nil {
				protoFile.Messages = append(protoFile.Messages, *currentMessage)
				currentMessage = nil
				inMessage = false
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading proto file: %w", err)
	}

	return protoFile, nil
}

func (p *ProtoParser) extractPackage(line string) string {
	re := regexp.MustCompile(`package\s+([^;]+);`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

func (p *ProtoParser) extractGoPackage(line string) string {
	re := regexp.MustCompile(`go_package\s*=\s*"([^"]+)"`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

func (p *ProtoParser) extractServiceName(line string) string {
	re := regexp.MustCompile(`service\s+(\w+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func (p *ProtoParser) extractMessageName(line string) string {
	re := regexp.MustCompile(`message\s+(\w+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func (p *ProtoParser) extractMethod(line string) *ProtoMethod {
	re := regexp.MustCompile(`rpc\s+(\w+)\s*\(\s*(\w+)\s*\)\s*returns\s*\(\s*(\w+)\s*\)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 3 {
		return &ProtoMethod{
			Name:         matches[1],
			RequestType:  matches[2],
			ResponseType: matches[3],
		}
	}
	return nil
}

func (p *ProtoParser) extractField(line string) *ProtoField {
	re := regexp.MustCompile(`(\w+)\s+(\w+)\s*=\s*(\d+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) > 3 {
		return &ProtoField{
			Type: matches[1],
			Name: matches[2],
			Tag:  parseInt(matches[3]),
		}
	}
	return nil
}

func parseInt(s string) int {
	// Simple int parsing, could use strconv.Atoi for better error handling
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}

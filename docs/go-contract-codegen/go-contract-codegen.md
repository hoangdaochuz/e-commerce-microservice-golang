# Go Contract Code Generator

This document describes the proto2dgo code generator that creates Go contract files from Protocol Buffer definitions.

## Overview

The proto2dgo generator creates a `.d.go` contract file that includes:

- NATS subject constants for messaging
- Service interface definitions
- Proxy implementations for service delegation
- Router implementations for NATS routing

The generator parses `.proto` files and generates Go code that provides a clean abstraction layer for microservice communication over NATS.

## Features

- **Automatic NATS Subject Generation**: Creates consistent subject naming based on service and method names
- **Type-Safe Interfaces**: Generates Go interfaces that match your protobuf service definitions
- **Proxy Pattern**: Implements the proxy pattern for service method delegation
- **Router Integration**: Provides NATS router registration for seamless message handling
- **Template-Based**: Uses customizable templates for code generation

## Usage

### Via Task Runner

The easiest way to generate contract code is using the task runner:

```bash
task backend:codegen:service PROTO_FILE=apps/order/proto/order.proto
```

This command will:
1. Parse the specified proto file
2. Generate the contract file at `apps/{service}/api/{service}/{service}.d.go`
3. Create all necessary interfaces, proxies, and routers

### Via CLI Tool

You can also use the CLI tool directly:

```bash
go run pkg/codegen/cli/main.go \
  -type=backend-contract \
  -protofilePath=apps/order/proto/order.proto \
  -dgoOutput=apps/order/api/order/order.d.go
```

### CLI Options

- `-type`: Code generation type (use `backend-contract` for Go contract generation)
- `-protofilePath`: Path to the `.proto` file to process
- `-dgoOutput`: Output path for the generated `.d.go` file
- `-help`: Show help message and available options

## Generated Code Structure

### NATS Subject Constants

The generator creates constants for NATS subjects based on the service structure:

```go
const (
    NATS_SUBJECT = "/api/v1/order"
    ORDER_CREATE_ORDER = NATS_SUBJECT + "/CreateOrder"
    ORDER_GET_ORDER_BY_ID = NATS_SUBJECT + "/GetOrderById"
)
```

### Service Interface

A Go interface is generated matching your protobuf service definition:

```go
type OrderService interface {
    CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error)
    GetOrderById(ctx context.Context, req *GetOrderByIdRequest) (*OrderResponse, error)
}
```

### Proxy Implementation

A proxy struct that wraps your service implementation:

```go
type OrderServiceProxy struct {
    service OrderService
}

func NewOrderServiceProxy(service OrderService) *OrderServiceProxy {
    return &OrderServiceProxy{service: service}
}
```

### Router Registration

A router that registers service methods with NATS:

```go
type OrderServiceRouter struct {
    proxy *OrderServiceProxy
}

func (r *OrderServiceRouter) Register(natsRouter custom_nats.Router) {
    natsRouter.RegisterRoute("POST", ORDER_CREATE_ORDER, r.proxy.CreateOrder)
    natsRouter.RegisterRoute("POST", ORDER_GET_ORDER_BY_ID, r.proxy.GetOrderById)
}
```

## Example Usage in Service

After generating the contract file, you can use it in your service:

```go
package main

import (
    "context"
    custom_nats "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/custom-nats"
    "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/order/api/order"
)

// Implement your service
type orderServiceImpl struct{}

func (s *orderServiceImpl) CreateOrder(ctx context.Context, req *order.CreateOrderRequest) (*order.CreateOrderResponse, error) {
    // Your business logic here
    return &order.CreateOrderResponse{OrderId: "order-123"}, nil
}

func (s *orderServiceImpl) GetOrderById(ctx context.Context, req *order.GetOrderByIdRequest) (*order.OrderResponse, error) {
    // Your business logic here
    return &order.OrderResponse{Id: req.Id, Name: "Sample Order"}, nil
}

func main() {
    // Create service implementation
    serviceImpl := &orderServiceImpl{}
    
    // Create proxy and router
    proxy := order.NewOrderServiceProxy(serviceImpl)
    router := order.NewOrderServiceRouter(proxy)
    
    // Register with NATS
    natsRouter := custom_nats.NewRouter()
    router.Register(natsRouter)
    
    // Start NATS server
    natsRouter.Start()
}
```

## Proto File Requirements

Your `.proto` file should follow these conventions:

1. **Package Declaration**: Must include a package name
2. **Go Package Option**: Must specify `option go_package`
3. **Service Definition**: Define your gRPC service with RPC methods
4. **Message Types**: Define request and response message types

Example proto file:

```protobuf
syntax = "proto3";

package order;

option go_package = "order/api/order";

service OrderService {
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);
  rpc GetOrderById(GetOrderByIdRequest) returns (OrderResponse);
}

message CreateOrderRequest {
  string customer_id = 1;
}

message CreateOrderResponse {
  string order_id = 1;
}

message GetOrderByIdRequest {
  string Id = 1;
}

message OrderResponse {
  string Id = 1;
  string Name = 2;
}
```

## Template Customization

The generator uses Go templates located in `pkg/codegen/proto2dgo/templates/`. You can modify the `generated.d.tmpl` file to customize the generated code structure.

## Troubleshooting

### Common Issues

1. **Proto file not found**: Ensure the proto file path is correct and the file exists
2. **Invalid go_package**: Make sure your proto file has a valid `option go_package` declaration
3. **Output directory doesn't exist**: The tool will create the output directory automatically
4. **Template parsing errors**: Check that the template file is valid and accessible

### Debug Mode

To see detailed logs during generation, you can modify the CLI tool or add debug output to trace the generation process.



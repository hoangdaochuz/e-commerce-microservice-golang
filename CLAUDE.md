# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a microservices-based e-commerce application built with Go, implementing modern cloud-native patterns including DDD (Domain Driven Design), NATS messaging, API Gateway pattern, and database-per-service architecture.

## Development Commands

### Building and Running Services

```bash
# Run API Gateway (main entry point)
go run cmd/main.go

# Run specific services
go run apps/auth/cmd/main.go
go run apps/order/cmd/main.go

# Run all backend services (via taskfile)
task backend:dev
```

### Testing

```bash
# Run all unit tests
task test:unit

# Run a specific test file
go test -v ./apps/order/services/order/test/order_service_test.go

# Generate mocks for interfaces
task mocks:generate
```

### Code Generation

```bash
# Generate Go code from proto files (protobuf + gRPC)
task backend:codegen -- apps/order/proto/order.proto

# Generate .d.go contract files (NATS subjects, interfaces, proxies, routers)
task backend:codegen:dgo -- apps/order/proto/order.proto

# Generate .d.go with custom output
task backend:codegen:dgo OUT_FILE=custom/path/output.d.go -- apps/order/proto/order.proto
```

### Linting

```bash
# Run golangci-lint (configured in .golangci.yaml)
golangci-lint run ./...
```

### Docker Infrastructure

```bash
# Start all infrastructure (NATS, PostgreSQL, Redis, Loki, Tempo)
docker compose -f ./infra/docker-compose.yml -p e-commerce-microservice-golang up
```

## Architecture

### Service Structure

The project uses Domain Driven Design with 11 main microservices, each with its own database:

| Service | Database Type | Description |
|---------|--------------|-------------|
| `apps/auth` | - | Authentication service |
| `apps/order` | PostgreSQL | Order management |
| `apps/product` | PostgreSQL | Product catalog |
| `apps/shop` | PostgreSQL | Shop information |
| `apps/user` | MongoDB | User profiles |
| `apps/address` | PostgreSQL | Address management |
| `apps/settings` | MongoDB | User settings |
| `apps/cart` | Redis/PostgreSQL | Shopping cart |
| `apps/voucher` | PostgreSQL | Voucher management |
| `apps/comment` | MongoDB | Product comments |
| `apps/notification` | MongoDB | Notifications |
| `apps/payment` | PostgreSQL | Payment processing |

### Communication Flow

```
Client → API Gateway (port 8080) → NATS → Microservices → Databases
```

- **API Gateway** (`cmd/main.go`, `api_gateway/`): Entry point that transforms HTTP requests to NATS messages
- **NATS** (`pkg/custom-nats/`): Message broker for inter-service communication using request-response pattern
- **Services** (`apps/*/`): Domain services that subscribe to NATS subjects

### Key Packages

| Package | Purpose |
|---------|---------|
| `pkg/custom-nats/` | NATS client, server, router, and HTTP↔NATS request transformation |
| `pkg/circuitbreaker/` | Circuit breaker pattern implementation using `sony/gobreaker` |
| `pkg/dependency-injection/` | DI container using `go.uber.org/dig` |
| `pkg/repo/` | Database client abstractions (PostgreSQL/sqlx, MongoDB, GORM) |
| `pkg/zitadel/` | Zitadel OAuth/OIDC authentication integration |
| `pkg/codegen/` | Code generators for proto contracts and frontend TypeScript |
| `pkg/tracing/` | OpenTelemetry tracing (OTLP) |
| `pkg/metric/` | Prometheus metrics collection |
| `pkg/cache/` | Redis and in-memory caching |

## Code Generation Patterns

### Proto to Go Contract (.d.go files)

Each microservice has:
1. **Proto file**: `apps/{service}/proto/{service}.proto`
2. **Generated code**: `apps/{service}/api/{service}/{service}.pb.go` (protoc)
3. **Contract file**: `apps/{service}/api/{service}/{service}.d.go` (custom generator)

The `.d.go` file contains:
- `NATS_SUBJECT` constant for the service
- Per-method NATS subject constants
- Service interface
- Proxy implementation
- Router with NATS registration

**After modifying proto files**, always regenerate both:
```bash
task backend:codegen -- apps/{service}/proto/{service}.proto
task backend:codegen:dgo -- apps/{service}/proto/{service}.proto
```

### Service Registration Pattern

Services use a consistent pattern (`apps/auth/cmd/main.go`, `apps/order/cmd/main.go`):

```go
// 1. Connect to NATS
natsConn, _ := nats.Connect(config.NatsAuth.NATSUrl, nats.UserInfo(...))

// 2. Create chi router
chiRouter := chi.NewRouter()
router := custom_nats.NewRouter(chiRouter)

// 3. Resolve service implementation from DI container
var serviceApp *MyServiceApp
_ = di.Resolve(func(impl *MyServiceApp) {
    serviceApp = impl
})

// 4. Create proxy and router
serviceProxy := api.NewMyServiceProxy(serviceApp)
serviceRouter := api.NewMyServiceRouter(serviceProxy)

// 5. Start NATS server
server := custom_nats.NewServer(natsConn, router, api.NATS_SUBJECT, serviceRouter, &custom_nats.ServerConfig{...})
server.Start()
```

## Dependency Injection

The project uses `go.uber.org/dig` via `pkg/dependency-injection/`:

```go
// Register a constructor
di.Make(func() *MyService {
    return &MyService{...}
})

// Resolve dependencies
di.Resolve(func(service *MyService) {
    // use service
})
```

Registration typically happens in `init()` functions via blank imports:
```go
import _ "github.com/hoangdaochuz/ecommerce-microservice-golang/apps/auth/handler/session"
```

## Circuit Breaker Configuration

Circuit breakers are configured in `configs/config.yaml` and protect:
- NATS service calls (per-service configuration)
- Database connections
- External API calls (Zitadel)

The implementation in `pkg/circuitbreaker/` uses `sony/gobreaker` with states: CLOSED → OPEN → HALF-OPEN.

## Configuration

Configuration is loaded from `configs/config.yaml` using Viper. Key sections:
- `nats_auth`: NATS connection credentials
- `service_registry`: Service discovery settings
- `circuit_breaker`: Circuit breaker thresholds per service
- `zitadel_configs`: OAuth/OIDC endpoints
- `apigateway`: API Gateway port (default: 8080)
- `order_database`, `redis`, etc.: Database connection strings

## Testing

- Unit tests use `testify` for assertions
- Mocks are generated with `mockery` (configured via `//go:generate` directives or `task mocks:generate`)
- Test files follow `*_test.go` naming convention

## Observability

- **Tracing**: OpenTelemetry with OTLP endpoint (config: `general_config.otlp_endpoint`)
- **Metrics**: Prometheus, exposed at `/metrics`
- **Logging**: Structured logging via `pkg/logging/`

## Linting Configuration

The project uses `golangci-lint` with configuration in `.golangci.yaml`:
- Excludes generated files (`*.d.go`, `gen/`)
- Disables some checks for test files
- Configures `revive`, `gosec`, and other linters

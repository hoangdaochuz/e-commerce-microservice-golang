# E-commerce Microservice in Go

A microservices-based e-commerce application built with Go, implementing modern cloud-native patterns including DDD (Domain Driven Design), NATS messaging, API Gateway pattern, and database-per-service architecture.

## Tech Stack

| Category | Technology |
|----------|-----------|
| Language | Go 1.25 |
| Messaging | NATS (with JetStream) |
| Databases | PostgreSQL, MongoDB, Redis |
| HTTP Router | Chi |
| Auth | Zitadel (OAuth2/OIDC) |
| Observability | OpenTelemetry, Prometheus, Grafana, Loki, Tempo |
| Resilience | Circuit Breaker (sony/gobreaker), Rate Limiting |
| DI | uber/dig |
| Frontend | Next.js 16, React 19, TypeScript, Tailwind CSS |

## Communication Flow

```
Client --> API Gateway (port 8080) --> NATS --> Microservices --> Databases
```

- **API Gateway** (`cmd/main.go`, `api_gateway/`): Translates HTTP requests to NATS messages with middleware (CORS, rate limiting, auth, metrics, tracing)
- **NATS** (`pkg/custom-nats/`): Message broker for inter-service communication using request-response pattern
- **Services** (`apps/*/`): Domain services that subscribe to NATS subjects

## Project Structure

```
.
├── api_gateway/          # API Gateway (HTTP -> NATS proxy, middleware)
├── apps/
│   ├── auth/             # Authentication service (OAuth2/OIDC, JWT, sessions)
│   ├── order/            # Order management service (PostgreSQL)
│   ├── product/          # Product service (in progress)
│   ├── nats_auth/        # NATS authentication callout service
│   └── main.go           # Placeholder
├── cmd/
│   └── main.go           # API Gateway entry point
├── configs/              # YAML configuration (Viper)
├── docs/                 # Architecture docs, DB design, implementation plans
├── frontend/
│   ├── apis/             # Generated TypeScript API clients
│   └── customer_app/     # Next.js customer-facing app
├── infra/                # Docker Compose, NATS config, observability configs, k8s
├── migrationer/          # Database migration scripts
├── mocks/                # Generated test mocks (mockery)
├── pkg/                  # Shared libraries (see below)
├── shared/               # Shared constants, error types, context keys
├── tools/                # Migration runner CLI
└── taskfile.yml          # Task runner commands
```

### Shared Packages (`pkg/`)

| Package | Purpose |
|---------|---------|
| `cache` | In-memory and Redis caching abstraction |
| `circuitbreaker` | Circuit breaker pattern (sony/gobreaker) |
| `codegen` | Code generation CLI (proto -> .d.go, frontend TypeScript) |
| `configs` | Viper-based YAML configuration loading |
| `custom-nats` | NATS client, server, router, HTTP-to-NATS transformation |
| `dependency-injection` | DI container using uber/dig |
| `httpclient` | HTTP client with circuit breaker |
| `logging` | Zap-based structured logging |
| `metric` | Prometheus metrics collection |
| `migration` | Database migration framework (MongoDB, PostgreSQL) |
| `rate_limiter` | Redis-based distributed rate limiting |
| `redis` | Redis client initialization |
| `repo` | Repository pattern abstraction (PostgreSQL/sqlx, MongoDB, GORM) |
| `tracing` | OpenTelemetry distributed tracing (OTLP) |
| `utils` | Utility functions (JSON, struct conversion) |
| `zitadel` | OAuth2/OIDC integration with Zitadel |

## Services

### Implemented

| Service | Path | Database | Description |
|---------|------|----------|-------------|
| Auth | `apps/auth` | Redis (sessions) | OAuth2/OIDC authentication via Zitadel, JWT tokens, session management |
| Order | `apps/order` | PostgreSQL | Order processing and management |
| NATS Auth | `apps/nats_auth` | - | Custom NATS authentication callout using Xkey encryption |

### Planned (per DDD design)

| Service | Database | Description |
|---------|----------|-------------|
| Product | PostgreSQL | Product catalog, categories, inventory |
| Shop | PostgreSQL | Shop information |
| User | MongoDB | User profiles |
| Address | PostgreSQL | Address management |
| Settings | MongoDB | User settings |
| Cart | Redis + PostgreSQL | Shopping cart with backup for analytics |
| Voucher | PostgreSQL | Voucher management |
| Comment | MongoDB | Product comments and ratings |
| Notification | MongoDB | User notifications |
| Payment | PostgreSQL | Payment processing |

## Database Design

We use the **Database per Service** pattern. Each service owns its database, chosen based on domain requirements:

- **PostgreSQL**: For services needing ACID transactions, complex joins, and relational integrity (Order, Product, Shop, Address, Voucher, Payment)
- **MongoDB**: For services with flexible/unstructured schemas (User, Settings, Comment, Notification)
- **Redis**: For high-frequency read/write data (Cart, Sessions)

The full database schema is in `docs/database-design/schema/Ecommerce-db.sql`.

## Getting Started

### Prerequisites

- Go 1.25+
- Docker & Docker Compose
- [Task](https://taskfile.dev/) (task runner)
- [Bun](https://bun.sh/) (for frontend)

### Start Infrastructure

```bash
docker compose -f ./infra/docker-compose.yml -p e-commerce-microservice-golang up
```

This starts: NATS, PostgreSQL, MongoDB, Redis, Prometheus, Grafana, Loki, Tempo, and the OpenTelemetry collector (Alloy).

### Run Backend Services

```bash
# Run all backend services (API Gateway + Auth + Order)
task backend:dev

# Or run individually
task backend:api_gateway
task backend:authenticate
task backend:order
```

### Run Frontend

```bash
cd frontend/customer_app
bun install
bun dev
```

## Development

### Code Generation

Each microservice uses proto files to define its API contract. After modifying a `.proto` file:

```bash
# Generate Go code from proto (protobuf + gRPC)
task backend:codegen -- apps/{service}/proto/{service}.proto

# Generate .d.go contract file (NATS subjects, interfaces, proxies, routers)
task backend:codegen:dgo -- apps/{service}/proto/{service}.proto

# Generate TypeScript clients for frontend
task frontend:codegen:service    # Single service
task frontend:codegen:all        # All services
```

### Testing

```bash
# Run all unit tests
task test:unit

# Generate mocks
task mocks:generate

# Clean and regenerate mocks
task mocks:regenerate
```

### Linting

```bash
golangci-lint run ./...
```

### Database Migrations

```bash
go run tools/migration/main.go
```

## Observability

The project includes a full observability stack:

| Tool | Port | Purpose |
|------|------|---------|
| Prometheus | 9090 | Metrics collection |
| Grafana | 3001 | Dashboards and visualization |
| Loki | 3100 | Log aggregation |
| Tempo | 3200 | Distributed tracing backend |
| Alloy | 4317, 4318 | OpenTelemetry collector |

- **Tracing**: OpenTelemetry with OTLP export to Tempo
- **Metrics**: Prometheus client, exposed at `/metrics`
- **Logging**: Structured logging via Zap, aggregated to Loki

## Resilience

- **Circuit Breaker**: Configured per service for NATS calls, database connections, and external APIs (Zitadel). Uses three states: CLOSED -> OPEN -> HALF-OPEN. Configuration in `configs/config.yaml`.
- **Rate Limiting**: Redis-based distributed rate limiter (default: 50 requests/minute) applied at the API Gateway.

## Configuration

All configuration is managed via `configs/config.yaml` using Viper. Key sections:

- `nats_auth` - NATS connection and authentication
- `apigateway` - Gateway port (default: 8080)
- `order_database` - PostgreSQL connection for order service
- `mongo_db` - MongoDB connection
- `redis` - Redis connection
- `zitadel_configs` - OAuth2/OIDC provider endpoints
- `circuit_breaker` - Per-service circuit breaker thresholds
- `general_config` - OTLP endpoint, backend/frontend URLs

# E-Commerce Microservice - Architectural Completion Plan

## Executive Summary

This document outlines a comprehensive architectural roadmap to complete the e-commerce microservice system. The plan focuses exclusively on architectural components, infrastructure, and system design patterns—excluding business logic implementation.

**Current Status:** 3/11 services implemented (Auth, Order, NATS Auth)  
**Target:** Complete microservice architecture with production-ready infrastructure  
**Timeline:** 16-20 weeks  
**Approach:** Incremental, layer-by-layer implementation

### Progress Overview (Updated: December 2025)

| Phase | Status | Progress |
|-------|--------|----------|
| Phase 1: Core Infrastructure | ✅ 80% Done | Circuit Breaker, Observability complete |
| Phase 2: Remaining Microservices | ❌ Not Started | 0/8 services |
| Phase 3: Advanced Patterns | ❌ Not Started | SAGA, CQRS, Events |
| Phase 4: Production Readiness | ❌ Not Started | K8s, CI/CD, Security |
| Phase 5: Optimization | ❌ Not Started | Load testing, Scaling |

**Completed Components:**
- ✅ **Observability Stack**: OpenTelemetry + Tempo (Traces), Prometheus (Metrics), Zap + Loki (Logs), Grafana (Dashboards)
- ✅ **Circuit Breaker**: gobreaker v2 with registry pattern
- ✅ **HTTP Client**: With circuit breaker and auto-tracing (otelhttp)
- ✅ **Rate Limiting**: Redis-based implementation

---

## Table of Contents

1. [Current State Analysis](#1-current-state-analysis)
2. [Target Architecture](#2-target-architecture)
3. [Implementation Roadmap](#3-implementation-roadmap)
4. [Phase 1: Core Infrastructure](#phase-1-core-infrastructure-weeks-1-4)
5. [Phase 2: Remaining Microservices](#phase-2-remaining-microservices-weeks-5-10)
6. [Phase 3: Advanced Patterns](#phase-3-advanced-patterns-weeks-11-14)
7. [Phase 4: Production Readiness](#phase-4-production-readiness-weeks-15-18)
8. [Phase 5: Optimization & Scaling](#phase-5-optimization--scaling-weeks-19-20)
9. [Deployment Architecture](#9-deployment-architecture)
10. [Success Criteria](#10-success-criteria)

---

## 1. Current State Analysis

### 1.1 Implemented Components ✅

```
Infrastructure Layer:
├── NATS (Message Broker)
├── PostgreSQL (for Orders)
├── MongoDB (planned for Users, Comments, Notifications, Settings)
├── Redis (Cache & Sessions)
├── Zitadel (External OAuth)
└── ✅ Observability Stack (NEW)
    ├── Grafana Alloy (OTLP Collector + Log Forwarder)
    ├── Grafana Tempo (Distributed Tracing)
    ├── Prometheus (Metrics)
    ├── Grafana Loki (Log Aggregation)
    └── Grafana (Visualization)

Application Layer:
├── API Gateway (HTTP → NATS)
│   ├── Middleware (CORS, Logging, Rate Limiting, Auth)
│   ├── Request/Response transformation
│   ├── ✅ Distributed Tracing (otelhttp)
│   ├── ✅ Metrics Middleware (Prometheus)
│   └── ✅ Circuit Breaker Integration
├── Auth Service (Complete)
│   ├── OAuth flow (Zitadel integration)
│   ├── Session management (Redis)
│   ├── JWT handling
│   └── ✅ Tracing enabled
└── Order Service (Partial)
    ├── Database schema
    ├── Repository layer
    ├── gRPC service definition
    └── ✅ Tracing enabled

Shared Packages:
├── Custom NATS (Client/Server framework)
│   └── ✅ Trace propagation support
├── Dependency Injection
├── Repository abstraction (PostgreSQL, MongoDB)
├── Redis client
├── Rate limiter
├── Cache abstraction
├── Code generators (Proto → TypeScript, Proto → .d.go)
├── ✅ pkg/tracing/ (OpenTelemetry SDK)
├── ✅ pkg/logging/ (Zap structured logging)
├── ✅ pkg/metric/ (Prometheus metrics)
├── ✅ pkg/circuitbreaker/ (gobreaker v2)
└── ✅ pkg/httpclient/ (HTTP client with circuit breaker)
```

### 1.2 Missing Components ❌ / Partially Done ⚠️

```
Services (8 missing):
├── Product Service
├── Shop Service
├── User Service
├── Address Service
├── Settings Service
├── Cart Service
├── Voucher Service
├── Comment Service
├── Notification Service
└── Payment Service

Infrastructure Components:
├── ✅ Circuit breakers (pkg/circuitbreaker/ - gobreaker v2)
├── ✅ Distributed tracing (OpenTelemetry + Tempo)
├── ✅ Centralized logging (Zap + Loki + Grafana Alloy)
├── ✅ Metrics & monitoring (Prometheus + Grafana)
├── ❌ API documentation (Swagger/OpenAPI)
├── ✅ Service registry/discovery (via NATS)
├── ⚠️ Event-driven patterns (NATS JetStream) - partial
└── ❌ Background job processing

DevOps:
├── ❌ Kubernetes manifests
├── ❌ Helm charts
├── ❌ CI/CD pipelines
├── ❌ Infrastructure as Code (Terraform)
├── ❌ Database migrations
└── ✅ Service health checks (/health endpoint)

Security:
├── ❌ Service-to-service authentication (mTLS)
├── ❌ API key management
├── ❌ Secrets management
├── ✅ Rate limiting (pkg/rate_limiter/)
└── ⚠️ Request validation (partial)

Observability: ✅ COMPLETED
├── ✅ Distributed tracing (OpenTelemetry + Grafana Tempo)
├── ✅ Metrics collection (Prometheus + HTTP Middleware)
├── ✅ Log aggregation (Zap + Grafana Loki + Alloy)
└── ✅ Service Graph & Span Metrics (via Alloy connectors)
```

### 1.3 Architecture Quality Assessment

| Component | Status | Quality | Notes |
|-----------|--------|---------|-------|
| API Gateway | ✅ Complete | Good | Includes circuit breaker, tracing, metrics |
| NATS Framework | ✅ Complete | Good | Custom implementation with trace propagation |
| Dependency Injection | ✅ Complete | Good | Using uber/dig |
| Repository Pattern | ✅ Complete | Good | Abstracted for SQL/NoSQL |
| Authentication | ✅ Complete | Good | OAuth2 with Zitadel |
| Authorization | ⚠️ Partial | Fair | Basic implementation, needs RBAC |
| Service Template | ⚠️ Partial | Fair | Only 2 services as reference |
| Testing | ⚠️ Partial | Fair | Basic unit tests (jwt, circuitbreaker) |
| **Observability** | ✅ Complete | Good | Full stack: Tracing, Metrics, Logging |
| **Circuit Breaker** | ✅ Complete | Good | gobreaker v2, registry pattern |
| **HTTP Client** | ✅ Complete | Good | With circuit breaker + otelhttp |
| Documentation | ⚠️ Partial | Fair | Added observability docs |

---

## 2. Target Architecture

### 2.1 High-Level Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                         CLIENT LAYER                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐             │
│  │   Web App    │  │  Mobile App  │  │  Admin Panel │             │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘             │
└─────────┼──────────────────┼──────────────────┼──────────────────────┘
          │                  │                  │
          └──────────────────┴──────────────────┘
                             │
                    ┌────────▼─────────┐
                    │   CDN / LB       │
                    └────────┬─────────┘
┌────────────────────────────┼─────────────────────────────────────────┐
│                    EDGE LAYER                                         │
│            ┌───────────────▼───────────────┐                         │
│            │      API Gateway               │                         │
│            │  ┌──────────────────────────┐ │                         │
│            │  │ Middleware Pipeline      │ │                         │
│            │  │ - Authentication         │ │                         │
│            │  │ - Rate Limiting          │ │                         │
│            │  │ - Request Validation     │ │                         │
│            │  │ - Circuit Breaker        │ │                         │
│            │  │ - Logging/Tracing        │ │                         │
│            │  └──────────────────────────┘ │                         │
│            └───────────────┬───────────────┘                         │
└────────────────────────────┼─────────────────────────────────────────┘
                             │
┌────────────────────────────┼─────────────────────────────────────────┐
│              MESSAGE BUS (NATS JetStream)                             │
│  ┌──────────────────────────┼────────────────────────────┐           │
│  │  Subjects/Channels       │                            │           │
│  │  - auth.*               │  Pub/Sub Streams            │           │
│  │  - order.*              │  - events.order             │           │
│  │  - product.*            │  - events.payment           │           │
│  │  - payment.*            │  - events.notification      │           │
│  │  - user.*               │  - events.inventory         │           │
│  │  - notification.*       │                             │           │
│  └─────────────────────────────────────────────────────┘            │
└────────────────────────────┼─────────────────────────────────────────┘
                             │
┌────────────────────────────┼─────────────────────────────────────────┐
│                    MICROSERVICES LAYER                                │
│                             │                                          │
│  ┌──────────┬──────────┬───┴────┬──────────┬──────────┬──────────┐  │
│  │          │          │        │          │          │          │  │
│  │  Auth    │  User    │ Product│   Shop   │  Order   │  Cart    │  │
│  │ Service  │ Service  │ Service│ Service  │ Service  │ Service  │  │
│  │          │          │        │          │          │          │  │
│  │  ┌────┐  │  ┌────┐  │ ┌────┐ │  ┌────┐  │  ┌────┐  │  ┌────┐  │  │
│  │  │API │  │  │API │  │ │API │ │  │API │  │  │API │  │  │API │  │  │
│  │  └─┬──┘  │  └─┬──┘  │ └─┬──┘ │  └─┬──┘  │  └─┬──┘  │  └─┬──┘  │  │
│  │  ┌─▼──┐  │  ┌─▼──┐  │ ┌─▼──┐ │  ┌─▼──┐  │  ┌─▼──┐  │  ┌─▼──┐  │  │
│  │  │BIZ │  │  │BIZ │  │ │BIZ │ │  │BIZ │  │  │BIZ │  │  │BIZ │  │  │
│  │  └─┬──┘  │  └─┬──┘  │ └─┬──┘ │  └─┬──┘  │  └─┬──┘  │  └─┬──┘  │  │
│  │  ┌─▼──┐  │  ┌─▼──┐  │ ┌─▼──┐ │  ┌─▼──┐  │  ┌─▼──┐  │  ┌─▼──┐  │  │
│  │  │REPO│  │  │REPO│  │ │REPO│ │  │REPO│  │  │REPO│  │  │REPO│  │  │
│  │  └────┘  │  └────┘  │ └────┘ │  └────┘  │  └────┘  │  └────┘  │  │
│  └────┬─────┴────┬─────┴────┬───┴────┬─────┴────┬─────┴────┬─────┘  │
│       │          │          │        │          │          │        │
│  ┌────┴────┬─────┴─────┬────┴───┬────┴─────┬────┴─────┬────┴─────┐  │
│  │         │           │        │          │          │          │  │
│  │ Address │  Voucher  │Comment │  Notif.  │ Payment  │ Settings │  │
│  │ Service │  Service  │Service │ Service  │ Service  │ Service  │  │
│  │         │           │        │          │          │          │  │
│  └────┬────┴─────┬─────┴────┬───┴────┬─────┴────┬─────┴────┬─────┘  │
└───────┼──────────┼──────────┼────────┼──────────┼──────────┼─────────┘
        │          │          │        │          │          │
┌───────┼──────────┼──────────┼────────┼──────────┼──────────┼─────────┐
│                      DATA LAYER                                       │
│  ┌────▼────┐  ┌──▼───────┐  ┌──▼─────┐  ┌───▼──────┐  ┌────▼─────┐  │
│  │PostgreSQL│ │PostgreSQL│  │MongoDB │  │  MongoDB │  │  Redis   │  │
│  │  (Auth)  │ │(Orders)  │  │(Users) │  │(Comments)│  │  (Cache) │  │
│  │          │ │          │  │        │  │          │  │          │  │
│  │(Products)│ │(Payments)│  │(Settings)│ │(Notif.) │  │ (Session)│  │
│  │(Vouchers)│ │(Addresses)│ │        │  │          │  │  (Cart)  │  │
│  │  (Shop)  │ │          │  │        │  │          │  │          │  │
│  └──────────┘ └──────────┘  └────────┘  └──────────┘  └──────────┘  │
└───────────────────────────────────────────────────────────────────────┘
┌───────────────────────────────────────────────────────────────────────┐
│                    OBSERVABILITY LAYER ✅ IMPLEMENTED                 │
│                                                                       │
│                    ┌──────────────────┐                              │
│                    │   Grafana Alloy  │  OTLP Collector              │
│                    │  (4317/4318)     │  Log Forwarder               │
│                    └────────┬─────────┘                              │
│                             │                                         │
│  ┌──────────────┐  ┌───────▼──────┐  ┌──────────────┐               │
│  │  Prometheus  │  │    Tempo     │  │     Loki     │               │
│  │  (Metrics)   │◄─│   (Traces)   │  │    (Logs)    │               │
│  │   :9090      │  │    :3200     │  │    :3100     │               │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘               │
│         │                  │                  │                       │
│         └──────────────────┴──────────────────┘                       │
│                            │                                          │
│                    ┌───────▼────────┐                                │
│                    │    Grafana     │                                │
│                    │   :3001        │                                │
│                    └────────────────┘                                │
└───────────────────────────────────────────────────────────────────────┘
```

### 2.2 Service Architecture Standard

Each microservice follows this structure:

```
apps/{service-name}/
├── cmd/
│   └── main.go                    # Service entry point
├── api/{service}/                 # Generated gRPC code
│   ├── {service}_grpc.pb.go
│   ├── {service}.pb.go
│   └── {service}.d.go             # TypeScript definitions
├── proto/
│   └── {service}.proto            # Service contract
├── handler/{service}/
│   └── {service}.go               # gRPC handlers
├── services/{service}/
│   └── {service}_service.go       # Business logic layer
├── repository/
│   ├── repo.go                    # Repository interface
│   └── {entity}_model.go          # Data models
├── domains/
│   └── {entity}_model.go          # Domain models
├── configs/
│   └── {service}_config.go        # Service configuration
├── db/
│   └── init_{service}_schema.sql  # Database initialization
├── middleware/
│   └── {service}_middleware.go    # Service-specific middleware
└── tests/
    ├── unit/
    ├── integration/
    └── e2e/
```

### 2.3 Communication Patterns

```
┌─────────────────────────────────────────────────────────────────┐
│                    Communication Patterns                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. SYNCHRONOUS (Request-Response via NATS)                     │
│     Client → API Gateway → NATS Request → Service → Response    │
│     Use cases: Queries, immediate responses                     │
│                                                                  │
│  2. ASYNCHRONOUS (Pub/Sub via NATS JetStream)                   │
│     Service → NATS Publish → Stream → Subscribers               │
│     Use cases: Events, notifications, eventual consistency      │
│                                                                  │
│  3. INTER-SERVICE (Service-to-Service via NATS)                 │
│     Service A → NATS → Service B                                │
│     Use cases: Data aggregation, cross-service queries          │
│                                                                  │
│  4. CACHE-ASIDE (Redis)                                         │
│     Service → Check Cache → [Miss] → DB → Update Cache          │
│     Use cases: Product catalog, user profiles                   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 3. Implementation Roadmap

### 3.1 Phased Approach

```
Phase 1: Core Infrastructure (4 weeks) - ✅ 80% COMPLETED
├── ✅ Circuit breaker pattern (pkg/circuitbreaker/)
├── ✅ Distributed tracing (OpenTelemetry + Tempo)
├── ✅ Metrics & monitoring (Prometheus + HTTP Middleware)
├── ✅ Structured logging (Zap + Loki)
├── ❌ Service template generator
└── ❌ Database migration system

Phase 2: Remaining Microservices (6 weeks)
├── Product Service (Week 5-6)
├── User Service (Week 6-7)
├── Shop Service (Week 7)
├── Cart Service (Week 8)
├── Payment Service (Week 8-9)
├── Voucher Service (Week 9)
├── Address Service (Week 9)
├── Comment Service (Week 10)
├── Notification Service (Week 10)
└── Settings Service (Week 10)

Phase 3: Advanced Patterns (4 weeks)
├── Event-driven architecture (JetStream)
├── SAGA pattern (distributed transactions)
├── CQRS pattern (read/write separation)
└── Background job processing

Phase 4: Production Readiness (4 weeks)
├── Kubernetes deployment
├── CI/CD pipelines
├── Security hardening
├── Performance optimization
└── Disaster recovery

Phase 5: Optimization & Scaling (2 weeks)
├── Load testing & tuning
├── Horizontal scaling
├── Database optimization
└── Documentation finalization
```

### 3.2 Priority Matrix

| Component | Priority | Complexity | Dependencies | Timeline | Status |
|-----------|----------|------------|--------------|----------|--------|
| Circuit Breaker | P0 | Medium | None | Week 1 | ✅ Done |
| Tracing | P0 | Low | None | Week 1-2 | ✅ Done |
| Metrics | P0 | Low | Tracing | Week 2 | ✅ Done |
| Logging (Loki) | P0 | Low | None | Week 2 | ✅ Done |
| Service Template | P0 | Medium | None | Week 3 | ❌ Pending |
| Product Service | P0 | High | Template | Week 5-6 | ❌ Pending |
| User Service | P0 | High | Template | Week 6-7 | ❌ Pending |
| Cart Service | P0 | Medium | Product, User | Week 8 | ❌ Pending |
| Payment Service | P0 | High | Order | Week 8-9 | ❌ Pending |
| Shop Service | P1 | Medium | Product | Week 7 | ❌ Pending |
| Event System | P1 | High | All services | Week 11-12 | ❌ Pending |
| SAGA Pattern | P1 | High | Event System | Week 13 | ❌ Pending |
| Kubernetes | P0 | High | All services | Week 15-16 | ❌ Pending |
| CI/CD | P0 | Medium | None | Week 17 | ❌ Pending |

---

## Phase 1: Core Infrastructure (Weeks 1-4) - ✅ MOSTLY COMPLETED

### Week 1: Resilience Patterns ✅ COMPLETED

#### 1.1 Circuit Breaker Implementation ✅
*See circuit-breaker-implementation-plan.md for details*

**Deliverables:** ✅ All completed
- ✅ `pkg/circuitbreaker/` package (using sony/gobreaker v2)
- ✅ NATS circuit breaker wrapper (`pkg/custom-nats/`)
- ✅ HTTP Client with circuit breaker (`pkg/httpclient/breaker_client.go`)
- ✅ Configuration management (`pkg/circuitbreaker/config.go`)
- ✅ Registry pattern (`pkg/circuitbreaker/registry.go`)
- ⚠️ Unit tests (basic tests exist)

#### 1.2 Retry Mechanism

```go
// pkg/retry/retry.go

package retry

import (
    "context"
    "time"
)

type Config struct {
    MaxAttempts     int
    InitialDelay    time.Duration
    MaxDelay        time.Duration
    Multiplier      float64
    RandomizationFactor float64
}

type Retryer struct {
    config Config
}

func NewRetryer(config Config) *Retryer {
    return &Retryer{config: config}
}

// ExecuteWithRetry executes function with exponential backoff
func (r *Retryer) ExecuteWithRetry(
    ctx context.Context,
    fn func() error,
    isRetryable func(error) bool,
) error {
    // Implementation with exponential backoff
}
```

**Integration Points:**
- NATS client
- Database operations
- External HTTP calls

### Week 2: Observability Foundation ✅ COMPLETED

#### 2.1 Distributed Tracing (OpenTelemetry + Tempo) ✅ IMPLEMENTED

**Actual Implementation (`pkg/tracing/main.go`):**

```go
// pkg/tracing/main.go - ACTUAL CODE

package tracing

type TracingConfig struct {
    ServiceName  string
    Attributes   map[string]string
    SamplingRate float64           // 0-1 ratio
    BatchTimeout time.Duration
    BatchMaxSize int
    OtelEndpoint string            // OTLP endpoint (Alloy: localhost:4317)
}

func InitializeTraceRegistry(cfg *TracingConfig) (func(), error)
func InjectTraceIntoHttpReq(ctx context.Context, req *http.Request)
func ExtractTraceFromHttpRequest(req *http.Request) context.Context
func SpanContext(ctx context.Context, header http.Header, spanName string) (context.Context, trace.Span)
func SetSpanError(span trace.Span, err error)
```

**Package Structure:** ✅ Implemented
```
pkg/tracing/
└── main.go              # Full implementation with OTLP gRPC exporter
```

**Integration Points:** ✅ All integrated
- ✅ API Gateway (otelhttp + custom middleware)
- ✅ NATS custom framework (trace propagation via headers)
- ✅ HTTP client (otelhttp.NewTransport)
- ✅ All services (auth, order) have tracing enabled

**Infrastructure:** ✅ Configured
- ✅ Grafana Alloy as OTLP Collector (`infra/alloy-config.alloy`)
- ✅ Grafana Tempo for trace storage (`infra/tempo.yaml`)
- ✅ Service Graph & Span Metrics generation

#### 2.2 Metrics Collection (Prometheus) ✅ IMPLEMENTED

**Actual Implementation:**

```go
// pkg/metric/main.go
type MetricWrapper struct {
    registry *prometheus.Registry
}

func NewMetricWrapper() *MetricWrapper
func (mw *MetricWrapper) RegisterCollectorDefault()  // Go + Process collectors
func (mw *MetricWrapper) GetRegistry() *prometheus.Registry

// pkg/metric/httpmiddleware/http_middleware.go
type Middleware struct {
    reqTotal    *prometheus.CounterVec    // http_request_total
    reqDuration *prometheus.HistogramVec  // http_request_duration_seconds
    reqSize     *prometheus.SummaryVec    // http_request_size_bytes
    resSize     *prometheus.SummaryVec    // http_response_size_bytes
}

func NewMiddleware(buckets []float64, registry *prometheus.Registry) *Middleware
func (m *Middleware) WrapHandler(path string, next http.Handler) http.HandlerFunc
```

**Metrics Exposed:** ✅ 
| Metric | Type | Labels |
|--------|------|--------|
| `http_request_total` | Counter | path, method, code |
| `http_request_duration_seconds` | Histogram | path, method, code |
| `http_request_size_bytes` | Summary | path, method, code |
| `http_response_size_bytes` | Summary | path, method, code |

**Additional Metrics via Alloy:** ✅
- Service Graph metrics (from traces)
- Span metrics with histogram buckets
- Remote write to Prometheus

**Endpoint:** `/metrics` on API Gateway (port 8080)

**Infrastructure:** ✅
- ✅ Prometheus (`infra/prometheus.yml`) - scrape config
- ✅ Alloy SpanMetrics connector
- ✅ Remote write enabled

#### 2.3 Structured Logging (Zap + Loki) ✅ IMPLEMENTED

**Actual Implementation (`pkg/logging/main.go`):**

```go
package logging

var logger *zap.Logger

func initLogger() {
    mode := viper.GetString("general_config.mode")
    
    // Development: colored, human-readable
    developmentConfig := zap.NewDevelopmentConfig()
    developmentConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    developmentConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
    
    // Production: JSON format, structured
    productionConfig := zap.NewProductionConfig()
    productionConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    // Keys: ts, level, msg, caller, stacktrace
}

func GetSugaredLogger() *zap.SugaredLogger
```

**Log Aggregation Infrastructure:** ✅
- ✅ Grafana Loki (`infra/loki-config.yaml`)
- ✅ Grafana Alloy as log forwarder (`infra/alloy-config.alloy`)

**Log Sources Configured:**
```
loki.source.file "console_logs":
├── /logs/api_gateway.log
├── /logs/auth.log
└── /logs/order.log

loki.source.docker "docker_logs":
└── All Docker containers via unix socket
```

**Log-Trace Correlation:** ✅
Logs include traceId for correlation:
```go
logging.GetSugaredLogger().Infof("%s %s %v statusCode: %v traceId: %s", 
    r.Method, r.URL.Path, duration, statusCode, span.SpanContext().TraceID().String())
```

> 📚 **Detailed Documentation:** See `/docs/observability/OBSERVABILITY.md` for comprehensive observability setup guide.

### Week 3: Service Template & Code Generation ❌ PENDING

#### 3.1 Service Generator CLI

```bash
# Create new service from template
go run pkg/servicegen/cli/main.go \
    --name=product \
    --database=postgres \
    --entities=product,category,inventory

# Generates:
# - Complete service structure
# - Proto file template
# - Repository boilerplate
# - Configuration files
# - Dockerfile
# - Kubernetes manifests
```

**Service Generator Package:**
```
pkg/servicegen/
├── cli/
│   └── main.go
├── templates/
│   ├── service.tmpl
│   ├── handler.tmpl
│   ├── repository.tmpl
│   ├── config.tmpl
│   ├── proto.tmpl
│   ├── dockerfile.tmpl
│   └── k8s/
│       ├── deployment.tmpl
│       ├── service.tmpl
│       └── configmap.tmpl
├── generator.go
└── validator.go
```

#### 3.2 Database Migration System

```go
// pkg/migration/migration.go

package migration

import (
    "database/sql"
    "embed"
)

//go:embed migrations/*.sql
var migrations embed.FS

type Migrator interface {
    Up(db *sql.DB) error
    Down(db *sql.DB) error
    Version(db *sql.DB) (int, error)
}

type PostgresMigrator struct {
    migrations embed.FS
    tableName  string
}

func NewPostgresMigrator(migrations embed.FS) *PostgresMigrator

// Apply pending migrations
func (m *PostgresMigrator) Up(db *sql.DB) error

// Rollback last migration
func (m *PostgresMigrator) Down(db *sql.DB) error
```

**Migration File Structure:**
```
apps/{service}/db/migrations/
├── 001_create_table.up.sql
├── 001_create_table.down.sql
├── 002_add_indexes.up.sql
└── 002_add_indexes.down.sql
```

### Week 4: Testing Infrastructure ❌ PENDING

#### 4.1 Testing Framework Setup

```
pkg/testing/
├── testcontainers/      # Docker containers for testing
│   ├── postgres.go
│   ├── mongodb.go
│   ├── redis.go
│   └── nats.go
├── fixtures/            # Test data
│   ├── loader.go
│   └── data/
├── mocks/               # Mock interfaces
│   └── generate.go
└── helpers/
    ├── assert.go
    └── http.go
```

#### 4.2 Test Categories

**Unit Tests:**
```go
// apps/order/services/order/order_service_test.go

func TestOrderService_CreateOrder(t *testing.T) {
    // Test business logic in isolation
    // Mock repository dependencies
}
```

**Integration Tests:**
```go
// apps/order/tests/integration/order_integration_test.go

func TestOrderEndToEnd(t *testing.T) {
    // Start test containers (DB, NATS, Redis)
    // Test full request flow
}
```

**Contract Tests:**
```go
// Test proto contract compatibility
func TestOrderServiceContract(t *testing.T) {
    // Verify API contract hasn't broken
}
```

---

## Phase 2: Remaining Microservices (Weeks 5-10)

### Service Implementation Priority & Dependencies

```
┌─────────────────────────────────────────────────────────────────┐
│                    Service Dependency Graph                      │
└─────────────────────────────────────────────────────────────────┘

        [Auth]     (Week 5 - Already exists, enhance)
           │
           ├──────────────────┐
           │                  │
        [User]             [Shop]
      (Week 6-7)         (Week 7)
           │                  │
           ├──────────┬───────┴─────────┐
           │          │                 │
      [Address]   [Product]         [Voucher]
      (Week 9)   (Week 5-6)        (Week 9)
           │          │                 │
           └────┬─────┴─────┬───────────┘
                │           │
             [Cart]      [Order]
            (Week 8)   (Already exists, enhance)
                │           │
                └─────┬─────┘
                      │
                  [Payment]
                 (Week 8-9)
                      │
          ┌───────────┴───────────┐
          │                       │
      [Comment]            [Notification]
      (Week 10)              (Week 10)
                                 │
                             [Settings]
                            (Week 10)
```

### Week 5-6: Product Service (Foundation)

**Database:** PostgreSQL  
**Entities:** Product, Category, Inventory, Product_Voucher

#### Architecture:
```
Product Service
├── Product Management
│   ├── CRUD operations
│   ├── Search & filtering
│   ├── Price management
│   └── Image handling
├── Category Management
│   ├── Hierarchical categories
│   └── Category assignment
├── Inventory Management
│   ├── Stock tracking
│   ├── Reservation system
│   └── Low stock alerts
└── Integration Points
    ├── Shop Service (product ownership)
    ├── Cart Service (availability check)
    ├── Order Service (inventory update)
    └── Voucher Service (discount application)
```

#### Proto Definition:
```protobuf
// apps/product/proto/product.proto

syntax = "proto3";

package product;

service ProductService {
    rpc CreateProduct(CreateProductRequest) returns (ProductResponse);
    rpc GetProduct(GetProductRequest) returns (ProductResponse);
    rpc UpdateProduct(UpdateProductRequest) returns (ProductResponse);
    rpc DeleteProduct(DeleteProductRequest) returns (EmptyResponse);
    rpc ListProducts(ListProductsRequest) returns (ListProductsResponse);
    rpc SearchProducts(SearchProductsRequest) returns (ListProductsResponse);
    
    // Inventory operations
    rpc CheckAvailability(CheckAvailabilityRequest) returns (AvailabilityResponse);
    rpc ReserveInventory(ReserveInventoryRequest) returns (ReservationResponse);
    rpc ReleaseInventory(ReleaseInventoryRequest) returns (EmptyResponse);
    
    // Category operations
    rpc CreateCategory(CreateCategoryRequest) returns (CategoryResponse);
    rpc ListCategories(ListCategoriesRequest) returns (ListCategoriesResponse);
}

message Product {
    string id = 1;
    string shop_id = 2;
    string name = 3;
    string description = 4;
    double price = 5;
    repeated string image_urls = 6;
    string category_id = 7;
    int32 stock = 8;
    string status = 9; // active, inactive, out_of_stock
    int64 created_at = 10;
    int64 updated_at = 11;
}
```

#### Database Schema:
```sql
-- apps/product/db/migrations/001_create_tables.up.sql

CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    parent_id UUID REFERENCES categories(id),
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID NOT NULL,
    category_id UUID REFERENCES categories(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE inventory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products(id) ON DELETE CASCADE,
    stock INT NOT NULL DEFAULT 0,
    reserved INT NOT NULL DEFAULT 0,
    available INT GENERATED ALWAYS AS (stock - reserved) STORED,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_products_shop ON products(shop_id);
CREATE INDEX idx_products_category ON products(category_id);
CREATE INDEX idx_inventory_product ON inventory(product_id);
```

### Week 6-7: User Service

**Database:** MongoDB  
**Entities:** User, Profile

#### Architecture:
```
User Service
├── User Management
│   ├── User registration (via Auth)
│   ├── Profile CRUD
│   ├── User preferences
│   └── Avatar management
├── User Queries
│   ├── Get user by ID
│   ├── Search users
│   └── Batch user info
└── Integration Points
    ├── Auth Service (user creation trigger)
    ├── Order Service (customer info)
    ├── Address Service (user addresses)
    └── Settings Service (user settings)
```

#### Proto Definition:
```protobuf
// apps/user/proto/user.proto

syntax = "proto3";

package user;

service UserService {
    rpc CreateUser(CreateUserRequest) returns (UserResponse);
    rpc GetUser(GetUserRequest) returns (UserResponse);
    rpc UpdateUser(UpdateUserRequest) returns (UserResponse);
    rpc DeleteUser(DeleteUserRequest) returns (EmptyResponse);
    rpc GetUserProfile(GetUserProfileRequest) returns (UserProfileResponse);
    rpc UpdateUserProfile(UpdateUserProfileRequest) returns (UserProfileResponse);
    rpc BatchGetUsers(BatchGetUsersRequest) returns (BatchGetUsersResponse);
}

message User {
    string id = 1;
    string email = 2;
    string username = 3;
    string status = 4; // active, suspended, deleted
    int64 created_at = 5;
    int64 updated_at = 6;
}

message UserProfile {
    string user_id = 1;
    string first_name = 2;
    string last_name = 3;
    string phone = 4;
    string avatar_url = 5;
    string date_of_birth = 6;
    string gender = 7;
}
```

#### MongoDB Schema:
```javascript
// User collection
{
  _id: ObjectId,
  email: String,
  username: String,
  status: String,
  created_at: ISODate,
  updated_at: ISODate
}

// Profile collection
{
  _id: ObjectId,
  user_id: String,
  first_name: String,
  last_name: String,
  phone: String,
  avatar_url: String,
  date_of_birth: ISODate,
  gender: String,
  preferences: {
    language: String,
    currency: String,
    notifications: {
      email: Boolean,
      sms: Boolean,
      push: Boolean
    }
  },
  created_at: ISODate,
  updated_at: ISODate
}

// Indexes
db.users.createIndex({ email: 1 }, { unique: true });
db.users.createIndex({ username: 1 }, { unique: true });
db.profiles.createIndex({ user_id: 1 }, { unique: true });
```

### Week 7: Shop Service

**Database:** PostgreSQL  
**Entities:** Shop_info

#### Architecture:
```
Shop Service
├── Shop Management
│   ├── Shop registration
│   ├── Shop profile
│   ├── Shop verification
│   └── Shop status management
├── Shop Queries
│   ├── Get shop details
│   ├── List shops
│   └── Search shops
└── Integration Points
    ├── User Service (shop owner)
    ├── Product Service (shop products)
    └── Order Service (shop orders)
```

### Week 8: Cart Service

**Database:** Redis (primary) + PostgreSQL (backup)  
**Entities:** Shopping_carts

#### Architecture:
```
Cart Service
├── Cart Operations
│   ├── Add to cart
│   ├── Remove from cart
│   ├── Update quantity
│   ├── Clear cart
│   └── Get cart
├── Cart Sync
│   ├── Redis (hot data)
│   ├── PostgreSQL (persistence)
│   └── Background sync
└── Integration Points
    ├── Product Service (product info, availability)
    ├── User Service (user cart)
    └── Order Service (checkout)
```

**Cart Architecture Pattern:**
```
┌──────────────────────────────────────────┐
│           Cart Request                   │
└─────────────┬────────────────────────────┘
              │
       ┌──────▼──────┐
       │   Service   │
       └──────┬──────┘
              │
       ┌──────▼──────┐
       │   Redis     │ ◄──── Write-Through
       │   (Cache)   │
       └──────┬──────┘
              │
              │ Background Sync (every 5min)
              │ or on cart abandon
              │
       ┌──────▼──────┐
       │  PostgreSQL │
       │  (Backup)   │
       └─────────────┘
```

### Week 8-9: Payment Service

**Database:** PostgreSQL  
**Entities:** Payment

#### Architecture:
```
Payment Service
├── Payment Processing
│   ├── Create payment intent
│   ├── Confirm payment
│   ├── Refund
│   └── Payment status
├── Payment Methods
│   ├── Credit/Debit card
│   ├── Digital wallet
│   ├── Bank transfer
│   └── COD (Cash on Delivery)
├── Integration Points
│   ├── Order Service (order payment)
│   ├── External Payment Gateway (Stripe, PayPal)
│   └── Notification Service (payment events)
└── Security
    ├── PCI compliance
    ├── Token management
    └── Fraud detection
```

### Week 9: Voucher, Address Services

**Voucher Service:**
- Database: PostgreSQL
- Discount management
- Voucher validation
- Usage tracking

**Address Service:**
- Database: PostgreSQL
- Address CRUD
- Default address management
- Address validation

### Week 10: Comment, Notification, Settings Services

**Comment Service:**
- Database: MongoDB
- Product reviews
- Rating system
- Comment moderation

**Notification Service:**
- Database: MongoDB
- Multi-channel notifications (Email, SMS, Push)
- Template management
- Notification preferences

**Settings Service:**
- Database: MongoDB
- System settings
- User preferences
- Feature flags

---

## Phase 3: Advanced Patterns (Weeks 11-14)

### Week 11-12: Event-Driven Architecture (NATS JetStream)

#### 11.1 Event System Design

```go
// pkg/events/events.go

package events

type Event struct {
    ID          string
    Type        string
    AggregateID string
    Version     int
    Timestamp   time.Time
    Payload     interface{}
    Metadata    map[string]string
}

type EventPublisher interface {
    Publish(ctx context.Context, event Event) error
    PublishBatch(ctx context.Context, events []Event) error
}

type EventSubscriber interface {
    Subscribe(ctx context.Context, eventType string, handler EventHandler) error
    SubscribeAll(ctx context.Context, handler EventHandler) error
}

type EventHandler func(ctx context.Context, event Event) error
```

#### 11.2 Event Types

```
Domain Events:
├── order.created
├── order.confirmed
├── order.shipped
├── order.delivered
├── order.cancelled
├── payment.initiated
├── payment.completed
├── payment.failed
├── inventory.reserved
├── inventory.released
├── user.registered
├── user.updated
├── product.created
├── product.updated
└── notification.sent
```

#### 11.3 Event Stream Configuration

```go
// pkg/events/jetstream.go

type StreamConfig struct {
    Name        string
    Subjects    []string
    Retention   RetentionPolicy
    MaxAge      time.Duration
    MaxMessages int64
    Storage     StorageType
}

// Order events stream
var OrderEventsStream = StreamConfig{
    Name:        "orders",
    Subjects:    []string{"order.*"},
    Retention:   RetentionPolicyWorkQueue,
    MaxAge:      7 * 24 * time.Hour,
    MaxMessages: 1_000_000,
    Storage:     StorageTypeFile,
}
```

### Week 13: SAGA Pattern (Distributed Transactions)

#### 13.1 SAGA Orchestrator

```go
// pkg/saga/orchestrator.go

package saga

type SagaDefinition struct {
    Name  string
    Steps []Step
}

type Step struct {
    Name         string
    Action       func(ctx context.Context, data interface{}) error
    Compensation func(ctx context.Context, data interface{}) error
}

type Orchestrator struct {
    eventPublisher EventPublisher
    stateStore     StateStore
}

func (o *Orchestrator) Execute(ctx context.Context, saga SagaDefinition, data interface{}) error {
    // Execute steps sequentially
    // On failure, execute compensations in reverse order
}
```

#### 13.2 Order Checkout SAGA Example

```
┌──────────────────────────────────────────────────────────┐
│              Order Checkout SAGA                          │
├──────────────────────────────────────────────────────────┤
│                                                           │
│  1. Reserve Inventory      ←→ Compensate: Release        │
│  2. Create Order           ←→ Compensate: Cancel Order   │
│  3. Process Payment        ←→ Compensate: Refund         │
│  4. Send Notification      ←→ Compensate: N/A            │
│  5. Update User Points     ←→ Compensate: Revert Points  │
│                                                           │
└──────────────────────────────────────────────────────────┘

Success Flow:
Step 1 → Step 2 → Step 3 → Step 4 → Step 5 → Complete

Failure at Step 3:
Step 1 → Step 2 → Step 3 (FAIL) → Compensate 2 → Compensate 1
```

### Week 14: CQRS Pattern

#### 14.1 Command/Query Separation

```go
// pkg/cqrs/command.go

type Command interface {
    CommandName() string
}

type CommandHandler interface {
    Handle(ctx context.Context, cmd Command) error
}

type CommandBus interface {
    Dispatch(ctx context.Context, cmd Command) error
    Register(cmdName string, handler CommandHandler)
}

// pkg/cqrs/query.go

type Query interface {
    QueryName() string
}

type QueryHandler interface {
    Handle(ctx context.Context, query Query) (interface{}, error)
}

type QueryBus interface {
    Execute(ctx context.Context, query Query) (interface{}, error)
    Register(queryName string, handler QueryHandler)
}
```

#### 14.2 Read Model Synchronization

```
┌───────────────────────────────────────────────────────┐
│                CQRS Architecture                       │
├───────────────────────────────────────────────────────┤
│                                                        │
│  Write Side:                                          │
│  Command → Handler → Domain Model → Events → DB      │
│                                            │           │
│                                            │           │
│  Read Side:                                │           │
│  Events ────────────────────────────────────→         │
│      │                                      │           │
│      └→ Projection Builder → Read Model → Cache      │
│                                                        │
│  Query → Handler → Read Model/Cache → Response       │
│                                                        │
└───────────────────────────────────────────────────────┘
```

---

## Phase 4: Production Readiness (Weeks 15-18)

### Week 15-16: Kubernetes Deployment

#### 15.1 Kubernetes Architecture

```yaml
# infra/k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: ecommerce
```

**Resource Structure:**
```
infra/k8s/
├── base/
│   ├── namespace.yaml
│   ├── configmap.yaml
│   ├── secrets.yaml
│   └── ingress.yaml
├── services/
│   ├── api-gateway/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   ├── hpa.yaml
│   │   └── configmap.yaml
│   ├── auth/
│   ├── order/
│   ├── product/
│   └── ... (other services)
├── databases/
│   ├── postgres/
│   │   ├── statefulset.yaml
│   │   ├── service.yaml
│   │   ├── pvc.yaml
│   │   └── configmap.yaml
│   ├── mongodb/
│   └── redis/
├── monitoring/ ✅ (in infra/docker-compose.yml)
│   ├── prometheus/
│   ├── grafana/
│   ├── tempo/      # Replaced Jaeger
│   ├── loki/
│   └── alloy/
└── nats/
    ├── statefulset.yaml
    └── service.yaml
```

#### 15.2 Service Deployment Template

```yaml
# infra/k8s/services/product/deployment.yaml

apiVersion: apps/v1
kind: Deployment
metadata:
  name: product-service
  namespace: ecommerce
  labels:
    app: product-service
    version: v1
spec:
  replicas: 3
  selector:
    matchLabels:
      app: product-service
  template:
    metadata:
      labels:
        app: product-service
        version: v1
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
        prometheus.io/path: "/metrics"
    spec:
      containers:
      - name: product-service
        image: ecommerce/product-service:latest
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: metrics
        env:
        - name: SERVICE_NAME
          value: "product-service"
        - name: NATS_URL
          valueFrom:
            configMapKeyRef:
              name: nats-config
              key: url
        - name: DB_HOST
          valueFrom:
            configMapKeyRef:
              name: postgres-config
              key: host
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-secret
              key: password
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: product-service
  namespace: ecommerce
spec:
  selector:
    app: product-service
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: ClusterIP
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: product-service-hpa
  namespace: ecommerce
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: product-service
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

#### 15.3 Helm Charts

```
infra/helm/
├── ecommerce/
│   ├── Chart.yaml
│   ├── values.yaml
│   ├── values-dev.yaml
│   ├── values-staging.yaml
│   ├── values-prod.yaml
│   └── templates/
│       ├── api-gateway/
│       ├── services/
│       ├── databases/
│       └── monitoring/
```

### Week 17: CI/CD Pipeline

#### 17.1 GitHub Actions Workflow

```yaml
# .github/workflows/ci-cd.yaml

name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

env:
  REGISTRY: ghcr.io
  IMAGE_PREFIX: ${{ github.repository }}

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      
      - name: Run tests
        run: |
          go test -v -race -coverprofile=coverage.out ./...
          go tool cover -html=coverage.out -o coverage.html
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out

  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest

  build:
    name: Build Services
    needs: [test, lint]
    runs-on: ubuntu-latest
    strategy:
      matrix:
        service: [api-gateway, auth, order, product, user, cart, payment]
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2
      
      - name: Login to Registry
        uses: docker/login-action@v2
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      
      - name: Build and push
        uses: docker/build-push-action@v4
        with:
          context: .
          file: ./apps/${{ matrix.service }}/Dockerfile
          push: ${{ github.event_name != 'pull_request' }}
          tags: |
            ${{ env.REGISTRY }}/${{ env.IMAGE_PREFIX }}/${{ matrix.service }}:${{ github.sha }}
            ${{ env.REGISTRY }}/${{ env.IMAGE_PREFIX }}/${{ matrix.service }}:latest
          cache-from: type=gha
          cache-to: type=gha,mode=max

  deploy-staging:
    name: Deploy to Staging
    needs: [build]
    if: github.ref == 'refs/heads/develop'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up kubectl
        uses: azure/setup-kubectl@v3
      
      - name: Deploy to staging
        run: |
          kubectl config use-context staging
          helm upgrade --install ecommerce ./infra/helm/ecommerce \
            --namespace ecommerce-staging \
            --values ./infra/helm/ecommerce/values-staging.yaml \
            --set image.tag=${{ github.sha }}

  deploy-production:
    name: Deploy to Production
    needs: [build]
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/checkout@v3
      
      - name: Deploy to production
        run: |
          kubectl config use-context production
          helm upgrade --install ecommerce ./infra/helm/ecommerce \
            --namespace ecommerce-prod \
            --values ./infra/helm/ecommerce/values-prod.yaml \
            --set image.tag=${{ github.sha }}
```

### Week 18: Security Hardening

#### 18.1 Security Checklist

```
Security Measures:
├── Authentication & Authorization
│   ├── [x] OAuth2 with Zitadel
│   ├── [ ] Service-to-service mTLS
│   ├── [ ] API key management
│   └── [ ] RBAC implementation
├── Data Protection
│   ├── [ ] Encryption at rest (database)
│   ├── [ ] Encryption in transit (TLS)
│   ├── [ ] PII data masking
│   └── [ ] Secrets management (Vault)
├── Network Security
│   ├── [ ] Network policies (K8s)
│   ├── [ ] WAF integration
│   ├── [ ] DDoS protection
│   └── [ ] API rate limiting per user
├── Application Security
│   ├── [ ] Input validation
│   ├── [ ] SQL injection prevention
│   ├── [ ] XSS prevention
│   ├── [ ] CSRF protection
│   └── [ ] Dependency scanning
└── Monitoring & Audit
    ├── [ ] Security event logging
    ├── [ ] Anomaly detection
    ├── [ ] Audit trail
    └── [ ] Compliance reporting
```

#### 18.2 Service-to-Service Authentication

```go
// pkg/auth/service_auth.go

package auth

import (
    "crypto/tls"
    "crypto/x509"
)

type ServiceAuthConfig struct {
    CertFile   string
    KeyFile    string
    CAFile     string
    ServerName string
}

// CreateTLSConfig creates TLS config for mTLS
func CreateTLSConfig(config ServiceAuthConfig) (*tls.Config, error) {
    // Load client cert
    cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
    if err != nil {
        return nil, err
    }
    
    // Load CA cert
    caCert, err := os.ReadFile(config.CAFile)
    if err != nil {
        return nil, err
    }
    
    caCertPool := x509.NewCertPool()
    caCertPool.AppendCertsFromPEM(caCert)
    
    return &tls.Config{
        Certificates: []tls.Certificate{cert},
        RootCAs:      caCertPool,
        ServerName:   config.ServerName,
    }, nil
}
```

---

## Phase 5: Optimization & Scaling (Weeks 19-20)

### Week 19: Performance Optimization

#### 19.1 Database Optimization

**PostgreSQL:**
```sql
-- Index optimization
CREATE INDEX CONCURRENTLY idx_products_search 
ON products USING GIN (to_tsvector('english', name || ' ' || description));

-- Partitioning large tables
CREATE TABLE orders_2024 PARTITION OF orders
FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');

-- Connection pooling
-- Configure pgbouncer for connection pooling
```

**MongoDB:**
```javascript
// Compound indexes
db.users.createIndex({ email: 1, status: 1 });

// Covered queries
db.products.createIndex({ category_id: 1, price: 1, name: 1 });

// Read preference
db.products.find().readPref("secondaryPreferred");
```

**Redis:**
```
# Cache strategies
- Cache hot products (top 1000 viewed)
- Cache user sessions
- Cache cart data
- Cache API responses (TTL: 5min)

# Eviction policy
maxmemory-policy: allkeys-lru
```

#### 19.2 Service Optimization

```go
// Response caching
func CacheMiddleware(cache cache.Cache, ttl time.Duration) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Check if request is cacheable
            if r.Method != "GET" {
                next.ServeHTTP(w, r)
                return
            }
            
            // Try cache
            cacheKey := generateCacheKey(r)
            if cached, ok := cache.Get(cacheKey); ok {
                w.Write(cached)
                return
            }
            
            // Capture response
            rec := httptest.NewRecorder()
            next.ServeHTTP(rec, r)
            
            // Cache response
            cache.Set(cacheKey, rec.Body.Bytes(), ttl)
            
            // Write response
            w.Write(rec.Body.Bytes())
        })
    }
}
```

### Week 20: Load Testing & Tuning

#### 20.1 Load Testing with K6

```javascript
// tests/load/checkout-flow.js

import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '2m', target: 100 },  // Ramp up
    { duration: '5m', target: 100 },  // Stay at 100 users
    { duration: '2m', target: 200 },  // Ramp up to 200
    { duration: '5m', target: 200 },  // Stay at 200
    { duration: '2m', target: 0 },    // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
    http_req_failed: ['rate<0.01'],
  },
};

export default function () {
  // 1. Browse products
  let productsRes = http.get('http://api-gateway/products');
  check(productsRes, {
    'products loaded': (r) => r.status === 200,
  });
  
  // 2. Add to cart
  let addToCartRes = http.post('http://api-gateway/cart/add', {
    product_id: 'test-product',
    quantity: 1,
  });
  check(addToCartRes, {
    'added to cart': (r) => r.status === 200,
  });
  
  // 3. Checkout
  let checkoutRes = http.post('http://api-gateway/orders/checkout');
  check(checkoutRes, {
    'checkout successful': (r) => r.status === 200,
  });
  
  sleep(1);
}
```

#### 20.2 Auto-scaling Configuration

```yaml
# HPA for services
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: product-service-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: product-service
  minReplicas: 2
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Pods
    pods:
      metric:
        name: http_requests_per_second
      target:
        type: AverageValue
        averageValue: "1000"
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
      - type: Percent
        value: 100
        periodSeconds: 15
      - type: Pods
        value: 4
        periodSeconds: 15
      selectPolicy: Max
```

---

## 9. Deployment Architecture

### 9.1 Multi-Environment Setup

```
Environments:
├── Development (Local)
│   ├── Docker Compose
│   └── Mock services
├── Staging (Cloud)
│   ├── Kubernetes cluster
│   ├── Real databases (smaller instances)
│   └── External service sandboxes
└── Production (Cloud)
    ├── Kubernetes cluster (multi-region)
    ├── High-availability databases
    ├── CDN integration
    └── Monitoring & alerting
```

### 9.2 Infrastructure as Code (Terraform)

```hcl
# infra/terraform/main.tf

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

# EKS Cluster
module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 19.0"

  cluster_name    = "ecommerce-cluster"
  cluster_version = "1.27"

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  eks_managed_node_groups = {
    general = {
      min_size     = 2
      max_size     = 10
      desired_size = 3

      instance_types = ["t3.large"]
      capacity_type  = "ON_DEMAND"
    }
  }
}

# RDS PostgreSQL
module "rds" {
  source  = "terraform-aws-modules/rds/aws"
  version = "~> 5.0"

  identifier = "ecommerce-postgres"

  engine               = "postgres"
  engine_version       = "15.3"
  family               = "postgres15"
  major_engine_version = "15"
  instance_class       = "db.r5.large"

  allocated_storage     = 100
  max_allocated_storage = 500

  multi_az               = true
  db_subnet_group_name   = module.vpc.database_subnet_group
  vpc_security_group_ids = [module.security_group.id]

  backup_retention_period = 7
  backup_window           = "03:00-06:00"
  maintenance_window      = "Mon:00:00-Mon:03:00"
}

# ElastiCache Redis
module "redis" {
  source  = "terraform-aws-modules/elasticache/aws"
  version = "~> 1.0"

  cluster_id           = "ecommerce-redis"
  engine               = "redis"
  engine_version       = "7.0"
  node_type            = "cache.r5.large"
  num_cache_nodes      = 2
  parameter_group_name = "default.redis7"
  
  subnet_group_name    = module.vpc.elasticache_subnet_group
  security_group_ids   = [module.security_group.id]
}
```

### 9.3 Disaster Recovery

```
DR Strategy:
├── Backup Schedule
│   ├── Database: Daily full + hourly incremental
│   ├── Files: Daily to S3
│   └── Configuration: Git versioning
├── Recovery Objectives
│   ├── RPO (Recovery Point Objective): 1 hour
│   └── RTO (Recovery Time Objective): 4 hours
├── Multi-Region Setup
│   ├── Primary: us-east-1
│   ├── Secondary: eu-west-1
│   └── Database replication
└── Disaster Recovery Procedures
    ├── Failover automation
    ├── Health checks
    └── Rollback procedures
```

---

## 10. Success Criteria

### 10.1 Technical Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Service Availability | 99.9% | Uptime monitoring |
| API Response Time (P95) | <200ms | APM |
| API Response Time (P99) | <500ms | APM |
| Error Rate | <0.1% | Logs/Metrics |
| Database Query Time (P95) | <50ms | DB monitoring |
| Test Coverage | >80% | Code coverage |
| Build Time | <10min | CI/CD pipeline |
| Deployment Time | <15min | CI/CD pipeline |

### 10.2 Architecture Quality Gates

**Before Phase Completion:**
- [ ] All services deployed and operational
- [ ] Health checks passing
- [ ] Monitoring dashboards configured
- [ ] Documentation updated
- [ ] Security scan passed
- [ ] Load test passed
- [ ] Disaster recovery tested

**Production Ready Checklist:**
- [ ] All 11 microservices implemented (3/11 done)
- [x] Circuit breakers in place ✅
- [x] Distributed tracing working ✅ (OpenTelemetry + Tempo)
- [x] Metrics collection active ✅ (Prometheus + Alloy)
- [x] Logging centralized ✅ (Zap + Loki + Alloy)
- [ ] Kubernetes deployment successful
- [ ] CI/CD pipeline operational
- [ ] Security audit passed
- [ ] Load testing completed
- [x] Observability documentation complete ✅
- [ ] API documentation complete
- [ ] Runbooks created
- [ ] Team training done

---

## Appendix

### A. Technology Stack Summary

```
Language: Go 1.24
Message Broker: NATS JetStream
Databases:
  - PostgreSQL 15+ (Orders, Products, Payments, etc.)
  - MongoDB 7+ (Users, Comments, Notifications, Settings)
  - Redis 7+ (Cache, Sessions, Cart)
API: gRPC (internal), REST (external)
Service Discovery: NATS
Orchestration: Kubernetes (planned)
CI/CD: GitHub Actions (planned)
Auth: Zitadel OAuth2

Observability Stack: ✅ IMPLEMENTED
  - Tracing: OpenTelemetry SDK → Grafana Alloy → Grafana Tempo
  - Metrics: Prometheus (scrape) + Alloy (remote write)
  - Logging: Zap → File/Console → Grafana Alloy → Grafana Loki
  - Visualization: Grafana (:3001)
  - OTLP Endpoint: Grafana Alloy (:4317 gRPC, :4318 HTTP)

Resilience Patterns: ✅ IMPLEMENTED
  - Circuit Breaker: sony/gobreaker v2
  - Rate Limiting: Redis-based sliding window
```

### B. Team Structure

```
Recommended Team:
├── Backend Engineers (3-4)
│   ├── Microservices development
│   └── API implementation
├── DevOps Engineer (1-2)
│   ├── Infrastructure setup
│   ├── CI/CD pipeline
│   └── Monitoring
├── QA Engineer (1-2)
│   ├── Test automation
│   └── Load testing
└── Tech Lead (1)
    ├── Architecture decisions
    └── Code reviews
```

### C. Estimated Effort

| Phase | Duration | Team Size | Total Person-Weeks |
|-------|----------|-----------|-------------------|
| Phase 1: Infrastructure | 4 weeks | 3 | 12 |
| Phase 2: Services | 6 weeks | 4 | 24 |
| Phase 3: Advanced Patterns | 4 weeks | 3 | 12 |
| Phase 4: Production | 4 weeks | 4 | 16 |
| Phase 5: Optimization | 2 weeks | 3 | 6 |
| **Total** | **20 weeks** | **3-4** | **70** |

### D. Risk Management

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Service complexity | High | Medium | Start with template, incremental dev |
| Integration issues | High | High | Early integration testing |
| Performance bottlenecks | Medium | Medium | Load testing, profiling |
| Security vulnerabilities | High | Low | Security audits, penetration testing |
| Team availability | High | Medium | Documentation, knowledge sharing |
| Third-party dependencies | Medium | Low | Vendor evaluation, fallback options |

---

## Conclusion

This architectural plan provides a comprehensive roadmap to complete the e-commerce microservice system. The phased approach ensures:

1. **Foundation First**: Core infrastructure before services
2. **Incremental Delivery**: Services delivered in priority order
3. **Quality Assurance**: Testing and monitoring at every phase
4. **Production Ready**: Full DevOps and security implementation
5. **Scalability**: Architecture designed for growth

**Next Steps:**
1. Review and approve this plan
2. Set up project tracking (Jira/Linear)
3. Assign team members to phases
4. Begin Phase 1 implementation
5. Schedule weekly architecture reviews

**Timeline:** 20 weeks with 3-4 engineers

---

**Document Version:** 1.1  
**Last Updated:** December 17, 2025  
**Author:** AI Assistant  
**Status:** Updated - Observability & Resilience Patterns Completed

---

## Changelog

### v1.1 (December 17, 2025)
- ✅ Updated status of Phase 1 infrastructure components
- ✅ Marked Observability stack as COMPLETED (Tracing, Metrics, Logging)
- ✅ Marked Circuit Breaker pattern as COMPLETED
- ✅ Updated architecture diagrams to reflect actual implementation
- ✅ Added reference to `/docs/observability/OBSERVABILITY.md` for detailed docs
- ✅ Updated Technology Stack with actual tools used

### v1.0 (November 12, 2025)
- Initial architectural plan


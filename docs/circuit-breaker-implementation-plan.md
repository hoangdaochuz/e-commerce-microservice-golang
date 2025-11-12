# Circuit Breaker Pattern Implementation Plan

## Executive Summary

This document outlines a comprehensive plan to implement the Circuit Breaker pattern across the e-commerce microservice architecture. The circuit breaker pattern will enhance system resilience by preventing cascading failures, reducing resource waste, and providing graceful degradation when services or dependencies become unavailable.

## Table of Contents

1. [Current Architecture Analysis](#current-architecture-analysis)
2. [Circuit Breaker Pattern Overview](#circuit-breaker-pattern-overview)
3. [Implementation Strategy](#implementation-strategy)
4. [Integration Points](#integration-points)
5. [Implementation Phases](#implementation-phases)
6. [Technical Specifications](#technical-specifications)
7. [Configuration Management](#configuration-management)
8. [Monitoring and Observability](#monitoring-and-observability)
9. [Testing Strategy](#testing-strategy)
10. [Rollout Plan](#rollout-plan)
11. [Success Metrics](#success-metrics)

---

## 1. Current Architecture Analysis

### 1.1 System Overview

The project consists of:
- **11 microservices**: auth, order, user, product, shop, cart, voucher, comment, notification, payment, address
- **API Gateway**: Entry point for all HTTP requests
- **Message Broker**: NATS for inter-service communication
- **Databases**: PostgreSQL (orders, products, payments), MongoDB (users, settings, comments, notifications), Redis (cache, sessions)
- **External Services**: Zitadel for authentication and authorization

### 1.2 Communication Patterns

```
Client → API Gateway → NATS → Microservices → Databases/External APIs
```

**Key observations:**
- API Gateway uses NATS request-response pattern with 30s timeout
- Services communicate asynchronously via NATS
- Direct database connections (PostgreSQL via sqlx, MongoDB native driver)
- Redis for caching and session management
- External HTTP calls to Zitadel OAuth endpoints

### 1.3 Current Resilience Features

✅ **Existing:**
- Rate limiting (50 requests per minute at API Gateway)
- Request timeouts (30s)
- Connection pooling for databases

❌ **Missing:**
- Circuit breakers for service calls
- Circuit breakers for database connections
- Circuit breakers for external API calls
- Fallback mechanisms
- Automatic recovery detection
- Metrics and monitoring for failures

---

## 2. Circuit Breaker Pattern Overview

### 2.1 What is Circuit Breaker?

The Circuit Breaker pattern prevents an application from repeatedly trying to execute an operation that's likely to fail, allowing it to continue without waiting for the fault to be fixed or wasting resources.

### 2.2 Circuit States

```
┌─────────┐                    ┌──────────┐                   ┌────────┐
│ CLOSED  │───[failures >= N]─→│   OPEN   │───[timeout]──────→│  HALF  │
│         │                    │          │                   │  OPEN  │
└─────────┘                    └──────────┘                   └────────┘
     ↑                              │                              │
     └──────────────────────────────┴──────[success]──────────────┘
                                    └──────[failure]───────→
```

1. **CLOSED** (Normal): Requests flow through normally. Failures are counted.
2. **OPEN** (Failing): Requests fail immediately without attempting the operation.
3. **HALF-OPEN** (Testing): Limited requests are allowed to test if the underlying issue is resolved.

### 2.3 Benefits

- **Fail Fast**: Immediately return errors instead of waiting for timeouts
- **Resource Conservation**: Don't waste resources on operations likely to fail
- **Cascading Failure Prevention**: Stop failures from propagating across services
- **Automatic Recovery**: Detect when services recover and resume normal operations
- **Better User Experience**: Faster error responses, fallback options

---

## 3. Implementation Strategy

### 3.1 Library Selection

**Recommended: `sony/gobreaker`**

**Rationale:**
- Production-ready and battle-tested
- Simple, clean API
- Configurable states and thresholds
- Thread-safe
- No external dependencies
- Active maintenance

**Alternative: `go-resiliency/circuitbreaker`**

### 3.2 Architecture Approach

```
┌─────────────────────────────────────────────────────────────┐
│                      Application Layer                      │
├─────────────────────────────────────────────────────────────┤
│                  Circuit Breaker Middleware                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ NATS Breaker │  │  DB Breaker  │  │ HTTP Breaker │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
├─────────────────────────────────────────────────────────────┤
│           Transport Layer (NATS, DB, HTTP)                  │
└─────────────────────────────────────────────────────────────┘
```

**Key Principles:**
1. **Wrapper Pattern**: Wrap existing clients/connections with circuit breaker logic
2. **Transparent Integration**: Minimal changes to existing service code
3. **Centralized Configuration**: Single source of truth for circuit breaker settings
4. **Per-Dependency Breakers**: Separate circuit breakers for each dependency type

---

## 4. Integration Points

### 4.1 Priority Matrix

| Integration Point | Priority | Impact | Complexity |
|------------------|----------|--------|------------|
| NATS Service Calls | **HIGH** | HIGH | Medium |
| PostgreSQL Connections | **HIGH** | HIGH | Medium |
| MongoDB Connections | **HIGH** | HIGH | Medium |
| Redis Connections | **MEDIUM** | MEDIUM | Low |
| Zitadel OAuth | **HIGH** | HIGH | Low |
| HTTP External APIs | **MEDIUM** | MEDIUM | Low |

### 4.2 NATS Communication (Priority 1)

**Current Implementation:**
```go
// api_gateway/main.go:90
msgResponse, err := gw.natsConn.Request(natsReq.Subject, natsReqByte, gw.timeout)
```

**Issues:**
- No circuit breaker protection
- Failed services cause timeout delays (30s)
- No automatic failure detection

**Implementation Location:**
- `pkg/custom-nats/client.go` - Create wrapped NATS client
- `api_gateway/main.go` - Use circuit breaker wrapper

### 4.3 Database Connections (Priority 1)

**PostgreSQL (sqlx):**
- Location: `pkg/repo/postgres_sqlx/postgres_client.go`
- Operations: FindOne, FindAll, Create, Update, Delete
- Risk: Connection pool exhaustion, slow queries, database failures

**MongoDB:**
- Location: `pkg/repo/mongo/mongo_client.go`
- Operations: FindOne, FindAll, InsertOne, UpdateOne, DeleteMany
- Risk: Replica set failures, network issues, slow queries

**Implementation Approach:**
- Wrap repository methods with circuit breaker
- Per-database circuit breaker instances
- Separate breakers for read vs write operations (optional)

### 4.4 Redis Connections (Priority 2)

**Location:** `pkg/redis/main.go`, `pkg/cache/redis_cache.go`

**Use Cases:**
- Session storage
- Rate limiting
- General caching

**Strategy:**
- Circuit breaker for Redis operations
- Fallback to in-memory cache for non-critical operations
- Fail-fast for critical operations (sessions)

### 4.5 External HTTP Calls (Priority 1)

**Zitadel Authentication:**
- Location: `pkg/zitadel/authentication/oidc.go`
- Endpoints: Token, UserInfo, Authorization, EndSession
- Risk: OAuth provider downtime, network issues, rate limiting

**Implementation:**
- Wrap HTTP client with circuit breaker
- Implement fallback strategies (cached user info, cached token validation)

---

## 5. Implementation Phases

### Phase 1: Foundation (Week 1)

**Goals:**
- Set up circuit breaker infrastructure
- Create reusable circuit breaker package
- Implement configuration management

**Deliverables:**

1. **Create Circuit Breaker Package**
   ```
   pkg/circuitbreaker/
   ├── breaker.go           # Core circuit breaker wrapper
   ├── config.go            # Configuration structures
   ├── registry.go          # Breaker instance registry
   ├── middleware.go        # Middleware helpers
   ├── metrics.go           # Metrics collection
   └── breaker_test.go      # Unit tests
   ```

2. **Configuration Structure**
   - Add circuit breaker configs to `configs/config.yaml`
   - Support per-service and per-dependency configuration
   - Implement config validation

3. **Metrics Foundation**
   - Define metrics interface
   - Implement basic logging
   - Prepare for future Prometheus integration

### Phase 2: NATS Circuit Breaker (Week 2)

**Goals:**
- Protect inter-service communication
- Implement service discovery with circuit breaker awareness

**Implementation Steps:**

1. **Create NATS Wrapper**
   ```
   pkg/custom-nats/
   ├── breaker_client.go    # Circuit breaker wrapped client
   └── breaker_test.go      # Integration tests
   ```

2. **Update API Gateway**
   - Replace direct NATS calls with circuit breaker wrapper
   - Add fallback responses
   - Implement error handling

3. **Per-Service Breakers**
   - Separate circuit breaker for each microservice
   - Independent failure tracking
   - Service-specific configuration

### Phase 3: Database Circuit Breakers (Week 3)

**Goals:**
- Protect database connections
- Prevent connection pool exhaustion
- Graceful degradation for read operations

**Implementation Steps:**

1. **PostgreSQL Wrapper**
   ```
   pkg/repo/postgres_sqlx/
   ├── breaker_client.go    # Circuit breaker wrapper
   └── fallback.go          # Fallback strategies
   ```

2. **MongoDB Wrapper**
   ```
   pkg/repo/mongo/
   ├── breaker_client.go    # Circuit breaker wrapper
   └── fallback.go          # Fallback strategies
   ```

3. **Repository Updates**
   - Update each service repository to use wrapped clients
   - Implement read/write split if needed
   - Add caching for read operations (fallback)

### Phase 4: Redis and Cache Circuit Breakers (Week 4)

**Goals:**
- Protect Redis operations
- Implement fallback caching strategy

**Implementation Steps:**

1. **Redis Wrapper**
   ```
   pkg/cache/
   ├── breaker_cache.go     # Circuit breaker cache wrapper
   └── fallback_cache.go    # In-memory fallback
   ```

2. **Cache Strategy**
   - Primary: Redis cache
   - Fallback: In-memory cache (go-cache)
   - TTL management for fallback

3. **Session Management**
   - Critical operations fail fast
   - Non-critical operations use fallback

### Phase 5: External API Circuit Breakers (Week 5)

**Goals:**
- Protect external HTTP calls
- Implement OAuth-specific fallbacks

**Implementation Steps:**

1. **HTTP Client Wrapper**
   ```
   pkg/httpclient/
   ├── breaker_client.go    # Circuit breaker HTTP client
   ├── config.go            # HTTP client configuration
   └── retry.go             # Retry logic
   ```

2. **Zitadel Integration**
   - Wrap Zitadel OAuth calls
   - Cache user info responses
   - Implement token validation fallback

3. **Rate Limiting Integration**
   - Coordinate with circuit breaker
   - Exponential backoff
   - Respect rate limit headers

### Phase 6: Monitoring and Testing (Week 6)

**Goals:**
- Comprehensive monitoring
- Chaos testing
- Production readiness

**Implementation Steps:**

1. **Metrics and Monitoring**
   - Prometheus metrics export
   - Grafana dashboards
   - Alerting rules

2. **Testing**
   - Unit tests for all circuit breakers
   - Integration tests with failure injection
   - Load testing with circuit breaker
   - Chaos engineering tests

3. **Documentation**
   - API documentation
   - Runbooks for operations
   - Configuration guide

---

## 6. Technical Specifications

### 6.1 Circuit Breaker Package Structure

```go
// pkg/circuitbreaker/breaker.go

package circuitbreaker

import (
    "context"
    "time"
    "github.com/sony/gobreaker"
)

// Config defines circuit breaker configuration
type Config struct {
    Name          string
    MaxRequests   uint32        // Max requests allowed in half-open state
    Interval      time.Duration // Window for failure rate calculation
    Timeout       time.Duration // Time to wait before entering half-open
    ReadyToTrip   func(counts gobreaker.Counts) bool
    OnStateChange func(name string, from gobreaker.State, to gobreaker.State)
}

// Breaker wraps gobreaker with additional functionality
type Breaker struct {
    cb     *gobreaker.CircuitBreaker
    config Config
    metrics MetricsCollector
}

// NewBreaker creates a new circuit breaker instance
func NewBreaker(config Config) *Breaker

// Execute runs the given function with circuit breaker protection
func (b *Breaker) Execute(ctx context.Context, fn func() (interface{}, error)) (interface{}, error)

// ExecuteWithFallback runs function with fallback on failure
func (b *Breaker) ExecuteWithFallback(
    ctx context.Context,
    fn func() (interface{}, error),
    fallback func() (interface{}, error),
) (interface{}, error)

// State returns current circuit breaker state
func (b *Breaker) State() gobreaker.State

// Metrics returns collected metrics
func (b *Breaker) Metrics() Metrics
```

### 6.2 Configuration Schema

```yaml
# configs/config.yaml

circuit_breaker:
  # Global defaults
  defaults:
    max_requests: 3              # Max requests in half-open state
    interval: 60s                # Rolling window for failure counting
    timeout: 30s                 # Duration before transitioning to half-open
    failure_threshold: 5         # Number of failures to open circuit
    failure_rate_threshold: 0.6  # Failure rate (60%) to open circuit
    min_requests: 10             # Minimum requests before calculating rate

  # NATS service calls
  nats:
    enabled: true
    services:
      auth:
        max_requests: 5
        interval: 30s
        timeout: 20s
        failure_threshold: 3
      order:
        max_requests: 5
        interval: 30s
        timeout: 20s
        failure_threshold: 3
      product:
        max_requests: 5
        interval: 30s
        timeout: 20s
        failure_threshold: 3
      # ... other services

  # Database connections
  databases:
    postgres:
      enabled: true
      max_requests: 3
      interval: 60s
      timeout: 30s
      failure_threshold: 5
      separate_read_write: false  # Use separate breakers for read/write
    
    mongodb:
      enabled: true
      max_requests: 3
      interval: 60s
      timeout: 30s
      failure_threshold: 5
    
    redis:
      enabled: true
      max_requests: 5
      interval: 30s
      timeout: 10s
      failure_threshold: 3
      fallback_to_memory: true

  # External APIs
  external_apis:
    zitadel:
      enabled: true
      max_requests: 3
      interval: 60s
      timeout: 45s
      failure_threshold: 3
      cache_fallback: true
      cache_ttl: 300s

  # Monitoring
  monitoring:
    metrics_enabled: true
    log_state_changes: true
    alert_on_open: true
```

### 6.3 NATS Circuit Breaker Implementation

```go
// pkg/custom-nats/breaker_client.go

package custom_nats

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/nats-io/nats.go"
    "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/circuitbreaker"
)

// BreakerClient wraps NATS connection with circuit breaker
type BreakerClient struct {
    conn     *nats.Conn
    breakers map[string]*circuitbreaker.Breaker
    timeout  time.Duration
}

// NewBreakerClient creates a new circuit breaker protected NATS client
func NewBreakerClient(conn *nats.Conn, configs map[string]circuitbreaker.Config, timeout time.Duration) *BreakerClient {
    breakers := make(map[string]*circuitbreaker.Breaker)
    for service, config := range configs {
        breakers[service] = circuitbreaker.NewBreaker(config)
    }
    
    return &BreakerClient{
        conn:     conn,
        breakers: breakers,
        timeout:  timeout,
    }
}

// Request sends a NATS request with circuit breaker protection
func (bc *BreakerClient) Request(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
    // Extract service name from subject (e.g., "auth.Login" -> "auth")
    service := extractServiceFromSubject(subject)
    
    breaker, ok := bc.breakers[service]
    if !ok {
        // No circuit breaker configured, use default
        return bc.conn.Request(subject, data, timeout)
    }
    
    // Execute with circuit breaker
    result, err := breaker.Execute(context.Background(), func() (interface{}, error) {
        return bc.conn.Request(subject, data, timeout)
    })
    
    if err != nil {
        return nil, fmt.Errorf("circuit breaker error for service %s: %w", service, err)
    }
    
    return result.(*nats.Msg), nil
}

// RequestWithFallback sends request with fallback response
func (bc *BreakerClient) RequestWithFallback(
    subject string,
    data []byte,
    timeout time.Duration,
    fallback func() (*nats.Msg, error),
) (*nats.Msg, error) {
    service := extractServiceFromSubject(subject)
    breaker := bc.breakers[service]
    
    result, err := breaker.ExecuteWithFallback(
        context.Background(),
        func() (interface{}, error) {
            return bc.conn.Request(subject, data, timeout)
        },
        func() (interface{}, error) {
            return fallback()
        },
    )
    
    if err != nil {
        return nil, err
    }
    
    return result.(*nats.Msg), nil
}

// GetBreakerState returns current state for a service
func (bc *BreakerClient) GetBreakerState(service string) string {
    if breaker, ok := bc.breakers[service]; ok {
        return breaker.State().String()
    }
    return "UNKNOWN"
}

func extractServiceFromSubject(subject string) string {
    // Subject format: "service.Method" (e.g., "auth.Login")
    parts := strings.Split(subject, ".")
    if len(parts) > 0 {
        return parts[0]
    }
    return "unknown"
}
```

### 6.4 Database Circuit Breaker Implementation

```go
// pkg/repo/breaker.go

package repo

import (
    "context"
    "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/circuitbreaker"
)

// BreakerDBClient wraps IDBClient with circuit breaker
type BreakerDBClient struct {
    client  IDBClient
    breaker *circuitbreaker.Breaker
    cache   Cache // Optional cache for fallback
}

// NewBreakerDBClient creates a circuit breaker protected DB client
func NewBreakerDBClient(
    client IDBClient,
    breaker *circuitbreaker.Breaker,
    cache Cache,
) IDBClient {
    return &BreakerDBClient{
        client:  client,
        breaker: breaker,
        cache:   cache,
    }
}

// FindOne executes FindOne with circuit breaker protection
func (b *BreakerDBClient) FindOne(ctx context.Context, out BaseModel, query interface{}, others ...interface{}) error {
    result, err := b.breaker.ExecuteWithFallback(
        ctx,
        func() (interface{}, error) {
            err := b.client.FindOne(ctx, out, query, others...)
            return nil, err
        },
        func() (interface{}, error) {
            // Try cache fallback for reads
            if b.cache != nil {
                cacheKey := generateCacheKey(query, others...)
                if cached, ok := b.cache.Get(cacheKey); ok {
                    // Populate out with cached data
                    return nil, populateFromCache(out, cached)
                }
            }
            return nil, ErrCacheUnavailable
        },
    )
    
    if err != nil {
        return err
    }
    
    return nil
}

// Create executes Create with circuit breaker protection (no fallback for writes)
func (b *BreakerDBClient) Create(ctx context.Context, query interface{}, data BaseModel, others ...interface{}) error {
    _, err := b.breaker.Execute(ctx, func() (interface{}, error) {
        return nil, b.client.Create(ctx, query, data, others...)
    })
    return err
}

// Implement other IDBClient methods similarly...
```

### 6.5 Metrics Structure

```go
// pkg/circuitbreaker/metrics.go

package circuitbreaker

import (
    "sync"
    "time"
)

// Metrics holds circuit breaker metrics
type Metrics struct {
    Name              string
    State             string
    TotalRequests     uint64
    SuccessfulRequests uint64
    FailedRequests    uint64
    RejectedRequests  uint64 // Rejected due to open circuit
    LastStateChange   time.Time
    LastFailure       time.Time
    ConsecutiveFailures uint64
}

// MetricsCollector collects and aggregates metrics
type MetricsCollector interface {
    RecordSuccess(name string)
    RecordFailure(name string, err error)
    RecordRejection(name string)
    RecordStateChange(name string, from, to State)
    GetMetrics(name string) Metrics
    GetAllMetrics() map[string]Metrics
}

// DefaultMetricsCollector is a thread-safe metrics collector
type DefaultMetricsCollector struct {
    mu      sync.RWMutex
    metrics map[string]*Metrics
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() MetricsCollector {
    return &DefaultMetricsCollector{
        metrics: make(map[string]*Metrics),
    }
}

// Implement MetricsCollector interface...
```

---

## 7. Configuration Management

### 7.1 Configuration Loading

```go
// pkg/circuitbreaker/config.go

package circuitbreaker

import (
    "time"
    "github.com/spf13/viper"
    "github.com/sony/gobreaker"
)

// LoadConfig loads circuit breaker configuration from viper
func LoadConfig() (*GlobalConfig, error) {
    var config GlobalConfig
    
    if err := viper.UnmarshalKey("circuit_breaker", &config); err != nil {
        return nil, fmt.Errorf("failed to load circuit breaker config: %w", err)
    }
    
    // Validate configuration
    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("invalid circuit breaker config: %w", err)
    }
    
    return &config, nil
}

// GlobalConfig holds all circuit breaker configurations
type GlobalConfig struct {
    Defaults     DefaultConfig
    Nats         NatsConfig
    Databases    DatabasesConfig
    ExternalAPIs ExternalAPIsConfig
    Monitoring   MonitoringConfig
}

// DefaultConfig holds default circuit breaker settings
type DefaultConfig struct {
    MaxRequests          uint32
    Interval             time.Duration
    Timeout              time.Duration
    FailureThreshold     uint64
    FailureRateThreshold float64
    MinRequests          uint64
}

// ToBreaker converts config to gobreaker settings
func (c DefaultConfig) ToBreaker(name string, onStateChange func(string, gobreaker.State, gobreaker.State)) gobreaker.Settings {
    return gobreaker.Settings{
        Name:        name,
        MaxRequests: c.MaxRequests,
        Interval:    c.Interval,
        Timeout:     c.Timeout,
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            // Open circuit if:
            // 1. Consecutive failures exceed threshold
            if counts.ConsecutiveFailures >= c.FailureThreshold {
                return true
            }
            
            // 2. Failure rate exceeds threshold (after minimum requests)
            if counts.Requests >= c.MinRequests {
                failureRate := float64(counts.TotalFailures) / float64(counts.Requests)
                return failureRate >= c.FailureRateThreshold
            }
            
            return false
        },
        OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
            if onStateChange != nil {
                onStateChange(name, from, to)
            }
        },
    }
}
```

### 7.2 Dynamic Configuration Updates

Implement configuration hot-reload using file watchers or configuration management systems (e.g., Consul, etcd).

```go
// pkg/circuitbreaker/reload.go

package circuitbreaker

import (
    "fmt"
    "sync"
)

// ConfigReloader handles dynamic configuration updates
type ConfigReloader struct {
    mu       sync.RWMutex
    config   *GlobalConfig
    breakers map[string]*Breaker
}

// Reload reloads configuration and updates breakers
func (cr *ConfigReloader) Reload() error {
    cr.mu.Lock()
    defer cr.mu.Unlock()
    
    newConfig, err := LoadConfig()
    if err != nil {
        return fmt.Errorf("failed to reload config: %w", err)
    }
    
    // Update existing breakers with new configuration
    for name, breaker := range cr.breakers {
        if newBreakerConfig, ok := findBreakerConfig(newConfig, name); ok {
            breaker.UpdateConfig(newBreakerConfig)
        }
    }
    
    cr.config = newConfig
    return nil
}
```

---

## 8. Monitoring and Observability

### 8.1 Metrics to Track

**Circuit Breaker Metrics:**
- Circuit state (closed, open, half-open)
- State transition count
- Request count by state
- Success/failure rate
- Consecutive failures
- Time in each state

**Service Health Metrics:**
- Service availability percentage
- Mean time to recovery (MTTR)
- Mean time between failures (MTBF)

**Performance Metrics:**
- Request latency (with/without circuit breaker)
- Timeout rate
- Fallback usage rate

### 8.2 Prometheus Integration

```go
// pkg/circuitbreaker/prometheus.go

package circuitbreaker

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/sony/gobreaker"
)

var (
    circuitBreakerState = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "circuit_breaker_state",
            Help: "Current state of circuit breaker (0=closed, 1=open, 2=half-open)",
        },
        []string{"name", "type"},
    )
    
    circuitBreakerRequests = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "circuit_breaker_requests_total",
            Help: "Total number of requests through circuit breaker",
        },
        []string{"name", "type", "result"}, // result: success, failure, rejected
    )
    
    circuitBreakerStateChanges = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "circuit_breaker_state_changes_total",
            Help: "Total number of circuit breaker state changes",
        },
        []string{"name", "type", "from_state", "to_state"},
    )
)

func init() {
    prometheus.MustRegister(circuitBreakerState)
    prometheus.MustRegister(circuitBreakerRequests)
    prometheus.MustRegister(circuitBreakerStateChanges)
}

// PrometheusMetricsCollector implements MetricsCollector with Prometheus
type PrometheusMetricsCollector struct {
    *DefaultMetricsCollector
}

func (p *PrometheusMetricsCollector) RecordSuccess(name string, cbType string) {
    p.DefaultMetricsCollector.RecordSuccess(name)
    circuitBreakerRequests.WithLabelValues(name, cbType, "success").Inc()
}

func (p *PrometheusMetricsCollector) RecordFailure(name string, cbType string, err error) {
    p.DefaultMetricsCollector.RecordFailure(name, err)
    circuitBreakerRequests.WithLabelValues(name, cbType, "failure").Inc()
}

func (p *PrometheusMetricsCollector) RecordRejection(name string, cbType string) {
    p.DefaultMetricsCollector.RecordRejection(name)
    circuitBreakerRequests.WithLabelValues(name, cbType, "rejected").Inc()
}

func (p *PrometheusMetricsCollector) RecordStateChange(name string, cbType string, from, to gobreaker.State) {
    p.DefaultMetricsCollector.RecordStateChange(name, from, to)
    
    // Update state gauge
    stateValue := stateToFloat(to)
    circuitBreakerState.WithLabelValues(name, cbType).Set(stateValue)
    
    // Increment state change counter
    circuitBreakerStateChanges.WithLabelValues(
        name,
        cbType,
        from.String(),
        to.String(),
    ).Inc()
}

func stateToFloat(state gobreaker.State) float64 {
    switch state {
    case gobreaker.StateClosed:
        return 0
    case gobreaker.StateOpen:
        return 1
    case gobreaker.StateHalfOpen:
        return 2
    default:
        return -1
    }
}
```

### 8.3 Grafana Dashboard

Create a Grafana dashboard with the following panels:

1. **Circuit Breaker States Overview**
   - Gauge showing state of all circuit breakers
   - Color-coded: Green (closed), Red (open), Yellow (half-open)

2. **Request Success/Failure Rates**
   - Time series graph showing success vs failure rates
   - Separate lines for each service

3. **Circuit Opens/Closes**
   - Time series showing state transitions
   - Annotations for when circuits open

4. **Rejected Requests**
   - Counter showing requests rejected due to open circuit
   - By service

5. **Mean Time Between Failures (MTBF)**
   - Gauge showing average time between circuit opens

6. **Recovery Time**
   - Time taken to transition from open → half-open → closed

### 8.4 Logging

```go
// pkg/circuitbreaker/logging.go

package circuitbreaker

import (
    "log"
    "github.com/sony/gobreaker"
)

// StateChangeLogger logs circuit breaker state changes
func StateChangeLogger(name string, from, to gobreaker.State) {
    log.Printf(
        "[CIRCUIT BREAKER] %s: State transition %s -> %s",
        name,
        from.String(),
        to.String(),
    )
    
    // Send alerts for critical state changes
    if to == gobreaker.StateOpen {
        alertCircuitOpened(name)
    } else if from == gobreaker.StateOpen && to == gobreaker.StateClosed {
        alertCircuitRecovered(name)
    }
}

func alertCircuitOpened(name string) {
    // Integrate with alerting system (PagerDuty, Slack, etc.)
    log.Printf("[ALERT] Circuit breaker %s has OPENED", name)
}

func alertCircuitRecovered(name string) {
    log.Printf("[INFO] Circuit breaker %s has RECOVERED", name)
}
```

---

## 9. Testing Strategy

### 9.1 Unit Tests

```go
// pkg/circuitbreaker/breaker_test.go

package circuitbreaker_test

import (
    "context"
    "errors"
    "testing"
    "time"
    
    "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/circuitbreaker"
    "github.com/stretchr/testify/assert"
)

func TestBreakerOpenAfterFailures(t *testing.T) {
    config := circuitbreaker.Config{
        Name:          "test-breaker",
        MaxRequests:   3,
        Interval:      10 * time.Second,
        Timeout:       5 * time.Second,
        FailureThreshold: 3,
    }
    
    breaker := circuitbreaker.NewBreaker(config)
    
    // Simulate failures
    for i := 0; i < 3; i++ {
        _, err := breaker.Execute(context.Background(), func() (interface{}, error) {
            return nil, errors.New("simulated failure")
        })
        assert.Error(t, err)
    }
    
    // Circuit should be open now
    assert.Equal(t, gobreaker.StateOpen, breaker.State())
    
    // Next request should be rejected immediately
    _, err := breaker.Execute(context.Background(), func() (interface{}, error) {
        return "success", nil
    })
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "circuit breaker is open")
}

func TestBreakerRecovery(t *testing.T) {
    config := circuitbreaker.Config{
        Name:          "test-breaker",
        MaxRequests:   1,
        Interval:      10 * time.Second,
        Timeout:       1 * time.Second,
        FailureThreshold: 2,
    }
    
    breaker := circuitbreaker.NewBreaker(config)
    
    // Open circuit
    for i := 0; i < 2; i++ {
        breaker.Execute(context.Background(), func() (interface{}, error) {
            return nil, errors.New("failure")
        })
    }
    
    assert.Equal(t, gobreaker.StateOpen, breaker.State())
    
    // Wait for timeout
    time.Sleep(2 * time.Second)
    
    // Circuit should be half-open
    // Successful request should close it
    result, err := breaker.Execute(context.Background(), func() (interface{}, error) {
        return "success", nil
    })
    
    assert.NoError(t, err)
    assert.Equal(t, "success", result)
    
    // After successful requests, should be closed
    time.Sleep(100 * time.Millisecond)
    assert.Equal(t, gobreaker.StateClosed, breaker.State())
}
```

### 9.2 Integration Tests

```go
// pkg/custom-nats/breaker_test.go

package custom_nats_test

import (
    "testing"
    "time"
    
    "github.com/nats-io/nats.go"
    "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/custom-nats"
    "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/circuitbreaker"
)

func TestNATSBreakerWithFailingService(t *testing.T) {
    // Setup NATS server
    ns := nats.StartServer()
    defer ns.Shutdown()
    
    conn, _ := nats.Connect(ns.ClientURL())
    defer conn.Close()
    
    // Setup circuit breaker client
    configs := map[string]circuitbreaker.Config{
        "test-service": {
            Name:             "test-service",
            MaxRequests:      1,
            Interval:         10 * time.Second,
            Timeout:          1 * time.Second,
            FailureThreshold: 2,
        },
    }
    
    client := custom_nats.NewBreakerClient(conn, configs, 2*time.Second)
    
    // Simulate failing service by not subscribing to subject
    
    // First failure
    _, err := client.Request("test-service.Method", []byte("request"), 1*time.Second)
    assert.Error(t, err)
    
    // Second failure should open circuit
    _, err = client.Request("test-service.Method", []byte("request"), 1*time.Second)
    assert.Error(t, err)
    
    // Third request should be rejected immediately (circuit open)
    start := time.Now()
    _, err = client.Request("test-service.Method", []byte("request"), 1*time.Second)
    elapsed := time.Since(start)
    
    assert.Error(t, err)
    assert.Less(t, elapsed, 100*time.Millisecond) // Should fail fast
    assert.Contains(t, err.Error(), "circuit breaker")
}
```

### 9.3 Chaos Testing

Implement chaos testing to validate circuit breaker behavior under various failure scenarios:

**Test Scenarios:**

1. **Service Unavailability**
   - Kill a microservice
   - Verify circuit opens
   - Verify requests fail fast
   - Restart service
   - Verify circuit closes

2. **Database Failures**
   - Simulate database connection errors
   - Verify circuit protection
   - Verify fallback cache usage

3. **Network Partitions**
   - Introduce network delays
   - Verify timeout behavior
   - Verify circuit breaker trips

4. **Cascading Failures**
   - Fail multiple services simultaneously
   - Verify independent circuit breaker behavior
   - Verify API gateway remains responsive

**Tools:**
- Chaos Mesh (Kubernetes)
- toxiproxy (network simulation)
- Custom failure injection

### 9.4 Load Testing

Validate circuit breaker performance under load:

```bash
# Load test with K6
k6 run --vus 100 --duration 60s load-test.js
```

**Metrics to collect:**
- Response time with circuit breaker
- Throughput impact
- Memory usage
- CPU usage

---

## 10. Rollout Plan

### 10.1 Rollout Strategy

**Approach: Gradual Rollout with Feature Flags**

```yaml
# configs/config.yaml
feature_flags:
  circuit_breaker_enabled: true
  circuit_breaker_services:
    - auth
    - order
    # Add more services gradually
```

### 10.2 Rollout Phases

#### Phase 1: Development Environment (Week 1-2)
- Deploy circuit breaker to dev environment
- Test all integration points
- Validate metrics collection
- Fix bugs and tune configuration

#### Phase 2: Staging Environment (Week 3)
- Deploy to staging
- Run comprehensive integration tests
- Perform chaos testing
- Load testing with realistic traffic patterns

#### Phase 3: Production Canary (Week 4)
- Enable circuit breaker for 10% of traffic
- Monitor metrics closely
- Compare performance with control group
- Gradually increase to 25%, 50%, 75%

#### Phase 4: Full Production (Week 5)
- Enable for 100% of traffic
- Continue monitoring
- Tune configurations based on real traffic

#### Phase 5: Optimization (Week 6)
- Analyze metrics and logs
- Fine-tune thresholds
- Optimize fallback strategies
- Document lessons learned

### 10.3 Rollback Plan

If issues are detected:

1. **Immediate Rollback**
   - Use feature flag to disable circuit breakers
   - No code deployment needed

2. **Partial Rollback**
   - Disable circuit breaker for specific services
   - Keep working integrations active

3. **Rollback Triggers**
   - Increased error rate (>5%)
   - Performance degradation (>20% latency increase)
   - Memory/CPU issues
   - Customer complaints

---

## 11. Success Metrics

### 11.1 Technical Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Mean Time to Detect Failure (MTTD) | <5 seconds | Time from service failure to circuit open |
| Mean Time to Recovery (MTTR) | <30 seconds | Time from failure detection to recovery |
| Failed Request Reduction | >80% | Reduction in timeout errors |
| Fallback Success Rate | >95% | Successful fallback responses |
| Circuit Breaker Overhead | <5ms | Latency added by circuit breaker |
| False Positives | <1% | Incorrect circuit opens |

### 11.2 Business Metrics

| Metric | Target | Impact |
|--------|--------|--------|
| API Response Time (P95) | <200ms | Improved with fail-fast |
| API Availability | >99.9% | Better resilience |
| User Error Rate | <0.1% | Fewer timeout errors |
| Resource Utilization | -20% | Reduced wasted connections |
| Customer Satisfaction | +10% | Better user experience |

### 11.3 Monitoring Dashboard

Create a real-time dashboard showing:

1. **System Health Overview**
   - All services status
   - Circuit breaker states
   - Overall system availability

2. **Circuit Breaker Activity**
   - Number of open circuits
   - State transition frequency
   - Top failing services

3. **Performance Impact**
   - Request latency comparison (with/without breaker)
   - Throughput trends
   - Resource usage

4. **Alerts**
   - Critical circuit opens
   - High failure rates
   - Recovery events

---

## Appendix

### A. Code Examples

#### A.1 API Gateway Integration

```go
// api_gateway/main.go (updated)

package apigateway

import (
    "context"
    "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/custom-nats"
    "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/circuitbreaker"
)

type APIGateway struct {
    natsClient *custom_nats.BreakerClient
    // ... other fields
}

func NewAPIGateway(natsConn *nats.Conn) *APIGateway {
    // Load circuit breaker configs
    cbConfig, _ := circuitbreaker.LoadConfig()
    
    // Create NATS client with circuit breaker
    natsClient := custom_nats.NewBreakerClient(
        natsConn,
        cbConfig.Nats.ToServiceConfigs(),
        30*time.Second,
    )
    
    return &APIGateway{
        natsClient: natsClient,
    }
}

func (gw *APIGateway) ServeHTTP() func(http.ResponseWriter, *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        // ... existing code ...
        
        // Use circuit breaker protected client
        msgResponse, err := gw.natsClient.RequestWithFallback(
            natsReq.Subject,
            natsReqByte,
            gw.timeout,
            func() (*nats.Msg, error) {
                // Fallback: return cached response or friendly error
                return createFallbackResponse(natsReq.Subject)
            },
        )
        
        if err != nil {
            // Check if circuit is open
            if circuitbreaker.IsCircuitOpen(err) {
                gw.sendErrorResponse(w, 
                    "Service temporarily unavailable. Please try again later.",
                    http.StatusServiceUnavailable,
                )
                return
            }
            gw.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
            return
        }
        
        // ... rest of the code ...
    }
}
```

#### A.2 Service Repository Integration

```go
// apps/order/repository/repo.go (updated)

package order_repository

import (
    "context"
    "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/repo"
    "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/circuitbreaker"
)

type OrderRepository struct {
    client repo.IDBClient
}

func NewOrderRepository(dbClient repo.IDBClient) *OrderRepository {
    // Load circuit breaker config
    cbConfig, _ := circuitbreaker.LoadConfig()
    
    // Wrap DB client with circuit breaker
    breakerConfig := cbConfig.Databases.Postgres.ToConfig("order-db")
    breaker := circuitbreaker.NewBreaker(breakerConfig)
    
    breakerClient := repo.NewBreakerDBClient(dbClient, breaker, nil)
    
    return &OrderRepository{
        client: breakerClient,
    }
}

// All existing methods work unchanged due to interface compatibility
func (r *OrderRepository) FindOrderById(ctx context.Context, id uuid.UUID) (*Order, error) {
    var order Order
    query := "SELECT * FROM orders WHERE id = $1"
    
    // Circuit breaker protection is transparent
    err := r.client.FindOne(ctx, &order, query, id)
    if err != nil {
        return nil, err
    }
    
    return &order, nil
}
```

### B. Configuration Examples

#### B.1 Development Configuration

```yaml
# configs/config.dev.yaml
circuit_breaker:
  defaults:
    max_requests: 5
    interval: 30s
    timeout: 10s
    failure_threshold: 3
    failure_rate_threshold: 0.5
    min_requests: 5

  nats:
    enabled: true
    services:
      auth:
        failure_threshold: 2
        timeout: 10s
      order:
        failure_threshold: 2
        timeout: 10s

  databases:
    postgres:
      enabled: true
      timeout: 20s
    mongodb:
      enabled: true
      timeout: 20s
    redis:
      enabled: true
      timeout: 5s
      fallback_to_memory: true

  monitoring:
    metrics_enabled: true
    log_state_changes: true
```

#### B.2 Production Configuration

```yaml
# configs/config.prod.yaml
circuit_breaker:
  defaults:
    max_requests: 3
    interval: 60s
    timeout: 30s
    failure_threshold: 5
    failure_rate_threshold: 0.6
    min_requests: 10

  nats:
    enabled: true
    services:
      auth:
        failure_threshold: 5
        timeout: 20s
        interval: 60s
      order:
        failure_threshold: 5
        timeout: 20s
        interval: 60s
      product:
        failure_threshold: 5
        timeout: 20s
      # ... all services

  databases:
    postgres:
      enabled: true
      timeout: 30s
      failure_threshold: 5
    mongodb:
      enabled: true
      timeout: 30s
      failure_threshold: 5
    redis:
      enabled: true
      timeout: 10s
      failure_threshold: 3
      fallback_to_memory: true

  external_apis:
    zitadel:
      enabled: true
      timeout: 45s
      failure_threshold: 3
      cache_fallback: true

  monitoring:
    metrics_enabled: true
    log_state_changes: true
    alert_on_open: true
```

### C. Runbook

#### C.1 Circuit Breaker Opened

**Scenario:** Circuit breaker has opened for a service

**Investigation Steps:**

1. **Check Service Health**
   ```bash
   # Check if service is running
   kubectl get pods -l app=order-service
   
   # Check service logs
   kubectl logs -l app=order-service --tail=100
   ```

2. **Check Metrics**
   - Open Grafana dashboard
   - Look at error rate for the service
   - Check service latency

3. **Verify Database Connectivity**
   ```bash
   # Test database connection
   psql -h <db-host> -U <user> -d <database> -c "SELECT 1"
   ```

4. **Check NATS Connectivity**
   ```bash
   # Check NATS connection
   nats sub "order.*" --count=1
   ```

**Resolution:**

1. Fix the underlying issue (restart service, fix database, etc.)
2. Circuit will automatically recover after timeout
3. Monitor recovery in dashboard
4. If manual intervention needed:
   ```bash
   # Reset circuit breaker via API
   curl -X POST http://api-gateway/admin/circuit-breaker/reset/order
   ```

#### C.2 High Rejection Rate

**Scenario:** Many requests are being rejected due to open circuits

**Actions:**

1. **Identify Affected Services**
   ```bash
   # Check circuit breaker states
   curl http://api-gateway/admin/circuit-breaker/status
   ```

2. **Assess Impact**
   - Check user-facing errors
   - Verify fallback responses are working

3. **Emergency Actions**
   - If false positive: Increase failure threshold temporarily
   - If legitimate: Scale up affected services
   - If database issue: Add read replicas

4. **Long-term Fix**
   - Tune circuit breaker thresholds
   - Improve service reliability
   - Add better fallbacks

### D. Dependencies

```go
// go.mod additions
require (
    github.com/sony/gobreaker v0.5.0
    github.com/prometheus/client_golang v1.19.0
)
```

### E. References

- Circuit Breaker Pattern: https://martinfowler.com/bliki/CircuitBreaker.html
- Sony Gobreaker: https://github.com/sony/gobreaker
- Resilience Patterns: https://docs.microsoft.com/en-us/azure/architecture/patterns/circuit-breaker
- NATS Documentation: https://docs.nats.io/

---

## Conclusion

This implementation plan provides a comprehensive strategy for adding circuit breaker pattern to the e-commerce microservice architecture. The phased approach ensures minimal disruption while maximizing resilience benefits.

**Key Takeaways:**

1. **Comprehensive Protection**: Circuit breakers at all integration points (NATS, databases, external APIs)
2. **Graceful Degradation**: Fallback mechanisms ensure user experience
3. **Observability**: Detailed metrics and monitoring for operations
4. **Flexibility**: Configuration-driven with hot-reload capability
5. **Production Ready**: Thorough testing and gradual rollout

**Next Steps:**

1. Review and approve this plan
2. Set up development environment
3. Begin Phase 1 implementation
4. Schedule weekly progress reviews

**Timeline:** 6 weeks from approval to production deployment

**Team Requirements:**
- 2 backend developers
- 1 DevOps engineer
- 1 QA engineer

---

**Document Version:** 1.0  
**Date:** November 10, 2025  
**Author:** AI Assistant  
**Status:** Draft - Pending Review


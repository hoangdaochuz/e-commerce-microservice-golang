# Observability Documentation

## Mục lục
1. [Tổng quan](#tổng-quan)
2. [Kiến trúc Observability](#kiến-trúc-observability)
3. [Tracing (Distributed Tracing)](#tracing-distributed-tracing)
4. [Metrics](#metrics)
5. [Logging](#logging)
6. [Visualization với Grafana](#visualization-với-grafana)
7. [Cấu hình Infrastructure](#cấu-hình-infrastructure)
8. [Hướng dẫn sử dụng](#hướng-dẫn-sử-dụng)

---

## Tổng quan

Dự án E-commerce Microservice triển khai đầy đủ **3 trụ cột của Observability**:

| Pillar | Công nghệ | Mục đích |
|--------|-----------|----------|
| **Tracing** | OpenTelemetry + Tempo | Theo dõi request flow xuyên suốt các services |
| **Metrics** | Prometheus + Alloy | Thu thập và lưu trữ metrics từ các services |
| **Logging** | Zap + Loki + Alloy | Thu thập và tập trung logs từ tất cả services |

### Sơ đồ tổng quan

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              OBSERVABILITY STACK                                 │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│   ┌──────────────┐    ┌──────────────┐    ┌──────────────┐                       │
│   │  API Gateway │    │ Auth Service │    │ Order Service│                       │
│   │              │    │              │    │              │                       │
│   │ • Traces     │    │ • Traces     │    │ • Traces     │                       │
│   │ • Metrics    │    │ • Logs       │    │ • Logs       │                       │
│   │ • Logs       │    │              │    │              │                       │
│   └──────┬───────┘    └──────┬───────┘    └──────┬───────┘                       │
│          │                   │                   │                               │
│          └───────────────────┼───────────────────┘                               │
│                              │                                                   │
│                              ▼                                                   │
│                    ┌─────────────────┐                                           │
│                    │      ALLOY      │ ◄── OpenTelemetry Collector               │
│                    │  (Port: 12345)  │     OTLP Receiver (4317/4318)             │
│                    └────────┬────────┘                                           │
│                             │                                                    │
│          ┌──────────────────┼──────────────────┐                                 │
│          │                  │                  │                                 │
│          ▼                  ▼                  ▼                                 │
│   ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                          │
│   │    TEMPO    │    │ PROMETHEUS  │    │    LOKI     │                          │
│   │ (Port: 3200)│    │ (Port: 9090)│    │ (Port: 3100)│                          │
│   │             │    │             │    │             │                          │
│   │   Traces    │    │   Metrics   │    │    Logs     │                          │
│   └──────┬──────┘    └──────┬──────┘    └──────┬──────┘                          │
│          │                  │                  │                                 │
│          └──────────────────┼──────────────────┘                                 │
│                             │                                                    │
│                             ▼                                                    │
│                    ┌─────────────────┐                                           │
│                    │     GRAFANA     │                                           │
│                    │  (Port: 3001)   │                                           │
│                    │                 │                                           │
│                    │  Visualization  │                                           │
│                    └─────────────────┘                                           │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## Kiến trúc Observability

### Components chính

| Component | Port | Mô tả |
|-----------|------|-------|
| **Grafana Alloy** | 12345 (UI), 4317 (gRPC), 4318 (HTTP) | OpenTelemetry Collector - thu thập traces, metrics, logs |
| **Grafana Tempo** | 3200 | Distributed tracing backend |
| **Prometheus** | 9090 | Time-series database cho metrics |
| **Grafana Loki** | 3100 | Log aggregation system |
| **Grafana** | 3001 | Visualization dashboard |

### Data Flow

1. **Traces Flow:**
   ```
   Application → OTLP (gRPC/HTTP) → Alloy → Tempo → Grafana
                                      ↓
                               Service Graph Metrics → Prometheus
   ```

2. **Metrics Flow:**
   ```
   Application → /metrics endpoint → Prometheus → Grafana
                      ↓
   Application → Alloy (SpanMetrics) → Prometheus → Grafana
   ```

3. **Logs Flow:**
   ```
   Application → Log Files → Alloy → Loki → Grafana
   Docker Containers → Alloy → Loki → Grafana
   ```

---

## Tracing (Distributed Tracing)

### Tổng quan

Hệ thống sử dụng **OpenTelemetry** để implement distributed tracing, cho phép theo dõi request xuyên suốt các microservices.

### Cấu hình Tracing (`pkg/tracing/main.go`)

```go
type TracingConfig struct {
    ServiceName  string            // Tên service
    Attributes   map[string]string // Custom attributes
    SamplingRate float64           // Tỷ lệ sampling (0-1)
    BatchTimeout time.Duration     // Timeout cho batch export
    BatchMaxSize int               // Max size của batch
    OtelEndpoint string            // OTLP endpoint (Alloy)
}
```

### Khởi tạo Tracing trong Service

**API Gateway (`api_gateway/main.go`):**
```go
shutdownTracing, err := tracing.InitializeTraceRegistry(&tracing.TracingConfig{
    ServiceName:  "api_gateway",
    SamplingRate: 1,                    // 100% sampling
    BatchTimeout: 5 * time.Second,
    BatchMaxSize: 512,
    OtelEndpoint: otlpEndpoint,         // localhost:4317
})
defer shutdownTracing()
```

**Auth Service (`apps/auth/cmd/main.go`):**
```go
server := custom_nats.NewServer(natsConn, router, authService_api.NATS_SUBJECT, authServiceClient, &custom_nats.ServerConfig{
    ServiceName:  "auth",
    OtelEndpoint: config.GeneralConfig.OTLP_Endpoint,
})
```

**Order Service (`apps/order/cmd/main.go`):**
```go
server := custom_nats.NewServer(natsConn, router, order_api.NATS_SUBJECT, orderRouterClient, &custom_nats.ServerConfig{
    ServiceName:  "order",
    OtelEndpoint: config.GeneralConfig.OTLP_Endpoint,
})
```

### Trace Propagation

Context propagation được thực hiện qua HTTP headers:

```go
// Inject trace context vào HTTP request
func InjectTraceIntoHttpReq(ctx context.Context, req *http.Request) {
    otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
}

// Extract trace context từ HTTP request
func ExtractTraceFromHttpRequest(req *http.Request) context.Context {
    return otel.GetTextMapPropagator().Extract(req.Context(), propagation.HeaderCarrier(req.Header))
}
```

### Tạo Span

```go
// Tạo span mới với context
ctx, span := tracing.SpanContext(timeoutCtx, r.Header, "outbound request: /api/v1/auth/Login")
defer span.End()

// Set span attributes
span.SetAttributes(
    semconv.HTTPMethod(r.Method),
    semconv.HTTPRoute(r.URL.Path),
    semconv.HTTPStatusCode(response.StatusCode),
)

// Ghi nhận lỗi vào span
if err != nil {
    tracing.SetSpanError(span, err)
}
```

### HTTP Client với Tracing

```go
// BreakerHTTPClient sử dụng otelhttp để tự động instrument HTTP calls
httpClient := &http.Client{
    Transport: otelhttp.NewTransport(&http.Transport{...}, 
        otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
            return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
        })),
}
```

### Tempo Configuration (`infra/tempo.yaml`)

```yaml
# Receiver configuration
distributor:
  receivers:
    otlp:
      protocols:
        grpc:
          endpoint: 0.0.0.0:4317
        http:
          endpoint: 0.0.0.0:4318

# Metrics generation từ traces
metrics_generator:
  processor:
    service_graphs:
      dimensions:
        - http.method
        - http.status_code
      enable_client_server_prefix: true
    span_metrics:
      dimensions:
        - http.method
        - http.status_code
        - http.route
      histogram_buckets: [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
```

---

## Metrics

### Tổng quan

Metrics được thu thập từ nhiều nguồn:
1. **Application Metrics** - expose qua `/metrics` endpoint
2. **Span Metrics** - generate từ traces bởi Alloy
3. **Service Graph Metrics** - generate từ traces để vẽ service topology

### Prometheus Configuration (`infra/prometheus.yml`)

```yaml
global:
  scrape_interval: 1m
  evaluation_interval: 1m

scrape_configs:
  - job_name: "prometheus"
    static_configs:
      - targets: ["localhost:9090"]

  - job_name: "nats"
    static_configs:
      - targets: ["nats:8222"]

  - job_name: "ecommerce-api-gateway"
    static_configs:
      - targets: ["host.docker.internal:8080"]
```

### HTTP Metrics Middleware (`pkg/metric/httpmiddleware/http_middleware.go`)

Middleware tự động thu thập các metrics cho mỗi HTTP request:

| Metric | Type | Description |
|--------|------|-------------|
| `http_request_total` | Counter | Tổng số HTTP requests |
| `http_request_duration_seconds` | Histogram | Latency của requests |
| `http_request_size_bytes` | Summary | Kích thước request |
| `http_response_size_bytes` | Summary | Kích thước response |

**Labels:** `path`, `method`, `code`

```go
// Khởi tạo Metric Middleware
func NewMiddleware(buckets []float64, registry *prometheus.Registry) *Middleware {
    return &Middleware{
        reqTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
            Name: "http_request_total",
            Help: "Total number of HTTP requests",
        }, []string{"path", "method", "code"}),
        
        reqDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "Tracks the latencies for HTTP requests.",
            Buckets: buckets,
        }, []string{"path", "method", "code"}),
        // ...
    }
}
```

### Sử dụng Metrics trong API Gateway

```go
// Khởi tạo registry
registryWrapper := metric.NewMetricWrapper()
registryWrapper.RegisterCollectorDefault()  // Go runtime + process metrics
registry := registryWrapper.GetRegistry()

// Apply MetricMiddleware cho handlers
protectResourceHandler := useMiddleware(rootHandler, 
    CorsMiddleware, 
    ContentTypeMiddleware, 
    RateLimitMiddleware(rateLimiter), 
    MetricMiddleware(registry),  // <-- Metrics middleware
    AuthMiddleware,
)

// Expose /metrics endpoint
gw.mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
```

### Span Metrics (Alloy Configuration)

Alloy tự động generate metrics từ traces:

```hcl
otelcol.connector.spanmetrics "default" {
  histogram {
    explicit {
      buckets = ["5ms", "10ms", "50ms", "100ms", "250ms", "500ms", "1s", "2s", "5s"]
    }
  }

  dimension {
    name = "service.name"
  }
  dimension {
    name = "span.name"
  }
  dimension {
    name = "span.kind"
  }
  dimension {
    name = "status.code"
  }

  output {
    metrics = [otelcol.exporter.prometheus.metrics.input]
  }
}
```

### Circuit Breaker Metrics

HTTP Client với Circuit Breaker cung cấp metrics về trạng thái breaker:

```go
func (c *BreakerHTTPClient) GetMetrics() map[string]interface{} {
    return map[string]interface{}{
        "name":             c.breaker.GetName(),
        "state":            c.GetBreakerState(),      // closed/open/half-open
        "total_requests":   c.breaker.GetCount(),
        "success_requests": c.breaker.GetCountSuccessRequest(),
        "failed_requests":  c.breaker.GetCountFailureRequest(),
    }
}
```

---

## Logging

### Tổng quan

Logging được implement với **Zap** (high-performance logger) và tập trung qua **Grafana Alloy** vào **Loki**.

### Zap Logger Configuration (`pkg/logging/main.go`)

```go
func initLogger() {
    mode := viper.GetString("general_config.mode")
    
    // Development mode: colored output, human-readable
    developmentConfig := zap.NewDevelopmentConfig()
    developmentConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    developmentConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
    
    // Production mode: JSON format, structured
    productionConfig := zap.NewProductionConfig()
    productionConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    productionConfig.EncoderConfig.StacktraceKey = "stacktrace"
    productionConfig.EncoderConfig.MessageKey = "msg"
    // ...
}

// Sử dụng logger
logging.GetSugaredLogger().Infof("Starting API Gateway in port %s", port)
logging.GetSugaredLogger().Errorf("Failed to process request: %v", err)
```

### Log Files

Applications ghi logs vào các file trong thư mục `logs/`:

| Service | Log File |
|---------|----------|
| API Gateway | `logs/api_gateway.log` |
| Auth Service | `logs/auth.log` |
| Order Service | `logs/order.log` |

### Alloy Log Collection (`infra/alloy-config.alloy`)

**1. File-based Log Collection:**

```hcl
loki.source.file "console_logs" {
  targets = [
    {
      __path__ = "/logs/api_gateway.log",
      job = "console",
    },
    {
      __path__ = "/logs/auth.log",
      job = "console",
    },
    {
      __path__ = "/logs/order.log",
      job = "console",
    },
  ]
  forward_to = [loki.process.console_pipeline.receiver]
}
```

**2. Docker Container Log Collection:**

```hcl
discovery.docker "containers" {
  host = "unix:///var/run/docker.sock"
}

loki.source.docker "docker_logs" {
  targets    = discovery.docker.containers.targets
  forward_to = [loki.process.docker_pipeline.receiver]
  host       = "unix:///var/run/docker.sock"
}
```

**3. Log Processing Pipeline:**

```hcl
loki.process "console_pipeline" {
  stage.labels {
    values = {
      source = "console",
    }
  }
  forward_to = [loki.write.default.receiver]
}

loki.process "docker_pipeline" {
  forward_to = [loki.write.default.receiver]

  stage.drop {
    older_than = "6h"  // Drop logs older than 6 hours
  }
}
```

### Loki Configuration (`infra/loki-config.yaml`)

```yaml
auth_enabled: false

server:
  http_listen_port: 3100

ingester:
  lifecycler:
    ring:
      kvstore:
        store: inmemory
      replication_factor: 1
  chunk_idle_period: 1h
  chunk_target_size: 1048576  # 1MB

storage_config:
  boltdb_shipper:
    active_index_directory: /loki/index
    cache_location: /loki/cache
  filesystem:
    directory: /loki/chunks

limits_config:
  ingestion_rate_mb: 20
  ingestion_burst_size_mb: 30
```

### Log với Trace ID

Logs được enrich với Trace ID để correlate với traces:

```go
logging.GetSugaredLogger().Infof("%s %s %v statusCode: %v traceId: %s", 
    r.Method, 
    r.URL.Path, 
    time.Since(start), 
    natsResponse.StatusCode, 
    span.SpanContext().TraceID().String(),  // <-- Trace ID
)
```

---

## Visualization với Grafana

### Truy cập Grafana

- **URL:** http://localhost:3001
- **Username:** admin
- **Password:** admin

### Cấu hình Data Sources

Thêm các data sources trong Grafana:

| Data Source | Type | URL |
|-------------|------|-----|
| Prometheus | Prometheus | http://prometheus:9090 |
| Loki | Loki | http://loki:3100 |
| Tempo | Tempo | http://tempo:3200 |

### Dashboards được đề xuất

1. **Service Overview Dashboard**
   - Request rate
   - Error rate
   - Latency percentiles (p50, p95, p99)
   
2. **Service Graph Dashboard**
   - Service topology từ traces
   - Dependencies visualization
   
3. **Logs Explorer**
   - Search logs by service
   - Correlate logs với traces

### Sample Queries

**Prometheus (Metrics):**
```promql
# Request rate per service
rate(http_request_total[5m])

# 95th percentile latency
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Error rate
sum(rate(http_request_total{code=~"5.."}[5m])) / sum(rate(http_request_total[5m]))
```

**Loki (Logs):**
```logql
# Logs from API Gateway
{job="console"} |= "api_gateway"

# Error logs
{job="console"} |= "ERROR"

# Logs with specific trace ID
{job="console"} |= "traceId=abc123"
```

---

## Cấu hình Infrastructure

### Docker Compose Services

```yaml
services:
  # OpenTelemetry Collector
  alloy:
    image: grafana/alloy:latest
    ports:
      - "12345:12345"  # UI
      - "4317:4317"    # OTLP gRPC
      - "4318:4318"    # OTLP HTTP
    volumes:
      - ./alloy-config.alloy:/etc/alloy/config.alloy:ro
      - ../logs:/logs

  # Distributed Tracing Backend
  tempo:
    image: grafana/tempo:latest
    ports:
      - "3200:3200"
    volumes:
      - ./tempo.yaml:/etc/tempo.yaml

  # Metrics Database
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    command:
      - "--config.file=/etc/prometheus/prometheus.yml"
      - "--web.enable-remote-write-receiver"  # Enable remote write

  # Log Aggregation
  loki:
    image: grafana/loki:latest
    ports:
      - "3100:3100"

  # Visualization
  grafana:
    image: grafana/grafana:latest
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
```

### Khởi động Infrastructure

```bash
# Start all observability services
cd infra
docker-compose up -d prometheus grafana loki alloy tempo

# Check services status
docker-compose ps

# View logs
docker-compose logs -f alloy
docker-compose logs -f tempo
```

---

## Hướng dẫn sử dụng

### 1. Thêm Tracing vào Service mới

```go
import "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/tracing"

func main() {
    // Initialize tracing
    shutdown, err := tracing.InitializeTraceRegistry(&tracing.TracingConfig{
        ServiceName:  "my-service",
        SamplingRate: 1,
        BatchTimeout: 5 * time.Second,
        BatchMaxSize: 512,
        OtelEndpoint: "localhost:4317",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer shutdown()
    
    // Your service logic...
}
```

### 2. Tạo Custom Span

```go
ctx, span := tracing.SpanContext(ctx, req.Header, "operation-name")
defer span.End()

// Add attributes
span.SetAttributes(
    attribute.String("custom.attribute", "value"),
)

// Record error if any
if err != nil {
    tracing.SetSpanError(span, err)
}
```

### 3. Thêm Metrics cho Handler

```go
// Wrap handler với MetricMiddleware
handler := MetricMiddleware(registry)(yourHandler)

// Expose metrics endpoint
mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
```

### 4. Structured Logging

```go
import "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/logging"

// Info log
logging.GetSugaredLogger().Infof("Processing request: %s", requestID)

// Error log with context
logging.GetSugaredLogger().Errorf("Failed to process: %v, traceId: %s", err, traceID)

// Warning log
logging.GetSugaredLogger().Warnf("Rate limit approaching: %d/%d", current, limit)
```

### 5. Correlate Logs với Traces

Luôn include Trace ID trong logs để dễ dàng correlate:

```go
traceID := span.SpanContext().TraceID().String()
logging.GetSugaredLogger().Infof("Request processed, traceId: %s", traceID)
```

---

## Best Practices

### Tracing
- ✅ Luôn đặt tên span có ý nghĩa
- ✅ Thêm relevant attributes vào span
- ✅ Propagate context qua service boundaries
- ✅ Record errors vào span
- ❌ Không tạo quá nhiều spans (over-instrumentation)

### Metrics
- ✅ Sử dụng labels có cardinality thấp
- ✅ Đặt tên metrics theo convention
- ✅ Chọn đúng metric type (Counter, Gauge, Histogram)
- ❌ Không sử dụng high-cardinality labels (user ID, request ID)

### Logging
- ✅ Sử dụng structured logging
- ✅ Include trace ID trong logs
- ✅ Log ở đúng level (INFO, WARN, ERROR)
- ❌ Không log sensitive data (passwords, tokens)

---

## Troubleshooting

### Traces không hiển thị trong Tempo

1. Kiểm tra Alloy logs: `docker-compose logs alloy`
2. Verify OTLP endpoint trong config
3. Kiểm tra Tempo receiver: `curl http://localhost:3200/ready`

### Metrics không được scrape

1. Kiểm tra `/metrics` endpoint: `curl http://localhost:8080/metrics`
2. Verify Prometheus targets: http://localhost:9090/targets
3. Kiểm tra prometheus.yml config

### Logs không xuất hiện trong Loki

1. Kiểm tra log files tồn tại trong `logs/`
2. Verify Alloy có mount đúng volumes
3. Kiểm tra Loki: `curl http://localhost:3100/ready`

---

## References

- [OpenTelemetry Go SDK](https://opentelemetry.io/docs/instrumentation/go/)
- [Grafana Tempo Documentation](https://grafana.com/docs/tempo/latest/)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Loki Documentation](https://grafana.com/docs/loki/latest/)
- [Grafana Alloy Documentation](https://grafana.com/docs/alloy/latest/)
- [Uber Zap Logger](https://github.com/uber-go/zap)

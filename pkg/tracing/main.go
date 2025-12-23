package tracing

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/logging"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TracingConfig struct {
	ServiceName string
	Attributes  map[string]string

	// sampling option config
	SamplingRate float64

	// batch option config
	BatchTimeout time.Duration
	BatchMaxSize int
	OtelEndpoint string
}

var tracer trace.Tracer
var traceProvider trace.TracerProvider

func InitializeTraceRegistry(cfg *TracingConfig) (func(), error) {
	// var initError error
	otlpEndpoint := cfg.OtelEndpoint
	var shutdownFunc func(ctx context.Context) error
	// once.Do(func() {
	ctx := context.Background()
	if cfg.BatchTimeout == 0 {
		cfg.BatchTimeout = 5 * time.Second
	}
	if cfg.BatchMaxSize == 0 {
		cfg.BatchMaxSize = 512
	}

	if cfg.SamplingRate == 0 {
		cfg.SamplingRate = 1.0
	}

	// create otlp exporter
	conn, err := grpc.NewClient(otlpEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("fail to create grpc tracing exporter connection: %w", err)

	}
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("fail to create a exporter: %w", err)
	}

	// build custom attributes
	var attributes = []attribute.KeyValue{
		semconv.ServiceName(cfg.ServiceName),
	}

	if len(cfg.Attributes) > 0 {
		for key, value := range cfg.Attributes {
			attributes = append(attributes, attribute.String(key, value))
		}
	}
	// Create a resource
	resource, err := resource.New(ctx,
		resource.WithAttributes(attributes...),
		resource.WithHost(),
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithContainer(),
	)

	if err != nil {
		return nil, fmt.Errorf("fail to create a resource: %w", err)
	}

	// Create a sampler
	var sampler sdktrace.Sampler
	switch {
	case cfg.SamplingRate >= 1:
		sampler = sdktrace.AlwaysSample()
	case cfg.SamplingRate < 0:
		sampler = sdktrace.NeverSample()
	default:
		sampler = sdktrace.TraceIDRatioBased(cfg.SamplingRate)
	}

	// Create trace provider
	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(cfg.BatchTimeout),
			sdktrace.WithMaxExportBatchSize(cfg.BatchMaxSize),
		),
		sdktrace.WithResource(resource),
		sdktrace.WithSampler(sampler),
	)

	// Set global providers
	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	shutdownFunc = traceProvider.Shutdown
	tracer = traceProvider.Tracer(cfg.ServiceName)
	logging.GetSugaredLogger().Infof("Tracing has inialized for service %s", cfg.ServiceName)
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := shutdownFunc(ctx)
		if err != nil {
			logging.GetSugaredLogger().Error("fail to shut down tracing of service %s", cfg.ServiceName)
		}
	}, nil
}

func GetTraceProvider() trace.TracerProvider {
	return traceProvider
}

func InjectTraceIntoHttpReq(ctx context.Context, req *http.Request) {
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
}

func ExtractTraceFromHttpRequest(req *http.Request) context.Context {
	return otel.GetTextMapPropagator().Extract(req.Context(), propagation.HeaderCarrier(req.Header))
}

func SpanContext(ctx context.Context, header http.Header, spanName string) (context.Context, trace.Span) {
	if tracer == nil {
		if ctx == nil {
			ctx = context.Background()
		}
		return ctx, trace.SpanFromContext(ctx)
	}

	nCtx := otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(header))
	spanCtx, span := tracer.Start(nCtx, spanName)
	return spanCtx, span
}

func SetSpanError(span trace.Span, err error) {
	span.SetStatus(codes.Error, err.Error())
	span.RecordError(err)
}

package httpmiddleware

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricHttpMiddleware interface {
	WrapHandler(handlerName string, handler http.Handler) http.HandlerFunc
}

type Middleware struct {
	buckets  []float64
	registry *prometheus.Registry
}

func NewMiddleware(buckets []float64, registry *prometheus.Registry) *Middleware {
	if buckets == nil {
		buckets = prometheus.ExponentialBuckets(0.1, 1.5, 5)
	}
	return &Middleware{
		buckets:  buckets,
		registry: registry,
	}
}

func (m *Middleware) WrapHandler(handlerName string, handler http.Handler) http.HandlerFunc {
	reg := prometheus.WrapRegistererWith(prometheus.Labels{
		"handler": handlerName,
	}, m.registry)

	requestTotal := promauto.With(reg).NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "code"},
	)

	requestDuration := promauto.With(reg).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Tracks the latencies for HTTP requests.",
			Buckets: m.buckets,
		},
		[]string{"method", "code"},
	)

	requestSize := promauto.With(reg).NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_request_size_bytes",
			Help: "Size of HTTP requests",
		},
		[]string{"method", "code"},
	)

	responseSize := promauto.With(reg).NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_response_size_bytes",
			Help: "Size of HTTP responses",
		},
		[]string{"method", "code"},
	)

	base := promhttp.InstrumentHandlerCounter(requestTotal,
		promhttp.InstrumentHandlerDuration(requestDuration,
			promhttp.InstrumentHandlerRequestSize(requestSize,
				promhttp.InstrumentHandlerResponseSize(responseSize,
					handler,
				),
			),
		),
	)
	return base.ServeHTTP
}

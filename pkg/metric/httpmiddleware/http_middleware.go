package httpmiddleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type MetricHttpMiddleware interface {
	WrapHandler(handlerName string, handler http.Handler) http.HandlerFunc
}

type Middleware struct {
	registry    *prometheus.Registry
	reqTotal    *prometheus.CounterVec
	reqDuration *prometheus.HistogramVec
	reqSize     *prometheus.SummaryVec
	resSize     *prometheus.SummaryVec
}

func NewMiddleware(buckets []float64, registry *prometheus.Registry) *Middleware {
	if buckets == nil {
		buckets = prometheus.DefBuckets
	}

	m := &Middleware{
		registry: registry,
		reqTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_request_total",
			Help: "Total number of HTTP requests",
		}, []string{"path", "method", "code"}),
		reqDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Tracks the latencies for HTTP requests.",
			Buckets: buckets,
		}, []string{"path", "method", "code"}),
		reqSize: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Name: "http_request_size_bytes",
			Help: "Size of HTTP requests",
		}, []string{"path", "method", "code"}),
		resSize: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Name: "http_response_size_bytes",
			Help: "Size of HTTP responses",
		}, []string{"path", "method", "code"}),
	}

	// Register metrics once
	// We use Register and ignore AlreadyRegisteredError to be safe if called multiple times,
	// though ideally this should be called once.
	if err := registry.Register(m.reqTotal); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			m.reqTotal = are.ExistingCollector.(*prometheus.CounterVec)
		}
	}
	if err := registry.Register(m.reqDuration); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			m.reqDuration = are.ExistingCollector.(*prometheus.HistogramVec)
		}
	}
	if err := registry.Register(m.reqSize); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			m.reqSize = are.ExistingCollector.(*prometheus.SummaryVec)
		}
	}
	if err := registry.Register(m.resSize); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			m.resSize = are.ExistingCollector.(*prometheus.SummaryVec)
		}
	}

	return m
}

// responseWriterDelegator captures the status code
type responseWriterDelegator struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

func (w *responseWriterDelegator) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriterDelegator) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.written += int64(n)
	return n, err
}

func (m *Middleware) WrapHandler(path string, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code and size
		d := &responseWriterDelegator{ResponseWriter: w, statusCode: http.StatusOK}

		// Process request
		next.ServeHTTP(d, r)

		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(d.statusCode)
		method := r.Method

		// Observe metrics
		reqSize := computeApproxRequestSize(r)

		m.reqTotal.WithLabelValues(path, method, statusCode).Inc()
		m.reqDuration.WithLabelValues(path, method, statusCode).Observe(duration)
		m.reqSize.WithLabelValues(path, method, statusCode).Observe(float64(reqSize))
		m.resSize.WithLabelValues(path, method, statusCode).Observe(float64(d.written))
	}
}

func computeApproxRequestSize(r *http.Request) int64 {
	s := 0
	if r.URL != nil {
		s += len(r.URL.String())
	}
	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)
	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}
	return int64(s)
}

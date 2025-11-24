package apigateway

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"time"

	ratelimiter "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/rate_limiter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		frontEndUserDomain := viper.GetString("general_config.frontend_user_endpoint")

		w.Header().Set("Access-Control-Allow-Origin", frontEndUserDomain)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func ContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Path
		skipAuthPaths := []string{
			"/api/v1/auth/Login",
			"/api/v1/auth/Callback",
			"/callback",
		}
		if slices.Contains(skipAuthPaths, url) {
			next.ServeHTTP(w, r)
			return
		}
		cookieKey := viper.GetString("zitadel_configs.cookie_name")
		cookie, err := r.Cookie(cookieKey)
		if err != nil || cookie.Value == "" {
			w.WriteHeader(http.StatusUnauthorized)
			bodyByte, err := json.Marshal(map[string]string{
				"error": "Not found cookie",
			})
			if err != nil {
				log.Fatal("fail to marshal err response")
			}
			w.Write(bodyByte)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RateLimitMiddleware(rateLimiter *ratelimiter.RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := r.RemoteAddr
			uri := r.URL.Path
			key := clientIP + uri
			isAllow, err := rateLimiter.IsAllow(key)
			if err != nil || !isAllow {

				// return 429 code
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]string{
					"error": err.Error(),
				})
				return
			}
			limit := rateLimiter.GetLimit()
			currentNumRequest := rateLimiter.GetCurrentNumberRequest()
			w.Header().Set("X-Request-Remaining", fmt.Sprintf("%d", limit-currentNumRequest))
			next.ServeHTTP(w, r)
		})
	}
}

func MiddlewareChain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

func MetricMiddleware(registry *prometheus.Registry) func(http.Handler) http.Handler {
	// 1. Define metrics with "path" label in addition to "method" and "code"
	requestTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path", "method", "code"},
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Tracks the latencies for HTTP requests.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method", "code"},
	)

	requestSize := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_request_size_bytes",
			Help: "Size of HTTP requests",
		},
		[]string{"path", "method", "code"},
	)

	responseSize := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_response_size_bytes",
			Help: "Size of HTTP responses",
		},
		[]string{"path", "method", "code"},
	)

	// 2. Register metrics (ignore error if already registered, or handle it)
	// Using Register instead of MustRegister to avoid panic if this function is called multiple times (though it shouldn't be)
	// Ideally, these should be package-level variables or registered in init(), but here we use the passed registry.
	if err := registry.Register(requestTotal); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			requestTotal = are.ExistingCollector.(*prometheus.CounterVec)
		} else {
			log.Printf("Failed to register requestTotal: %v", err)
		}
	}
	if err := registry.Register(requestDuration); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			requestDuration = are.ExistingCollector.(*prometheus.HistogramVec)
		} else {
			log.Printf("Failed to register requestDuration: %v", err)
		}
	}
	if err := registry.Register(requestSize); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			requestSize = are.ExistingCollector.(*prometheus.SummaryVec)
		} else {
			log.Printf("Failed to register requestSize: %v", err)
		}
	}
	if err := registry.Register(responseSize); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			responseSize = are.ExistingCollector.(*prometheus.SummaryVec)
		} else {
			log.Printf("Failed to register responseSize: %v", err)
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			// 3. Curry the metrics with the path label
			// MustCurryWith returns a new Vector with the label fixed.
			// The remaining labels should be "method" and "code" which promhttp expects.
			curriedReqTotal := requestTotal.MustCurryWith(prometheus.Labels{"path": path})
			curriedReqDuration := requestDuration.MustCurryWith(prometheus.Labels{"path": path})
			curriedReqSize := requestSize.MustCurryWith(prometheus.Labels{"path": path})
			curriedResSize := responseSize.MustCurryWith(prometheus.Labels{"path": path})

			// 4. Wrap the handler with promhttp instrumenters
			handler := promhttp.InstrumentHandlerCounter(curriedReqTotal,
				promhttp.InstrumentHandlerDuration(curriedReqDuration,
					promhttp.InstrumentHandlerRequestSize(curriedReqSize,
						promhttp.InstrumentHandlerResponseSize(curriedResSize,
							next,
						),
					),
				),
			)
			handler.ServeHTTP(w, r)
		})
	}
}

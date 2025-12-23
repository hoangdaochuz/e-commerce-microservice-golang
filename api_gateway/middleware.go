package apigateway

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/logging"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/metric/httpmiddleware"
	ratelimiter "github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/rate_limiter"
	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/tracing"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logging.GetSugaredLogger().Infof("%s %s %v", r.Method, r.URL.Path, time.Since(start))
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
			_, _ = w.Write(bodyByte)
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
				_ = json.NewEncoder(w).Encode(map[string]string{
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
	// Initialize the middleware logic from pkg
	mw := httpmiddleware.NewMiddleware(nil, registry)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use r.URL.Path as the path label since user confirmed no dynamic IDs
			mw.WrapHandler(r.URL.Path, next).ServeHTTP(w, r)
		})
	}
}

func ApiGatewayTracing(ctx context.Context, tracer trace.Tracer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tracing.InjectTraceIntoHttpReq(ctx, r)
			ctx, span := tracer.Start(ctx, r.URL.Path, trace.WithAttributes(
				semconv.HTTPMethod(r.Method),
				semconv.HTTPRoute(r.URL.Path),
				semconv.HTTPScheme(r.URL.Scheme),
			))
			defer span.End()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

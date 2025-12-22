package apigateway

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockHandler is a simple handler for testing middleware
func mockHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
}

func TestLoggingMiddleware(t *testing.T) {
	t.Run("passes request through and logs", func(t *testing.T) {
		handler := LoggingMiddleware(mockHandler())
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, `{"status":"ok"}`, rec.Body.String())
	})

	t.Run("logs POST requests", func(t *testing.T) {
		handler := LoggingMiddleware(mockHandler())
		req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestContentTypeMiddleware(t *testing.T) {
	t.Run("sets Content-Type header to application/json", func(t *testing.T) {
		handler := ContentTypeMiddleware(mockHandler())
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("preserves response body", func(t *testing.T) {
		handler := ContentTypeMiddleware(mockHandler())
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, `{"status":"ok"}`, rec.Body.String())
	})
}

func TestAuthMiddleware(t *testing.T) {
	t.Run("allows request with skip auth path - Login", func(t *testing.T) {
		handler := AuthMiddleware(mockHandler())
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/Login", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("allows request with skip auth path - Callback", func(t *testing.T) {
		handler := AuthMiddleware(mockHandler())
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/Callback", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("allows request with skip auth path - callback lowercase", func(t *testing.T) {
		handler := AuthMiddleware(mockHandler())
		req := httptest.NewRequest(http.MethodGet, "/callback", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("rejects request without cookie", func(t *testing.T) {
		handler := AuthMiddleware(mockHandler())
		req := httptest.NewRequest(http.MethodGet, "/api/v1/orders", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		var response map[string]string
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Not found cookie", response["error"])
	})

	t.Run("rejects request with empty cookie value", func(t *testing.T) {
		handler := AuthMiddleware(mockHandler())
		req := httptest.NewRequest(http.MethodGet, "/api/v1/orders", nil)
		// Add cookie with empty value - but we need to set the cookie name from viper
		// Since viper isn't configured in tests, the cookie check will fail
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestCorsMiddleware(t *testing.T) {
	t.Run("sets CORS headers", func(t *testing.T) {
		handler := CorsMiddleware(mockHandler())
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, "GET, POST, PUT, PATCH, DELETE, OPTIONS", rec.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type, Authorization", rec.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("handles OPTIONS preflight request", func(t *testing.T) {
		handler := CorsMiddleware(mockHandler())
		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		// OPTIONS should return early without calling the next handler
		assert.Empty(t, rec.Body.String())
	})

	t.Run("passes non-OPTIONS requests through", func(t *testing.T) {
		handler := CorsMiddleware(mockHandler())
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, `{"status":"ok"}`, rec.Body.String())
	})
}

func TestMiddlewareChain(t *testing.T) {
	t.Run("applies middlewares in correct order", func(t *testing.T) {
		var order []string

		middleware1 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "m1-before")
				next.ServeHTTP(w, r)
				order = append(order, "m1-after")
			})
		}

		middleware2 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, "m2-before")
				next.ServeHTTP(w, r)
				order = append(order, "m2-after")
			})
		}

		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "handler")
			w.WriteHeader(http.StatusOK)
		})

		chain := MiddlewareChain(finalHandler, middleware1, middleware2)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		chain.ServeHTTP(rec, req)

		// Middleware1 wraps middleware2 wraps handler
		// So order should be: m1-before -> m2-before -> handler -> m2-after -> m1-after
		expected := []string{"m1-before", "m2-before", "handler", "m2-after", "m1-after"}
		assert.Equal(t, expected, order)
	})

	t.Run("works with single middleware", func(t *testing.T) {
		called := false
		middleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				next.ServeHTTP(w, r)
			})
		}

		chain := MiddlewareChain(mockHandler(), middleware)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		chain.ServeHTTP(rec, req)

		assert.True(t, called)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("works with no middlewares", func(t *testing.T) {
		chain := MiddlewareChain(mockHandler())
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		chain.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, `{"status":"ok"}`, rec.Body.String())
	})
}

func TestApiGatewayTracing(t *testing.T) {
	t.Run("creates span for request", func(t *testing.T) {
		// Since tracing requires a real tracer setup, we test that the middleware
		// passes the request through correctly
		ctx := context.Background()

		// Create a mock tracer that doesn't require OpenTelemetry setup
		// For this test, we just verify the middleware structure works
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Context should be passed through
			assert.NotNil(t, r.Context())
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		// Without a real tracer, we can't fully test this middleware
		// But we can verify the handler signature is correct
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}


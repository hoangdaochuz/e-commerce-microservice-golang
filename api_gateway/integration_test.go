//go:build integration

package apigateway

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAPIGateway_Integration tests the API Gateway routing and middleware chain
func TestAPIGateway_Integration(t *testing.T) {
	// Create a test router that mimics the API Gateway
	router := chi.NewRouter()

	// Apply middlewares (except those requiring external services)
	router.Use(LoggingMiddleware)
	router.Use(ContentTypeMiddleware)
	router.Use(CorsMiddleware)

	// Mock endpoints
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	router.Post("/api/v1/auth/Login", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"isSuccess":   true,
			"redirectURL": "https://auth.example.com/authorize",
		})
	})

	router.Get("/api/v1/orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"id":   id,
			"name": "Test Order",
		})
	})

	router.Post("/api/v1/orders", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		customerId, ok := body["customer_id"].(string)
		if !ok || customerId == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "customer_id is required"})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"order_id": "order_" + customerId + "_123456",
		})
	})

	server := httptest.NewServer(router)
	defer server.Close()

	client := &http.Client{}

	t.Run("health check endpoint", func(t *testing.T) {
		resp, err := client.Get(server.URL + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var body map[string]string
		json.NewDecoder(resp.Body).Decode(&body)
		assert.Equal(t, "healthy", body["status"])
	})

	t.Run("CORS headers are set", func(t *testing.T) {
		resp, err := client.Get(server.URL + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, "GET, POST, PUT, PATCH, DELETE, OPTIONS", resp.Header.Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type, Authorization", resp.Header.Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "true", resp.Header.Get("Access-Control-Allow-Credentials"))
	})

	t.Run("OPTIONS preflight request", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodOptions, server.URL+"/api/v1/orders", nil)
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("login endpoint returns redirect URL", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/v1/auth/Login", nil)
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		assert.True(t, body["isSuccess"].(bool))
		assert.Contains(t, body["redirectURL"].(string), "auth.example.com")
	})

	t.Run("get order by ID", func(t *testing.T) {
		resp, err := client.Get(server.URL + "/api/v1/orders/550e8400-e29b-41d4-a716-446655440001")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body map[string]string
		json.NewDecoder(resp.Body).Decode(&body)
		assert.Equal(t, "550e8400-e29b-41d4-a716-446655440001", body["id"])
		assert.Equal(t, "Test Order", body["name"])
	})
}

// TestAPIGateway_MiddlewareOrder verifies middleware execution order
func TestAPIGateway_MiddlewareOrder(t *testing.T) {
	var executionOrder []string

	router := chi.NewRouter()

	// Custom middlewares to track execution order
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			executionOrder = append(executionOrder, "middleware1-before")
			next.ServeHTTP(w, r)
			executionOrder = append(executionOrder, "middleware1-after")
		})
	})

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			executionOrder = append(executionOrder, "middleware2-before")
			next.ServeHTTP(w, r)
			executionOrder = append(executionOrder, "middleware2-after")
		})
	})

	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		executionOrder = append(executionOrder, "handler")
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(router)
	defer server.Close()

	resp, err := http.Get(server.URL + "/test")
	require.NoError(t, err)
	defer resp.Body.Close()

	expected := []string{
		"middleware1-before",
		"middleware2-before",
		"handler",
		"middleware2-after",
		"middleware1-after",
	}
	assert.Equal(t, expected, executionOrder)
}

// TestAPIGateway_ErrorHandling tests error response formatting
func TestAPIGateway_ErrorHandling(t *testing.T) {
	router := chi.NewRouter()
	router.Use(ContentTypeMiddleware)

	router.Get("/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
	})

	router.Get("/not-found", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "resource not found"})
	})

	server := httptest.NewServer(router)
	defer server.Close()

	t.Run("500 error response", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/error")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var body map[string]string
		json.NewDecoder(resp.Body).Decode(&body)
		assert.Equal(t, "internal server error", body["error"])
	})

	t.Run("404 error response", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/not-found")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		var body map[string]string
		json.NewDecoder(resp.Body).Decode(&body)
		assert.Equal(t, "resource not found", body["error"])
	})
}

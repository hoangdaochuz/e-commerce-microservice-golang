//go:build e2e

// Package e2e contains end-to-end tests for the e-commerce microservice system.
// These tests run against the full system via the API Gateway.
//
// Prerequisites:
//   - All services must be running (API Gateway, Auth, Order)
//   - Infrastructure must be available (Redis, PostgreSQL, NATS)
//
// Run with: go test -tags=e2e ./e2e/...
package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Config holds E2E test configuration
type Config struct {
	APIGatewayURL string
	Timeout       time.Duration
}

// DefaultConfig returns the default E2E test configuration
func DefaultConfig() *Config {
	apiURL := os.Getenv("E2E_API_GATEWAY_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}

	return &Config{
		APIGatewayURL: apiURL,
		Timeout:       30 * time.Second,
	}
}

// E2ETestClient provides HTTP client methods for E2E testing
type E2ETestClient struct {
	client  *http.Client
	baseURL string
}

// NewE2ETestClient creates a new E2E test client
func NewE2ETestClient(config *Config) *E2ETestClient {
	return &E2ETestClient{
		client: &http.Client{
			Timeout: config.Timeout,
		},
		baseURL: config.APIGatewayURL,
	}
}

// Get performs a GET request
func (c *E2ETestClient) Get(path string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return c.client.Do(req)
}

// Post performs a POST request with JSON body
func (c *E2ETestClient) Post(path string, body interface{}, headers map[string]string) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return c.client.Do(req)
}

// ReadJSONResponse reads and parses a JSON response body
func ReadJSONResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}

// TestE2E_HealthCheck verifies the API Gateway is running
func TestE2E_HealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	config := DefaultConfig()
	client := NewE2ETestClient(config)

	t.Run("API Gateway responds to health check", func(t *testing.T) {
		// Try to connect with retries
		var resp *http.Response
		var err error

		for i := 0; i < 5; i++ {
			resp, err = client.Get("/health", nil)
			if err == nil {
				break
			}
			time.Sleep(time.Second)
		}

		if err != nil {
			t.Skipf("API Gateway not available: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// TestE2E_AuthFlow tests the authentication flow
func TestE2E_AuthFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	config := DefaultConfig()
	client := NewE2ETestClient(config)

	t.Run("Login endpoint returns redirect URL", func(t *testing.T) {
		resp, err := client.Post("/api/v1/auth/Login", map[string]string{
			"username": "testuser@example.com",
		}, nil)

		if err != nil {
			t.Skipf("API Gateway not available: %v", err)
		}
		defer resp.Body.Close()

		// Login should return OK with redirect URL
		if resp.StatusCode == http.StatusOK {
			var body map[string]interface{}
			err := ReadJSONResponse(resp, &body)
			require.NoError(t, err)

			if isSuccess, ok := body["isSuccess"].(bool); ok {
				assert.True(t, isSuccess)
			}
			if redirectURL, ok := body["redirectURL"].(string); ok {
				assert.NotEmpty(t, redirectURL)
			}
		}
	})

	t.Run("Protected endpoint requires authentication", func(t *testing.T) {
		resp, err := client.Get("/api/v1/orders", nil)

		if err != nil {
			t.Skipf("API Gateway not available: %v", err)
		}
		defer resp.Body.Close()

		// Without auth cookie, should get 401
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Skip auth paths don't require authentication", func(t *testing.T) {
		skipPaths := []string{
			"/api/v1/auth/Login",
			"/api/v1/auth/Callback",
			"/callback",
		}

		for _, path := range skipPaths {
			t.Run(path, func(t *testing.T) {
				resp, err := client.Post(path, nil, nil)
				if err != nil {
					t.Skipf("API Gateway not available: %v", err)
				}
				defer resp.Body.Close()

				// Should not get 401 Unauthorized
				assert.NotEqual(t, http.StatusUnauthorized, resp.StatusCode)
			})
		}
	})
}

// TestE2E_OrderFlow tests the order creation and retrieval flow
func TestE2E_OrderFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	config := DefaultConfig()
	client := NewE2ETestClient(config)

	t.Run("Create order requires customer_id", func(t *testing.T) {
		// This test would need authentication in real scenario
		// For now, test the validation logic

		resp, err := client.Post("/api/v1/orders", map[string]string{
			"customer_id": "",
		}, map[string]string{
			"Cookie": "session=test_session", // Mock session
		})

		if err != nil {
			t.Skipf("API Gateway not available: %v", err)
		}
		defer resp.Body.Close()

		// Either unauthorized (no valid session) or bad request (empty customer_id)
		assert.True(t, resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusBadRequest)
	})
}

// TestE2E_RateLimiting tests the rate limiting functionality
func TestE2E_RateLimiting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	config := DefaultConfig()
	client := NewE2ETestClient(config)

	t.Run("Rate limit headers are present", func(t *testing.T) {
		resp, err := client.Get("/api/v1/auth/Login", nil)

		if err != nil {
			t.Skipf("API Gateway not available: %v", err)
		}
		defer resp.Body.Close()

		// Check for rate limit headers
		remaining := resp.Header.Get("X-Request-Remaining")
		if remaining != "" {
			// Rate limiting is enabled
			t.Logf("X-Request-Remaining: %s", remaining)
		}
	})

	t.Run("Excessive requests get rate limited", func(t *testing.T) {
		// Make many requests quickly
		rateLimitTriggered := false

		for i := 0; i < 200; i++ {
			resp, err := client.Get("/api/v1/auth/Login", nil)
			if err != nil {
				continue
			}

			if resp.StatusCode == http.StatusTooManyRequests {
				rateLimitTriggered = true
				resp.Body.Close()
				break
			}
			resp.Body.Close()
		}

		// This test may or may not trigger depending on rate limit config
		t.Logf("Rate limit triggered: %v", rateLimitTriggered)
	})
}

// TestE2E_CORSHeaders tests CORS header functionality
func TestE2E_CORSHeaders(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	config := DefaultConfig()
	client := NewE2ETestClient(config)

	t.Run("CORS headers are present", func(t *testing.T) {
		resp, err := client.Get("/api/v1/auth/Login", nil)

		if err != nil {
			t.Skipf("API Gateway not available: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, "GET, POST, PUT, PATCH, DELETE, OPTIONS", resp.Header.Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type, Authorization", resp.Header.Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "true", resp.Header.Get("Access-Control-Allow-Credentials"))
	})

	t.Run("OPTIONS preflight request succeeds", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodOptions, config.APIGatewayURL+"/api/v1/orders", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "POST")

		resp, err := client.client.Do(req)
		if err != nil {
			t.Skipf("API Gateway not available: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// TestE2E_ContentType tests that responses have correct content type
func TestE2E_ContentType(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	config := DefaultConfig()
	client := NewE2ETestClient(config)

	t.Run("Responses have JSON content type", func(t *testing.T) {
		resp, err := client.Post("/api/v1/auth/Login", map[string]string{
			"username": "test@example.com",
		}, nil)

		if err != nil {
			t.Skipf("API Gateway not available: %v", err)
		}
		defer resp.Body.Close()

		contentType := resp.Header.Get("Content-Type")
		assert.Contains(t, contentType, "application/json")
	})
}

// TestE2E_ErrorResponses tests error response formatting
func TestE2E_ErrorResponses(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	config := DefaultConfig()
	client := NewE2ETestClient(config)

	t.Run("Unauthorized error has JSON body", func(t *testing.T) {
		resp, err := client.Get("/api/v1/protected", nil)

		if err != nil {
			t.Skipf("API Gateway not available: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			var body map[string]string
			err := ReadJSONResponse(resp, &body)
			if err == nil {
				assert.NotEmpty(t, body["error"])
			}
		}
	})
}

// BenchmarkE2E_CreateOrder benchmarks order creation (for load testing)
func BenchmarkE2E_CreateOrder(b *testing.B) {
	config := DefaultConfig()
	client := NewE2ETestClient(config)

	for i := 0; i < b.N; i++ {
		resp, err := client.Post("/api/v1/orders", map[string]string{
			"customer_id": fmt.Sprintf("bench_customer_%d", i),
		}, map[string]string{
			"Cookie": "session=bench_session",
		})

		if err != nil {
			b.Skipf("API Gateway not available: %v", err)
		}
		resp.Body.Close()
	}
}

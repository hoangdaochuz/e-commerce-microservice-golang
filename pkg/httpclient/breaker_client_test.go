package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/circuitbreaker"
	"github.com/sony/gobreaker/v2"
)

func TestBreakerHTTPClient_SuccessfulRequest(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	}))
	defer server.Close()

	// Create breaker client
	config := DefaultConfig()
	breakerConfig := &circuitbreaker.Config{
		Name:             "test-breaker",
		MaxRequests:      3,
		Interval:         10,
		Timeout:          5,
		FailureThreshold: 3,
	}
	breaker := circuitbreaker.NewBreaker[*HTTPResponse](breakerConfig)

	client := NewBreakerHTTPClient(config, breaker)

	// Make request
	ctx := context.Background()
	resp, err := client.Get(ctx, server.URL, nil)

	// Assertions
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", resp.StatusCode)
	}

	if string(resp.Body) != `{"status":"success"}` {
		t.Errorf("Unexpected body: %s", string(resp.Body))
	}

	// Check circuit is closed
	if client.IsCircuitOpen() {
		t.Error("Circuit should be closed after successful request")
	}
}

func TestBreakerHTTPClient_CircuitOpensAfterFailures(t *testing.T) {
	// Create test server that always returns 500
	failureCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		failureCount++
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}`))
	}))
	defer server.Close()

	// Create breaker client with low threshold
	config := DefaultConfig()
	breakerConfig := &circuitbreaker.Config{
		Name:             "test-breaker-failures",
		MaxRequests:      1,
		Interval:         10,
		Timeout:          2,
		FailureThreshold: 2, // Open after 2 failures
	}

	breaker := circuitbreaker.NewBreaker[*HTTPResponse](breakerConfig)
	client := NewBreakerHTTPClient(config, breaker)
	ctx := context.Background()

	// Make 3 requests (should fail all)
	for i := 0; i < 3; i++ {
		_, err := client.Get(ctx, server.URL, nil)
		if err == nil {
			t.Errorf("Expected error on request %d", i+1)
		}

		// Small delay between requests
		time.Sleep(100 * time.Millisecond)
	}

	// Circuit should be open now
	if !client.IsCircuitOpen() {
		t.Error("Circuit should be open after threshold failures")
	}

	// Next request should fail immediately without hitting server
	initialFailureCount := failureCount
	_, err := client.Get(ctx, server.URL, nil)

	if err == nil {
		t.Error("Expected error when circuit is open")
	}

	if failureCount != initialFailureCount {
		t.Error("Request should not reach server when circuit is open")
	}
}

func TestBreakerHTTPClient_CircuitRecovery(t *testing.T) {
	requestCount := 0

	// Create test server that fails initially, then succeeds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount <= 3 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"recovered"}`))
		}
	}))
	defer server.Close()

	// Create breaker client
	config := DefaultConfig()
	breakerConfig := &circuitbreaker.Config{
		Name:             "test-breaker-recovery",
		MaxRequests:      1,
		Interval:         10,
		Timeout:          1, // Short timeout for testing
		FailureThreshold: 2,
	}
	breaker := circuitbreaker.NewBreaker[*HTTPResponse](breakerConfig)
	client := NewBreakerHTTPClient(config, breaker)
	ctx := context.Background()

	// Trigger failures to open circuit
	for i := 0; i < 3; i++ {
		client.Get(ctx, server.URL, nil)
		time.Sleep(100 * time.Millisecond)
	}

	// Verify circuit is open
	if !client.IsCircuitOpen() {
		t.Error("Circuit should be open")
	}

	// Wait for timeout (circuit moves to half-open)
	time.Sleep(2 * time.Second)

	// Make successful request (should close circuit)
	resp, err := client.Get(ctx, server.URL, nil)
	if err != nil {
		t.Fatalf("Expected successful request after recovery, got error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", resp.StatusCode)
	}

	// Give some time for state transition
	time.Sleep(500 * time.Millisecond)

	// Circuit should be closed or half-open (recovering)
	state := client.GetBreakerState()
	if state == gobreaker.StateOpen.String() {
		t.Error("Circuit should not be open after successful recovery")
	}
}

func TestBreakerHTTPClient_WithFallback(t *testing.T) {
	// Create test server that returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Create breaker client
	config := DefaultConfig()
	breakerConfig := &circuitbreaker.Config{
		Name:             "test-breaker-fallback",
		MaxRequests:      1,
		Interval:         10,
		Timeout:          2,
		FailureThreshold: 1,
	}

	breaker := circuitbreaker.NewBreaker[*HTTPResponse](breakerConfig)

	client := NewBreakerHTTPClient(config, breaker)
	ctx := context.Background()

	// Make request with fallback
	resp, err := client.GetWithFallback(
		ctx,
		server.URL,
		nil,
		func() (*HTTPResponse, error) {
			// Fallback returns cached data
			return &HTTPResponse{
				StatusCode: http.StatusOK,
				Body:       []byte(`{"status":"fallback"}`),
				Header:     http.Header{},
			}, nil
		},
	)

	if err != nil {
		t.Fatalf("Expected no error with fallback, got: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected fallback status 200, got: %d", resp.StatusCode)
	}

	if string(resp.Body) != `{"status":"fallback"}` {
		t.Errorf("Expected fallback body, got: %s", string(resp.Body))
	}
}

func TestBreakerHTTPClient_Metrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := DefaultConfig()
	breakerConfig := &circuitbreaker.Config{
		Name:             "test-breaker-metrics",
		MaxRequests:      3,
		Interval:         10,
		Timeout:          5,
		FailureThreshold: 3,
	}

	breaker := circuitbreaker.NewBreaker[*HTTPResponse](breakerConfig)

	client := NewBreakerHTTPClient(config, breaker)
	ctx := context.Background()

	// Make 3 successful requests
	for i := 0; i < 3; i++ {
		_, err := client.Get(ctx, server.URL, nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
	}

	// Check metrics
	metrics := client.GetMetrics()

	if metrics["name"] != "test-breaker-metrics" {
		t.Errorf("Expected name 'test-breaker-metrics', got: %v", metrics["name"])
	}

	totalRequests := metrics["total_requests"].(int)
	if totalRequests != 3 {
		t.Errorf("Expected 3 total requests, got: %d", totalRequests)
	}

	successRequests := metrics["success_requests"].(int)
	if successRequests != 3 {
		t.Errorf("Expected 3 success requests, got: %d", successRequests)
	}

	failedRequests := metrics["failed_requests"].(int)
	if failedRequests != 0 {
		t.Errorf("Expected 0 failed requests, got: %d", failedRequests)
	}
}

func TestBreakerHTTPClient_PostRequest(t *testing.T) {
	receivedBody := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got: %s", r.Method)
		}

		// Read body
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		receivedBody = string(body)

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"123"}`))
	}))
	defer server.Close()

	config := DefaultConfig()
	breakerConfig := &circuitbreaker.Config{
		Name:             "test-breaker-post",
		MaxRequests:      3,
		Interval:         10,
		Timeout:          5,
		FailureThreshold: 3,
	}
	breaker := circuitbreaker.NewBreaker[*HTTPResponse](breakerConfig)
	client := NewBreakerHTTPClient(config, breaker)
	ctx := context.Background()

	// Make POST request
	requestBody := []byte(`{"name":"test"}`)
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	resp, err := client.Post(ctx, server.URL, requestBody, headers)

	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got: %d", resp.StatusCode)
	}

	if receivedBody != `{"name":"test"}` {
		t.Errorf("Server received unexpected body: %s", receivedBody)
	}

	if string(resp.Body) != `{"id":"123"}` {
		t.Errorf("Unexpected response body: %s", string(resp.Body))
	}
}

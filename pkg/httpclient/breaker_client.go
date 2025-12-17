package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/pkg/circuitbreaker"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// HTTPResponse wraps http.Response with body content
type HTTPResponse struct {
	StatusCode int
	Body       []byte
	Header     http.Header
}

// BreakerHTTPClient wraps http.Client with circuit breaker protection
type BreakerHTTPClient struct {
	client  *http.Client
	breaker *circuitbreaker.Breaker[*HTTPResponse]
	config  *Config
}

// NewBreakerHTTPClient creates a new circuit breaker protected HTTP client
func NewBreakerHTTPClient(config *Config, breaker *circuitbreaker.Breaker[*HTTPResponse]) *BreakerHTTPClient {
	httpClient := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
		Transport: otelhttp.NewTransport(&http.Transport{
			MaxIdleConns:        config.MaxIdleConns,
			MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
			IdleConnTimeout:     time.Duration(config.IdleConnTimeout) * time.Second,
		}, otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		})),
	}

	return &BreakerHTTPClient{
		client:  httpClient,
		breaker: breaker,
		config:  config,
	}
}

// Do executes HTTP request with circuit breaker protection
func (c *BreakerHTTPClient) Do(ctx context.Context, req *http.Request) (*HTTPResponse, error) {
	result, err := c.breaker.Do(ctx, func() (*HTTPResponse, error) {
		return c.doRequest(ctx, req)
	})

	if err != nil {
		return nil, fmt.Errorf("circuit breaker error for %s %s: %w", req.Method, req.URL.String(), err)
	}

	return *result, nil
}

// DoWithFallback executes HTTP request with fallback mechanism
func (c *BreakerHTTPClient) DoWithFallback(
	ctx context.Context,
	req *http.Request,
	fallback func() (*HTTPResponse, error),
) (*HTTPResponse, error) {
	result, err := c.breaker.DoWithCallback(
		ctx,
		func() (*HTTPResponse, error) {
			return c.doRequest(ctx, req)
		},
		fallback,
	)

	if err != nil {
		return nil, fmt.Errorf("circuit breaker error with fallback for %s %s: %w", req.Method, req.URL.String(), err)
	}

	return *result, nil
}

// Get performs GET request with circuit breaker protection
func (c *BreakerHTTPClient) Get(ctx context.Context, url string, headers map[string]string) (*HTTPResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GET request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return c.Do(ctx, req)
}

// Post performs POST request with circuit breaker protection
func (c *BreakerHTTPClient) Post(ctx context.Context, url string, body []byte, headers map[string]string) (*HTTPResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create POST request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return c.Do(ctx, req)
}

// GetWithFallback performs GET request with fallback
func (c *BreakerHTTPClient) GetWithFallback(
	ctx context.Context,
	url string,
	headers map[string]string,
	fallback func() (*HTTPResponse, error),
) (*HTTPResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GET request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return c.DoWithFallback(ctx, req, fallback)
}

// PostWithFallback performs POST request with fallback
func (c *BreakerHTTPClient) PostWithFallback(
	ctx context.Context,
	url string,
	body []byte,
	headers map[string]string,
	fallback func() (*HTTPResponse, error),
) (*HTTPResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create POST request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return c.DoWithFallback(ctx, req, fallback)
}

// doRequest executes the actual HTTP request
func (c *BreakerHTTPClient) doRequest(ctx context.Context, req *http.Request) (*HTTPResponse, error) {
	req = req.WithContext(ctx)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Consider 5xx status codes as failures to trip the circuit breaker
	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("server error: status code %d", resp.StatusCode)
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       body,
		Header:     resp.Header,
	}, nil
}

// GetBreakerState returns current circuit breaker state
func (c *BreakerHTTPClient) GetBreakerState() string {
	return c.breaker.GetCurrentState().String()
}

// IsCircuitOpen checks if circuit breaker is open
func (c *BreakerHTTPClient) IsCircuitOpen() bool {
	return c.breaker.IsOpen()
}

// GetMetrics returns circuit breaker metrics
func (c *BreakerHTTPClient) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"name":             c.breaker.GetName(),
		"state":            c.GetBreakerState(),
		"total_requests":   c.breaker.GetCount(),
		"success_requests": c.breaker.GetCountSuccessRequest(),
		"failed_requests":  c.breaker.GetCountFailureRequest(),
	}
}

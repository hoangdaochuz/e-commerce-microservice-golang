// Package testutil provides shared testing utilities for the e-commerce microservice project
package testutil

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hoangdaochuz/ecommerce-microservice-golang/shared"
	"github.com/stretchr/testify/require"
)

// TestContext creates a context with a timeout for tests
func TestContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// TestContextWithHTTPRequest creates a context with an HTTP request attached (for handlers that need it)
func TestContextWithHTTPRequest(t *testing.T, r *http.Request) context.Context {
	ctx := TestContext(t)
	return context.WithValue(ctx, shared.HTTPRequest_ContextKey, r)
}

// NewTestRequest creates a new HTTP request for testing
func NewTestRequest(t *testing.T, method, path string, body interface{}) *http.Request {
	var req *http.Request
	var err error

	if body != nil {
		bodyBytes, err := json.Marshal(body)
		require.NoError(t, err)
		req = httptest.NewRequest(method, path, nil)
		req.Body = NewReadCloser(bodyBytes)
		req.GetBody = func() (r io.ReadCloser, err error) {
			return NewReadCloser(bodyBytes), nil
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, path, nil)
		require.NoError(t, err)
	}

	return req
}

// ReadCloser is a helper for creating an io.ReadCloser from bytes
type ReadCloser struct {
	data   []byte
	offset int
}

// NewReadCloser creates a new ReadCloser from bytes
func NewReadCloser(data []byte) *ReadCloser {
	return &ReadCloser{data: data}
}

// Read implements io.Reader
func (r *ReadCloser) Read(p []byte) (n int, err error) {
	if r.offset >= len(r.data) {
		return 0, nil
	}
	n = copy(p, r.data[r.offset:])
	r.offset += n
	return n, nil
}

// Close implements io.Closer
func (r *ReadCloser) Close() error {
	return nil
}

// AssertJSONEqual compares two JSON strings for equality
func AssertJSONEqual(t *testing.T, expected, actual string) {
	var expectedJSON, actualJSON interface{}
	require.NoError(t, json.Unmarshal([]byte(expected), &expectedJSON))
	require.NoError(t, json.Unmarshal([]byte(actual), &actualJSON))
	require.Equal(t, expectedJSON, actualJSON)
}

// ResponseRecorder is a wrapper around httptest.ResponseRecorder with additional helpers
type ResponseRecorder struct {
	*httptest.ResponseRecorder
}

// NewResponseRecorder creates a new ResponseRecorder
func NewResponseRecorder() *ResponseRecorder {
	return &ResponseRecorder{
		ResponseRecorder: httptest.NewRecorder(),
	}
}

// JSONBody returns the response body as parsed JSON
func (r *ResponseRecorder) JSONBody(t *testing.T, target interface{}) {
	require.NoError(t, json.Unmarshal(r.Body.Bytes(), target))
}

// AssertStatus asserts the response status code
func (r *ResponseRecorder) AssertStatus(t *testing.T, expected int) {
	require.Equal(t, expected, r.Code, "unexpected status code")
}

// AssertHeader asserts a response header value
func (r *ResponseRecorder) AssertHeader(t *testing.T, key, expected string) {
	require.Equal(t, expected, r.Header().Get(key), "unexpected header value for %s", key)
}

// TestServer wraps httptest.Server with additional helpers
type TestServer struct {
	*httptest.Server
	t *testing.T
}

// NewTestServer creates a new test HTTP server
func NewTestServer(t *testing.T, handler http.Handler) *TestServer {
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return &TestServer{Server: server, t: t}
}

// URL returns the test server URL
func (s *TestServer) URL() string {
	return s.Server.URL
}

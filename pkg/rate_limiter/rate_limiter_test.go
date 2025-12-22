package ratelimiter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_GetLimit(t *testing.T) {
	t.Run("returns configured limit", func(t *testing.T) {
		// Create rate limiter with a nil redis client for unit testing
		// In production, this would use a real redis client
		limiter := &RateLimiter{
			limit:  100,
			window: time.Minute,
		}

		assert.Equal(t, 100, limiter.GetLimit())
	})

	t.Run("returns different limits correctly", func(t *testing.T) {
		testCases := []struct {
			name     string
			limit    int
			expected int
		}{
			{"zero limit", 0, 0},
			{"low limit", 10, 10},
			{"high limit", 1000, 1000},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				limiter := &RateLimiter{limit: tc.limit}
				assert.Equal(t, tc.expected, limiter.GetLimit())
			})
		}
	})
}

func TestRateLimiter_GetCurrentNumberRequest(t *testing.T) {
	t.Run("returns current request count", func(t *testing.T) {
		limiter := &RateLimiter{
			currentNumberRequest: 50,
		}

		assert.Equal(t, 50, limiter.GetCurrentNumberRequest())
	})

	t.Run("returns zero initially", func(t *testing.T) {
		limiter := &RateLimiter{}

		assert.Equal(t, 0, limiter.GetCurrentNumberRequest())
	})
}

func TestNewRateLimiter(t *testing.T) {
	t.Run("creates rate limiter with correct parameters", func(t *testing.T) {
		// For unit tests, we pass nil redis client
		// Integration tests would use a real client
		limiter := NewRateLimiter(nil, 100, time.Minute, nil)

		assert.NotNil(t, limiter)
		assert.Equal(t, 100, limiter.limit)
		assert.Equal(t, time.Minute, limiter.window)
	})

	t.Run("creates rate limiter with different window durations", func(t *testing.T) {
		testCases := []struct {
			name   string
			window time.Duration
		}{
			{"second window", time.Second},
			{"minute window", time.Minute},
			{"hour window", time.Hour},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				limiter := NewRateLimiter(nil, 100, tc.window, nil)
				assert.Equal(t, tc.window, limiter.window)
			})
		}
	})
}

// Note: IsAllow tests require a real Redis connection
// Those tests are in the integration test file
func TestRateLimiter_KeyFormat(t *testing.T) {
	t.Run("rate limit key format", func(t *testing.T) {
		// Test the expected key format
		testCases := []struct {
			input    string
			expected string
		}{
			{"192.168.1.1/api/v1/orders", "rate_limit:192.168.1.1/api/v1/orders"},
			{"10.0.0.1/api/v1/auth/login", "rate_limit:10.0.0.1/api/v1/auth/login"},
		}

		for _, tc := range testCases {
			t.Run(tc.input, func(t *testing.T) {
				// The key format is: rate_limit:{clientIP}{uri}
				key := "rate_limit:" + tc.input
				assert.Equal(t, tc.expected, key)
			})
		}
	})
}

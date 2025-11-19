package httpclient

// Config holds HTTP client configuration
type Config struct {
	// Timeout is the timeout for HTTP requests in seconds
	Timeout int

	// MaxIdleConns controls the maximum number of idle (keep-alive) connections
	MaxIdleConns int

	// MaxIdleConnsPerHost controls the maximum idle connections per host
	MaxIdleConnsPerHost int

	// IdleConnTimeout is the maximum time an idle connection will remain idle before closing in seconds
	IdleConnTimeout int
}

// DefaultConfig returns default HTTP client configuration
func DefaultConfig() *Config {
	return &Config{
		Timeout:             30,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90,
	}
}

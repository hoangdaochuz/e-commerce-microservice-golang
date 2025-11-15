package circuitbreaker

import "github.com/hoangdaochuz/ecommerce-microservice-golang/configs"

type Config struct {
	Name                 string
	MaxRequests          int
	Interval             int
	Timeout              int
	FailureThreshold     int
	FailureRateThreshold float64
}

func ToCircuitBreakerConfig(circuitBreakerName string, config *configs.CircuitBreakerCommon) *Config {
	return &Config{
		Name:                 circuitBreakerName,
		MaxRequests:          config.MaxRequest,
		Interval:             config.Interval,
		Timeout:              config.Timeout,
		FailureThreshold:     config.FailureThreshold,
		FailureRateThreshold: config.FailureRateThreshold,
	}
}

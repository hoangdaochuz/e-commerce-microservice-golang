package circuitbreaker

type Config struct {
	Name                 string
	MaxRequests          int
	Interval             int
	Timeout              int
	FailureThreshold     int
	FailureRateThreshold float64
}

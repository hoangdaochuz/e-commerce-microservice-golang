package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

type MetricWrapper struct {
	registry *prometheus.Registry
}

func NewMetricWrapper() *MetricWrapper {
	return &MetricWrapper{
		registry: prometheus.NewRegistry(),
	}
}

func (mw *MetricWrapper) RegisterCollectorDefault() {
	mw.registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
}

func (mw *MetricWrapper) GetRegistry() *prometheus.Registry {
	return mw.registry
}

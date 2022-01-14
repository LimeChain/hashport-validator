package service

import "github.com/prometheus/client_golang/prometheus"

type Prometheus interface {
	// NewGaugeMetric creates new Gauge Metric
	NewGaugeMetric(name string, help string) prometheus.Gauge
	// RegisterGaugeMetric registering new Gauge Metric
	RegisterGaugeMetric(gauge prometheus.Gauge)
	// GetGauge retrieves Gauge by name with flag for existence
	GetGauge(name string) prometheus.Gauge
}

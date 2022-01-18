package service

import "github.com/prometheus/client_golang/prometheus"

type Prometheus interface {
	// CreateAndRegisterGaugeMetric creates new Gauge Metric and registers it in Prometheus
	CreateAndRegisterGaugeMetric(name string, help string) prometheus.Gauge
	// GetGauge retrieves Gauge by name with flag for existence
	GetGauge(name string) prometheus.Gauge
}

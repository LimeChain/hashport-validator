package service

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Prometheus interface {
	// CreateAndRegisterGaugeMetric creates new Gauge Metric and registers it in Prometheus
	CreateAndRegisterGaugeMetric(name string, help string) prometheus.Gauge
	// CreateAndRegisterGaugeMetricForSuccessRate creates new Gauge Metric for Success Rate and registers it in Prometheus
	CreateAndRegisterGaugeMetricForSuccessRate(transactionId string, sourceChainId int64, targetChainId int64, asset, metricNameSuffix, metricHelp string) (prometheus.Gauge, error)
	// GetGauge retrieves Gauge by name with flag for existence
	GetGauge(name string) prometheus.Gauge
	// UnregisterGauge unregisters Gauge with the passed name
	UnregisterGauge(name string)
	// CreateAndRegisterCounterMetric creates new Counter Metric and registers it in Prometheus
	CreateAndRegisterCounterMetric(name string, help string) prometheus.Counter
	// GetCounter retrieves Counter by name with flag for existence
	GetCounter(name string) prometheus.Counter
	// UnregisterCounter unregisters Counter with the passed name
	UnregisterCounter(name string)
	// ConstructNameForSuccessRateMetric constructing name for success rate metric
	ConstructNameForSuccessRateMetric(sourceNetworkId, targetNetworkId uint64, asset, transactionId, metricTarget string) (string, error)
}

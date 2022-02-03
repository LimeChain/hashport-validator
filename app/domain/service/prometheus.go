package service

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Prometheus interface {
	// CreateAndRegisterGaugeMetricIfNotExists creates new Gauge Metric and registers it in Prometheus if not exists
	CreateAndRegisterGaugeMetricIfNotExists(name string, help string, labels prometheus.Labels) prometheus.Gauge
	// CreateAndRegisterSuccessRateGaugeMetricIfNotExists creates new Gauge Metric for Success Rate and registers it in Prometheus if not exists
	CreateAndRegisterSuccessRateGaugeMetricIfNotExists(transactionId string, sourceChainId int64, targetChainId int64, asset, metricNameSuffix, metricHelp string) (prometheus.Gauge, error)
	// GetGauge retrieves Gauge by name with flag for existence
	GetGauge(name string) prometheus.Gauge
	// UnregisterAndDeleteGauge unregisters and deletes Gauge with the passed name
	UnregisterAndDeleteGauge(name string)
	// CreateAndRegisterCounterMetricIfNotExists creates new Counter Metric and registers it in Prometheus
	CreateAndRegisterCounterMetricIfNotExists(name string, help string, labels prometheus.Labels) prometheus.Counter
	// GetCounter retrieves Counter by name with flag for existence
	GetCounter(name string) prometheus.Counter
	// UnregisterAndDeleteCounter unregisters and deletes Counter with the passed name
	UnregisterAndDeleteCounter(name string)
	// ConstructMetricName constructing name for metric
	ConstructMetricName(sourceNetworkId, targetNetworkId uint64, asset, transactionId, metricTarget string) (string, error)
	// GetIsMonitoringEnabled returns if the monitoring is enabled
	GetIsMonitoringEnabled() bool
}

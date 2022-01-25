package service

import (
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/prometheus/client_golang/prometheus"
)

type Prometheus interface {
	// CreateAndRegisterGaugeMetric creates new Gauge Metric and registers it in Prometheus
	CreateAndRegisterGaugeMetric(name string, help string) prometheus.Gauge
	// GetGauge retrieves Gauge by name with flag for existence
	GetGauge(name string) prometheus.Gauge
	// CreateAndRegisterCounterMetric creates new Counter Metric and registers it in Prometheus
	CreateAndRegisterCounterMetric(name string, help string) prometheus.Counter
	// GetCounter retrieves Counter by name with flag for existence
	GetCounter(name string) prometheus.Counter
	// ConstructNameForSuccessRateMetric constructing name for success rate metric
	ConstructNameForSuccessRateMetric(sourceNetworkId, targetNetworkId uint64, asset, transactionId, metricTarget string) (string, error)
	// IsNative Returns true if the asset is native to the passed network
	IsNative(networkId int64, asset string) bool
	// NativeToWrapped Getting Wrapped token from Native token for the given network/chain ids
	NativeToWrapped(nativeAsset string, nativeChainId, targetChainId int64) string
	// WrappedToNative Getting Native token from Wrapped token for the given network/chain ids
	WrappedToNative(wrappedAsset string, wrappedChainId int64) *config.NativeAsset
}

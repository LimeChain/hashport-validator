package prometheus

import (
	"fmt"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	testConstants "github.com/limechain/hedera-eth-bridge-validator/test/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	service             *Service
	gauge               prometheus.Gauge
	counter             prometheus.Counter
	isMonitoringEnabled = true
	gaugeName           = "GaugeName"
	gaugeSuffix         = "gauge_suffix"
	gaugeHelp           = "GaugeHelp"
	counterName         = "CounterName"
	counterSuffix       = "counter_suffix"
	counterHelp         = "CounterHelp"
	assets              = config.LoadAssets(testConstants.Networks)
)

func Test_New(t *testing.T) {
	setup()

	actualService := NewService(assets, isMonitoringEnabled)

	assert.Equal(t, service, actualService)
}

func Test_CreateAndRegisterGaugeMetric(t *testing.T) {
	setup()

	gauge = service.CreateAndRegisterGaugeMetric(gaugeName, gaugeHelp)
	defer service.UnregisterGauge(gaugeName)

	assert.NotNil(t, gauge)
}

func Test_ConstructMetricName_Native(t *testing.T) {
	setup()

	expectedNative := fmt.Sprintf("%v_%v_to_%v_%v_%v", constants.Native, constants.Hedera, constants.Ethereum, "0_0_1234_1234_1234", constants.MajorityReachedNameSuffix)
	actual, err := service.ConstructMetricName(0, 3, constants.Hbar, "0.0.1234-1234-1234", constants.MajorityReachedNameSuffix)

	assert.Equal(t, err, nil)
	assert.Equal(t, expectedNative, actual)
}

func Test_ConstructMetricName_Wrapped(t *testing.T) {
	setup()

	expectedNative := fmt.Sprintf("%v_%v_to_%v_%v_%v", constants.Wrapped, constants.Ethereum, constants.Hedera, "0_0_1234_1234_1234", constants.MajorityReachedNameSuffix)
	actual, err := service.ConstructMetricName(3, 0, constants.Hbar, "0.0.1234-1234-1234", constants.MajorityReachedNameSuffix)

	assert.Equal(t, err, nil)
	assert.Equal(t, expectedNative, actual)
}

func Test_ConstructMetricName_ShouldThrow(t *testing.T) {
	setup()

	_, err := service.ConstructMetricName(10, 0, constants.Hbar, "0.0.1234-1234-1234", constants.MajorityReachedNameSuffix)

	expectedError := fmt.Sprintf("Network id %v is missing in id to name mapping.", 10)
	assert.Errorf(t, err, expectedError)
}

func Test_CreateAndRegisterGaugeMetricForSuccessRate(t *testing.T) {
	setup()

	transactionId := "0.0.1234"
	sourceChainId := int64(constants.HederaNetworkId)
	targetChainId := int64(constants.EthereumChainId)
	asset := constants.Hbar
	gauge, err := service.CreateAndRegisterGaugeMetricForSuccessRate(
		transactionId,
		sourceChainId,
		targetChainId,
		asset,
		gaugeSuffix,
		gaugeHelp)

	fullGaugeName, err2 := service.ConstructMetricName(uint64(sourceChainId), uint64(targetChainId), asset, transactionId, gaugeSuffix)

	defer service.UnregisterGauge(fullGaugeName)

	assert.NotNil(t, gauge)
	assert.Nil(t, err)
	assert.Nil(t, err2)
}

func Test_GetGauge(t *testing.T) {
	setup()

	gauge = service.CreateAndRegisterGaugeMetric(gaugeName, gaugeHelp)
	defer service.UnregisterGauge(gaugeName)
	gaugeInMapping := service.GetGauge(gaugeName)

	assert.NotNil(t, gaugeInMapping)
}

func Test_UnregisterGauge(t *testing.T) {
	setup()

	gauge = service.CreateAndRegisterGaugeMetric(gaugeName, gaugeHelp)
	service.UnregisterGauge(gaugeName)

	gaugeInMapping := service.GetGauge(gaugeName)

	assert.Nil(t, gaugeInMapping)
}

func Test_CreateAndRegisterCounterMetric(t *testing.T) {
	setup()

	counter = service.CreateAndRegisterCounterMetric(counterName, counterHelp)
	defer service.UnregisterCounter(counterName)

	assert.NotNil(t, counter)
}

func Test_GetCounter(t *testing.T) {
	setup()

	counter = service.CreateAndRegisterCounterMetric(counterName, counterHelp)
	defer service.UnregisterCounter(counterName)
	counterInMapping := service.GetCounter(counterName)

	assert.NotNil(t, counterInMapping)
}

func Test_UnregisterCounter(t *testing.T) {
	setup()

	counter = service.CreateAndRegisterCounterMetric(counterName, counterHelp)
	service.UnregisterCounter(counterName)

	counterInMapping := service.GetCounter(counterName)

	assert.Nil(t, counterInMapping)
}

func setup() {
	mocks.Setup()

	service = &Service{
		logger:              config.GetLoggerFor("Prometheus Service"),
		gauges:              map[string]prometheus.Gauge{},
		counters:            map[string]prometheus.Counter{},
		assetsConfig:        assets,
		isMonitoringEnabled: isMonitoringEnabled,
	}
}

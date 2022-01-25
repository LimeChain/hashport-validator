package prometheus

import (
	"fmt"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	service     *Service
	gauge       prometheus.Gauge
	counter     prometheus.Counter
	gaugeName   = "GaugeName"
	gaugeHelp   = "GaugeHelp"
	counterName = "CounterName"
	counterHelp = "CounterHelp"
	networks    = map[int64]*parser.Network{
		0: {
			Tokens: map[string]parser.Token{
				constants.Hbar: {
					Networks: map[int64]string{
						33: "0x0000000000000000000000000000000000000001",
					},
				},
			},
		},
		1: {
			Tokens: map[string]parser.Token{
				"0xsomeethaddress": {
					Networks: map[int64]string{
						33: "0x0000000000000000000000000000000000000123",
					},
				},
			},
		},
		2: {
			Tokens: map[string]parser.Token{
				"0x0000000000000000000000000000000000000000": {
					Networks: map[int64]string{
						0: "",
					},
				},
			},
		},
		3: {
			Tokens: map[string]parser.Token{
				"0x0000000000000000000000000000000000000000": {
					Networks: map[int64]string{
						0: "",
					},
				},
			},
		},
		32: {
			Tokens: map[string]parser.Token{
				"0x0000000000000000000000000000000000000000": {
					Networks: map[int64]string{
						0: "",
					},
				},
			},
		},
		33: {
			Tokens: map[string]parser.Token{
				"0x0000000000000000000000000000000000000000": {
					Networks: map[int64]string{
						0: constants.Hbar,
						1: "0xsome-other-eth-address",
					},
				},
			}},
	}
	assets = config.LoadAssets(networks)
)

func Test_New(t *testing.T) {
	setup()

	actualService := NewService(assets)

	assert.Equal(t, service, actualService)
}

func Test_CreateAndRegisterGaugeMetric(t *testing.T) {
	setup()

	gauge = service.CreateAndRegisterGaugeMetric(gaugeName, gaugeHelp)
	defer prometheus.Unregister(gauge)

	assert.NotNil(t, gauge)
}

func Test_GetGauge(t *testing.T) {
	setup()

	gauge = service.CreateAndRegisterGaugeMetric(gaugeName, gaugeHelp)
	defer prometheus.Unregister(gauge)
	gaugeInMapping := service.GetGauge(gaugeName)

	assert.NotNil(t, gaugeInMapping)
}

func Test_CreateAndRegisterCounterMetric(t *testing.T) {
	setup()

	counter = service.CreateAndRegisterCounterMetric(counterName, counterHelp)
	defer prometheus.Unregister(counter)

	assert.NotNil(t, gauge)
}

func Test_GetCounter(t *testing.T) {
	setup()

	counter = service.CreateAndRegisterCounterMetric(counterName, counterHelp)
	defer prometheus.Unregister(counter)
	counterInMapping := service.GetCounter(counterName)

	assert.NotNil(t, counterInMapping)
}

func Test_ConstructNameForSuccessRateMetric_Native(t *testing.T) {
	setup()

	expectedNative := fmt.Sprintf("%v_%v_to_%v_%v_%v", constants.Native, constants.Hedera, constants.Ethereum, "0_0_1234_1234_1234", constants.MajorityReachedNameSuffix)
	actual, err := service.ConstructNameForSuccessRateMetric(0, 3, constants.Hbar, "0.0.1234-1234-1234", constants.MajorityReachedNameSuffix)

	assert.Equal(t, err, nil)
	assert.Equal(t, expectedNative, actual)
}

func Test_ConstructNameForSuccessRateMetric_Wrapped(t *testing.T) {
	setup()

	expectedNative := fmt.Sprintf("%v_%v_to_%v_%v_%v", constants.Wrapped, constants.Ethereum, constants.Hedera, "0_0_1234_1234_1234", constants.MajorityReachedNameSuffix)
	actual, err := service.ConstructNameForSuccessRateMetric(3, 0, constants.Hbar, "0.0.1234-1234-1234", constants.MajorityReachedNameSuffix)

	assert.Equal(t, err, nil)
	assert.Equal(t, expectedNative, actual)
}

func Test_ConstructNameForSuccessRateMetric_ShouldThrow(t *testing.T) {
	setup()

	_, err := service.ConstructNameForSuccessRateMetric(10, 0, constants.Hbar, "0.0.1234-1234-1234", constants.MajorityReachedNameSuffix)

	expectedError := fmt.Sprintf("Network id %v is missing in id to name mapping.", 10)
	assert.Errorf(t, err, expectedError)
}

func Test_NativeToWrapped(t *testing.T) {
	setup()

	actual := service.NativeToWrapped(constants.Hbar, 0, 33)
	expected := "0x0000000000000000000000000000000000000001"

	assert.Equal(t, expected, actual)
}

func Test_WrappedToNative(t *testing.T) {
	setup()

	actual := service.WrappedToNative("0x0000000000000000000000000000000000000001", 33)
	expected := constants.Hbar

	assert.NotNil(t, actual)
	assert.Equal(t, expected, actual.Asset)
}

func setup() {
	mocks.Setup()

	service = &Service{
		logger:       config.GetLoggerFor("Prometheus Service"),
		gauges:       map[string]prometheus.Gauge{},
		counters:     map[string]prometheus.Counter{},
		assetsConfig: assets,
	}
}

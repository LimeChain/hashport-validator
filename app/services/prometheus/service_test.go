package prometheus

import (
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	gauge     prometheus.Gauge
	service   *Service
	gaugeName = "GaugeName"
	gaugeHelp = "GaugeHelp"
)

func Test_New(t *testing.T) {
	setup()

	actualService := NewService()

	assert.Equal(t, service, actualService)
}

func Test_CreateAndRegisterGaugeMetric(t *testing.T) {
	setup()

	gauge = service.CreateAndRegisterGaugeMetric(gaugeName, gaugeHelp)
	defer prometheus.Unregister(gauge)
	gaugeInMapping := service.GetGauge(gaugeName)

	assert.NotNil(t, gauge)
	assert.NotNil(t, gaugeInMapping)
}

func Test_GetGauge(t *testing.T) {
	setup()

	gauge = service.CreateAndRegisterGaugeMetric(gaugeName, gaugeHelp)
	defer prometheus.Unregister(gauge)
	gaugeInMapping := service.GetGauge(gaugeName)

	assert.NotNil(t, gaugeInMapping)
}

func setup() {
	mocks.Setup()

	service = &Service{
		logger: config.GetLoggerFor("Prometheus Service"),
		gauges: map[string]prometheus.Gauge{},
	}
}

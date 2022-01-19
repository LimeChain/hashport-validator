package prometheus

import (
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	gauge           prometheus.Gauge
	service         *Service
	gaugeName       = "GaugeName"
	gaugeHelp       = "GaugeHelp"
	pollingInterval = 5 * time.Second
)

func Test_New(t *testing.T) {
	setup()

	actualService := NewService(
		mocks.MHederaMirrorClient,
		mocks.MTransferRepository,
		pollingInterval,
	)

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
		mirrorNode:         mocks.MHederaMirrorClient,
		transferRepository: mocks.MTransferRepository,
		pollingInterval:    pollingInterval,
		logger:             config.GetLoggerFor("Prometheus Service"),
		gauges:             map[string]prometheus.Gauge{},
	}
}

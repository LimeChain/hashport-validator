package calculator

import (
	"testing"

	"github.com/gookit/event"
	bridge_config_event "github.com/limechain/hedera-eth-bridge-validator/app/model/bridge-config-event"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/stretchr/testify/assert"
)

var feePercentages = map[string]int64{
	"hbar":       10000,
	"0.0.123321": 1213,
}

func Test_New(t *testing.T) {
	newService := New(feePercentages)

	expectedService := &Service{
		feePercentages: feePercentages,
		logger:         config.GetLoggerFor("Fee Service"),
	}

	assert.Equal(t, expectedService, newService)
}

func Test_CalculateFee(t *testing.T) {
	service := New(feePercentages)

	fee, remainder := service.CalculateFee("hbar", 20)

	expectedFee := int64(2)
	expectedRemainder := int64(18)

	assert.Equal(t, expectedFee, fee)
	assert.Equal(t, expectedRemainder, remainder)
}

func Test_CalculatePercentageFee(t *testing.T) {
	service := New(feePercentages)

	fee, remainder := service.CalculatePercentageFee(20, 10000)

	expectedFee := int64(2)
	expectedRemainder := int64(18)

	assert.Equal(t, expectedFee, fee)
	assert.Equal(t, expectedRemainder, remainder)
}

func Test_bridgeCfgUpdateEventHandler(t *testing.T) {
	service := New(feePercentages)

	newFeePercentages := make(map[string]int64)
	for tokenName, feeAmount := range service.feePercentages {
		newFeePercentages[tokenName] = feeAmount + 1
	}
	event.MustFire(constants.EventBridgeConfigUpdate, event.M{constants.BridgeConfigUpdateEventParamsKey: &bridge_config_event.Params{
		Bridge: &config.Bridge{
			Hedera: &config.BridgeHedera{
				FeePercentages: newFeePercentages,
			},
		},
	}})

	assert.Equal(t, newFeePercentages, service.feePercentages)
}

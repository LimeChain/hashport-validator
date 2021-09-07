package calculator

import (
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/stretchr/testify/assert"
	"testing"
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

func Test_CalculateFee_WithRemainder(t *testing.T) {
	service := New(feePercentages)

	fee, remainder := service.CalculateFee("0.0.123321", 2000)

	expectedFee := int64(2)
	expectedRemainder := int64(19)

	assert.Equal(t, expectedFee, fee)
	assert.Equal(t, expectedRemainder, remainder)
}

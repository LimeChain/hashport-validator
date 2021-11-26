package auth_message

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	sourceChainId   = int64(0)
	targetChainId   = int64(1)
	txId            = "0.0.123-123321-123321"
	asset           = "0x00000"
	receiverAddress = "0xsomeaddress"
	invalidAmount   = "invalidamount"
	amount          = "100"
)

func Test_EncodeFungibleBytesFromWithInvalidAmount(t *testing.T) {
	actualResult, err := EncodeFungibleBytesFrom(
		sourceChainId,
		targetChainId,
		txId,
		asset,
		receiverAddress,
		invalidAmount)

	assert.Error(t, err)
	assert.Nil(t, actualResult)
}

func Test_EncodeFungibleBytesFromWorks(t *testing.T) {
	actualResult, err := EncodeFungibleBytesFrom(
		sourceChainId,
		targetChainId,
		txId,
		asset,
		receiverAddress,
		amount)

	assert.Nil(t, err)
	assert.NotNil(t, actualResult)
}

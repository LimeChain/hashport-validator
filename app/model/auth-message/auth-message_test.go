package auth_message

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	txId            = "0.0.123-123321-123321"
	routerAddress   = "0xsomerouteraddress"
	wrappedAsset    = "0x00000"
	receiverAddress = "0xsomeaddress"
	invalidAmount   = "invalidamount"
	amount          = "100"
)

func Test_EncodeBytesFromWithInvalidAmount(t *testing.T) {
	actualResult, err := EncodeBytesFrom(txId,
		routerAddress,
		wrappedAsset,
		receiverAddress,
		invalidAmount)

	assert.Error(t, err)
	assert.Nil(t, actualResult)
}

func Test_EncodeBytesFromWorks(t *testing.T) {
	actualResult, err := EncodeBytesFrom(txId,
		routerAddress,
		wrappedAsset,
		receiverAddress,
		amount)

	assert.Nil(t, err)
	assert.NotNil(t, actualResult)
}

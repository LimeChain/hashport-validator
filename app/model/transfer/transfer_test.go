package transfer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	txId          = "0.0.123123-123321-420"
	receiver      = "0xreceiver"
	amount        = "100"
	nativeAsset   = "0.0.123"
	wrappedAsset  = "0xwrapped00123"
	routerAddress = "0xrouteraddress"
)

func Test_New(t *testing.T) {
	expectedTransfer := &Transfer{
		TransactionId: txId,
		Receiver:      receiver,
		Amount:        amount,
		NativeAsset:   nativeAsset,
		WrappedAsset:  wrappedAsset,
		RouterAddress: routerAddress,
	}
	actualTransfer := New(txId,
		receiver,
		nativeAsset,
		wrappedAsset,
		amount,
		routerAddress)
	assert.Equal(t, expectedTransfer, actualTransfer)
}

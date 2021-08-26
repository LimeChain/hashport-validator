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
	chainId       = 3
)

func Test_New(t *testing.T) {
	expectedTransfer := &NativeTransfer{
		Transfer: Transfer{
			TransactionId: txId,
			Receiver:      receiver,
			Amount:        amount,
			NativeAsset:   nativeAsset,
			WrappedAsset:  wrappedAsset,
			RouterAddress: routerAddress,
			TargetChainID: chainId,
		},
	}
	actualTransfer := NewNative(txId,
		receiver,
		nativeAsset,
		wrappedAsset,
		amount,
		routerAddress,
		chainId)
	assert.Equal(t, expectedTransfer, actualTransfer)
}

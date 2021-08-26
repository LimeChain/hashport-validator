package transfer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	txId          = "0.0.123123-123321-420"
	sourceChainId = uint64(0)
	targetChainId = uint64(1)
	receiver      = "0xreceiver"
	amount        = "100"
	nativeAsset   = "0.0.123"
	wrappedAsset  = "0xwrapped00123"
	routerAddress = "0xrouteraddress"
)

func Test_New(t *testing.T) {
	expectedTransfer := &Transfer{
		TransactionId: txId,
		SourceChainId: sourceChainId,
		TargetChainId: targetChainId,
		Receiver:      receiver,
		Amount:        amount,
		NativeAsset:   nativeAsset,
		WrappedAsset:  wrappedAsset,
		RouterAddress: routerAddress,
	}
	actualTransfer := New(txId,
		sourceChainId,
		targetChainId,
		receiver,
		nativeAsset,
		wrappedAsset,
		amount,
		routerAddress)
	assert.Equal(t, expectedTransfer, actualTransfer)
}

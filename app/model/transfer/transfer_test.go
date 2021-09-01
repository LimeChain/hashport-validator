package transfer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	txId          = "0.0.123123-123321-420"
	sourceChainId = int64(0)
	targetChainId = int64(1)
	nativeChainId = int64(0)
	receiver      = "0xreceiver"
	amount        = "100"
	sourceAsset   = "0.0.123"
	nativeAsset   = "0.0.123"
	targetAsset   = "0xwrapped00123"
	routerAddress = "0xrouteraddress"
	chainId       = 3
)

func Test_New(t *testing.T) {
	expectedTransfer := &Transfer{
		TransactionId: txId,
		SourceChainId: sourceChainId,
		TargetChainId: targetChainId,
		Receiver:      receiver,
		Amount:        amount,
		SourceAsset:   sourceAsset,
		NativeAsset:   nativeAsset,
		TargetAsset:   targetAsset,
		RouterAddress: routerAddress,
	}
	actualTransfer := New(txId,
		sourceChainId,
		targetChainId,
		nativeChainId,
		receiver,
		nativeAsset,
		targetAsset,
		sourceAsset,
		amount,
		routerAddress,
		chainId)
	assert.Equal(t, expectedTransfer, actualTransfer)
}

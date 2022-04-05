package transfer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	txId          = "0.0.123123-123321-420"
	sourceChainId = 0
	targetChainId = 1
	nativeChainId = 0
	receiver      = "0xreceiver"
	amount        = "100"
	sourceAsset   = "0.0.123"
	nativeAsset   = "0.0.123"
	targetAsset   = "0xwrapped00123"
	serialNum     = 0
	metadata      = "42"
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
	}
	actualTransfer := New(txId,
		sourceChainId,
		targetChainId,
		nativeChainId,
		receiver,
		nativeAsset,
		targetAsset,
		sourceAsset,
		amount)
	assert.Equal(t, expectedTransfer, actualTransfer)
}

func Test_NewNft(t *testing.T) {
	expected := &Transfer{
		TransactionId: txId,
		SourceChainId: sourceChainId,
		TargetChainId: targetChainId,
		NativeChainId: nativeChainId,
		SourceAsset:   sourceAsset,
		TargetAsset:   targetAsset,
		NativeAsset:   nativeAsset,
		Receiver:      receiver,
		SerialNum:     serialNum,
		Metadata:      metadata,
		IsNft:         true,
		Amount:        "",
		Timestamp:     "",
	}

	actual := NewNft(txId,
		sourceChainId,
		targetChainId,
		nativeChainId,
		receiver,
		sourceAsset,
		targetAsset,
		nativeAsset,
		serialNum,
		metadata)
	assert.Equal(t, expected, actual)
}

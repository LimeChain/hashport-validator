package message

import (
	"encoding/base64"
	"github.com/golang/protobuf/proto"
	timestampHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	model "github.com/limechain/hedera-eth-bridge-validator/proto"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	invalidBytes            = []byte{1, 2}
	ts                int64 = 123
	stringTs                = "123.123"
	invalidStringTs         = "invalidTs"
	invalidStringData       = "1-1-1-1-1-1"
	now                     = time.Now()
)

func expectedSignature() *model.TopicEthSignatureMessage {
	return &model.TopicEthSignatureMessage{
		SourceChainId: 0,
		TargetChainId: 1,
		TransferID:    "0.0.123321-123321-420",
		Asset:         "0xasset",
		Recipient:     "0xsomereceiver",
		Amount:        "100",
		Signature:     "somesigneddatahere",
	}
}

func Test_FromBytesWorks(t *testing.T) {
	expectedBytes, err := proto.Marshal(expectedSignature())
	if err != nil {
		t.Fatal(err)
	}
	actualSignature, err := FromBytes(expectedBytes)
	assert.Nil(t, err)
	signatureEqualFields(t, expectedSignature(), actualSignature.TopicMessage.GetFungibleSignatureMessage())
}

func Test_FromBytesWithInvalidBytes(t *testing.T) {
	result, err := FromBytes(invalidBytes)
	assert.Nil(t, result)
	assert.Error(t, err)
}

func Test_FromBytesWithTSWorks(t *testing.T) {
	expectedSignature := expectedSignature()
	expectedBytes, err := proto.Marshal(expectedSignature)
	if err != nil {
		t.Fatal(err)
	}
	actualSignature, err := FromBytesWithTS(expectedBytes, now.UnixNano())
	assert.Nil(t, err)
	signatureEqualFields(t, expectedSignature, actualSignature.TopicMessage.GetFungibleSignatureMessage())
}

func Test_FromBytesWithTSWithInvalidBytes(t *testing.T) {
	result, err := FromBytesWithTS(invalidBytes, ts)
	assert.Nil(t, result)
	assert.Error(t, err)
}

func Test_NewSignatureWorks(t *testing.T) {
	actualSignature := NewFungibleSignature(
		0,
		1,
		"0.0.123321-123321-420",
		"0xasset",
		"0xsomereceiver",
		"100",
		"somesigneddatahere")
	signatureEqualFields(t, expectedSignature(), actualSignature.TopicMessage.GetFungibleSignatureMessage())
}

func Test_FromStringWithInvalidTS(t *testing.T) {
	result, err := FromString(invalidStringData, invalidStringTs)
	assert.Nil(t, result)
	assert.Error(t, err)
}

func Test_FromStringWithInvalidData(t *testing.T) {
	result, err := FromString(invalidStringData, stringTs)
	assert.Nil(t, result)
	assert.Error(t, err)
}

func Test_FromStringWorks(t *testing.T) {
	expected := expectedSignature()

	bytes, err := proto.Marshal(expectedSignature())
	if err != nil {
		t.Fatal(err)
	}

	validData := base64.StdEncoding.EncodeToString(bytes)

	result, err := FromString(validData, timestampHelper.String(now.UnixNano()))
	assert.Nil(t, err)
	signatureEqualFields(t, expected, result.TopicMessage.GetFungibleSignatureMessage())
}

//
//func Test_ToBytes(t *testing.T) {
//	expectedBytes, err := proto.Marshal(expectedSignature())
//	if err != nil {
//		t.Fatal(err)
//	}
//	expectedMessage := &Message{TopicMessage: expectedSignature()}
//	actualBytes, err := expectedMessage.ToBytes()
//	assert.Nil(t, err)
//	assert.Equal(t, expectedBytes, actualBytes)
//}

func signatureEqualFields(t *testing.T, expected, actual *model.TopicEthSignatureMessage) {
	identical :=
		expected.SourceChainId == actual.SourceChainId &&
			expected.TargetChainId == actual.TargetChainId &&
			expected.TransferID == actual.TransferID &&
			expected.Asset == actual.Asset &&
			expected.Amount == actual.Amount &&
			expected.Recipient == actual.Recipient &&
			expected.Signature == actual.Signature
	assert.True(t, identical, "Signature fields were not equal.")
}

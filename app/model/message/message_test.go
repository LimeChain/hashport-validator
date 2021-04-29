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
		TransferID:           "0.0.123321-123321-420",
		RouterAddress:        "0xsomerouteraddress",
		WrappedAsset:         "0xwrappedasset",
		Receiver:             "0xsomereceiver",
		Amount:               "100",
		Signature:            "somesigneddatahere",
		TransactionTimestamp: 0,
	}
}

func Test_FromBytesWorks(t *testing.T) {
	expectedBytes, err := proto.Marshal(expectedSignature())
	if err != nil {
		t.Fatal(err)
	}
	actualSignature, err := FromBytes(expectedBytes)
	assert.Nil(t, err)
	signatureEqualFields(t, expectedSignature(), actualSignature.TopicEthSignatureMessage)
}

func Test_FromBytesWithInvalidBytes(t *testing.T) {
	result, err := FromBytes(invalidBytes)
	assert.Nil(t, result)
	assert.Error(t, err)
}

func Test_FromBytesWithTSWorks(t *testing.T) {
	expectedSignature := expectedSignature()
	expectedSignature.TransactionTimestamp = now.UnixNano()
	expectedBytes, err := proto.Marshal(expectedSignature)
	if err != nil {
		t.Fatal(err)
	}
	actualSignature, err := FromBytesWithTS(expectedBytes, now.UnixNano())
	assert.Nil(t, err)
	signatureEqualFields(t, expectedSignature, actualSignature.TopicEthSignatureMessage)
}

func Test_FromBytesWithTSWithInvalidBytes(t *testing.T) {
	result, err := FromBytesWithTS(invalidBytes, ts)
	assert.Nil(t, result)
	assert.Error(t, err)
}

func Test_NewSignatureWorks(t *testing.T) {
	actualSignature := NewSignature("0.0.6969-123321-420",
		"0xsomerouteraddress",
		"0xsomereceiver",
		"100",
		"somesigneddatahere",
		"0xwrappedasset")
	signatureEqualFields(t, expectedSignature(), actualSignature.TopicEthSignatureMessage)
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
	expected.TransactionTimestamp = now.UnixNano()

	bytes, err := proto.Marshal(expectedSignature())
	if err != nil {
		t.Fatal(err)
	}

	validData := base64.StdEncoding.EncodeToString(bytes)

	result, err := FromString(validData, timestampHelper.String(now.UnixNano()))
	assert.Nil(t, err)
	signatureEqualFields(t, expected, result.TopicEthSignatureMessage)
}

func Test_ToBytes(t *testing.T) {
	expectedBytes, err := proto.Marshal(expectedSignature())
	if err != nil {
		t.Fatal(err)
	}
	expectedMessage := &Message{expectedSignature()}
	actualBytes, err := expectedMessage.ToBytes()
	assert.Nil(t, err)
	assert.Equal(t, expectedBytes, actualBytes)
}

func signatureEqualFields(t *testing.T, expected, actual *model.TopicEthSignatureMessage) {
	identical := expected.Amount == actual.Amount &&
		expected.Receiver == actual.Receiver &&
		expected.WrappedAsset == actual.WrappedAsset &&
		expected.TransactionTimestamp == actual.TransactionTimestamp &&
		expected.Signature == actual.Signature &&
		expected.RouterAddress == actual.RouterAddress &&
		expected.TransferID == actual.TransferID
	assert.True(t, identical, "Signature fields were not equal.")
}

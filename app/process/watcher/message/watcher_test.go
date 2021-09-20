package message

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	p "github.com/limechain/hedera-eth-bridge-validator/proto"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"testing"
)

var (
	w       *Watcher
	topicID = hedera.TopicID{
		Shard: 0,
		Realm: 0,
		Topic: 1,
	}
)

func Test_UpdateStatusTimestamp(t *testing.T) {
	setup()
	mocks.MStatusRepository.On("Update", topicID.String(), int64(1)).Return(nil)
	w.updateStatusTimestamp(1)
}

func Test_GetAndProcessMessages_Fails(t *testing.T) {
	setup()
	mocks.MStatusRepository.On("Get", topicID.String()).Return(int64(1), nil)
	mocks.MHederaMirrorClient.On("GetMessagesAfterTimestamp", topicID, int64(1)).Return([]model.Message{}, errors.New("some-error"))
	err := w.getAndProcessMessages(mocks.MQueue, 1)
	assert.Equal(t, errors.New("some-error"), err)
}

func Test_ProcessMessage_FromString_Fails(t *testing.T) {
	setup()
	w.processMessage(model.Message{Contents: "invalid-data"}, mocks.MQueue)
	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything)
}

func Test_GetAndProcessMessages_HappyPath(t *testing.T) {
	setup()
	actualMessage := &p.TopicEthSignatureMessage{
		SourceChainId:        0,
		TargetChainId:        1,
		TransferID:           "some-id",
		Asset:                constants.Hbar,
		Recipient:            "0xsomeethereumaddress",
		Amount:               "100",
		TransactionTimestamp: 10000000000000,
	}
	bytes, e := proto.Marshal(actualMessage)
	if e != nil {
		t.Fatal(e)
	}
	contents := base64.StdEncoding.EncodeToString(bytes)
	m := model.Message{
		ConsensusTimestamp: "10000.0",
		TopicId:            topicID.String(),
		Contents:           contents,
	}
	mocks.MStatusRepository.On("Update", topicID.String(), int64(10000000000000)).Return(nil)
	mocks.MStatusRepository.On("Get", topicID.String()).Return(int64(1), nil)
	mocks.MHederaMirrorClient.On("GetMessagesAfterTimestamp", topicID, int64(1)).Return([]model.Message{m}, nil)
	mocks.MQueue.On("Push", mock.Anything)
	err := w.getAndProcessMessages(mocks.MQueue, 1)
	assert.Nil(t, err)
	mocks.MQueue.AssertCalled(t, "Push", mock.Anything)
}

func Test_GetAndProcessMessages_InvalidConsensusTimestamp(t *testing.T) {
	setup()
	m := model.Message{
		ConsensusTimestamp: "aaaa",
		TopicId:            topicID.String(),
		Contents:           "some-content-here",
	}
	mocks.MStatusRepository.On("Get", topicID.String()).Return(int64(1), nil)
	mocks.MHederaMirrorClient.On("GetMessagesAfterTimestamp", topicID, int64(1)).Return([]model.Message{m}, nil)
	err := w.getAndProcessMessages(mocks.MQueue, 1)
	assert.Nil(t, err)
	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything, mock.Anything)
	mocks.MStatusRepository.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func Test_NewWatcher(t *testing.T) {
	mocks.Setup()
	mocks.MStatusRepository.On("Get", topicID.String()).Return(int64(0), nil)
	NewWatcher(mocks.MHederaMirrorClient, "0.0.1", mocks.MStatusRepository, 1, 0)
}

func Test_NewWatcher_Get_Error(t *testing.T) {
	mocks.Setup()
	mocks.MStatusRepository.On("Get", topicID.String()).Return(int64(0), gorm.ErrRecordNotFound)
	mocks.MStatusRepository.On("Create", topicID.String(), mock.Anything).Return(nil)
	NewWatcher(mocks.MHederaMirrorClient, "0.0.1", mocks.MStatusRepository, 1, 0)
}

func Test_NewWatcher_WithTS(t *testing.T) {
	mocks.Setup()
	mocks.MStatusRepository.On("Get", topicID.String()).Return(int64(1), nil)
	mocks.MStatusRepository.On("Update", topicID.String(), int64(1)).Return(nil)
	NewWatcher(mocks.MHederaMirrorClient, "0.0.1", mocks.MStatusRepository, 1, 1)
}

func setup() {
	mocks.Setup()
	w = &Watcher{
		client:           mocks.MHederaMirrorClient,
		topicID:          topicID,
		statusRepository: mocks.MStatusRepository,
		pollingInterval:  1,
		logger:           config.GetLoggerFor(fmt.Sprintf("[%s] Topic Watcher", topicID)),
	}
}

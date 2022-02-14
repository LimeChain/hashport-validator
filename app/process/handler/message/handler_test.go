/*
 * Copyright 2021 LimeChain Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package message

import (
	"errors"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	auth_message "github.com/limechain/hedera-eth-bridge-validator/app/model/auth-message"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"math/big"
	"testing"
)

var (
	h       *Handler
	topicId = hedera.TopicID{
		Shard: 0,
		Realm: 0,
		Topic: 1,
	}

	tesm = &proto.TopicEthSignatureMessage{
		SourceChainId: 0,
		TargetChainId: 1,
		TransferID:    "some-transfer-id",
		Asset:         "0.0.1",
		Recipient:     "0xb083879B1e10C8476802016CB12cd2F25a896691",
		Amount:        "100",
		Signature:     "custom-signature",
	}
	tsm = message.Message{
		TopicMessage: &proto.TopicMessage{
			Message: &proto.TopicMessage_FungibleSignatureMessage{
				FungibleSignatureMessage: tesm,
			},
		},
	}
	transactionTimestamp = int64(0)
	authMsgBytes, _      = auth_message.EncodeFungibleBytesFrom(int64(tesm.SourceChainId), int64(tesm.TargetChainId), tesm.TransferID, tesm.Asset, tesm.Recipient, tesm.Amount)
)

func Test_NewHandler(t *testing.T) {
	setup()
	assert.Equal(t, h, NewHandler(topicId.String(), mocks.MTransferRepository, mocks.MMessageRepository, map[int64]service.Contracts{1: mocks.MBridgeContractService}, mocks.MMessageService))
}

func Test_Handle_Fails(t *testing.T) {
	setup()
	h.Handle("invalid-payload")
	mocks.MMessageService.AssertNotCalled(t, "ProcessSignature", mock.Anything)
	mocks.MMessageRepository.AssertNotCalled(t, "Get", mock.Anything)
	mocks.MBridgeContractService.AssertNotCalled(t, "GetMembers")
}

func Test_HandleSignatureMessage_SanityCheckFails(t *testing.T) {
	setup()
	mocks.MMessageService.On("SanityCheckFungibleSignature", tsm.GetFungibleSignatureMessage()).Return(false, errors.New("some-error"))
	h.handleFungibleSignatureMessage(tsm.GetFungibleSignatureMessage(), transactionTimestamp)
	mocks.MMessageService.AssertNotCalled(t, "ProcessSignature", tsm)
}

func Test_HandleSignatureMessage_SanityCheckIsNotValid(t *testing.T) {
	setup()
	mocks.MMessageService.On("SanityCheckFungibleSignature", tsm.GetFungibleSignatureMessage()).Return(false, nil)
	h.handleFungibleSignatureMessage(tsm.GetFungibleSignatureMessage(), transactionTimestamp)
	mocks.MMessageService.AssertNotCalled(t, "ProcessSignature", tsm)
}

func Test_HandleSignatureMessage_ProcessSignatureFails(t *testing.T) {
	setup()
	mocks.MMessageService.On("SanityCheckFungibleSignature", tsm.GetFungibleSignatureMessage()).Return(true, nil)
	mocks.MMessageService.On("ProcessSignature", tsm.GetFungibleSignatureMessage().TransferID, tsm.GetFungibleSignatureMessage().Signature, int64(tsm.GetFungibleSignatureMessage().TargetChainId), transactionTimestamp, authMsgBytes).Return(errors.New("some-error"))
	h.handleFungibleSignatureMessage(tsm.GetFungibleSignatureMessage(), transactionTimestamp)
	mocks.MTransferRepository.AssertNotCalled(t, "Update", mock.Anything)
	mocks.MMessageRepository.AssertNotCalled(t, "Get", mock.Anything)
	mocks.MBridgeContractService.AssertNotCalled(t, "GetMembers")
}

func Test_HandleSignatureMessage_MajorityReached(t *testing.T) {
	setup()
	mocks.MMessageService.On("SanityCheckFungibleSignature", tsm.GetFungibleSignatureMessage()).Return(true, nil)
	mocks.MMessageService.On("ProcessSignature", tsm.GetFungibleSignatureMessage().TransferID, tsm.GetFungibleSignatureMessage().Signature, int64(tsm.GetFungibleSignatureMessage().TargetChainId), transactionTimestamp, authMsgBytes).Return(nil)
	mocks.MMessageRepository.On("Get", tsm.GetFungibleSignatureMessage().TransferID).Return([]entity.Message{{}, {}, {}}, nil)
	mocks.MBridgeContractService.On("GetMembers").Return([]string{"", "", ""})
	mocks.MBridgeContractService.On("HasValidSignaturesLength", big.NewInt(3)).Return(true, nil)
	mocks.MTransferRepository.On("UpdateStatusCompleted", tsm.GetFungibleSignatureMessage().TransferID).Return(nil)
	h.handleFungibleSignatureMessage(tsm.GetFungibleSignatureMessage(), transactionTimestamp)
	mocks.MBridgeContractService.AssertCalled(t, "HasValidSignaturesLength", big.NewInt(3))
	mocks.MTransferRepository.AssertCalled(t, "UpdateStatusCompleted", tsm.GetFungibleSignatureMessage().TransferID)
}

func Test_Handle(t *testing.T) {
	setup()
	mocks.MMessageService.On("SanityCheckFungibleSignature", tsm.GetFungibleSignatureMessage()).Return(true, nil)
	mocks.MMessageService.On("ProcessSignature", tsm.GetFungibleSignatureMessage().TransferID, tsm.GetFungibleSignatureMessage().Signature, int64(tsm.GetFungibleSignatureMessage().TargetChainId), transactionTimestamp, authMsgBytes).Return(nil)
	mocks.MMessageRepository.On("Get", tsm.GetFungibleSignatureMessage().TransferID).Return([]entity.Message{{}, {}, {}}, nil)
	mocks.MBridgeContractService.On("GetMembers").Return([]string{"", "", ""})
	mocks.MBridgeContractService.On("HasValidSignaturesLength", big.NewInt(3)).Return(true, nil)
	mocks.MTransferRepository.On("UpdateStatusCompleted", tsm.GetFungibleSignatureMessage().TransferID).Return(nil)
	h.Handle(&tsm)
	mocks.MBridgeContractService.AssertCalled(t, "HasValidSignaturesLength", big.NewInt(3))
	mocks.MTransferRepository.AssertCalled(t, "UpdateStatusCompleted", tsm.GetFungibleSignatureMessage().TransferID)
}

func Test_HandleSignatureMessage_UpdateStatusCompleted_Fails(t *testing.T) {
	setup()
	mocks.MMessageService.On("SanityCheckFungibleSignature", tsm.GetFungibleSignatureMessage()).Return(true, nil)
	mocks.MMessageService.On("ProcessSignature", tsm.GetFungibleSignatureMessage().TransferID, tsm.GetFungibleSignatureMessage().Signature, int64(tsm.GetFungibleSignatureMessage().TargetChainId), transactionTimestamp, authMsgBytes).Return(nil)
	mocks.MMessageRepository.On("Get", tsm.GetFungibleSignatureMessage().TransferID).Return([]entity.Message{{}, {}, {}}, nil)
	mocks.MBridgeContractService.On("GetMembers").Return([]string{"", "", ""})
	mocks.MBridgeContractService.On("HasValidSignaturesLength", big.NewInt(3)).Return(true, nil)
	mocks.MTransferRepository.On("UpdateStatusCompleted", tsm.GetFungibleSignatureMessage().TransferID).Return(errors.New("some-error"))
	h.handleFungibleSignatureMessage(tsm.GetFungibleSignatureMessage(), transactionTimestamp)
	mocks.MBridgeContractService.AssertCalled(t, "HasValidSignaturesLength", big.NewInt(3))
	mocks.MTransferRepository.AssertNotCalled(t, "UpdateStatusCompleted")
}

func Test_HandleSignatureMessage_CheckMajority_Fails(t *testing.T) {
	setup()
	mocks.MMessageService.On("SanityCheckFungibleSignature", tsm.GetFungibleSignatureMessage()).Return(true, nil)
	mocks.MMessageService.On("ProcessSignature", tsm.GetFungibleSignatureMessage().TransferID, tsm.GetFungibleSignatureMessage().Signature, int64(tsm.GetFungibleSignatureMessage().TargetChainId), transactionTimestamp, authMsgBytes).Return(nil)
	mocks.MMessageRepository.On("Get", tsm.GetFungibleSignatureMessage().TransferID).Return([]entity.Message{{}, {}, {}}, errors.New("some-error"))
	h.handleFungibleSignatureMessage(tsm.GetFungibleSignatureMessage(), transactionTimestamp)
	mocks.MBridgeContractService.AssertNotCalled(t, "GetMembers")
	mocks.MTransferRepository.AssertNotCalled(t, "UpdateStatusCompleted", tsm.GetFungibleSignatureMessage().TransferID)
}

func setup() {
	mocks.Setup()
	h = &Handler{
		transferRepository: mocks.MTransferRepository,
		messageRepository:  mocks.MMessageRepository,
		contracts:          map[int64]service.Contracts{1: mocks.MBridgeContractService},
		messages:           mocks.MMessageService,
		logger:             config.GetLoggerFor(fmt.Sprintf("Topic [%s] Handler", topicId.String())),
	}
}

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
	"github.com/limechain/hedera-eth-bridge-validator/app/model/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
	"github.com/limechain/hedera-eth-bridge-validator/test/constants"
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
		SourceChainId:        0,
		TargetChainId:        1,
		TransferID:           "some-transfer-id",
		Asset:                "0.0.1",
		Recipient:            "0xsomeethaddress",
		Amount:               "100",
		Signature:            "custom-signature",
		TransactionTimestamp: 0,
	}
	tsm = message.Message{
		TopicEthSignatureMessage: tesm,
	}

	assets config.Assets
)

func Test_NewHandler(t *testing.T) {
	setup()
	assert.Equal(t, h, NewHandler(topicId.String(), mocks.MTransferRepository, mocks.MMessageRepository, map[int64]service.Contracts{1: mocks.MBridgeContractService}, mocks.MMessageService, mocks.MPrometheusService, assets))
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
	mocks.MMessageService.On("SanityCheckSignature", tsm).Return(false, errors.New("some-error"))
	h.handleSignatureMessage(tsm)
	mocks.MMessageService.AssertNotCalled(t, "ProcessSignature", tsm)
}

func Test_HandleSignatureMessage_SanityCheckIsNotValid(t *testing.T) {
	setup()
	mocks.MMessageService.On("SanityCheckSignature", tsm).Return(false, nil)
	h.handleSignatureMessage(tsm)
	mocks.MMessageService.AssertNotCalled(t, "ProcessSignature", tsm)
}

func Test_HandleSignatureMessage_ProcessSignatureFails(t *testing.T) {
	setup()
	mocks.MMessageService.On("SanityCheckSignature", tsm).Return(true, nil)
	mocks.MMessageService.On("ProcessSignature", tsm).Return(errors.New("some-error"))
	h.handleSignatureMessage(tsm)
	mocks.MTransferRepository.AssertNotCalled(t, "Update", mock.Anything)
	mocks.MMessageRepository.AssertNotCalled(t, "Get", mock.Anything)
	mocks.MBridgeContractService.AssertNotCalled(t, "GetMembers")
}

func Test_HandleSignatureMessage_MajorityReached(t *testing.T) {
	setup()
	mocks.MMessageService.On("SanityCheckSignature", tsm).Return(true, nil)
	mocks.MMessageService.On("ProcessSignature", tsm).Return(nil)
	mocks.MMessageRepository.On("Get", tsm.TransferID).Return([]entity.Message{{}, {}, {}}, nil)
	mocks.MBridgeContractService.On("GetMembers").Return([]string{"", "", ""})
	mocks.MBridgeContractService.On("HasValidSignaturesLength", big.NewInt(3)).Return(true, nil)
	mocks.MTransferRepository.On("UpdateStatusCompleted", tsm.TransferID).Return(nil)
	h.handleSignatureMessage(tsm)
	mocks.MBridgeContractService.AssertCalled(t, "HasValidSignaturesLength", big.NewInt(3))
	mocks.MTransferRepository.AssertCalled(t, "UpdateStatusCompleted", tsm.TransferID)
}

func Test_Handle(t *testing.T) {
	setup()
	mocks.MMessageService.On("SanityCheckSignature", tsm).Return(true, nil)
	mocks.MMessageService.On("ProcessSignature", tsm).Return(nil)
	mocks.MMessageRepository.On("Get", tsm.TransferID).Return([]entity.Message{{}, {}, {}}, nil)
	mocks.MBridgeContractService.On("GetMembers").Return([]string{"", "", ""})
	mocks.MBridgeContractService.On("HasValidSignaturesLength", big.NewInt(3)).Return(true, nil)
	mocks.MTransferRepository.On("UpdateStatusCompleted", tsm.TransferID).Return(nil)
	h.Handle(&tsm)
	mocks.MBridgeContractService.AssertCalled(t, "HasValidSignaturesLength", big.NewInt(3))
	mocks.MTransferRepository.AssertCalled(t, "UpdateStatusCompleted", tsm.TransferID)
}

func Test_HandleSignatureMessage_UpdateStatusCompleted_Fails(t *testing.T) {
	setup()
	mocks.MMessageService.On("SanityCheckSignature", tsm).Return(true, nil)
	mocks.MMessageService.On("ProcessSignature", tsm).Return(nil)
	mocks.MMessageRepository.On("Get", tsm.TransferID).Return([]entity.Message{{}, {}, {}}, nil)
	mocks.MBridgeContractService.On("GetMembers").Return([]string{"", "", ""})
	mocks.MBridgeContractService.On("HasValidSignaturesLength", big.NewInt(3)).Return(true, nil)
	mocks.MTransferRepository.On("UpdateStatusCompleted", tsm.TransferID).Return(errors.New("some-error"))
	h.handleSignatureMessage(tsm)
	mocks.MBridgeContractService.AssertCalled(t, "HasValidSignaturesLength", big.NewInt(3))
	mocks.MTransferRepository.AssertNotCalled(t, "UpdateStatusCompleted")
}

func Test_HandleSignatureMessage_CheckMajority_Fails(t *testing.T) {
	setup()
	mocks.MMessageService.On("SanityCheckSignature", tsm).Return(true, nil)
	mocks.MMessageService.On("ProcessSignature", tsm).Return(nil)
	mocks.MMessageRepository.On("Get", tsm.TransferID).Return([]entity.Message{{}, {}, {}}, errors.New("some-error"))
	h.handleSignatureMessage(tsm)
	mocks.MBridgeContractService.AssertNotCalled(t, "GetMembers")
	mocks.MTransferRepository.AssertNotCalled(t, "UpdateStatusCompleted", tsm.TransferID)
}

func setup() {
	mocks.Setup()
	mocks.MPrometheusService.On("GetIsMonitoringEnabled").Return(false)

	assets = config.LoadAssets(constants.Networks)
	h = &Handler{
		transferRepository:     mocks.MTransferRepository,
		messageRepository:      mocks.MMessageRepository,
		contracts:              map[int64]service.Contracts{1: mocks.MBridgeContractService},
		messages:               mocks.MMessageService,
		logger:                 config.GetLoggerFor(fmt.Sprintf("Topic [%s] Handler", topicId.String())),
		prometheusService:      mocks.MPrometheusService,
		assetsConfig:           assets,
		participationRateGauge: nil,
	}
}

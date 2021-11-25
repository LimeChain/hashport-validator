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

package message_submission

import (
	"errors"
	"github.com/hashgraph/hedera-sdk-go/v2"
	hederahelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/hedera"
	auth_message "github.com/limechain/hedera-eth-bridge-validator/app/model/auth-message"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

var (
	msHandler *Handler
	tr        = transfer.Transfer{
		TransactionId: "some-transaction-id",
		SourceChainId: 0,
		TargetChainId: 1,
		NativeChainId: 0,
		SourceAsset:   "0.0.1123",
		TargetAsset:   "0xethaddress",
		NativeAsset:   "0.0.1123",
		Receiver:      "0xethaddress-2",
		Amount:        "100",
	}

	transferRecord = &entity.Transfer{
		TransactionID: tr.TransactionId,
		SourceChainID: tr.SourceChainId,
		TargetChainID: tr.TargetChainId,
		NativeChainID: tr.NativeChainId,
		SourceAsset:   tr.SourceAsset,
		TargetAsset:   tr.TargetAsset,
		NativeAsset:   tr.NativeAsset,
		Receiver:      tr.Receiver,
		Amount:        tr.Amount,
		Status:        status.Initial,
		Messages:      nil,
		Fees:          []entity.Fee{},
		Schedules:     nil,
	}

	topicId = hedera.TopicID{
		Shard: 0,
		Realm: 0,
		Topic: 10,
	}
	fungibleMessage = &proto.TopicEthSignatureMessage{
		SourceChainId: uint64(transferRecord.SourceChainID),
		TargetChainId: uint64(transferRecord.TargetChainID),
		TransferID:    transferRecord.TransactionID,
		Asset:         transferRecord.NativeAsset,
		Recipient:     transferRecord.Receiver,
		Amount:        transferRecord.Amount,
		Signature:     "signature",
	}
	//m = &message.Message{
	//	TopicMessage: &proto.TopicMessage{
	//		Message: &proto.TopicMessage_FungibleSignatureMessage{
	//			FungibleSignatureMessage: fungibleMessage},
	//	},
	//	TransactionTimestamp: 0,
	//}
	authMsgBytes, _ = auth_message.EncodeBytesFrom(transferRecord.SourceChainID, transferRecord.TargetChainID, transferRecord.TransactionID, transferRecord.TargetAsset, transferRecord.Receiver, transferRecord.Amount)
	date            = time.Date(2001, time.June, 1, 1, 1, 1, 1, time.UTC)
	txId            = &hedera.TransactionID{
		AccountID: &hedera.AccountID{
			Shard:   0,
			Realm:   0,
			Account: 2,
		},
		ValidStart: &date,
	}
)

func Test_NewHandler(t *testing.T) {
	mocks.Setup()
	h := NewHandler(mocks.MHederaNodeClient, mocks.MHederaMirrorClient, mocks.MTransferService, mocks.MTransferRepository, mocks.MMessageService, "0.0.1111")
	assert.Equal(t, &Handler{
		hederaNode:         mocks.MHederaNodeClient,
		mirrorNode:         mocks.MHederaMirrorClient,
		transfersService:   mocks.MTransferService,
		transferRepository: mocks.MTransferRepository,
		topicID: hedera.TopicID{
			Shard: 0,
			Realm: 0,
			Topic: 1111,
		},
		messageService: mocks.MMessageService,
		logger:         config.GetLoggerFor("Topic Message Submission Handler"),
	}, h)
}

func Test_Handle_Encoding_Fails(t *testing.T) {
	setup()

	invalidTransferPayload := []byte{1, 2, 1}

	msHandler.Handle(invalidTransferPayload)

	mocks.MLockService.AssertNotCalled(t, "ProcessEvent")
}

func Test_Invalid_Payload(t *testing.T) {
	setup()
	msHandler.Handle(tr)
	mocks.MTransferService.AssertNotCalled(t, "InitiateNewTransfer", mock.Anything)
}

func Test_AuthMessageSubmissionCallbacks(t *testing.T) {
	setup()
	onSuccess, onFail := msHandler.authMessageSubmissionCallbacks("some-tx-id")
	onSuccess()
	onFail()
}

func Test_Handle(t *testing.T) {
	setup()
	mocks.MTransferService.On("InitiateNewTransfer", tr).Return(transferRecord, nil)
	mocks.MMessageService.On("SignMessage", mock.Anything).Return(authMsgBytes, nil)
	mocks.MHederaNodeClient.On("SubmitTopicConsensusMessage", topicId, mock.Anything).Return(txId, nil)
	mocks.MHederaMirrorClient.On("WaitForTransaction", hederahelper.ToMirrorNodeTransactionID(txId.String()), mock.Anything, mock.Anything)
	msHandler.Handle(&tr)
}

func Test_Handle_SubmitTopicConsensusMessageFails(t *testing.T) {
	setup()
	mocks.MTransferService.On("InitiateNewTransfer", tr).Return(transferRecord, nil)
	mocks.MMessageService.On("SignMessage", mock.Anything).Return(authMsgBytes, nil)
	mocks.MHederaNodeClient.On("SubmitTopicConsensusMessage", topicId, mock.Anything).Return(txId, errors.New("some-error"))
	msHandler.Handle(&tr)
	mocks.MHederaMirrorClient.AssertNotCalled(t, "WaitForTransaction", hederahelper.ToMirrorNodeTransactionID(txId.String()), mock.Anything, mock.Anything)
}

func Test_Handle_InitiateNewTransfer_Fails(t *testing.T) {
	setup()
	mocks.MTransferService.On("InitiateNewTransfer", tr).Return(transferRecord, errors.New("some-error"))
	msHandler.Handle(&tr)
	mocks.MSignerService.AssertNotCalled(t, "Sign", mock.Anything)
	mocks.MHederaNodeClient.AssertNotCalled(t, "SubmitTopicConsensusMessage", topicId, mock.Anything)
	mocks.MHederaMirrorClient.AssertNotCalled(t, "WaitForTransaction", hederahelper.ToMirrorNodeTransactionID(txId.String()), mock.Anything, mock.Anything)
}

func Test_Handle_InitiateNewTransfer_NotInitial(t *testing.T) {
	setup()
	transferRecord.Status = "not-initial"

	mocks.MTransferService.On("InitiateNewTransfer", tr).Return(transferRecord, nil)
	msHandler.Handle(&tr)
	mocks.MSignerService.AssertNotCalled(t, "Sign", mock.Anything)
	mocks.MHederaNodeClient.AssertNotCalled(t, "SubmitTopicConsensusMessage", topicId, mock.Anything)
	mocks.MHederaMirrorClient.AssertNotCalled(t, "WaitForTransaction", hederahelper.ToMirrorNodeTransactionID(txId.String()), mock.Anything, mock.Anything)

	transferRecord.Status = status.Initial
}

func Test_Handle_SignMessage_Fails(t *testing.T) {
	setup()
	mocks.MTransferService.On("InitiateNewTransfer", tr).Return(transferRecord, nil)
	mocks.MMessageService.On("SignMessage", mock.Anything).Return([]byte{}, errors.New("some-error"))
	msHandler.Handle(&tr)
	mocks.MHederaNodeClient.AssertNotCalled(t, "SubmitTopicConsensusMessage", topicId, mock.Anything)
	mocks.MHederaMirrorClient.AssertNotCalled(t, "WaitForTransaction", hederahelper.ToMirrorNodeTransactionID(txId.String()), mock.Anything, mock.Anything)
}

func setup() {
	mocks.Setup()
	msHandler = &Handler{
		hederaNode:         mocks.MHederaNodeClient,
		mirrorNode:         mocks.MHederaMirrorClient,
		transfersService:   mocks.MTransferService,
		transferRepository: mocks.MTransferRepository,
		messageService:     mocks.MMessageService,
		topicID:            topicId,
		logger:             config.GetLoggerFor("Hedera Mint and Transfer Handler"),
	}
}

/*
 * Copyright 2022 LimeChain Ltd.
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

package transfer

import (
	"database/sql"
	"errors"
	"github.com/hashgraph/hedera-sdk-go/v2"
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/schedule"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	testConstants "github.com/limechain/hedera-eth-bridge-validator/test/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	test_config "github.com/limechain/hedera-eth-bridge-validator/test/test-config"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	handler       *Handler
	bridgeAccount = test_config.TestConfig.Bridge.Hedera.BridgeAccount
	transactionId = "1234"
	sourceChainId = constants.HederaNetworkId
	targetChainId = testConstants.EthereumNetworkId
	nativeChainId = constants.HederaNetworkId
	sourceAsset   = testConstants.NetworkEthereumNFTWrappedTokenForNetworkHedera
	targetAsset   = testConstants.NetworkHederaNonFungibleNativeToken
	nativeAsset   = testConstants.NetworkHederaNonFungibleNativeToken
	receiver      = "0.0.455300"
	amount        = "1000000000000000000"
	serialNum     = int64(123)
	metadata      = "SomeMetadata"
	fee           = "10000"
	isNft         = true
	timestamp     = time.Now().UTC().String()
	entityStatus  = status.Initial
	nilErr        error

	payload = &model.Transfer{
		TransactionId: transactionId,
		SourceChainId: sourceChainId,
		TargetChainId: targetChainId,
		NativeChainId: nativeChainId,
		SourceAsset:   sourceAsset,
		TargetAsset:   targetAsset,
		NativeAsset:   nativeAsset,
		Receiver:      receiver,
		Amount:        amount,
		SerialNum:     serialNum,
		Metadata:      metadata,
		IsNft:         isNft,
		Timestamp:     timestamp,
	}

	resultEntityTransfer = &entity.Transfer{
		TransactionID: transactionId,
		SourceChainID: sourceChainId,
		TargetChainID: targetChainId,
		NativeChainID: nativeChainId,
		SourceAsset:   sourceAsset,
		TargetAsset:   targetAsset,
		NativeAsset:   nativeAsset,
		Receiver:      receiver,
		Amount:        amount,
		Fee:           fee,
		Status:        entityStatus,
		SerialNumber:  serialNum,
		Metadata:      metadata,
		IsNft:         isNft,
		Messages:      make([]entity.Message, 0),
		Fees:          make([]entity.Fee, 0),
		Schedules:     make([]entity.Schedule, 0),
	}

	bridgeAccountId   hedera.AccountID
	receiverAccountId hedera.AccountID

	scheduleId              = "0.0.675300"
	onSuccessScheduleEntity = &entity.Schedule{
		TransactionID: transactionId,
		ScheduleID:    scheduleId,
		HasReceiver:   true,
		Operation:     schedule.TRANSFER,
		Status:        status.Submitted,
		TransferID:    sql.NullString{String: transactionId, Valid: true},
	}
	onFailureScheduleEntity = &entity.Schedule{
		TransactionID: transactionId,
		ScheduleID:    "",
		HasReceiver:   true,
		Operation:     "",
		Status:        status.Failed,
		TransferID:    sql.NullString{String: transactionId, Valid: true},
	}
)

func Test_NewHandler(t *testing.T) {
	setup(t)

	actualHandler := NewHandler(
		bridgeAccount,
		mocks.MTransferRepository,
		mocks.MScheduleRepository,
		mocks.MTransferService,
		mocks.MScheduledService)

	assert.Equal(t, handler, actualHandler)
}

func Test_NewHandler_BridgeAccountError(t *testing.T) {
	setup(t)

	defer func() { log.StandardLogger().ExitFunc = nil }()
	fatal := false
	log.StandardLogger().ExitFunc = func(int) { fatal = true }

	_ = NewHandler(
		"",
		mocks.MTransferRepository,
		mocks.MScheduleRepository,
		mocks.MTransferService,
		mocks.MScheduledService)

	assert.Equal(t, true, fatal)
}

func Test_Handle(t *testing.T) {
	setup(t)

	token, err := hedera.TokenIDFromString(targetAsset)
	if err != nil {
		t.Fatal(err)
	}
	nftID := hedera.NftID{
		TokenID:      token,
		SerialNumber: serialNum,
	}

	mocks.MTransferService.On("InitiateNewTransfer", *payload).Return(resultEntityTransfer, nilErr)
	mocks.MScheduledService.On("ExecuteScheduledNftTransferTransaction",
		transactionId,
		nftID,
		bridgeAccountId,
		receiverAccountId).Return()

	handler.Handle(payload)

	mocks.MTransferService.AssertCalled(t, "InitiateNewTransfer", *payload)
	mocks.MScheduledService.AssertCalled(t, "ExecuteScheduledNftTransferTransaction",
		transactionId,
		nftID,
		bridgeAccountId,
		receiverAccountId)
}

func Test_Handle_CastError(t *testing.T) {
	setup(t)
	brokenPayload := "Not a transfer"

	handler.Handle(brokenPayload)

	mocks.MTransferService.AssertNotCalled(t, "InitiateNewTransfer")
	mocks.MScheduledService.AssertNotCalled(t, "ExecuteScheduledNftTransferTransaction")
}

func Test_Handle_ReceiverError(t *testing.T) {
	setup(t)
	payload.Receiver = ""

	handler.Handle(payload)

	payload.Receiver = receiver

	mocks.MTransferService.AssertNotCalled(t, "InitiateNewTransfer")
	mocks.MScheduledService.AssertNotCalled(t, "ExecuteScheduledNftTransferTransaction")
}

func Test_Handle_TokenError(t *testing.T) {
	setup(t)
	payload.TargetAsset = ""

	handler.Handle(payload)

	payload.TargetAsset = targetAsset

	mocks.MTransferService.AssertNotCalled(t, "InitiateNewTransfer")
	mocks.MScheduledService.AssertNotCalled(t, "ExecuteScheduledNftTransferTransaction")
}

func Test_Handle_TransactionError(t *testing.T) {
	setup(t)
	mocks.MTransferService.On("InitiateNewTransfer", *payload).Return(resultEntityTransfer, errors.New("failed to create record"))

	handler.Handle(payload)

	mocks.MTransferService.AssertCalled(t, "InitiateNewTransfer", *payload)
	mocks.MScheduledService.AssertNotCalled(t, "ExecuteScheduledNftTransferTransaction")
}

func Test_Handle_NotInitialStatus(t *testing.T) {
	setup(t)
	resultEntityTransfer.Status = status.Submitted
	mocks.MTransferService.On("InitiateNewTransfer", *payload).Return(resultEntityTransfer, nilErr)

	handler.Handle(payload)

	resultEntityTransfer.Status = entityStatus

	mocks.MTransferService.AssertCalled(t, "InitiateNewTransfer", *payload)
	mocks.MScheduledService.AssertNotCalled(t, "ExecuteScheduledNftTransferTransaction")
}

func Test_scheduledTxMinedCallbacks(t *testing.T) {
	setup(t)

	mocks.MTransferRepository.On("UpdateStatusCompleted", transactionId).Return(nilErr)
	mocks.MTransferRepository.On("UpdateStatusFailed", transactionId).Return(nilErr)
	mocks.MScheduleRepository.On("UpdateStatusCompleted", transactionId).Return(nilErr)
	mocks.MScheduleRepository.On("UpdateStatusFailed", transactionId).Return(nilErr)

	onSuccess, onFailure := handler.scheduledTxMinedCallbacks(transactionId)
	onSuccess(transactionId)
	onFailure(transactionId)

	mocks.MTransferRepository.AssertCalled(t, "UpdateStatusCompleted", transactionId)
	mocks.MTransferRepository.AssertCalled(t, "UpdateStatusFailed", transactionId)
	mocks.MScheduleRepository.AssertCalled(t, "UpdateStatusCompleted", transactionId)
	mocks.MScheduleRepository.AssertCalled(t, "UpdateStatusFailed", transactionId)
}

func Test_scheduledTxMinedCallbacks_TransferRepoErrorOnSuccess(t *testing.T) {
	setup(t)

	err := errors.New("some error")
	mocks.MTransferRepository.On("UpdateStatusCompleted", transactionId).Return(err)

	onSuccess, _ := handler.scheduledTxMinedCallbacks(transactionId)
	onSuccess(transactionId)

	mocks.MTransferRepository.AssertCalled(t, "UpdateStatusCompleted", transactionId)
	mocks.MScheduleRepository.AssertNotCalled(t, "UpdateStatusCompleted")
}

func Test_scheduledTxMinedCallbacks_ScheduleRepoErrorOnSuccess(t *testing.T) {
	setup(t)

	err := errors.New("some error")
	mocks.MTransferRepository.On("UpdateStatusCompleted", transactionId).Return(nilErr)
	mocks.MScheduleRepository.On("UpdateStatusCompleted", transactionId).Return(err)

	onSuccess, _ := handler.scheduledTxMinedCallbacks(transactionId)
	onSuccess(transactionId)

	mocks.MTransferRepository.AssertCalled(t, "UpdateStatusCompleted", transactionId)
	mocks.MScheduleRepository.AssertCalled(t, "UpdateStatusCompleted", transactionId)
}

func Test_scheduledTxMinedCallbacks_TransferRepoErrorOnFailure(t *testing.T) {
	setup(t)

	err := errors.New("some error")
	mocks.MScheduleRepository.On("UpdateStatusFailed", transactionId).Return(nilErr)
	mocks.MTransferRepository.On("UpdateStatusFailed", transactionId).Return(err)

	_, onFailure := handler.scheduledTxMinedCallbacks(transactionId)
	onFailure(transactionId)

	mocks.MScheduleRepository.AssertCalled(t, "UpdateStatusFailed", transactionId)
	mocks.MTransferRepository.AssertCalled(t, "UpdateStatusFailed", transactionId)
}

func Test_scheduledTxMinedCallbacks_ScheduledRepoErrorOnFailure(t *testing.T) {
	setup(t)

	err := errors.New("some error")
	mocks.MScheduleRepository.On("UpdateStatusFailed", transactionId).Return(err)

	_, onFailure := handler.scheduledTxMinedCallbacks(transactionId)
	onFailure(transactionId)

	mocks.MScheduleRepository.AssertCalled(t, "UpdateStatusFailed", transactionId)
	mocks.MTransferRepository.AssertNotCalled(t, "UpdateStatusFailed")
}

func Test_scheduledTxExecutionCallbacks_OnSuccess(t *testing.T) {
	setup(t)

	mocks.MScheduleRepository.On("Create", onSuccessScheduleEntity).Return(nilErr)

	onSuccess, _ := handler.scheduledTxExecutionCallbacks(transactionId, true)
	onSuccess(transactionId, scheduleId)

	mocks.MScheduleRepository.AssertCalled(t, "Create", onSuccessScheduleEntity)

}

func Test_scheduledTxExecutionCallbacks_OnSuccess_Err(t *testing.T) {
	setup(t)

	mocks.MScheduleRepository.On("Create", onSuccessScheduleEntity).Return(errors.New("some error"))

	onSuccess, _ := handler.scheduledTxExecutionCallbacks(transactionId, true)
	onSuccess(transactionId, scheduleId)

	mocks.MScheduleRepository.AssertCalled(t, "Create", onSuccessScheduleEntity)

}

func Test_scheduledTxExecutionCallbacks_OnFailure(t *testing.T) {
	setup(t)

	mocks.MScheduleRepository.On("Create", onFailureScheduleEntity).Return(nilErr)
	mocks.MTransferRepository.On("UpdateStatusFailed", transactionId).Return(nilErr)

	_, OnFailure := handler.scheduledTxExecutionCallbacks(transactionId, true)
	OnFailure(transactionId)

	mocks.MScheduleRepository.AssertCalled(t, "Create", onFailureScheduleEntity)
	mocks.MTransferRepository.AssertCalled(t, "UpdateStatusFailed", transactionId)
}

func Test_scheduledTxExecutionCallbacks_OnFailure_CreateEntityErr(t *testing.T) {
	setup(t)

	mocks.MScheduleRepository.On("Create", onFailureScheduleEntity).Return(errors.New("some error"))

	_, OnFailure := handler.scheduledTxExecutionCallbacks(transactionId, true)
	OnFailure(transactionId)

	mocks.MScheduleRepository.AssertCalled(t, "Create", onFailureScheduleEntity)
}

func Test_scheduledTxExecutionCallbacks_OnFailure_UpdateStatusErr(t *testing.T) {
	setup(t)

	mocks.MScheduleRepository.On("Create", onFailureScheduleEntity).Return(nilErr)
	mocks.MTransferRepository.On("UpdateStatusFailed", transactionId).Return(errors.New("some error"))

	_, OnFailure := handler.scheduledTxExecutionCallbacks(transactionId, true)
	OnFailure(transactionId)

	mocks.MScheduleRepository.AssertCalled(t, "Create", onFailureScheduleEntity)
	mocks.MTransferRepository.AssertCalled(t, "UpdateStatusFailed", transactionId)
}

func setup(t *testing.T) {
	mocks.Setup()
	var err error

	bridgeAccountId, err = hedera.AccountIDFromString(bridgeAccount)
	if err != nil {
		t.Fatalf("Invalid bridge account id [%s]. Error: [%s]", bridgeAccount, err)
	}

	receiverAccountId, err = hedera.AccountIDFromString(receiver)
	if err != nil {
		t.Fatalf("Invalid receiver account id [%s]. Error: [%s]", receiver, err)
	}

	handler = &Handler{
		bridgeAccountId,
		mocks.MTransferRepository,
		mocks.MScheduleRepository,
		mocks.MScheduledService,
		mocks.MTransferService,
		config.GetLoggerFor("Hedera Native Scheduled Nft Transfer Handler"),
	}

}

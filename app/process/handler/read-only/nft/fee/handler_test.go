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

package fee

import (
	"database/sql"
	"errors"
	"github.com/hashgraph/hedera-sdk-go/v2"
	mirrorNodeErr "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/error"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/transaction"
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/schedule"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	testConstants "github.com/limechain/hedera-eth-bridge-validator/test/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"strconv"
	"testing"
	"time"
)

var (
	handler            *Handler
	bridgeAccountAsStr = "0.0.111111"
	bridgeAccount      hedera.AccountID
	transactionId      = "1234"
	sourceChainId      = constants.HederaNetworkId
	targetChainId      = testConstants.EthereumNetworkId
	nativeChainId      = constants.HederaNetworkId
	sourceAsset        = testConstants.NetworkHederaNonFungibleNativeToken
	targetAsset        = testConstants.NetworkEthereumNFTWrappedTokenForNetworkHedera
	nativeAsset        = testConstants.NetworkHederaNonFungibleNativeToken
	receiver           = "0.0.455300"
	amount             = "1000000000000000000"
	serialNum          = int64(123)
	metadata           = "SomeMetadata"
	fee                = "10000"
	isNft              = true
	timestamp          = time.Now().UTC().String()
	entityStatus       = status.Initial
	nilErr             error

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

	entityTransfer = &entity.Transfer{
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
	validatorAccountIdsAsStr = []string{"0.0.26306401", "0.0.26306402", "0.0.26306403"}
	countOfValidators        = int64(len(validatorAccountIdsAsStr))
	validatorAccountIds      = make([]hedera.AccountID, countOfValidators)
	hederaFeeForSourceAsset  = testConstants.HederaNftFees[sourceAsset]
	validFee                 = (hederaFeeForSourceAsset / countOfValidators) * countOfValidators
	formattedValidFee        = strconv.FormatInt(validFee, 10)
	feePerValidator          = validFee / countOfValidators
	hederaTransfers          = make([]model.Hedera, countOfValidators)
	splitTransfers           = make([][]model.Hedera, 1)
	scheduleId               = "33333"

	scheduleEntity = &entity.Schedule{
		TransactionID: transactionId,
		ScheduleID:    scheduleId,
		HasReceiver:   false,
		Operation:     schedule.TRANSFER,
		Status:        status.Completed,
		TransferID:    sql.NullString{String: transactionId, Valid: true},
	}
	feeEntity = &entity.Fee{
		TransactionID: transactionId,
		ScheduleID:    scheduleId,
		Amount:        strconv.FormatInt(-validFee, 10),
		Status:        status.Completed,
		TransferID:    sql.NullString{String: transactionId, Valid: true},
	}
)

func Test_NewHandler(t *testing.T) {
	setup(t, false)

	actualHandler := NewHandler(
		mocks.MTransferRepository,
		mocks.MFeeRepository,
		mocks.MScheduleRepository,
		mocks.MHederaMirrorClient,
		bridgeAccountAsStr,
		mocks.MDistributorService,
		mocks.MTransferService,
		testConstants.HederaNftFees,
		mocks.MReadOnlyService,
	)

	assert.Equal(t, handler, actualHandler)
}

func Test_NewHandler_ErrOnBridgeAccountParse(t *testing.T) {
	setup(t, false)

	defer func() { log.StandardLogger().ExitFunc = nil }()
	fatal := false
	log.StandardLogger().ExitFunc = func(int) { fatal = true }

	_ = NewHandler(
		mocks.MTransferRepository,
		mocks.MFeeRepository,
		mocks.MScheduleRepository,
		mocks.MHederaMirrorClient,
		"invalid account",
		mocks.MDistributorService,
		mocks.MTransferService,
		testConstants.HederaNftFees,
		mocks.MReadOnlyService,
	)

	assert.Equal(t, true, fatal)
}

func Test_Handle(t *testing.T) {
	setup(t, true)

	mocks.MTransferService.On("InitiateNewTransfer", *payload).Return(entityTransfer, nil)
	mocks.MDistributorService.On("ValidAmount", hederaFeeForSourceAsset).Return(validFee)
	mocks.MTransferRepository.On("UpdateFee", transactionId, formattedValidFee).Return(nil)
	mocks.MDistributorService.On("CalculateMemberDistribution", validFee).Return(hederaTransfers, nilErr)
	mocks.MReadOnlyService.On("FindAssetTransfer", transactionId, constants.Hbar, splitTransfers[0], mock.Anything, mock.Anything)

	handler.Handle(payload)

	mocks.MTransferService.AssertCalled(t, "InitiateNewTransfer", *payload)
	mocks.MDistributorService.AssertCalled(t, "ValidAmount", hederaFeeForSourceAsset)
	mocks.MTransferRepository.AssertCalled(t, "UpdateFee", transactionId, formattedValidFee)
	mocks.MDistributorService.AssertCalled(t, "CalculateMemberDistribution", validFee)
	mocks.MReadOnlyService.AssertCalled(t, "FindAssetTransfer", transactionId, constants.Hbar, splitTransfers[0], mock.Anything, mock.Anything)
}

func Test_Handle_ErrOnCast(t *testing.T) {
	setup(t, true)
	brokenPayload := "not a transfer"

	handler.Handle(brokenPayload)

	mocks.MTransferService.AssertNotCalled(t, "InitiateNewTransfer", *payload)
	mocks.MDistributorService.AssertNotCalled(t, "ValidAmount", hederaFeeForSourceAsset)
	mocks.MTransferRepository.AssertNotCalled(t, "UpdateFee", transactionId, formattedValidFee)
	mocks.MDistributorService.AssertNotCalled(t, "CalculateMemberDistribution", validFee)
	mocks.MReadOnlyService.AssertNotCalled(t, "FindAssetTransfer", transactionId, constants.Hbar, splitTransfers[0], mock.Anything, mock.Anything)
}

func Test_Handle_ErrOnTransactionRecord(t *testing.T) {
	setup(t, true)
	var nilTransfer *entity.Transfer
	mocks.MTransferService.On("InitiateNewTransfer", *payload).Return(nilTransfer, errors.New("failed to create transaction record"))

	handler.Handle(payload)

	mocks.MTransferService.AssertCalled(t, "InitiateNewTransfer", *payload)
	mocks.MDistributorService.AssertNotCalled(t, "ValidAmount", hederaFeeForSourceAsset)
	mocks.MTransferRepository.AssertNotCalled(t, "UpdateFee", transactionId, formattedValidFee)
	mocks.MDistributorService.AssertNotCalled(t, "CalculateMemberDistribution", validFee)
	mocks.MReadOnlyService.AssertNotCalled(t, "FindAssetTransfer", transactionId, constants.Hbar, splitTransfers[0], mock.Anything, mock.Anything)
}

func Test_Handle_TransactionRecordNotInitialStatus(t *testing.T) {
	setup(t, true)
	entityTransfer.Status = status.Completed
	mocks.MTransferService.On("InitiateNewTransfer", *payload).Return(entityTransfer, nilErr)

	handler.Handle(payload)

	entityTransfer.Status = entityStatus

	mocks.MTransferService.AssertCalled(t, "InitiateNewTransfer", *payload)
	mocks.MDistributorService.AssertNotCalled(t, "ValidAmount", hederaFeeForSourceAsset)
	mocks.MTransferRepository.AssertNotCalled(t, "UpdateFee", transactionId, formattedValidFee)
	mocks.MDistributorService.AssertNotCalled(t, "CalculateMemberDistribution", validFee)
	mocks.MReadOnlyService.AssertNotCalled(t, "FindAssetTransfer", transactionId, constants.Hbar, splitTransfers[0], mock.Anything, mock.Anything)
}

func Test_Handle_ErrOnUpdateFee(t *testing.T) {
	setup(t, true)

	mocks.MTransferService.On("InitiateNewTransfer", *payload).Return(entityTransfer, nil)
	mocks.MDistributorService.On("ValidAmount", hederaFeeForSourceAsset).Return(validFee)
	mocks.MTransferRepository.On("UpdateFee", transactionId, formattedValidFee).Return(errors.New("failed to create transaction record"))

	handler.Handle(payload)

	mocks.MTransferService.AssertCalled(t, "InitiateNewTransfer", *payload)
	mocks.MDistributorService.AssertCalled(t, "ValidAmount", hederaFeeForSourceAsset)
	mocks.MTransferRepository.AssertCalled(t, "UpdateFee", transactionId, formattedValidFee)
	mocks.MDistributorService.AssertNotCalled(t, "CalculateMemberDistribution", validFee)
	mocks.MReadOnlyService.AssertNotCalled(t, "FindAssetTransfer", transactionId, constants.Hbar, splitTransfers[0], mock.Anything, mock.Anything)
}

func Test_fetch(t *testing.T) {
	setup(t, false)
	expectedResponse := &transaction.Response{
		Transactions: []transaction.Transaction{},
		Status:       mirrorNodeErr.Status{},
	}
	mocks.MHederaMirrorClient.On(
		"GetAccountDebitTransactionsAfterTimestampString",
		bridgeAccount,
		payload.Timestamp,
	).Return(expectedResponse, nilErr)

	actualResponse, err := handler.fetch(payload)

	assert.Nil(t, err)
	assert.Equal(t, expectedResponse, actualResponse)
	mocks.MHederaMirrorClient.AssertCalled(t, "GetAccountDebitTransactionsAfterTimestampString", bridgeAccount, payload.Timestamp)
}

func Test_save(t *testing.T) {
	setup(t, false)
	mocks.MScheduleRepository.On("Create", scheduleEntity).Return(nilErr)
	mocks.MFeeRepository.On("Create", feeEntity).Return(nilErr)

	err := handler.save(transactionId, scheduleId, status.Completed, payload, -validFee)

	assert.Nil(t, err)
	mocks.MScheduleRepository.AssertCalled(t, "Create", scheduleEntity)
	mocks.MFeeRepository.AssertCalled(t, "Create", feeEntity)
}

func Test_save_ErrOnCreateScheduleEntity(t *testing.T) {
	setup(t, false)
	expectedErr := errors.New("failed to create schedule record")
	mocks.MScheduleRepository.On("Create", scheduleEntity).Return(expectedErr)

	err := handler.save(transactionId, scheduleId, status.Completed, payload, -validFee)

	assert.Equal(t, expectedErr, err)
	mocks.MScheduleRepository.AssertCalled(t, "Create", scheduleEntity)
	mocks.MFeeRepository.AssertNotCalled(t, "Create", feeEntity)
}

func Test_save_ErrOnCreateFeeEntity(t *testing.T) {
	setup(t, false)
	expectedErr := errors.New("failed to create fee record")
	mocks.MScheduleRepository.On("Create", scheduleEntity).Return(nilErr)
	mocks.MFeeRepository.On("Create", feeEntity).Return(expectedErr)

	err := handler.save(transactionId, scheduleId, status.Completed, payload, -validFee)

	assert.Equal(t, expectedErr, err)
	mocks.MScheduleRepository.AssertCalled(t, "Create", scheduleEntity)
	mocks.MFeeRepository.AssertCalled(t, "Create", feeEntity)
}

func setup(t *testing.T, setupForHandle bool) {
	mocks.Setup()

	var err error
	bridgeAccount, err = hedera.AccountIDFromString(bridgeAccountAsStr)
	if err != nil {
		t.Fatalf("Invalid account id [%s]. Error: [%s]", bridgeAccount, err)
	}
	splitTransfers[0] = make([]model.Hedera, countOfValidators+1)
	if setupForHandle {
		for index, accountIdStr := range validatorAccountIdsAsStr {
			accountId, err := hedera.AccountIDFromString(accountIdStr)
			if err != nil {
				t.Fatalf("Invalid account id [%s]. Error: [%s]", accountIdStr, err)
			}
			validatorAccountIds[index] = accountId
			transferPerValidator := model.Hedera{AccountID: accountId, Amount: feePerValidator}
			hederaTransfers[index] = transferPerValidator
			splitTransfers[0][index] = transferPerValidator
		}
		splitTransfers[0][len(splitTransfers[0])-1] = model.Hedera{AccountID: bridgeAccount, Amount: -validFee}
	}

	handler = &Handler{
		transferRepository: mocks.MTransferRepository,
		feeRepository:      mocks.MFeeRepository,
		scheduleRepository: mocks.MScheduleRepository,
		mirrorNode:         mocks.MHederaMirrorClient,
		bridgeAccount:      bridgeAccount,
		distributor:        mocks.MDistributorService,
		transfersService:   mocks.MTransferService,
		readOnlyService:    mocks.MReadOnlyService,
		hederaNftFees:      testConstants.HederaNftFees,
		logger:             config.GetLoggerFor("Hedera Transfer and Topic Submission Read-only Handler"),
	}
}

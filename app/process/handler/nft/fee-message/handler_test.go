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

package fee_message

import (
	"errors"
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	testConstants "github.com/limechain/hedera-eth-bridge-validator/test/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	handler       *Handler
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
)

func Test_NewHandler(t *testing.T) {
	setup()

	actualHandler := NewHandler(mocks.MTransferService)

	assert.Equal(t, handler, actualHandler)
}

func Test_Handle(t *testing.T) {
	setup()

	mocks.MTransferService.On("InitiateNewTransfer", *payload).Return(resultEntityTransfer, nilErr)
	mocks.MTransferService.On("ProcessNativeNftTransfer", *payload).Return(nilErr)

	handler.Handle(payload)

	mocks.MTransferService.AssertCalled(t, "InitiateNewTransfer", *payload)
	mocks.MTransferService.AssertCalled(t, "ProcessNativeNftTransfer", *payload)
}

func Test_Handle_CastError(t *testing.T) {
	setup()
	brokenPayload := "Not a transfer"

	handler.Handle(brokenPayload)

	mocks.MTransferService.AssertNotCalled(t, "InitiateNewTransfer", *payload)
	mocks.MTransferService.AssertNotCalled(t, "ProcessNativeNftTransfer", *payload)
}

func Test_Handle_TransactionError(t *testing.T) {
	setup()

	mocks.MTransferService.On("InitiateNewTransfer", *payload).Return(resultEntityTransfer, errors.New("failed to create record"))

	handler.Handle(payload)

	mocks.MTransferService.AssertCalled(t, "InitiateNewTransfer", *payload)
	mocks.MTransferService.AssertNotCalled(t, "ProcessNativeNftTransfer", *payload)
}

func Test_Handle_ProcessNativeNftTransferError(t *testing.T) {
	setup()

	mocks.MTransferService.On("InitiateNewTransfer", *payload).Return(resultEntityTransfer, nilErr)
	mocks.MTransferService.On("ProcessNativeNftTransfer", *payload).Return(errors.New("failed to process native NFT transfer"))

	handler.Handle(payload)

	mocks.MTransferService.AssertCalled(t, "InitiateNewTransfer", *payload)
	mocks.MTransferService.AssertCalled(t, "ProcessNativeNftTransfer", *payload)
}

func Test_Handle_NotInitial(t *testing.T) {
	setup()

	resultEntityTransfer.Status = status.Submitted
	mocks.MTransferService.On("InitiateNewTransfer", *payload).Return(resultEntityTransfer, nilErr)
	mocks.MTransferService.On("ProcessNativeNftTransfer", *payload).Return(nilErr)

	handler.Handle(payload)

	resultEntityTransfer.Status = entityStatus

	mocks.MTransferService.AssertCalled(t, "InitiateNewTransfer", *payload)
	mocks.MTransferService.AssertNotCalled(t, "ProcessNativeNftTransfer", *payload)

}

func setup() {
	mocks.Setup()

	handler = &Handler{
		transfersService: mocks.MTransferService,
		logger:           config.GetLoggerFor("Hedera Transfer and Topic Submission Handler"),
	}

}

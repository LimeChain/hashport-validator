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

package fee_transfer

import (
	"errors"
	"github.com/hashgraph/hedera-sdk-go/v2"
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

var (
	h  *Handler
	tr = &model.Transfer{
		TransactionId: "some-tx-id",
		SourceChainId: 0,
		TargetChainId: 1,
		NativeChainId: 0,
		SourceAsset:   constants.Hbar,
		TargetAsset:   "0xsomeethaddress",
		NativeAsset:   constants.Hbar,
		Receiver:      "0xsomeotherethaddress",
		Amount:        "100",
		Timestamp:     "1",
	}
	accountId = hedera.AccountID{
		Shard:   0,
		Realm:   0,
		Account: 1,
	}
)

func Test_NewHandler(t *testing.T) {
	setup()
	assert.Equal(t, h, NewHandler(mocks.MFeeRepository, mocks.MScheduleRepository, mocks.MHederaMirrorClient, accountId.String(), mocks.MDistributorService, mocks.MFeeService, mocks.MTransferService, mocks.MReadOnlyService))
}

func Test_Handle(t *testing.T) {
	setup()
	tr := &entity.Transfer{
		TransactionID: "some-txn-id",
		SourceChainID: 0,
		TargetChainID: 1,
		NativeChainID: 0,
		SourceAsset:   constants.Hbar,
		TargetAsset:   "0xethaddress",
		NativeAsset:   constants.Hbar,
		Receiver:      "0xsomeethreceiver",
		Amount:        "100",
		Status:        status.Initial,
		Messages:      nil,
		Fee:           entity.Fee{},
		Schedules:     nil,
	}
	mocks.MTransferService.On("InitiateNewTransfer", *tr).Return(tr, nil)
	mocks.MFeeService.On("CalculateFee", tr.TargetAsset, int64(100)).Return(int64(10), int64(0))
	mocks.MDistributorService.On("ValidAmount", 10).Return(int64(3))
	mocks.MReadOnlyService.On("FindTransfer", mock.Anything, mock.Anything, mock.Anything)
	h.Handle(tr)
}

func Test_Handle_FindTransfer(t *testing.T) {
	setup()
	mocks.MTransferService.On("InitiateNewTransfer", *tr).Return(&entity.Transfer{Status: status.Initial}, nil)
	mocks.MFeeService.On("CalculateFee", tr.TargetAsset, int64(100)).Return(int64(10), int64(0))
	mocks.MDistributorService.On("ValidAmount", int64(10)).Return(int64(3))
	mocks.MReadOnlyService.On("FindTransfer", mock.Anything, mock.Anything, mock.Anything)
	h.Handle(tr)
}

func Test_Handle_NotInitialFails(t *testing.T) {
	setup()
	mocks.MTransferService.On("InitiateNewTransfer", *tr).Return(&entity.Transfer{Status: "not-initial"}, nil)
	h.Handle(tr)
	mocks.MReadOnlyService.AssertNotCalled(t, "FindTransfer", mock.Anything, mock.Anything, mock.Anything)
	mocks.MFeeService.AssertNotCalled(t, "CalculateFee", mock.Anything, mock.Anything)
	mocks.MDistributorService.AssertNotCalled(t, "ValidAmount", mock.Anything)
}

func Test_Handle_InvalidPayload(t *testing.T) {
	setup()
	h.Handle("invalid-payload")
	mocks.MTransferService.AssertNotCalled(t, "InitiateNewTransfer", *tr)
	mocks.MReadOnlyService.AssertNotCalled(t, "FindTransfer", mock.Anything, mock.Anything, mock.Anything)
	mocks.MFeeService.AssertNotCalled(t, "CalculateFee", mock.Anything, mock.Anything)
	mocks.MDistributorService.AssertNotCalled(t, "ValidAmount", mock.Anything)
}

func Test_Handle_InitiateNewTransferFails(t *testing.T) {
	setup()
	mocks.MTransferService.On("InitiateNewTransfer", *tr).Return(nil, errors.New("some-error"))
	h.Handle(tr)
	mocks.MReadOnlyService.AssertNotCalled(t, "FindTransfer", mock.Anything, mock.Anything, mock.Anything)
	mocks.MFeeService.AssertNotCalled(t, "CalculateFee", mock.Anything, mock.Anything)
	mocks.MDistributorService.AssertNotCalled(t, "ValidAmount", mock.Anything)
}

func setup() {
	mocks.Setup()
	h = &Handler{
		feeRepository:      mocks.MFeeRepository,
		scheduleRepository: mocks.MScheduleRepository,
		mirrorNode:         mocks.MHederaMirrorClient,
		bridgeAccount:      accountId,
		feeService:         mocks.MFeeService,
		distributorService: mocks.MDistributorService,
		transfersService:   mocks.MTransferService,
		readOnlyService:    mocks.MReadOnlyService,
		logger:             config.GetLoggerFor("Hedera Fee and Schedule Transfer Read-only Handler"),
	}
}
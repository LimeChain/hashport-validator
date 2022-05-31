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

package burn

import (
	"errors"
	"testing"

	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/payload"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	h  *Handler
	tr = &payload.Transfer{
		TransactionId:    "some-tx-id",
		SourceChainId:    0,
		TargetChainId:    1,
		NativeChainId:    0,
		SourceAsset:      constants.Hbar,
		TargetAsset:      "0xb083879B1e10C8476802016CB12cd2F25a896691",
		NativeAsset:      constants.Hbar,
		Receiver:         "0xsomeotherethaddress",
		Amount:           "100",
		NetworkTimestamp: "1",
	}
	accountId = hedera.AccountID{
		Shard:   0,
		Realm:   0,
		Account: 1,
	}
)

func Test_NewHandler(t *testing.T) {
	setup()
	assert.Equal(t, h, NewHandler(accountId.String(), mocks.MHederaMirrorClient, mocks.MScheduleRepository, mocks.MTransferService, mocks.MReadOnlyService))
}

func Test_Handle(t *testing.T) {
	setup()
	mocks.MTransferService.On("InitiateNewTransfer", *tr).Return(&entity.Transfer{Status: status.Initial}, nil)
	mocks.MReadOnlyService.On("FindTransfer", mock.Anything, mock.Anything, mock.Anything)
	h.Handle(tr)
}

func Test_Handle_NotInitialFails(t *testing.T) {
	setup()
	mocks.MTransferService.On("InitiateNewTransfer", *tr).Return(&entity.Transfer{Status: "not-initial"}, nil)
	h.Handle(tr)
	mocks.MReadOnlyService.AssertNotCalled(t, "FindTransfer", mock.Anything, mock.Anything, mock.Anything)
}

func Test_Handle_InvalidPayload(t *testing.T) {
	setup()
	h.Handle("invalid-payload")
	mocks.MTransferService.AssertNotCalled(t, "InitiateNewTransfer", *tr)
}

func Test_Handle_InitiateNewTransferFails(t *testing.T) {
	setup()
	mocks.MTransferService.On("InitiateNewTransfer", *tr).Return(nil, errors.New("some-error"))
	h.Handle(tr)
	mocks.MReadOnlyService.AssertNotCalled(t, "FindTransfer", mock.Anything, mock.Anything, mock.Anything)
}

func setup() {
	mocks.Setup()
	h = &Handler{
		bridgeAccount:      accountId,
		transfersService:   mocks.MTransferService,
		scheduleRepository: mocks.MScheduleRepository,
		mirrorNode:         mocks.MHederaMirrorClient,
		readOnlyService:    mocks.MReadOnlyService,
		logger:             config.GetLoggerFor("Hedera Burn and Topic Message Read-only Handler"),
	}
}

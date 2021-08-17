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

package ethereum

import (
	"github.com/hashgraph/hedera-sdk-go/v2"
	burn_event "github.com/limechain/hedera-eth-bridge-validator/app/model/burn-event"
	lock_event "github.com/limechain/hedera-eth-bridge-validator/app/model/lock-event"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

var (
	ethHandler *Handler
)

func Test_NewHandler(t *testing.T) {
	mocks.Setup()
	h := NewHandler(mocks.MBurnService, mocks.MLockService)
	assert.Equal(t, &Handler{
		lockService: mocks.MLockService,
		burnService: mocks.MBurnService,
		logger:      config.GetLoggerFor("EVM Event Handler"),
	}, h)
}

func Test_Handle_Lock(t *testing.T) {
	setup()
	someLockEvent := &lock_event.LockEvent{
		Id:            "",
		Amount:        0,
		Recipient:     hedera.AccountID{},
		NativeAsset:   "",
		WrappedAsset:  "",
		SourceChainId: nil,
		TargetChainId: nil,
	}
	mocks.MLockService.On("ProcessEvent", *someLockEvent).Return()
	ethHandler.Handle(someLockEvent)
	mocks.MLockService.AssertCalled(t, "ProcessEvent", *someLockEvent)
	mocks.MBurnService.AssertNotCalled(t, "ProcessEvent", mock.Anything)
}

func Test_Handle_Burn(t *testing.T) {
	setup()
	someBurnEvent := &burn_event.BurnEvent{
		Id:           "",
		Amount:       0,
		Recipient:    hedera.AccountID{},
		NativeAsset:  "",
		WrappedAsset: "",
	}
	mocks.MBurnService.On("ProcessEvent", *someBurnEvent).Return()
	ethHandler.Handle(someBurnEvent)
	mocks.MBurnService.AssertCalled(t, "ProcessEvent", *someBurnEvent)
	mocks.MLockService.AssertNotCalled(t, "ProcessEvent", mock.Anything)
}

func Test_Handle_InvalidPayload(t *testing.T) {
	setup()
	ethHandler.Handle("this-is-invalid-payload")
	mocks.MBurnService.AssertNotCalled(t, "ProcessEvent", mock.Anything)
	mocks.MLockService.AssertNotCalled(t, "ProcessEvent", mock.Anything)
}

func setup() {
	mocks.Setup()
	ethHandler = &Handler{
		lockService: mocks.MLockService,
		burnService: mocks.MBurnService,
		logger:      config.GetLoggerFor("EVM Event Handler"),
	}
}

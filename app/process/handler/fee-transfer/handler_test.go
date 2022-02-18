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

package fee_transfer

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	feeTransferHandler *Handler
)

func Test_NewHandler(t *testing.T) {
	mocks.Setup()
	h := NewHandler(mocks.MBurnService)
	assert.Equal(t, &Handler{
		burnService: mocks.MBurnService,
		logger:      config.GetLoggerFor("Hedera Fee and Schedule Transfer Handler"),
	}, h)
}

func Test_Handle_Burn(t *testing.T) {
	setup()
	someEvent := &transfer.Transfer{
		TransactionId: "",
		SourceChainId: 0,
		TargetChainId: 0,
		NativeChainId: 0,
		SourceAsset:   "",
		TargetAsset:   "",
		NativeAsset:   "",
		Receiver:      "",
		Amount:        "0",
	}
	mocks.MBurnService.On("ProcessEvent", *someEvent).Return()
	feeTransferHandler.Handle(someEvent)
	mocks.MBurnService.AssertCalled(t, "ProcessEvent", *someEvent)
}

func Test_Handle_Encoding_Fails(t *testing.T) {
	setup()

	invalidTransferPayload := []byte{1, 2, 1}

	feeTransferHandler.Handle(invalidTransferPayload)

	mocks.MBurnService.AssertNotCalled(t, "ProcessEvent")
}

func setup() {
	mocks.Setup()
	feeTransferHandler = &Handler{
		burnService: mocks.MBurnService,
		logger:      config.GetLoggerFor("Hedera Fee and Schedule Transfer Handler"),
	}
}

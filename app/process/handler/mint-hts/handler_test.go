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

package mint_hts

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	mintHtsHandler *Handler
)

func Test_NewHandler(t *testing.T) {
	mocks.Setup()
	h := NewHandler(mocks.MLockService)
	assert.Equal(t, &Handler{
		lockService: mocks.MLockService,
		logger:      config.GetLoggerFor("Hedera Mint and Transfer Handler"),
	}, h)
}

func Test_Handle_Lock(t *testing.T) {
	setup()
	tr := &transfer.Transfer{
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
	mocks.MLockService.On("ProcessEvent", *tr).Return()
	mintHtsHandler.Handle(tr)
	mocks.MLockService.AssertCalled(t, "ProcessEvent", *tr)
}

func Test_Handle_Encoding_Fails(t *testing.T) {
	setup()

	invalidTransferPayload := []byte{1, 2, 1}

	mintHtsHandler.Handle(invalidTransferPayload)

	mocks.MLockService.AssertNotCalled(t, "ProcessEvent")
}

func setup() {
	mocks.Setup()
	mintHtsHandler = &Handler{
		lockService: mocks.MLockService,
		logger:      config.GetLoggerFor("Hedera Mint and Transfer Handler"),
	}
}

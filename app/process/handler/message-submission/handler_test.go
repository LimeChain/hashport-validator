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
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	msHandler *Handler
)

func Test_NewHandler(t *testing.T) {
	mocks.Setup()
	h := NewHandler(nil, nil, nil, mocks.MTransferService, nil, "0.0.1111")
	assert.Equal(t, &Handler{
		logger:           config.GetLoggerFor("Topic Message Submission Handler"),
		transfersService: mocks.MTransferService,
		topicID: hedera.TopicID{
			Shard: 0,
			Realm: 0,
			Topic: 1111,
		},
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
	tr := transfer.Transfer{
		TransactionId: "",
		SourceChainId: 0,
		TargetChainId: 0,
		NativeChainId: 0,
		SourceAsset:   "",
		TargetAsset:   "",
		NativeAsset:   "",
		Receiver:      "",
		Amount:        "0",
		RouterAddress: "",
	}
	mocks.MTransferService.On("InitiateNewTransfer", tr).Return(tr, nil)
	msHandler.Handle(tr)
}

func setup() {
	mocks.Setup()
	msHandler = &Handler{
		transfersService: mocks.MTransferService,
		logger:           config.GetLoggerFor("Hedera Mint and Transfer Handler"),
	}
}

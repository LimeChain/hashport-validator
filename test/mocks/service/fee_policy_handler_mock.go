/*
 * Copyright 2022 LimeChain Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service

import (
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/stretchr/testify/mock"
)

type MockFeePolicyHandler struct {
	mock.Mock
}

func (m MockFeePolicyHandler) ProcessLatestFeePolicyConfig(feePolicyTopicID hedera.TopicID) (*parser.FeePolicy, error) {
	args := m.Called(feePolicyTopicID)
	return args[0].(*parser.FeePolicy), args.Error(1)
}

func (m MockFeePolicyHandler) FeeAmountFor(networkId uint64, account string, token string, amount int64) (feeAmount int64, exist bool) {
	// TODO implement me
	panic("implement me")
}

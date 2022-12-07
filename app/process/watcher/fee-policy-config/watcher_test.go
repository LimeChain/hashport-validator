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

package fee_policy_config

import (
	"errors"
	"testing"
	"time"

	"github.com/hashgraph/hedera-sdk-go/v2"
	qi "github.com/limechain/hedera-eth-bridge-validator/app/domain/queue"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	testConstants "github.com/limechain/hedera-eth-bridge-validator/test/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
)

var (
	watcher *Watcher
	topicId = hedera.TopicID{
		Shard: 0,
		Realm: 0,
		Topic: 42,
	}
	pollingInterval = time.Duration(0)
	nilParser       *parser.FeePolicy
)

func Test_NewWatcher(t *testing.T) {
	setup()

	actualWatcher := NewWatcher(mocks.MBFeePolicyHandler, topicId, pollingInterval)

	assert.Equal(t, watcher, actualWatcher)
}

func Test_watchIteration(t *testing.T) {
	setup()
	mocks.MBFeePolicyHandler.On("ProcessLatestConfig", topicId).Return(&testConstants.ParsedFeePolicyConfig, nil)

	watcher.watchIteration()

	mocks.MBFeePolicyHandler.AssertCalled(t, "ProcessLatestConfig", topicId)
}

func Test_watchIteration_Error(t *testing.T) {
	setup()
	mocks.MBFeePolicyHandler.On("ProcessLatestConfig", topicId).Return(nilParser, errors.New("some error"))

	watcher.watchIteration()

	mocks.MBFeePolicyHandler.On("ProcessLatestConfig", topicId)
}

func Test_Watch(t *testing.T) {
	setup()
	mocks.MBFeePolicyHandler.On("ProcessLatestConfig", topicId).Return(&testConstants.ParsedFeePolicyConfig, nil)

	watcher.Watch(qi.Queue(nil))
}

func setup() {
	mocks.Setup()

	watcher = &Watcher{
		feePolicyHandler: mocks.MBFeePolicyHandler,
		topicID:          topicId,
		logger:           config.GetLoggerFor("Fee Policy Config Watcher"),
	}
}

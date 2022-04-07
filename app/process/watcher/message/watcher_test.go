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

package message

import (
	"errors"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go/v2"
	message2 "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/core/queue"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/message"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"testing"
)

var (
	w       *Watcher
	topicID = hedera.TopicID{
		Shard: 0,
		Realm: 0,
		Topic: 1,
	}
	consensusTimestamp    = "1633633534.108746000"
	milestoneTimestamp, _ = timestamp.FromString(consensusTimestamp)
)

func Test_UpdateStatusTimestamp(t *testing.T) {
	setup()
	mocks.MStatusRepository.On("Update", topicID.String(), int64(1)).Return(nil)
	w.updateStatusTimestamp(1)
}

func Test_ProcessMessage_FromString_Fails(t *testing.T) {
	setup()
	w.processMessage(message2.Message{Contents: "invalid-data"}, mocks.MQueue)
	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything)
}

func Test_NewWatcher(t *testing.T) {
	mocks.Setup()
	mocks.MStatusRepository.On("Get", topicID.String()).Return(int64(0), nil)
	NewWatcher(mocks.MHederaMirrorClient, "0.0.1", mocks.MStatusRepository, 1, 0)
}

func Test_NewWatcher_Get_Error(t *testing.T) {
	mocks.Setup()
	mocks.MStatusRepository.On("Get", topicID.String()).Return(int64(0), gorm.ErrRecordNotFound)
	mocks.MStatusRepository.On("Create", topicID.String(), mock.Anything).Return(nil)
	NewWatcher(mocks.MHederaMirrorClient, "0.0.1", mocks.MStatusRepository, 1, 0)
}

func Test_NewWatcher_WithTS(t *testing.T) {
	mocks.Setup()
	mocks.MStatusRepository.On("Get", topicID.String()).Return(int64(6), nil)
	mocks.MStatusRepository.On("Update", topicID.String(), int64(6)).Return(nil)
	NewWatcher(mocks.MHederaMirrorClient, "0.0.1", mocks.MStatusRepository, 1, 6)
}

func Test_BeginWatch_FailsMessagesRetrieval(t *testing.T) {
	setup()
	mocks.MStatusRepository.On("Get", topicID.String()).Return(int64(5), nil)
	mocks.MHederaMirrorClient.On("GetMessagesAfterTimestamp", topicID, int64(5)).Return([]message2.Message{}, errors.New("some-error"))
	w.beginWatching(mocks.MQueue)

	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything)
	mocks.MStatusRepository.AssertNotCalled(t, "Update", mock.Anything)
}

func Test_BeginWatch_SuccessfulExecution(t *testing.T) {
	m := message2.Message{
		ConsensusTimestamp: consensusTimestamp,
		TopicId:            "0.0.4321",
		Contents: "EIHxBBodMC4wLjE4OTMtMTYzMTI2MDg5MC05NDgyMDg5NDkiKjB4MDg3MkI5RjY1OUYwYjQ" +
			"xNGU1M2ZEYWIyQjY2OThDMzRCYWMxY0I5MCoqMHgwZjJGNjYyM2FDNGI5NGUxZDYxQjRDZD" +
			"E5NUE2YzI4OTkyMzEwRjk2Mgk5MDAwMDAwMDE6ggE0YThiZmNhMmY2MGVkN2M5NDkwZDBhZ" +
			"DNiZWNmODk2YmVjMGYxYmYxZmFiOTlhNWQwMmY4ZjZiYzU1NWZmNTA2NzdiOWRkMWJmOTg4" +
			"OGIxMzZhYjhlMzMzMjE0NjJjMGRkZWNiNWQ5NzE3YTY1OGQxYjYyZTliYTkyY2Q4OTlmYjFj",
		RunningHash: "0xff3",
	}
	payload, _ := message.FromString(m.Contents, m.ConsensusTimestamp)
	queueMessage := &queue.Message{
		Payload: payload,
		Topic:   constants.TopicMessageValidation,
	}

	setup()
	mocks.MStatusRepository.On("Get", topicID.String()).Return(int64(2), nil).Once()
	mocks.MStatusRepository.On("Get", topicID.String()).Return(milestoneTimestamp, nil)
	mocks.MHederaMirrorClient.On("GetMessagesAfterTimestamp", topicID, int64(2)).Return([]message2.Message{m}, nil).Once()
	mocks.MHederaMirrorClient.On("GetMessagesAfterTimestamp", topicID, milestoneTimestamp).Return([]message2.Message{}, errors.New("some-error"))
	mocks.MQueue.On("Push", queueMessage)
	mocks.MStatusRepository.On("Update", topicID.String(), milestoneTimestamp).Return(nil)

	w.beginWatching(mocks.MQueue)

	mocks.MQueue.AssertCalled(t, "Push", queueMessage)
	mocks.MStatusRepository.AssertCalled(t, "Update", topicID.String(), milestoneTimestamp)
}

func setup() {
	mocks.Setup()
	w = &Watcher{
		client:           mocks.MHederaMirrorClient,
		topicID:          topicID,
		statusRepository: mocks.MStatusRepository,
		pollingInterval:  1,
		logger:           config.GetLoggerFor(fmt.Sprintf("[%s] Topic Watcher", topicID)),
	}
}

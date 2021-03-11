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

package consensusmessage

import (
	"errors"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/encoding"
	"github.com/limechain/hedera-eth-bridge-validator/app/process"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/publisher"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-watcher-sdk/queue"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

type Watcher struct {
	nodeClient       client.HederaNode
	topicID          hedera.TopicID
	typeMessage      string
	statusRepository repository.Status
	startTimestamp   int64
	logger           *log.Entry
}

func NewWatcher(nodeClient client.HederaNode, topicID hedera.TopicID, repository repository.Status, startTimestamp int64) *Watcher {
	return &Watcher{
		nodeClient:       nodeClient,
		topicID:          topicID,
		typeMessage:      process.HCSMessageType,
		statusRepository: repository,
		startTimestamp:   startTimestamp,
		logger:           config.GetLoggerFor(fmt.Sprintf("Topic [%s] Watcher", topicID.String())),
	}
}

func (cmw Watcher) Watch(q *queue.Queue) {
	topic := cmw.topicID.String()
	_, err := cmw.statusRepository.GetLastFetchedTimestamp(topic)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			cmw.logger.Debug("No Topic Watcher Timestamp found in DB")
			err := cmw.statusRepository.CreateTimestamp(topic, cmw.startTimestamp)
			if err != nil {
				cmw.logger.Fatalf("Failed to create Topic Watcher Status timestamp. Error %s", err)
			}
		} else {
			cmw.logger.Fatalf("Failed to fetch last Topic Watcher timestamp. Err: %s", err)
		}
	}
	cmw.subscribeToTopic(q)
}

func (cmw Watcher) subscribeToTopic(q *queue.Queue) {
	_, err := hedera.NewTopicMessageQuery().
		SetStartTime(time.Unix(0, cmw.startTimestamp)).
		SetTopicID(cmw.topicID).
		Subscribe(
			cmw.nodeClient.GetClient(),
			func(topicMsg hedera.TopicMessage) {
				cmw.processMessage(topicMsg, q)
			},
		)

	if err != nil {
		cmw.logger.Error("Failed to subscribe to topic")
		return
	}
	cmw.logger.Infof("Subscribed to Topic successfully.")
}

func (cmw Watcher) processMessage(topicMsg hedera.TopicMessage, q *queue.Queue) {
	cmw.logger.Debugf("Received new Message for Topic. Timestamp: [%s] Contents: [%s]", topicMsg.ConsensusTimestamp, topicMsg.Contents)
	messageTimestamp := topicMsg.ConsensusTimestamp.UnixNano()
	msg, err := encoding.NewTopicMessageFromBytesWithTS(topicMsg.Contents, messageTimestamp)
	if err != nil {
		cmw.logger.Errorf("Could not decode incoming message [%s]. Error: [%s]", topicMsg.Contents, err)
		return
	}

	publisher.Publish(msg, cmw.typeMessage, cmw.topicID, q)
	err = cmw.statusRepository.UpdateLastFetchedTimestamp(cmw.topicID.String(), messageTimestamp)
	if err != nil {
		cmw.logger.Errorf("Could not update last fetched timestamp - [%d]", messageTimestamp)
	}
}

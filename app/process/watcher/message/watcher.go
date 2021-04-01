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

package message

import (
	"errors"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/core/pair"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/encoding"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/publisher"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

type Watcher struct {
	nodeClient       client.HederaNode
	topicID          hedera.TopicID
	statusRepository repository.Status
	startTimestamp   int64
	logger           *log.Entry
}

func NewWatcher(nodeClient client.HederaNode, topicID string, repository repository.Status, startTimestamp int64) *Watcher {
	id, err := hedera.TopicIDFromString(topicID)
	if err != nil {
		log.Fatalf("Could not start Consensus Topic Watcher for topic [%s] - Error: [%s]", topicID, err)
	}

	return &Watcher{
		nodeClient:       nodeClient,
		topicID:          id,
		statusRepository: repository,
		startTimestamp:   startTimestamp,
		logger:           config.GetLoggerFor(fmt.Sprintf("[%s] Topic Watcher", topicID)),
	}
}

func (cmw Watcher) Watch(q *pair.Queue) {
	topic := cmw.topicID.String()
	_, err := cmw.statusRepository.GetLastFetchedTimestamp(topic)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err := cmw.statusRepository.CreateTimestamp(topic, cmw.startTimestamp)
			if err != nil {
				cmw.logger.Fatalf("Failed to create Topic Watcher Status timestamp. Error [%s]", err)
			}
			cmw.logger.Tracef("Created new Topic Watcher status timestamp [%s]", timestamp.ToHumanReadable(cmw.startTimestamp))
		} else {
			cmw.logger.Fatalf("Failed to fetch last Topic Watcher timestamp. Error [%s]", err)
		}
	} else {
		cmw.updateStatusTimestamp(cmw.startTimestamp)
	}
	cmw.subscribeToTopic(q)
}

func (cmw Watcher) updateStatusTimestamp(ts int64) {
	err := cmw.statusRepository.UpdateLastFetchedTimestamp(cmw.topicID.String(), ts)
	if err != nil {
		cmw.logger.Fatalf("Failed to update Topic Watcher Status timestamp. Error [%s]", err)
	}
	cmw.logger.Tracef("Updated Topic Watcher timestamp to [%s]", timestamp.ToHumanReadable(ts))
}

func (cmw Watcher) subscribeToTopic(q *pair.Queue) {
	_, err := hedera.NewTopicMessageQuery().
		SetStartTime(time.Unix(0, cmw.startTimestamp)).
		SetTopicID(cmw.topicID).
		Subscribe(
			cmw.nodeClient.GetClient(),
			func(topicMsg hedera.TopicMessage) {
				cmw.processMessage(topicMsg, q)
				cmw.updateStatusTimestamp(topicMsg.ConsensusTimestamp.UnixNano())
			},
		)

	if err != nil {
		cmw.logger.Error("Failed to subscribe to topic")
		return
	}
	cmw.logger.Infof("Subscribed to Messages after Timestamp [%s]", timestamp.ToHumanReadable(cmw.startTimestamp))
}

func (cmw Watcher) processMessage(topicMsg hedera.TopicMessage, q *pair.Queue) {
	cmw.logger.Info("New Message Received")

	messageTimestamp := topicMsg.ConsensusTimestamp.UnixNano()
	msg, err := encoding.NewTopicMessageFromBytesWithTS(topicMsg.Contents, messageTimestamp)
	if err != nil {
		cmw.logger.Errorf("Could not decode incoming message [%s]. Error: [%s]", topicMsg.Contents, err)
		return
	}

	publisher.Publish(msg, q)
}

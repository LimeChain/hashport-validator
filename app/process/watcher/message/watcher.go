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
	mirror_node "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	"github.com/limechain/hedera-eth-bridge-validator/app/core/pair"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/message"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

type Watcher struct {
	client           client.MirrorNode
	topicID          hedera.TopicID
	statusRepository repository.Status
	pollingInterval  time.Duration
	maxRetries       int
	startTimestamp   int64
	logger           *log.Entry
}

func NewWatcher(client client.MirrorNode, topicID string, repository repository.Status, pollingInterval time.Duration, maxRetries int, startTimestamp int64) *Watcher {
	id, err := hedera.TopicIDFromString(topicID)
	if err != nil {
		log.Fatalf("Could not start Consensus Topic Watcher for topic [%s] - Error: [%s]", topicID, err)
	}

	return &Watcher{
		client:           client,
		topicID:          id,
		statusRepository: repository,
		startTimestamp:   startTimestamp,
		pollingInterval:  pollingInterval,
		maxRetries:       maxRetries,
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
	cmw.beginWatching(q)
}

func (cmw Watcher) updateStatusTimestamp(ts int64) {
	err := cmw.statusRepository.UpdateLastFetchedTimestamp(cmw.topicID.String(), ts)
	if err != nil {
		cmw.logger.Fatalf("Failed to update Topic Watcher Status timestamp. Error [%s]", err)
	}
	cmw.logger.Tracef("Updated Topic Watcher timestamp to [%s]", timestamp.ToHumanReadable(ts))
}

func (cmw Watcher) beginWatching(q *pair.Queue) {
	milestoneTimestamp := cmw.startTimestamp

	for {
		messages, err := cmw.client.GetMessagesAfterTimestamp(cmw.topicID, milestoneTimestamp)
		if err != nil {
			cmw.logger.Errorf("Error while retrieving messages from mirror node. Error [%s]", err)
			cmw.restart(q)
			return
		}

		cmw.logger.Tracef("Polling found [%d] Messages", len(messages))

		for _, msg := range messages {
			milestoneTimestamp, err = timestamp.FromString(msg.ConsensusTimestamp)
			if err != nil {
				cmw.logger.Errorf("Unable to parse latest message timestamp. Error - [%s].", err)
				continue
			}
			cmw.processMessage(msg, q)
			cmw.updateStatusTimestamp(milestoneTimestamp)
		}
		time.Sleep(cmw.pollingInterval * time.Second)
	}
}

func (cmw Watcher) processMessage(topicMsg mirror_node.Message, q *pair.Queue) {
	cmw.logger.Info("New Message Received")

	msg, err := message.FromString(topicMsg.Contents, topicMsg.ConsensusTimestamp)
	if err != nil {
		cmw.logger.Errorf("Could not decode incoming message [%s]. Error: [%s]", topicMsg.Contents, err)
		return
	}

	q.Push(&pair.Message{Payload: msg})
}

func (cmw *Watcher) restart(q *pair.Queue) {
	if cmw.maxRetries > 0 {
		cmw.maxRetries--
		cmw.logger.Infof("Watcher is trying to reconnect. Connections left [%d]", cmw.maxRetries)
		go cmw.Watch(q)
		return
	}
	cmw.logger.Errorf("Watcher failed: [Too many retries]")
}

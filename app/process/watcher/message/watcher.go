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
	mirrorNodeMsg "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/core/queue"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	qi "github.com/limechain/hedera-eth-bridge-validator/app/domain/queue"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/message"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

type Watcher struct {
	client           client.MirrorNode
	topicID          hedera.TopicID
	statusRepository repository.Status
	pollingInterval  time.Duration
	logger           *log.Entry
}

func NewWatcher(
	client client.MirrorNode,
	topicID string,
	repository repository.Status,
	pollingInterval time.Duration,
	startTimestamp int64) *Watcher {
	id, err := hedera.TopicIDFromString(topicID)
	if err != nil {
		log.Fatalf("Could not start Consensus Topic Watcher for topic [%s] - Error: [%s]", topicID, err)
	}

	targetTimestamp := time.Now().UnixNano()
	timeStamp := startTimestamp
	if startTimestamp == 0 {
		_, err := repository.Get(topicID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err := repository.Create(topicID, targetTimestamp)
				if err != nil {
					log.Fatalf("Failed to create Transfer Watcher timestamp. Error: [%s]", err)
				}
				log.Tracef("Created new Transfer Watcher timestamp [%s]", timestamp.ToHumanReadable(targetTimestamp))
			} else {
				log.Fatalf("Failed to fetch last Transfer Watcher timestamp. Error: [%s]", err)
			}
		}
	} else {
		err := repository.Update(topicID, timeStamp)
		if err != nil {
			log.Fatalf("Failed to update Transfer Watcher Status timestamp. Error [%s]", err)
		}
		targetTimestamp = timeStamp
		log.Tracef("Updated Transfer Watcher timestamp to [%s]", timestamp.ToHumanReadable(timeStamp))
	}

	return &Watcher{
		client:           client,
		topicID:          id,
		statusRepository: repository,
		pollingInterval:  pollingInterval,
		logger:           config.GetLoggerFor(fmt.Sprintf("[%s] Topic Watcher", topicID)),
	}
}

func (cmw Watcher) Watch(q qi.Queue) {
	if !cmw.client.TopicExists(cmw.topicID) {
		cmw.logger.Errorf("Could not start monitoring topic [%s] - Topic not found.", cmw.topicID.String())
		return
	}

	cmw.beginWatching(q)
}

func (cmw Watcher) updateStatusTimestamp(ts int64) {
	err := cmw.statusRepository.Update(cmw.topicID.String(), ts)
	if err != nil {
		cmw.logger.Fatalf("Failed to update Topic Watcher Status timestamp. Error [%s]", err)
	}
	cmw.logger.Tracef("Updated Topic Watcher timestamp to [%s]", timestamp.ToHumanReadable(ts))
}

func (cmw Watcher) beginWatching(q qi.Queue) {
	milestoneTimestamp, err := cmw.statusRepository.Get(cmw.topicID.String())
	if err != nil {
		cmw.logger.Fatalf("Failed to retrieve Topic Watcher Status timestamp. Error [%s]", err)
	}
	cmw.logger.Infof("Watching for Messages after Timestamp [%s]", timestamp.ToHumanReadable(milestoneTimestamp))

	for {
		messages, err := cmw.client.GetMessagesAfterTimestamp(cmw.topicID, milestoneTimestamp)
		if err != nil {
			cmw.logger.Errorf("Error while retrieving messages from mirror node. Error [%s]", err)
			go cmw.beginWatching(q)
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

func (cmw Watcher) processMessage(topicMsg mirrorNodeMsg.Message, q qi.Queue) {
	cmw.logger.Info("New Message Received")

	msg, err := message.FromString(topicMsg.Contents, topicMsg.ConsensusTimestamp)
	if err != nil {
		cmw.logger.Errorf("Could not decode incoming message [%s]. Error: [%s]", topicMsg.Contents, err)
		return
	}

	q.Push(&queue.Message{Payload: msg, Topic: constants.TopicMessageValidation})
}

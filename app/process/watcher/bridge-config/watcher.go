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

package bridge_config

import (
	"github.com/hashgraph/hedera-sdk-go/v2"
	qi "github.com/limechain/hedera-eth-bridge-validator/app/domain/queue"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
	"time"
)

type Watcher struct {
	svc             service.BridgeConfig
	pollingInterval time.Duration
	topicID         hedera.TopicID
	logger          *log.Entry
}

func NewWatcher(svc service.BridgeConfig, topicID hedera.TopicID) *Watcher {
	return &Watcher{
		svc:     svc,
		topicID: topicID,
		logger:  config.GetLoggerFor("Bridge Config Watcher"),
	}
}

func (w *Watcher) Watch(q qi.Queue) {
	// there will be no handler, so the q is to implement the interface
	go func() {
		for {
			w.watchIteration()
			time.Sleep(w.pollingInterval * time.Second)
		}
	}()
}

func (w *Watcher) watchIteration() {
	w.logger.Infof("Checking for new bridge config ...")
	parsedBridge, err := w.svc.ProcessLatestConfig(w.topicID)

	if err != nil {
		w.logger.Errorf(err.Error())
		return
	}

	if parsedBridge != nil {
		if parsedBridge.PollingInterval != w.pollingInterval {
			w.pollingInterval = parsedBridge.PollingInterval
		}
	}
}

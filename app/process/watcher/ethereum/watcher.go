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

package ethereum

import (
	bridgeContract "github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum/contracts/bridge"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	c "github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-watcher-sdk/queue"
	log "github.com/sirupsen/logrus"
)

type Watcher struct {
	config    config.Ethereum
	contracts service.Contracts
	logger    *log.Entry
}

func NewWatcher(contracts service.Contracts, config config.Ethereum) *Watcher {
	return &Watcher{
		config:    config,
		contracts: contracts,
		logger:    c.GetLoggerFor("Ethereum Watcher"),
	}
}

func (ew *Watcher) Watch(queue *queue.Queue) {
	go ew.listenForEvents(queue)
	log.Infof("Listening for events at contract [%s]", ew.config.BridgeContractAddress)
}

func (ew *Watcher) listenForEvents(q *queue.Queue) {
	events := make(chan *bridgeContract.BridgeBurn)
	sub, err := ew.contracts.WatchBurnEventLogs(nil, events)
	if err != nil {
		log.Errorf("Failed to subscribe for Burn Event Logs for contract address [%s]. Error [%s].", ew.config.BridgeContractAddress, err)
	}

	for {
		select {
		case err := <-sub.Err():
			log.Errorf("Burn Event Logs subscription failed. Error [%s].", err)
			return
		case eventLog := <-events:
			ew.handleLog(eventLog, q)
		}
	}
}

func (ew *Watcher) handleLog(eventLog *bridgeContract.BridgeBurn, q *queue.Queue) {
	log.Infof("New Burn Event Log for [%s], Amount [%s], Receiver Address [%s] has been found.",
		eventLog.Account.Hex(),
		eventLog.Amount.String(),
		eventLog.Receiver)
	// TODO: push to queue with message type, corresponding to ETH Handler
}

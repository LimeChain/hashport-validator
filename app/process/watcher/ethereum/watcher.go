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
	"errors"
	ethereum2 "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	routerContract "github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/core/pair"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	c "github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

type Watcher struct {
	config    config.Ethereum
	contracts service.Contracts
	ethClient client.Ethereum
	logger    *log.Entry
}

func NewWatcher(contracts service.Contracts, ethClient client.Ethereum, config config.Ethereum) *Watcher {
	return &Watcher{
		config:    config,
		contracts: contracts,
		ethClient: ethClient,
		logger:    c.GetLoggerFor("Ethereum Watcher"),
	}
}

func (ew *Watcher) Watch(queue *pair.Queue) {
	go ew.listenForEvents(queue)
	ew.logger.Infof("Listening for events at contract [%s]", ew.config.RouterContractAddress)
}

func (ew *Watcher) listenForEvents(q *pair.Queue) {
	events := make(chan *routerContract.RouterBurn)
	sub, err := ew.contracts.WatchBurnEventLogs(nil, events)
	if err != nil {
		ew.logger.Errorf("Failed to subscribe for Burn Event Logs for contract address [%s]. Error [%s].", ew.config.RouterContractAddress, err)
	}

	for {
		select {
		case err := <-sub.Err():
			ew.logger.Errorf("Burn Event Logs subscription failed. Error: [%s].", err)
			return
		case eventLog := <-events:
			go ew.handleLog(eventLog, q)
		}
	}
}

func (ew *Watcher) handleLog(eventLog *routerContract.RouterBurn, q *pair.Queue) {
	ew.logger.Infof("New Burn Event Log for [%s], Amount [%s], Receiver Address [%s] has been found.",
		eventLog.Account.Hex(),
		eventLog.Amount.String(),
		eventLog.Receiver)

	if eventLog.Raw.Removed {
		ew.logger.Infof("[%s] Uncle block transaction was removed.", eventLog.Raw.TxHash)
		return
	}

	done, err := ew.ethClient.WaitForConfirmations(eventLog.Raw)
	<-done
	if *err != nil {
		ew.logger.Errorf("[%s] Failed waiting for confirmation before processing. Error: %s", eventLog.Raw.TxHash, *err)
	}

	onSuccess, onRevert, onError := ew.ethTxCallbacks(eventLog.Raw, q)
	ew.ethClient.WaitForTransaction(eventLog.Raw.TxHash.String(), onSuccess, onRevert, onError)
}

func (ew *Watcher) ethTxCallbacks(eventLog types.Log, q *pair.Queue) (onSuccess, onRevert func(), onError func(error)) {
	onSuccess = func() {
		ew.logger.Infof("[%s] Ethereum TX was successfully mined. Processing continues.", eventLog.TxHash)
		// TODO: push to queue with message type, corresponding to ETH Handler
		// TODO: what is the format of the message payload?
		message := &pair.Message{Payload: eventLog.Data}
		q.Push(message)
	}

	onRevert = func() {
		ew.logger.Infof("[%s] Ethereum TX reverted - transaction went into an uncle block.", eventLog.TxHash)
	}

	onError = func(err error) {
		if errors.Is(err, ethereum2.NotFound) {
			ew.logger.Infof("[%s] Ethereum TX went into an uncle block.", eventLog.TxHash)
		}
	}
	return onSuccess, onRevert, onError
}

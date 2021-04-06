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
	"fmt"
	"github.com/hashgraph/hedera-sdk-go/v2"
	routerContract "github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/core/pair"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	burn_event "github.com/limechain/hedera-eth-bridge-validator/app/model/burn-event"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	c "github.com/limechain/hedera-eth-bridge-validator/config"
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
		logger:    c.GetLoggerFor(fmt.Sprintf("Ethereum Router Watcher [%s]", config.RouterContractAddress)),
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
		log.Errorf("Failed to subscribe for Burn Event Logs for contract address [%s]. Error [%s].", ew.config.RouterContractAddress, err)
	}

	for {
		select {
		case err := <-sub.Err():
			ew.logger.Errorf("Burn Event Logs subscription failed. Error: [%s].", err)
			return
		case eventLog := <-events:
			ew.handleLog(eventLog, q)
		}
	}
}

func (ew *Watcher) handleLog(eventLog *routerContract.RouterBurn, q *pair.Queue) {
	eventAccount := string(eventLog.Receiver)
	ew.logger.Infof("[%s] - New Burn Event Log from [%s], with Amount [%s], ServiceFee [%s], Receiver Address [%s] has been found.",
		eventLog.Account.Hex(),
		eventLog.Raw.TxHash.String(),
		eventLog.Amount.String(),
		eventLog.ServiceFee.String(),
		eventAccount)

	recipientAccount, err := hedera.AccountIDFromString(eventAccount)
	if err != nil {
		ew.logger.Errorf("[%s] - Failed to parse acount [%s]. Error [%s]", eventLog.Raw.TxHash, eventAccount, err)
		return
	}
	burnEvent := &burn_event.BurnEvent{
		Amount:    eventLog.Amount.Int64(),
		TxHash:    eventLog.Raw.TxHash.String(),
		Recipient: recipientAccount,
	}

	q.Push(&pair.Message{Payload: burnEvent})
}

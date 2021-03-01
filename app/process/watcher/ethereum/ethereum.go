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
	bridgecontract "github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum/contracts/bridge"
	"github.com/limechain/hedera-eth-bridge-validator/app/process"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/publisher"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/ethereum/bridge"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	c "github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
	"github.com/limechain/hedera-watcher-sdk/queue"
	log "github.com/sirupsen/logrus"
)

type EthWatcher struct {
	config          config.Ethereum
	contractService *bridge.BridgeContractService
	logger          *log.Entry
}

func NewEthereumWatcher(contractService *bridge.BridgeContractService, config config.Ethereum) *EthWatcher {
	return &EthWatcher{
		config:          config,
		contractService: contractService,
		logger:          c.GetLoggerFor("Ethereum Watcher"),
	}
}

func (ew *EthWatcher) Watch(queue *queue.Queue) {
	log.Infof("[Ethereum Watcher] - Start listening for events for contract address [%s].", ew.config.BridgeContractAddress)
	go ew.listenForEvents(queue)
}

func (ew *EthWatcher) listenForEvents(q *queue.Queue) {
	events := make(chan *bridgecontract.BridgeBurn)
	sub, err := ew.contractService.WatchBurnEventLogs(nil, events)
	if err != nil {
		log.Errorf("Failed to subscribe for Burn Event Logs for contract address [%s]. Error [%s].", ew.config.BridgeContractAddress, err)
	}

	for {
		select {
		case err := <-sub.Err():
			log.Errorf("Burn Event Logs subscription failed. Error [%s].", err)
		case eventLog := <-events:
			ew.handleLog(eventLog, q)
		}
	}
}

func (ew *EthWatcher) handleLog(eventLog *bridgecontract.BridgeBurn, q *queue.Queue) {
	receiverAccount := string(eventLog.Receiver)

	log.Infof("New Burn Event Log from [%s], TxID [%s], Amount [%s], Service fee [%s], Receiver Address [%s] has been found.",
		eventLog.Account.Hex(),
		eventLog.Raw.TxHash.String(),
		eventLog.Amount.String(),
		eventLog.ServiceFee.String(),
		receiverAccount)

	_, err := hedera.AccountIDFromString(receiverAccount)
	if err != nil {
		ew.logger.Warnf("[%s] - Failed to parse receiver account [%s]. Error [%s].", eventLog.Account.String(), receiverAccount, err)
		return
	}
	nonce := fmt.Sprintf("%s-%d", eventLog.Raw.TxHash.String(), eventLog.Raw.Index)

	scheduledTransactionMessage := &proto.ScheduledTransactionMessage{
		Amount:    eventLog.Amount.Int64(),
		Recipient: receiverAccount,
		Nonce:     nonce,
	}

	publisher.Publish(scheduledTransactionMessage, process.HCSSheduledTxMessage, nil, q)
}

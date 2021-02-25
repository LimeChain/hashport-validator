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
	"github.com/hashgraph/hedera-sdk-go/v2"
	bridgecontract "github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum/contracts/bridge"
	hederaclient "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/ethereum/bridge"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	c "github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-watcher-sdk/queue"
	log "github.com/sirupsen/logrus"
)

type EthWatcher struct {
	config           config.Ethereum
	contractService  *bridge.BridgeContractService
	hederaClient     *hederaclient.HederaNodeClient
	logger           *log.Entry
	custodialAccount hedera.AccountID
}

func NewEthereumWatcher(contractService *bridge.BridgeContractService, config config.Ethereum, hederaClient *hederaclient.HederaNodeClient) *EthWatcher {
	custodialAccount, err := hedera.AccountIDFromString(config.CustodialAccount)
	if err != nil {
		log.Fatalf("Invalid custodial account: [%s]", config.CustodialAccount)
	}

	return &EthWatcher{
		config:           config,
		contractService:  contractService,
		custodialAccount: custodialAccount,
		hederaClient:     hederaClient,
		logger:           c.GetLoggerFor("Ethereum Watcher"),
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
	log.Infof("New Burn Event Log for [%s], Amount [%s], Service fee [%s], Receiver Address [%s] has been found.",
		eventLog.Account.Hex(),
		eventLog.Amount.String(),
		eventLog.ServiceFee.String(),
		eventLog.Receiver.String())

	recipientAccountID, err := hedera.AccountIDFromString("0.0.2678")
	if err != nil {
		ew.logger.Warnf("[%s] - Failed to parse receiver account [%s]. Error [%s].", eventLog.Account.String(), eventLog.Receiver.String(), err)
		return
	}

	transactionID, err := ew.hederaClient.SubmitScheduledTransaction(eventLog.Amount.Int64(), recipientAccountID, ew.custodialAccount, eventLog.Raw.TxHash.String())
	if err != nil {
		ew.logger.Errorf("Failed to submit scheduled transaction. Error [%s]", err)
		return
	}

	ew.logger.Infof("[%s] - Successfully submitted scheduled transaction for [%s] to receive [%s] tinybars.",
		transactionID.String(), recipientAccountID, eventLog.Amount.String())

	// TODO: push to queue with message type, corresponding to ETH Handler
	// TODO: upon handling, add information to database (use eth tx hash as unique identifier)
	// TODO: submit scheduled transaction
	// TODO: update status and txn id of the scheduled transaction for the corresponding log
	// TODO: query mirror node for the final status (similar to topic submission)
}

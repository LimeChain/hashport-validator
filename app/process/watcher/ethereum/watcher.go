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
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/core/pair"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	hederahelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/hedera"
	burn_event "github.com/limechain/hedera-eth-bridge-validator/app/model/burn-event"
	lock_event "github.com/limechain/hedera-eth-bridge-validator/app/model/lock-event"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	c "github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
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
		logger:    c.GetLoggerFor(fmt.Sprintf("Ethereum Router Watcher [%s]", config.RouterContractAddress)),
	}
}

func (ew *Watcher) Watch(queue *pair.Queue) {
	go ew.listenForEvents(queue)
	ew.logger.Infof("Listening for events at contract [%s]", ew.config.RouterContractAddress)
}

func (ew *Watcher) listenForEvents(q *pair.Queue) {
	burnEvents := make(chan *router.RouterBurn)
	burnSubscription, err := ew.contracts.WatchBurnEventLogs(nil, burnEvents)
	if err != nil {
		ew.logger.Errorf("Failed to subscribe for Burn Event Logs for contract address [%s]. Error [%s].", ew.config.RouterContractAddress, err)
		return
	}

	lockEvents := make(chan *router.RouterLock)
	lockSubscription, err := ew.contracts.WatchLockEventLogs(nil, lockEvents)
	if err != nil {
		ew.logger.Errorf("Failed to subscribe for Lock Event Logs for contract address [%s]. Error [%s].", ew.config.RouterContractAddress, err)
		return
	}

	for {
		select {
		case err := <-burnSubscription.Err():
			ew.logger.Errorf("Burn Event Logs subscription failed. Error: [%s].", err)
			go ew.listenForEvents(q)
			return
		case err := <-lockSubscription.Err():
			ew.logger.Errorf("Lock Event Logs subscription failed. Error: [%s].", err)
			go ew.listenForEvents(q)
			return
		case eventLog := <-burnEvents:
			go ew.handleBurnLog(eventLog, q)
		case eventLog := <-lockEvents:
			go ew.handleLockLog(eventLog, q)
		}
	}
}

func (ew *Watcher) handleBurnLog(eventLog *router.RouterBurn, q *pair.Queue) {
	ew.logger.Debugf("[%s] - New Burn Event Log received. Waiting block confirmations", eventLog.Raw.TxHash)

	if eventLog.Raw.Removed {
		ew.logger.Debugf("[%s] - Uncle block transaction was removed.", eventLog.Raw.TxHash)
		return
	}

	recipientAccount, err := hedera.AccountIDFromBytes(eventLog.Receiver)
	if err != nil {
		ew.logger.Errorf("[%s] - Failed to parse account from bytes [%v]. Error: [%s].", eventLog.Raw.TxHash, eventLog.Receiver, err)
		return
	}

	// TODO: Figure out the mapping
	nativeAsset, err := ew.contracts.ToNative(eventLog.Token)
	if err != nil {
		ew.logger.Errorf("[%s] - Failed to retrieve native asset of [%s]. Error: [%s].", eventLog.Raw.TxHash, eventLog.Token, err)
		return
	}

	if nativeAsset != constants.Hbar && !hederahelper.IsTokenID(nativeAsset) {
		ew.logger.Errorf("[%s] - Invalid Native Token [%s].", eventLog.Raw.TxHash, nativeAsset)
		return
	}

	burnEvent := &burn_event.BurnEvent{
		Amount:       eventLog.Amount.Int64(),
		Id:           fmt.Sprintf("%s-%d", eventLog.Raw.TxHash, eventLog.Raw.Index),
		Recipient:    recipientAccount,
		NativeAsset:  nativeAsset,
		WrappedAsset: eventLog.Token.String(),
	}

	err = ew.ethClient.WaitForConfirmations(eventLog.Raw)
	if err != nil {
		ew.logger.Errorf("[%s] - Failed waiting for confirmation before processing. Error: %s", eventLog.Raw.TxHash, err)
		return
	}

	ew.logger.Infof("[%s] - New Burn Event Log with Amount [%s], Receiver Address [%s] has been found.",
		eventLog.Raw.TxHash.String(),
		eventLog.Amount.String(),
		recipientAccount.String())

	q.Push(&pair.Message{Payload: burnEvent})
}

func (ew *Watcher) handleLockLog(eventLog *router.RouterLock, q *pair.Queue) {
	ew.logger.Debugf("[%s] - New Lock Event Log received. Waiting block confirmations", eventLog.Raw.TxHash)

	if eventLog.Raw.Removed {
		ew.logger.Debugf("[%s] - Uncle block transaction was removed.", eventLog.Raw.TxHash)
		return
	}

	recipientAccount, err := hedera.AccountIDFromBytes(eventLog.Receiver)
	if err != nil {
		ew.logger.Errorf("[%s] - Failed to parse account from bytes [%v]. Error: [%s].", eventLog.Raw.TxHash, eventLog.Receiver, err)
		return
	}

	// TODO: Figure out the mapping
	nativeAsset, err := ew.contracts.ToNative(eventLog.Token)
	if err != nil {
		ew.logger.Errorf("[%s] - Failed to retrieve native asset of [%s]. Error: [%s].", eventLog.Raw.TxHash, eventLog.Token, err)
		return
	}

	if nativeAsset != constants.Hbar && !hederahelper.IsTokenID(nativeAsset) {
		ew.logger.Errorf("[%s] - Invalid Native Token [%s].", eventLog.Raw.TxHash, nativeAsset)
		return
	}

	lockEvent := &lock_event.LockEvent{
		Amount:       eventLog.Amount.Int64(),
		Id:           fmt.Sprintf("%s-%d", eventLog.Raw.TxHash, eventLog.Raw.Index),
		Recipient:    recipientAccount,
		NativeAsset:  nativeAsset,
		WrappedAsset: eventLog.Token.String(),
		ChainId:      eventLog.TargetChain,
	}

	err = ew.ethClient.WaitForConfirmations(eventLog.Raw)
	if err != nil {
		ew.logger.Errorf("[%s] - Failed waiting for confirmation before processing. Error: %s", eventLog.Raw.TxHash, err)
		return
	}

	ew.logger.Infof("[%s] - New Lock Event Log with Amount [%s], Receiver Address [%s] has been found.",
		eventLog.Raw.TxHash.String(),
		eventLog.Amount.String(),
		recipientAccount.String())

	q.Push(&pair.Message{Payload: lockEvent})
}

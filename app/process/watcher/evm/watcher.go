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

package evm

import (
	"fmt"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/core/queue"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	qi "github.com/limechain/hedera-eth-bridge-validator/app/domain/queue"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	hederahelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/hedera"
	burn_event "github.com/limechain/hedera-eth-bridge-validator/app/model/burn-event"
	lock_event "github.com/limechain/hedera-eth-bridge-validator/app/model/lock-event"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	c "github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	log "github.com/sirupsen/logrus"
	"strings"
)

type Watcher struct {
	routerContractAddress string
	contracts             service.Contracts
	evmClient             client.EVM
	logger                *log.Entry
	mappings              config.AssetMappings
}

func NewWatcher(contracts service.Contracts, evmClient client.EVM, mappings c.AssetMappings) *Watcher {
	return &Watcher{
		routerContractAddress: evmClient.GetRouterContractAddress(),
		contracts:             contracts,
		evmClient:             evmClient,
		logger:                c.GetLoggerFor(fmt.Sprintf("EVM Router Watcher [%s]", evmClient.GetRouterContractAddress())),
		mappings:              mappings,
	}
}

func (ew *Watcher) Watch(queue qi.Queue) {
	go ew.listenForEvents(queue)
	ew.logger.Infof("Listening for events at contract [%s]", ew.routerContractAddress)
}

func (ew *Watcher) listenForEvents(q qi.Queue) {
	burnEvents := make(chan *router.RouterBurn)
	burnSubscription, err := ew.contracts.WatchBurnEventLogs(nil, burnEvents)
	if err != nil {
		ew.logger.Errorf("Failed to subscribe for Burn Event Logs for contract address [%s]. Error [%s].", ew.routerContractAddress, err)
		return
	}

	lockEvents := make(chan *router.RouterLock)
	lockSubscription, err := ew.contracts.WatchLockEventLogs(nil, lockEvents)
	if err != nil {
		ew.logger.Errorf("Failed to subscribe for Lock Event Logs for contract address [%s]. Error [%s].", ew.routerContractAddress, err)
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

func (ew *Watcher) handleBurnLog(eventLog *router.RouterBurn, q qi.Queue) {
	ew.logger.Debugf("[%s] - New Burn Event Log received. Waiting block confirmations", eventLog.Raw.TxHash)

	if eventLog.Raw.Removed {
		ew.logger.Debugf("[%s] - Uncle block transaction was removed.", eventLog.Raw.TxHash)
		return
	}

	if len(eventLog.Receiver) == 0 {
		ew.logger.Errorf("[%s] - Empty receiver account.", eventLog.Raw.TxHash)
		return
	}

	recipientAccount, err := hedera.AccountIDFromBytes(eventLog.Receiver)
	if err != nil {
		ew.logger.Errorf("[%s] - Failed to parse account from bytes [%v]. Error: [%s].", eventLog.Raw.TxHash, eventLog.Receiver, err)
		return
	}

	// TODO: Replace with external configuration service. Ask whether ew.evmClient.ChainID() is a correct way of chainID recognition
	nativeAsset := ew.mappings.WrappedToNative[fmt.Sprintf("%d-%s", ew.evmClient.ChainID().Int64(), eventLog.Token.String())]
	if nativeAsset == "" {
		ew.logger.Errorf("[%s] - Failed to retrieve native asset of [%s].", eventLog.Raw.TxHash, eventLog.Token)
		return
	}
	// TODO: refactor
	nativeAsset = strings.Split(nativeAsset, "-")[0]

	if nativeAsset != constants.Hbar && !hederahelper.IsTokenID(nativeAsset) {
		ew.logger.Errorf("[%s] - Invalid Native Token [%s].", eventLog.Raw.TxHash, nativeAsset)
		return
	}

	// TODO: We will need to parse the targetChain as-well in order to be able to transfer from Polygon to Ethereum f.e. not only Polygon to Hedera.
	burnEvent := &burn_event.BurnEvent{
		Amount:       eventLog.Amount.Int64(),
		Id:           fmt.Sprintf("%s-%d", eventLog.Raw.TxHash, eventLog.Raw.Index),
		Recipient:    recipientAccount,
		NativeAsset:  nativeAsset,
		WrappedAsset: eventLog.Token.String(),
	}

	err = ew.evmClient.WaitForConfirmations(eventLog.Raw)
	if err != nil {
		ew.logger.Errorf("[%s] - Failed waiting for confirmation before processing. Error: %s", eventLog.Raw.TxHash, err)
		return
	}

	ew.logger.Infof("[%s] - New Burn Event Log with Amount [%s], Receiver Address [%s] has been found.",
		eventLog.Raw.TxHash.String(),
		eventLog.Amount.String(),
		recipientAccount.String())

	q.Push(&queue.Message{Payload: burnEvent, ChainId: ew.evmClient.ChainID().Int64()})
}

func (ew *Watcher) handleLockLog(eventLog *router.RouterLock, q qi.Queue) {
	ew.logger.Debugf("[%s] - New Lock Event Log received. Waiting block confirmations", eventLog.Raw.TxHash)

	if eventLog.Raw.Removed {
		ew.logger.Errorf("[%s] - Uncle block transaction was removed.", eventLog.Raw.TxHash)
		return
	}

	if len(eventLog.Receiver) == 0 {
		ew.logger.Errorf("[%s] - Empty receiver account.", eventLog.Raw.TxHash)
		return
	}

	recipientAccount, err := hedera.AccountIDFromBytes(eventLog.Receiver)
	if err != nil {
		ew.logger.Errorf("[%s] - Failed to parse account from bytes [%v]. Error: [%s].", eventLog.Raw.TxHash, eventLog.Receiver, err)
		return
	}

	// TODO: Replace with external configuration service
	wrappedAsset := ew.mappings.NativeToWrappedByNetwork[ew.evmClient.ChainID().Int64()].NativeAssets[eventLog.Token.String()][eventLog.TargetChain.Int64()]
	if wrappedAsset == "" {
		ew.logger.Errorf("[%s] - Failed to retrieve native asset of [%s].", eventLog.Raw.TxHash, eventLog.Token)
		return
	}

	// TODO: This must be removed when we want to have support for multiple chains not only Hedera 1:1 EVM chain.
	if wrappedAsset != constants.Hbar && !hederahelper.IsTokenID(wrappedAsset) {
		ew.logger.Errorf("[%s] - Invalid Native Token [%s].", eventLog.Raw.TxHash, wrappedAsset)
		return
	}

	lockEvent := &lock_event.LockEvent{
		Amount:        eventLog.Amount.Int64(),
		Id:            fmt.Sprintf("%s-%d", eventLog.Raw.TxHash, eventLog.Raw.Index),
		Recipient:     recipientAccount,
		NativeAsset:   eventLog.Token.String(),
		WrappedAsset:  wrappedAsset,
		SourceChainId: ew.evmClient.ChainID(),
		TargetChainId: eventLog.TargetChain,
	}

	err = ew.evmClient.WaitForConfirmations(eventLog.Raw)
	if err != nil {
		ew.logger.Errorf("[%s] - Failed waiting for confirmation before processing. Error: %s", eventLog.Raw.TxHash, err)
		return
	}

	ew.logger.Infof("[%s] - New Lock Event Log with Amount [%s], Receiver Address [%s], Source Chain [%d] and Target Chain [%d] has been found.",
		eventLog.Raw.TxHash.String(),
		eventLog.Amount.String(),
		recipientAccount.String(),
		ew.evmClient.ChainID().Int64(),
		eventLog.TargetChain.Int64())

	q.Push(&queue.Message{Payload: lockEvent, ChainId: ew.evmClient.ChainID().Int64()})
}

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
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/core/queue"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	qi "github.com/limechain/hedera-eth-bridge-validator/app/domain/queue"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	c "github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"math/big"
	"strconv"
)

type Watcher struct {
	repository  repository.Status
	contracts   service.Contracts
	evmClient   client.EVM
	logger      *log.Entry
	mappings    c.Assets
	targetBlock uint64
	validator   bool
}

func NewWatcher(
	repository repository.Status,
	contracts service.Contracts,
	evmClient client.EVM,
	mappings c.Assets,
	startBlock int64,
	validator bool) *Watcher {
	targetBlock, err := evmClient.GetClient().BlockNumber(context.Background())
	if err != nil {
		log.Fatalf("Could not retrieve latest block. Error: [%s].", err)
	}

	if startBlock == 0 {
		_, err := repository.GetLastFetchedTimestamp(contracts.Address().String())
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err := repository.CreateTimestamp(contracts.Address().String(), int64(targetBlock))
				if err != nil {
					log.Fatalf("[%s] - Failed to create Transfer Watcher timestamp. Error: [%s]", contracts.Address(), err)
				}
				log.Tracef("[%s] - Created new Transfer Watcher timestamp [%s]", contracts.Address(), timestamp.ToHumanReadable(int64(targetBlock)))
			} else {
				log.Fatalf("[%s] - Failed to fetch last Transfer Watcher timestamp. Error: [%s]", contracts.Address(), err)
			}
		}
	} else {
		err := repository.UpdateLastFetchedTimestamp(contracts.Address().String(), startBlock)
		if err != nil {
			log.Fatalf("[%s] - Failed to update Transfer Watcher Status timestamp. Error [%s]", contracts.Address(), err)
		}
		targetBlock = uint64(startBlock)
		log.Tracef("[%s] - Updated Transfer Watcher timestamp to [%s]", contracts.Address(), timestamp.ToHumanReadable(startBlock))
	}
	return &Watcher{
		repository:  repository,
		contracts:   contracts,
		evmClient:   evmClient,
		logger:      c.GetLoggerFor(fmt.Sprintf("EVM Router Watcher [%s]", contracts.Address())),
		mappings:    mappings,
		targetBlock: targetBlock,
		validator:   validator,
	}
}

func (ew *Watcher) Watch(queue qi.Queue) {
	go ew.listenForEvents(queue)

	ew.processPastLogs(queue)
	ew.logger.Infof("Listening for events at contract [%s]", ew.contracts.Address())
}

func (ew Watcher) processPastLogs(queue qi.Queue) {
	fromBlock, err := ew.repository.GetLastFetchedTimestamp(ew.contracts.Address().String())
	if err != nil {
		ew.logger.Fatalf("Failed to retrieve EVM Watcher Status fromBlock. Error [%s]", err)
		return
	}

	ew.logger.Infof("Processing events from [%d]", fromBlock)

	// TODO: Figure out a way to dynamically get the hash of the event (ABI)
	burnHash := common.HexToHash("97715804dcd62a721835eaba4356dc90eaf6d442a12fe944f01bbf5f8c0b8992")
	lockHash := common.HexToHash("aa3a3bc72b8c754ca6ee8425a5531bafec37569ec012d62d5f682ca909ae06f1")

	topics := [][]common.Hash{
		{
			burnHash,
			lockHash,
		},
	}
	addresses := []common.Address{
		ew.contracts.Address(),
	}
	query := &ethereum.FilterQuery{
		FromBlock: new(big.Int).SetInt64(fromBlock),
		Addresses: addresses,
		Topics:    topics,
	}
	logs, err := ew.evmClient.GetClient().FilterLogs(context.Background(), *query)
	if err != nil {
		ew.logger.Errorf("Failed to to filter logs. Error: [%s]", err)
		return
	}

	for _, log := range logs {
		if len(log.Topics) > 0 {
			if log.Topics[0] == lockHash {
				lock, err := ew.contracts.ParseLockLog(log)
				if err != nil {
					ew.logger.Errorf("Could not parse lock log [%s]. Error [%s].", lock.Raw.TxHash.String(), err)
					continue
				}
				go ew.handleLockLog(lock, queue)
			} else if log.Topics[0] == burnHash {
				burn, err := ew.contracts.ParseBurnLog(log)
				if err != nil {
					ew.logger.Errorf("Could not parse lock log [%s]. Error [%s].", burn.Raw.TxHash.String(), err)
				}
				go ew.handleBurnLog(burn, queue)
			}
		}
	}
}

func (ew *Watcher) listenForEvents(q qi.Queue) {
	burnEvents := make(chan *router.RouterBurn)

	burnSubscription, err := ew.contracts.WatchBurnEventLogs(nil, burnEvents)
	if err != nil {
		ew.logger.Errorf("Failed to subscribe for Burn Event Logs for contract address [%s]. Error [%s].", ew.contracts.Address(), err)
		return
	}

	lockEvents := make(chan *router.RouterLock)
	lockSubscription, err := ew.contracts.WatchLockEventLogs(nil, lockEvents)
	if err != nil {
		ew.logger.Errorf("Failed to subscribe for Lock Event Logs for contract address [%s]. Error [%s].", ew.contracts.Address(), err)
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

	nativeAsset := ew.mappings.WrappedToNative(eventLog.Token.String(), ew.evmClient.ChainID().Int64())
	if nativeAsset == nil {
		ew.logger.Errorf("[%s] - Failed to retrieve native asset of [%s].", eventLog.Raw.TxHash, eventLog.Token)
		return
	}

	targetAsset := nativeAsset.Asset
	// This is the case when you are bridging wrapped to wrapped
	if eventLog.TargetChain.Int64() != nativeAsset.ChainId {
		ew.logger.Errorf("[%s] - Wrapped to Wrapped transfers currently not supported [%s] - [%d] for [%d]", eventLog.Raw.TxHash, nativeAsset.Asset, nativeAsset.ChainId, eventLog.TargetChain.Int64())
		return
	}

	recipientAccount := ""
	var err error
	if eventLog.TargetChain.Int64() == 0 {
		recipient, err := hedera.AccountIDFromBytes(eventLog.Receiver)
		if err != nil {
			ew.logger.Errorf("[%s] - Failed to parse account from bytes [%v]. Error: [%s].", eventLog.Raw.TxHash, eventLog.Receiver, err)
			return
		}
		recipientAccount = recipient.String()
	} else {
		recipientAccount = common.BytesToAddress(eventLog.Receiver).String()
	}

	burnEvent := &transfer.Transfer{
		TransactionId: fmt.Sprintf("%s-%d", eventLog.Raw.TxHash, eventLog.Raw.Index),
		SourceChainId: ew.evmClient.ChainID().Int64(),
		TargetChainId: eventLog.TargetChain.Int64(),
		NativeChainId: nativeAsset.ChainId,
		SourceAsset:   eventLog.Token.String(),
		TargetAsset:   targetAsset,
		NativeAsset:   nativeAsset.Asset,
		Receiver:      recipientAccount,
		Amount:        eventLog.Amount.String(),
		// TODO: set router address
	}

	err = ew.evmClient.WaitForConfirmations(eventLog.Raw)
	if err != nil {
		ew.logger.Errorf("[%s] - Failed waiting for confirmation before processing. Error: [%s]", eventLog.Raw.TxHash, err)
		return
	}

	ew.logger.Infof("[%s] - New Burn Event Log with Amount [%s], Receiver Address [%s] has been found.",
		eventLog.Raw.TxHash.String(),
		eventLog.Amount.String(),
		recipientAccount)

	currentBlockNumber := eventLog.Raw.BlockNumber

	err = ew.repository.UpdateLastFetchedTimestamp(ew.contracts.Address().String(), int64(eventLog.Raw.BlockNumber))
	if err != nil {
		ew.logger.Errorf("[%s] - Failed to update latest processed block. Error: [%s]", eventLog.Raw.TxHash, err)
		return
	}

	if ew.validator && currentBlockNumber >= ew.targetBlock {
		if burnEvent.TargetChainId == 0 {
			q.Push(&queue.Message{Payload: burnEvent, Topic: constants.HederaFeeTransfer})
		} else {
			q.Push(&queue.Message{Payload: burnEvent, Topic: constants.TopicMessageSubmission})
		}
	} else {
		blockTimestamp, err := ew.evmClient.GetBlockTimestamp(big.NewInt(int64(eventLog.Raw.BlockNumber)))
		if err != nil {
			ew.logger.Errorf("[%s] - Failed to retrieve block timestamp. Error: [%s]", eventLog.Raw.TxHash.String(), err)
			return
		}

		burnEvent.Timestamp = strconv.FormatUint(blockTimestamp, 10)
		if burnEvent.TargetChainId == 0 {
			q.Push(&queue.Message{Payload: burnEvent, Topic: constants.ReadOnlyHederaTransfer})
		} else {
			q.Push(&queue.Message{Payload: burnEvent, Topic: constants.ReadOnlyTransferSave})
		}
	}
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

	recipientAccount := ""
	var err error
	if eventLog.TargetChain.Int64() == 0 {
		recipient, err := hedera.AccountIDFromBytes(eventLog.Receiver)
		if err != nil {
			ew.logger.Errorf("[%s] - Failed to parse account from bytes [%v]. Error: [%s].", eventLog.Raw.TxHash, eventLog.Receiver, err)
			return
		}
		recipientAccount = recipient.String()
	} else {
		recipientAccount = common.BytesToAddress(eventLog.Receiver).String()
	}

	// TODO: Replace with external configuration service
	wrappedAsset := ew.mappings.NativeToWrapped(eventLog.Token.String(), ew.evmClient.ChainID().Int64(), eventLog.TargetChain.Int64())
	if wrappedAsset == "" {
		ew.logger.Errorf("[%s] - Failed to retrieve native asset of [%s].", eventLog.Raw.TxHash, eventLog.Token)
		return
	}

	tr := &transfer.Transfer{
		TransactionId: fmt.Sprintf("%s-%d", eventLog.Raw.TxHash, eventLog.Raw.Index),
		SourceChainId: ew.evmClient.ChainID().Int64(),
		TargetChainId: eventLog.TargetChain.Int64(),
		NativeChainId: ew.evmClient.ChainID().Int64(),
		SourceAsset:   eventLog.Token.String(),
		TargetAsset:   wrappedAsset,
		NativeAsset:   eventLog.Token.String(),
		Receiver:      recipientAccount,
		Amount:        eventLog.Amount.String(),
		// TODO: set router address
	}

	err = ew.evmClient.WaitForConfirmations(eventLog.Raw)
	if err != nil {
		ew.logger.Errorf("[%s] - Failed waiting for confirmation before processing. Error: %s", eventLog.Raw.TxHash, err)
		return
	}

	ew.logger.Infof("[%s] - New Lock Event Log with Amount [%s], Receiver Address [%s], Source Chain [%d] and Target Chain [%d] has been found.",
		eventLog.Raw.TxHash.String(),
		eventLog.Amount.String(),
		recipientAccount,
		ew.evmClient.ChainID().Int64(),
		eventLog.TargetChain.Int64())

	currentBlockNumber := eventLog.Raw.BlockNumber

	err = ew.repository.UpdateLastFetchedTimestamp(ew.contracts.Address().String(), int64(eventLog.Raw.BlockNumber))
	if err != nil {
		ew.logger.Errorf("[%s] - Failed to update latest processed block. Error: [%s]", eventLog.Raw.TxHash, err)
		return
	}

	// TODO: Extend for recoverability
	if ew.validator && currentBlockNumber >= ew.targetBlock {
		if tr.TargetChainId == 0 {
			q.Push(&queue.Message{Payload: tr, Topic: constants.HederaMintHtsTransfer})
		} else {
			q.Push(&queue.Message{Payload: tr, Topic: constants.TopicMessageSubmission})
		}
	} else {
		blockTimestamp, err := ew.evmClient.GetBlockTimestamp(big.NewInt(int64(eventLog.Raw.BlockNumber)))
		if err != nil {
			ew.logger.Errorf("[%s] - Failed to retrieve block timestamp. Error [%s]", eventLog.Raw.TxHash.String(), err)
			return
		}
		tr.Timestamp = strconv.FormatUint(blockTimestamp, 10)
		if tr.TargetChainId == 0 {
			q.Push(&queue.Message{Payload: tr, Topic: constants.ReadOnlyHederaMintHtsTransfer})
		} else {
			q.Push(&queue.Message{Payload: tr, Topic: constants.ReadOnlyTransferSave})
		}
	}
}

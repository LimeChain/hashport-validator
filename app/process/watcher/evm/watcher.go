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
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/core/queue"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	qi "github.com/limechain/hedera-eth-bridge-validator/app/domain/queue"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	helper "github.com/limechain/hedera-eth-bridge-validator/app/helper/big-numbers"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	c "github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"math/big"
	"strconv"
	"strings"
	"time"
)

type Watcher struct {
	repository    repository.Status
	contracts     service.Contracts
	evmClient     client.EVM
	logger        *log.Entry
	mappings      c.Assets
	targetBlock   uint64
	sleepDuration time.Duration
	validator     bool
	filterConfig  FilterConfig
}

// Certain node providers (Alchemy, Infura) have a limitation on how many blocks
// eth_getLogs can process at once. For this to be mitigated, a maximum amount of blocks
// is introduced, splitting the request into chunks with a range of N.
// For example, a query for events with a range of 5 000 blocks, will be split into 10 queries, each having
// a range of 500 blocks
const defaultMaxLogsBlocks = int64(500)

// The default polling interval (in seconds) when querying for upcoming events/logs
const defaultSleepDuration = 15 * time.Second

type FilterConfig struct {
	abi               abi.ABI
	topics            [][]common.Hash
	addresses         []common.Address
	burnHash          common.Hash
	lockHash          common.Hash
	burnERC721Hash    common.Hash
	memberUpdatedHash common.Hash
	maxLogsBlocks     int64
}

func NewWatcher(
	repository repository.Status,
	contracts service.Contracts,
	evmClient client.EVM,
	mappings c.Assets,
	startBlock int64,
	validator bool,
	pollingInterval time.Duration,
	maxLogsBlocks int64) *Watcher {
	currentBlock, err := evmClient.BlockNumber(context.Background())
	if err != nil {
		log.Fatalf("Could not retrieve latest block. Error: [%s].", err)
	}
	targetBlock := helper.Max(0, currentBlock-evmClient.BlockConfirmations())

	abi, err := abi.JSON(strings.NewReader(router.RouterABI))
	if err != nil {
		log.Fatalf("Failed to parse router ABI. Error: [%s]", err)
	}

	burnHash := abi.Events["Burn"].ID
	lockHash := abi.Events["Lock"].ID
	memberUpdatedHash := abi.Events["MemberUpdated"].ID
	burnERC721Hash := abi.Events["BurnERC721"].ID

	topics := [][]common.Hash{
		{
			burnHash,
			lockHash,
			memberUpdatedHash,
			burnERC721Hash,
		},
	}

	addresses := []common.Address{
		contracts.Address(),
	}

	if maxLogsBlocks == 0 {
		maxLogsBlocks = defaultMaxLogsBlocks
	}

	filterConfig := FilterConfig{
		abi:               abi,
		topics:            topics,
		addresses:         addresses,
		burnHash:          burnHash,
		lockHash:          lockHash,
		memberUpdatedHash: memberUpdatedHash,
		burnERC721Hash:    burnERC721Hash,
		maxLogsBlocks:     maxLogsBlocks,
	}

	if pollingInterval == 0 {
		pollingInterval = defaultSleepDuration
	} else {
		pollingInterval = pollingInterval * time.Second
	}

	if startBlock == 0 {
		_, err := repository.Get(contracts.Address().String())
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err := repository.Create(contracts.Address().String(), int64(targetBlock))
				if err != nil {
					log.Fatalf("[%s] - Failed to create Transfer Watcher timestamp. Error: [%s]", contracts.Address(), err)
				}
				log.Tracef("[%s] - Created new Transfer Watcher timestamp [%s]", contracts.Address(), timestamp.ToHumanReadable(int64(targetBlock)))
			} else {
				log.Fatalf("[%s] - Failed to fetch last Transfer Watcher timestamp. Error: [%s]", contracts.Address(), err)
			}
		}
	} else {
		err := repository.Update(contracts.Address().String(), startBlock)
		if err != nil {
			log.Fatalf("[%s] - Failed to update Transfer Watcher Status timestamp. Error [%s]", contracts.Address(), err)
		}
		targetBlock = uint64(startBlock)
		log.Tracef("[%s] - Updated Transfer Watcher timestamp to [%s]", contracts.Address(), timestamp.ToHumanReadable(startBlock))
	}
	return &Watcher{
		repository:    repository,
		contracts:     contracts,
		evmClient:     evmClient,
		logger:        c.GetLoggerFor(fmt.Sprintf("EVM Router Watcher [%s]", contracts.Address())),
		mappings:      mappings,
		targetBlock:   targetBlock,
		validator:     validator,
		sleepDuration: pollingInterval,
		filterConfig:  filterConfig,
	}
}

func (ew *Watcher) Watch(queue qi.Queue) {
	go ew.beginWatching(queue)

	ew.logger.Infof("Listening for events at contract [%s]", ew.contracts.Address())
}

func (ew Watcher) beginWatching(queue qi.Queue) {
	fromBlock, err := ew.repository.Get(ew.contracts.Address().String())
	if err != nil {
		ew.logger.Errorf("Failed to retrieve EVM Watcher Status fromBlock. Error: [%s]", err)
		time.Sleep(ew.sleepDuration)
		ew.beginWatching(queue)
		return
	}

	ew.logger.Infof("Processing events from [%d]", fromBlock)

	for {
		fromBlock, err := ew.repository.Get(ew.contracts.Address().String())
		if err != nil {
			ew.logger.Errorf("Failed to retrieve EVM Watcher Status fromBlock. Error: [%s]", err)
			continue
		}

		currentBlock, err := ew.evmClient.BlockNumber(context.Background())
		if err != nil {
			ew.logger.Errorf("Failed to retrieve latest block number. Error [%s]", err)
			time.Sleep(ew.sleepDuration)
			continue
		}

		toBlock := int64(currentBlock - ew.evmClient.BlockConfirmations())
		if fromBlock > toBlock {
			time.Sleep(ew.sleepDuration)
			continue
		}

		if toBlock-fromBlock > ew.filterConfig.maxLogsBlocks {
			toBlock = fromBlock + ew.filterConfig.maxLogsBlocks
		}

		err = ew.processLogs(fromBlock, toBlock, queue)
		if err != nil {
			ew.logger.Errorf("Failed to process logs. Error: [%s].", err)
			time.Sleep(ew.sleepDuration)
			continue
		}

		time.Sleep(ew.sleepDuration)
	}
}

func (ew Watcher) processLogs(fromBlock, endBlock int64, queue qi.Queue) error {
	query := &ethereum.FilterQuery{
		FromBlock: new(big.Int).SetInt64(fromBlock),
		ToBlock:   new(big.Int).SetInt64(endBlock),
		Addresses: ew.filterConfig.addresses,
		Topics:    ew.filterConfig.topics,
	}

	logs, err := ew.evmClient.FilterLogs(context.Background(), *query)
	if err != nil {
		ew.logger.Errorf("Failed to to filter logs. Error: [%s]", err)
		return err
	}

	for _, log := range logs {
		if len(log.Topics) > 0 {
			if log.Topics[0] == ew.filterConfig.lockHash {
				lock, err := ew.contracts.ParseLockLog(log)
				if err != nil {
					ew.logger.Errorf("Could not parse lock log [%s]. Error [%s].", lock.Raw.TxHash.String(), err)
					continue
				}
				ew.handleLockLog(lock, queue)
			} else if log.Topics[0] == ew.filterConfig.burnHash {
				burn, err := ew.contracts.ParseBurnLog(log)
				if err != nil {
					ew.logger.Errorf("Could not parse burn log [%s]. Error [%s].", burn.Raw.TxHash.String(), err)
					continue
				}
				ew.handleBurnLog(burn, queue)
			} else if log.Topics[0] == ew.filterConfig.memberUpdatedHash {
				go ew.contracts.ReloadMembers()
			} else if log.Topics[0] == ew.filterConfig.burnERC721Hash {
				event, err := ew.contracts.ParseBurnERC721Log(log)
				if err != nil {
					ew.logger.Errorf("Could not parse burn ERC-721 log [%s]. Error [%s].", event.Raw.TxHash.String(), err)
					continue
				}
				ew.handleBurnERC721(event, queue)
			}
		}
	}

	// Given that the log filtering boundaries are inclusive,
	// the next time log filtering is done will start from the next block,
	// so that processing of duplicate events does not occur
	blockToBeUpdated := endBlock + 1

	err = ew.repository.Update(ew.contracts.Address().String(), blockToBeUpdated)
	if err != nil {
		ew.logger.Errorf("Failed to update latest processed block [%d]. Error: [%s]", blockToBeUpdated, err)
		return err
	}

	return nil
}

func (ew *Watcher) handleBurnLog(eventLog *router.RouterBurn, q qi.Queue) {
	ew.logger.Debugf("[%s] - New Burn Event Log received.", eventLog.Raw.TxHash)

	if eventLog.Raw.Removed {
		ew.logger.Debugf("[%s] - Uncle block transaction was removed.", eventLog.Raw.TxHash)
		return
	}

	if len(eventLog.Receiver) == 0 {
		ew.logger.Errorf("[%s] - Empty receiver account.", eventLog.Raw.TxHash)
		return
	}

	var chain *big.Int
	chain, e := ew.evmClient.ChainID(context.Background())
	if e != nil {
		ew.logger.Errorf("[%s] - Failed to retrieve chain ID.", eventLog.Raw.TxHash)
		return
	}
	nativeAsset := ew.mappings.WrappedToNative(eventLog.Token.String(), chain.Int64())
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

	properAmount := eventLog.Amount
	if eventLog.TargetChain.Int64() == 0 {
		properAmount, err = ew.contracts.RemoveDecimals(properAmount, eventLog.Token.String())
		if err != nil {
			ew.logger.Errorf("[%s] - Failed to adjust [%s] amount [%s] decimals between chains.", eventLog.Raw.TxHash, eventLog.Token, properAmount)
			return
		}
	}
	if properAmount.Cmp(big.NewInt(0)) == 0 {
		ew.logger.Errorf("[%s] - Insufficient amount provided: Event Amount [%s] and Proper Amount [%s].", eventLog.Raw.TxHash, eventLog.Amount, properAmount)
		return
	}
	if properAmount.Cmp(nativeAsset.MinAmount) < 0 {
		ew.logger.Errorf("[%s] - Transfer Amount [%s] less than Minimum Amount [%s].", eventLog.Raw.TxHash, properAmount, nativeAsset.MinAmount)
		return
	}

	burnEvent := &transfer.Transfer{
		TransactionId: fmt.Sprintf("%s-%d", eventLog.Raw.TxHash, eventLog.Raw.Index),
		SourceChainId: chain.Int64(),
		TargetChainId: eventLog.TargetChain.Int64(),
		NativeChainId: nativeAsset.ChainId,
		SourceAsset:   eventLog.Token.String(),
		TargetAsset:   targetAsset,
		NativeAsset:   nativeAsset.Asset,
		Receiver:      recipientAccount,
		Amount:        properAmount.String(),
	}

	ew.logger.Infof("[%s] - New Burn Event Log with Amount [%s], Receiver Address [%s] has been found.",
		eventLog.Raw.TxHash.String(),
		eventLog.Amount.String(),
		recipientAccount)

	currentBlockNumber := eventLog.Raw.BlockNumber

	if ew.validator && currentBlockNumber >= ew.targetBlock {
		if burnEvent.TargetChainId == 0 {
			q.Push(&queue.Message{Payload: burnEvent, Topic: constants.HederaFeeTransfer})
		} else {
			q.Push(&queue.Message{Payload: burnEvent, Topic: constants.TopicMessageSubmission})
		}
	} else {
		blockTimestamp := ew.evmClient.GetBlockTimestamp(big.NewInt(int64(eventLog.Raw.BlockNumber)))

		burnEvent.Timestamp = strconv.FormatUint(blockTimestamp, 10)
		if burnEvent.TargetChainId == 0 {
			q.Push(&queue.Message{Payload: burnEvent, Topic: constants.ReadOnlyHederaTransfer})
		} else {
			q.Push(&queue.Message{Payload: burnEvent, Topic: constants.ReadOnlyTransferSave})
		}
	}
}

func (ew *Watcher) handleLockLog(eventLog *router.RouterLock, q qi.Queue) {
	ew.logger.Debugf("[%s] - New Lock Event Log received.", eventLog.Raw.TxHash)

	if eventLog.Raw.Removed {
		ew.logger.Errorf("[%s] - Uncle block transaction was removed.", eventLog.Raw.TxHash)
		return
	}

	if len(eventLog.Receiver) == 0 {
		ew.logger.Errorf("[%s] - Empty receiver account.", eventLog.Raw.TxHash)
		return
	}
	var chain *big.Int
	chain, e := ew.evmClient.ChainID(context.Background())
	if e != nil {
		ew.logger.Errorf("[%s] - Failed to retrieve chain ID.", eventLog.Raw.TxHash)
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

	wrappedAsset := ew.mappings.NativeToWrapped(eventLog.Token.String(), chain.Int64(), eventLog.TargetChain.Int64())
	if wrappedAsset == "" {
		ew.logger.Errorf("[%s] - Failed to retrieve native asset of [%s].", eventLog.Raw.TxHash, eventLog.Token)
		return
	}
	nativeAsset := ew.mappings.FungibleNativeAsset(chain.Int64(), eventLog.Token.String())
	if eventLog.Amount.Cmp(nativeAsset.MinAmount) < 0 {
		ew.logger.Errorf("[%s] - Transfer Amount [%s] less than Minimum Amount [%s].", eventLog.Raw.TxHash, eventLog.Amount, nativeAsset.MinAmount)
		return
	}

	properAmount := new(big.Int).Sub(eventLog.Amount, eventLog.ServiceFee)
	if eventLog.TargetChain.Int64() == 0 {
		properAmount, err = ew.contracts.RemoveDecimals(properAmount, eventLog.Token.String())
		if err != nil {
			ew.logger.Errorf("[%s] - Failed to adjust [%s] amount [%s] decimals between chains.", eventLog.Raw.TxHash, eventLog.Token, properAmount)
			return
		}
	}
	if properAmount.Cmp(big.NewInt(0)) == 0 {
		ew.logger.Errorf("[%s] - Insufficient amount provided: Event Amount [%s] and Proper Amount [%s].", eventLog.Raw.TxHash, eventLog.Amount, properAmount)
		return
	}

	tr := &transfer.Transfer{
		TransactionId: fmt.Sprintf("%s-%d", eventLog.Raw.TxHash, eventLog.Raw.Index),
		SourceChainId: chain.Int64(),
		TargetChainId: eventLog.TargetChain.Int64(),
		NativeChainId: chain.Int64(),
		SourceAsset:   eventLog.Token.String(),
		TargetAsset:   wrappedAsset,
		NativeAsset:   eventLog.Token.String(),
		Receiver:      recipientAccount,
		Amount:        properAmount.String(),
	}

	ew.logger.Infof("[%s] - New Lock Event Log with Amount [%s], Receiver Address [%s], Source Chain [%d] and Target Chain [%d] has been found.",
		eventLog.Raw.TxHash.String(),
		properAmount,
		recipientAccount,
		chain.Int64(),
		eventLog.TargetChain.Int64())

	currentBlockNumber := eventLog.Raw.BlockNumber

	if ew.validator && currentBlockNumber >= ew.targetBlock {
		if tr.TargetChainId == 0 {
			q.Push(&queue.Message{Payload: tr, Topic: constants.HederaMintHtsTransfer})
		} else {
			q.Push(&queue.Message{Payload: tr, Topic: constants.TopicMessageSubmission})
		}
	} else {
		blockTimestamp := ew.evmClient.GetBlockTimestamp(big.NewInt(int64(eventLog.Raw.BlockNumber)))

		tr.Timestamp = strconv.FormatUint(blockTimestamp, 10)
		if tr.TargetChainId == 0 {
			q.Push(&queue.Message{Payload: tr, Topic: constants.ReadOnlyHederaMintHtsTransfer})
		} else {
			q.Push(&queue.Message{Payload: tr, Topic: constants.ReadOnlyTransferSave})
		}
	}
}

func (ew *Watcher) handleBurnERC721(eventLog *router.RouterBurnERC721, q qi.Queue) {
	ew.logger.Debugf("[%s] - New Burn ERC-721 Event Log received.", eventLog.Raw.TxHash)

	if eventLog.Raw.Removed {
		ew.logger.Debugf("[%s] - Uncle block transaction was removed.", eventLog.Raw.TxHash)
		return
	}

	if len(eventLog.Receiver) == 0 {
		ew.logger.Errorf("[%s] - Empty receiver account.", eventLog.Raw.TxHash)
		return
	}

	var chain *big.Int
	chain, e := ew.evmClient.ChainID(context.Background())
	if e != nil {
		ew.logger.Errorf("[%s] - Failed to retrieve chain ID.", eventLog.Raw.TxHash)
		return
	}
	nativeAsset := ew.mappings.WrappedToNative(eventLog.WrappedToken.String(), chain.Int64())
	if nativeAsset == nil {
		ew.logger.Errorf("[%s] - Failed to retrieve native asset of [%s].", eventLog.Raw.TxHash, eventLog.WrappedToken)
		return
	}

	targetAsset := nativeAsset.Asset
	// This is the case when you are bridging wrapped to wrapped
	if eventLog.TargetChain.Int64() != nativeAsset.ChainId {
		ew.logger.Errorf("[%s] - Wrapped to Wrapped transfers currently not supported [%s] - [%d] for [%d]", eventLog.Raw.TxHash, nativeAsset.Asset, nativeAsset.ChainId, eventLog.TargetChain.Int64())
		return
	}

	recipientAccount := ""
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

	transfer := &transfer.Transfer{
		TransactionId: fmt.Sprintf("%s-%d", eventLog.Raw.TxHash, eventLog.Raw.Index),
		SourceChainId: chain.Int64(),
		TargetChainId: eventLog.TargetChain.Int64(),
		NativeChainId: nativeAsset.ChainId,
		SourceAsset:   eventLog.WrappedToken.String(),
		TargetAsset:   targetAsset,
		NativeAsset:   nativeAsset.Asset,
		Receiver:      recipientAccount,
		IsNft:         true,
		SerialNum:     eventLog.TokenId.Int64(),
	}

	ew.logger.Infof("[%s] - New ERC-721Burn ERC-721 Event Log with TokenId [%d], Receiver Address [%s] has been found.",
		eventLog.Raw.TxHash.String(),
		eventLog.TokenId.Int64(),
		recipientAccount)

	currentBlockNumber := eventLog.Raw.BlockNumber

	if ew.validator && currentBlockNumber >= ew.targetBlock {
		if transfer.TargetChainId == 0 {
			q.Push(&queue.Message{Payload: transfer, Topic: constants.HederaNftTransfer})
		} else {
			ew.logger.Errorf("[%s] - NFT Transfer to TargetChain different than [%d]. Not supported.", transfer.TransactionId, 0)
			return
		}
	} else {
		blockTimestamp := ew.evmClient.GetBlockTimestamp(big.NewInt(int64(eventLog.Raw.BlockNumber)))

		transfer.Timestamp = strconv.FormatUint(blockTimestamp, 10)
		if transfer.TargetChainId == 0 {
			q.Push(&queue.Message{Payload: transfer, Topic: constants.ReadOnlyHederaUnlockNftTransfer})
		} else {
			ew.logger.Errorf("[%s] - Read-only NFT Transfer to TargetChain different than [%d]. Not supported.", transfer.TransactionId, 0)
			return
		}
	}
}

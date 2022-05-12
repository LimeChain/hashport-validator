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

package evm

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

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
	bigNumbersHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/big-numbers"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/decimal"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/evm"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/metrics"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	c "github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Watcher struct {
	repository repository.Status
	// A unique database identifier, used as a key to track the progress
	// of the given EVM watcher. Given that addresses between different
	// EVM networks might be the same, a concatenation between
	// <chain-id>-<contract-address> removes possible duplication.
	dbIdentifier      string
	contracts         service.Contracts
	prometheusService service.Prometheus
	pricingService    service.Pricing
	evmClient         client.EVM
	logger            *log.Entry
	assetsService     service.Assets
	targetBlock       uint64
	sleepDuration     time.Duration
	validator         bool
	filterConfig      FilterConfig
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
	mintHash          common.Hash
	burnHash          common.Hash
	lockHash          common.Hash
	unlockHash        common.Hash
	burnERC721Hash    common.Hash
	memberUpdatedHash common.Hash
	maxLogsBlocks     int64
}

func NewWatcher(
	repository repository.Status,
	contracts service.Contracts,
	prometheusService service.Prometheus,
	pricingService service.Pricing,
	evmClient client.EVM,
	assetsService service.Assets,
	dbIdentifier string,
	startBlock int64,
	validator bool,
	pollingInterval time.Duration,
	maxLogsBlocks int64) *Watcher {
	currentBlock, err := evmClient.RetryBlockNumber()
	if err != nil {
		log.Fatalf("Could not retrieve latest block. Error: [%s].", err)
	}
	targetBlock := bigNumbersHelper.Max(0, currentBlock-evmClient.BlockConfirmations())

	abi, err := abi.JSON(strings.NewReader(router.RouterABI))
	if err != nil {
		log.Fatalf("Failed to parse router ABI. Error: [%s]", err)
	}

	mintHash := abi.Events["Mint"].ID
	burnHash := abi.Events["Burn"].ID
	lockHash := abi.Events["Lock"].ID
	unlockHash := abi.Events["Unlock"].ID
	memberUpdatedHash := abi.Events["MemberUpdated"].ID
	burnERC721Hash := abi.Events["BurnERC721"].ID

	topics := [][]common.Hash{
		{
			mintHash,
			burnHash,
			lockHash,
			unlockHash,
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
		mintHash:          mintHash,
		burnHash:          burnHash,
		lockHash:          lockHash,
		unlockHash:        unlockHash,
		burnERC721Hash:    burnERC721Hash,
		memberUpdatedHash: memberUpdatedHash,
		maxLogsBlocks:     maxLogsBlocks,
	}

	if pollingInterval == 0 {
		pollingInterval = defaultSleepDuration
	} else {
		pollingInterval = pollingInterval * time.Second
	}

	if startBlock == 0 {
		_, err := repository.Get(dbIdentifier)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err := repository.Create(dbIdentifier, int64(targetBlock))
				if err != nil {
					log.Fatalf("[%s] - Failed to create Transfer Watcher timestamp. Error: [%s]", dbIdentifier, err)
				}
				log.Tracef("[%s] - Created new Transfer Watcher timestamp [%s]", dbIdentifier, timestamp.ToHumanReadable(int64(targetBlock)))
			} else {
				log.Fatalf("[%s] - Failed to fetch last Transfer Watcher timestamp. Error: [%s]", dbIdentifier, err)
			}
		}
	} else {
		err := repository.Update(dbIdentifier, startBlock)
		if err != nil {
			log.Fatalf("[%s] - Failed to update Transfer Watcher Status timestamp. Error [%s]", dbIdentifier, err)
		}
		targetBlock = uint64(startBlock)
		log.Tracef("[%s] - Updated Transfer Watcher timestamp to [%s]", dbIdentifier, timestamp.ToHumanReadable(startBlock))
	}
	return &Watcher{
		repository:        repository,
		dbIdentifier:      dbIdentifier,
		contracts:         contracts,
		prometheusService: prometheusService,
		pricingService:    pricingService,
		evmClient:         evmClient,
		logger:            c.GetLoggerFor(fmt.Sprintf("EVM Router Watcher [%s]", dbIdentifier)),
		assetsService:     assetsService,
		targetBlock:       targetBlock,
		validator:         validator,
		sleepDuration:     pollingInterval,
		filterConfig:      filterConfig,
	}
}

func (ew *Watcher) Watch(queue qi.Queue) {
	go ew.beginWatching(queue)

	ew.logger.Infof("Listening for events at contract [%s]", ew.dbIdentifier)
}

func (ew Watcher) beginWatching(queue qi.Queue) {
	fromBlock, err := ew.repository.Get(ew.dbIdentifier)
	if err != nil {
		ew.logger.Errorf("Failed to retrieve EVM Watcher Status fromBlock. Error: [%s]", err)
		time.Sleep(ew.sleepDuration)
		ew.beginWatching(queue)
		return
	}

	ew.logger.Infof("Processing events from [%d]", fromBlock)

	for {
		fromBlock, err := ew.repository.Get(ew.dbIdentifier)
		if err != nil {
			ew.logger.Errorf("Failed to retrieve EVM Watcher Status fromBlock. Error: [%s]", err)
			continue
		}

		currentBlock, err := ew.evmClient.RetryBlockNumber()
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
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetInt64(fromBlock),
		ToBlock:   new(big.Int).SetInt64(endBlock),
		Addresses: ew.filterConfig.addresses,
		Topics:    ew.filterConfig.topics,
	}

	logs, err := ew.evmClient.RetryFilterLogs(query)
	if err != nil {
		ew.logger.Errorf("Failed to filter logs. Error: [%s]", err)
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
			} else if log.Topics[0] == ew.filterConfig.unlockHash {
				unlock, err := ew.contracts.ParseUnlockLog(log)
				if err != nil {
					ew.logger.Errorf("Could not parse unlock log [%s]. Error [%s].", unlock.Raw.TxHash.String(), err)
					continue
				}
				ew.handleUnlockLog(unlock)
			} else if log.Topics[0] == ew.filterConfig.mintHash {
				mint, err := ew.contracts.ParseMintLog(log)
				if err != nil {
					ew.logger.Errorf("Could not parse mint log [%s]. Error [%s].", mint.Raw.TxHash.String(), err)
					continue
				}
				ew.handleMintLog(mint)
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

	err = ew.repository.Update(ew.dbIdentifier, blockToBeUpdated)
	if err != nil {
		ew.logger.Errorf("Failed to update latest processed block [%d]. Error: [%s]", blockToBeUpdated, err)
		return err
	}

	return nil
}

func (ew *Watcher) handleMintLog(eventLog *router.RouterMint) {
	ew.logger.Infof("[%s] - New Mint Event Log received.", eventLog.Raw.TxHash)

	if eventLog.Raw.Removed {
		ew.logger.Debugf("[%s] - Uncle block transaction was removed.", eventLog.Raw.TxHash)
		return
	}

	transactionId := string(eventLog.TransactionId)
	sourceChainId := eventLog.SourceChain.Uint64()
	targetChainId := ew.evmClient.GetChainID()
	oppositeToken := ew.assetsService.OppositeAsset(sourceChainId, targetChainId, eventLog.Token.String())

	metrics.SetUserGetHisTokens(sourceChainId, targetChainId, oppositeToken, transactionId, ew.prometheusService, ew.logger)
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

	sourceChainId := ew.evmClient.GetChainID()
	nativeAsset := ew.assetsService.WrappedToNative(eventLog.Token.String(), sourceChainId)
	if nativeAsset == nil {
		ew.logger.Errorf("[%s] - Failed to retrieve native asset of [%s].", eventLog.Raw.TxHash, eventLog.Token)
		return
	}

	targetChainId := eventLog.TargetChain.Uint64()
	transactionId := fmt.Sprintf("%s-%d", eventLog.Raw.TxHash, eventLog.Raw.Index)
	token := eventLog.Token.String()

	if ew.prometheusService.GetIsMonitoringEnabled() {
		if targetChainId != constants.HederaNetworkId {
			metrics.CreateMajorityReachedIfNotExists(sourceChainId, targetChainId, token, transactionId, ew.prometheusService, ew.logger)
		} else {
			metrics.CreateFeeTransferredIfNotExists(sourceChainId, targetChainId, token, transactionId, ew.prometheusService, ew.logger)
		}

		metrics.CreateUserGetHisTokensIfNotExists(sourceChainId, targetChainId, token, transactionId, ew.prometheusService, ew.logger)
	}

	// This is the case when you are bridging wrapped to wrapped
	if targetChainId != nativeAsset.ChainId {
		ew.logger.Errorf("[%s] - Wrapped to Wrapped transfers currently not supported [%s] - [%d] for [%d]", eventLog.Raw.TxHash, nativeAsset.Asset, nativeAsset.ChainId, eventLog.TargetChain.Int64())
		return
	}

	recipientAccount := ""
	var err error
	if targetChainId == constants.HederaNetworkId {
		recipient, err := hedera.AccountIDFromBytes(eventLog.Receiver)
		if err != nil {
			ew.logger.Errorf("[%s] - Failed to parse account from bytes [%v]. Error: [%s].", eventLog.Raw.TxHash, eventLog.Receiver, err)
			return
		}
		recipientAccount = recipient.String()
	} else {
		recipientAccount = common.BytesToAddress(eventLog.Receiver).String()
	}

	targetAmount, err := ew.convertTargetAmount(sourceChainId, targetChainId, token, nativeAsset.Asset, eventLog.Amount)
	if err != nil {
		ew.logger.Errorf("[%s] - Failed to convert to target amount. Error: [%s]", eventLog.Raw.TxHash, err)
		return
	}

	tokenPriceInfo, exist := ew.pricingService.GetTokenPriceInfo(targetChainId, nativeAsset.Asset)
	if !exist {
		ew.logger.Errorf("[%s] - Couldn't get price info in USD for asset [%s].", eventLog.Raw.TxHash, nativeAsset.Asset)
		return
	}

	if targetAmount.Cmp(tokenPriceInfo.MinAmountWithFee) < 0 {
		ew.logger.Errorf("[%s] - Transfer Amount [%s] less than Minimum Amount [%s].", eventLog.Raw.TxHash, targetAmount, tokenPriceInfo.MinAmountWithFee)
		return
	}

	blockTimestamp := ew.evmClient.GetBlockTimestamp(big.NewInt(int64(eventLog.Raw.BlockNumber)))
	tx, err := ew.evmClient.WaitForTransaction(eventLog.Raw.TxHash)
	if err != nil {
		ew.logger.Errorf("[%s] - Failed to get transaction receipt. Error: [%s]", eventLog.Raw.TxHash, err)
		return
	}
	originator, err := evm.OriginatorFromTx(tx)
	if err != nil {
		ew.logger.Errorf("[%s] - Failed to get originator. Error: [%s]", eventLog.Raw.TxHash, err)
		return
	}

	burnEvent := &transfer.Transfer{
		TransactionId: transactionId,
		SourceChainId: sourceChainId,
		TargetChainId: targetChainId,
		NativeChainId: nativeAsset.ChainId,
		SourceAsset:   token,
		TargetAsset:   nativeAsset.Asset,
		NativeAsset:   nativeAsset.Asset,
		Receiver:      recipientAccount,
		Amount:        targetAmount.String(),
		Originator:    originator,
		Timestamp:     time.Unix(int64(blockTimestamp), 0),
	}

	ew.logger.Infof("[%s] - New Burn Event Log with Amount [%s], Receiver Address [%s] has been found.",
		eventLog.Raw.TxHash.String(),
		eventLog.Amount.String(),
		recipientAccount)

	currentBlockNumber := eventLog.Raw.BlockNumber

	if ew.validator && currentBlockNumber >= ew.targetBlock {
		if burnEvent.TargetChainId == constants.HederaNetworkId {
			q.Push(&queue.Message{Payload: burnEvent, Topic: constants.HederaFeeTransfer})
		} else {
			q.Push(&queue.Message{Payload: burnEvent, Topic: constants.TopicMessageSubmission})
		}
	} else {
		burnEvent.NetworkTimestamp = strconv.FormatUint(blockTimestamp, 10)
		if burnEvent.TargetChainId == constants.HederaNetworkId {
			q.Push(&queue.Message{Payload: burnEvent, Topic: constants.ReadOnlyHederaTransfer})
		} else {
			q.Push(&queue.Message{Payload: burnEvent, Topic: constants.ReadOnlyTransferSave})
		}
	}
}

func (ew *Watcher) handleLockLog(eventLog *router.RouterLock, q qi.Queue) {
	ew.logger.Debugf("[%s] - New Lock Event Log received.", eventLog.Raw.TxHash)

	transactionId := fmt.Sprintf("%s-%d", eventLog.Raw.TxHash, eventLog.Raw.Index)
	targetChainId := eventLog.TargetChain.Uint64()
	token := eventLog.Token.String()

	if eventLog.Raw.Removed {
		ew.logger.Errorf("[%s] - Uncle block transaction was removed.", eventLog.Raw.TxHash)
		return
	}

	if len(eventLog.Receiver) == 0 {
		ew.logger.Errorf("[%s] - Empty receiver account.", eventLog.Raw.TxHash)
		return
	}

	sourceChainId := ew.evmClient.GetChainID()
	if targetChainId != constants.HederaNetworkId {
		metrics.CreateMajorityReachedIfNotExists(sourceChainId, targetChainId, token, transactionId, ew.prometheusService, ew.logger)
	}
	metrics.CreateUserGetHisTokensIfNotExists(sourceChainId, targetChainId, token, transactionId, ew.prometheusService, ew.logger)

	recipientAccount := ""
	var err error
	if targetChainId == constants.HederaNetworkId {
		recipient, err := hedera.AccountIDFromBytes(eventLog.Receiver)
		if err != nil {
			ew.logger.Errorf("[%s] - Failed to parse account from bytes [%v]. Error: [%s].", eventLog.Raw.TxHash, eventLog.Receiver, err)
			return
		}
		recipientAccount = recipient.String()
	} else {
		recipientAccount = common.BytesToAddress(eventLog.Receiver).String()
	}

	wrappedAsset := ew.assetsService.NativeToWrapped(token, sourceChainId, targetChainId)
	if wrappedAsset == "" {
		ew.logger.Errorf("[%s] - Failed to retrieve wrapped asset of [%s].", eventLog.Raw.TxHash, eventLog.Token)
		return
	}

	amount := new(big.Int).Sub(eventLog.Amount, eventLog.ServiceFee)
	targetAmount, err := ew.convertTargetAmount(sourceChainId, targetChainId, token, wrappedAsset, amount)
	if err != nil {
		ew.logger.Errorf("[%s] - Failed to convert to target amount. Error: [%s]", eventLog.Raw.TxHash, err)
		return
	}

	nativeAsset := ew.assetsService.FungibleNativeAsset(sourceChainId, token)
	tokenPriceInfo, exist := ew.pricingService.GetTokenPriceInfo(sourceChainId, nativeAsset.Asset)
	if !exist {
		ew.logger.Errorf("[%s] - Couldn't get price info in USD for asset [%s].", eventLog.Raw.TxHash, nativeAsset.Asset)
		return
	}

	if eventLog.Amount.Cmp(tokenPriceInfo.MinAmountWithFee) < 0 {
		ew.logger.Errorf("[%s] - Transfer Amount [%s] less than Minimum Amount [%s].", eventLog.Raw.TxHash, eventLog.Amount, tokenPriceInfo.MinAmountWithFee)
		return
	}

	blockTimestamp := ew.evmClient.GetBlockTimestamp(big.NewInt(int64(eventLog.Raw.BlockNumber)))
	tx, err := ew.evmClient.WaitForTransaction(eventLog.Raw.TxHash)
	if err != nil {
		ew.logger.Errorf("[%s] - Failed to get transaction receipt. Error: [%s]", eventLog.Raw.TxHash, err)
		return
	}
	originator, err := evm.OriginatorFromTx(tx)
	if err != nil {
		ew.logger.Errorf("[%s] - Failed to get originator. Error: [%s]", eventLog.Raw.TxHash, err)
		return
	}

	tr := &transfer.Transfer{
		TransactionId: transactionId,
		SourceChainId: sourceChainId,
		TargetChainId: targetChainId,
		NativeChainId: sourceChainId,
		SourceAsset:   token,
		TargetAsset:   wrappedAsset,
		NativeAsset:   token,
		Receiver:      recipientAccount,
		Amount:        targetAmount.String(),
		Originator:    originator,
		Timestamp:     time.Unix(int64(blockTimestamp), 0),
	}

	ew.logger.Infof("[%s] - New Lock Event Log with Amount [%s], Receiver Address [%s], Source Chain [%d] and Target Chain [%d] has been found.",
		eventLog.Raw.TxHash.String(),
		targetAmount,
		recipientAccount,
		sourceChainId,
		eventLog.TargetChain.Int64())

	currentBlockNumber := eventLog.Raw.BlockNumber

	if ew.validator && currentBlockNumber >= ew.targetBlock {
		if tr.TargetChainId == constants.HederaNetworkId {
			q.Push(&queue.Message{Payload: tr, Topic: constants.HederaMintHtsTransfer})
		} else {
			q.Push(&queue.Message{Payload: tr, Topic: constants.TopicMessageSubmission})
		}
	} else {
		tr.NetworkTimestamp = strconv.FormatUint(blockTimestamp, 10)
		if tr.TargetChainId == constants.HederaNetworkId {
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

	sourceChainId := ew.evmClient.GetChainID()
	nativeAsset := ew.assetsService.WrappedToNative(eventLog.WrappedToken.String(), sourceChainId)
	if nativeAsset == nil {
		ew.logger.Errorf("[%s] - Failed to retrieve native asset of [%s].", eventLog.Raw.TxHash, eventLog.WrappedToken)
		return
	}

	targetAsset := nativeAsset.Asset
	// This is the case when you are bridging wrapped to wrapped
	if eventLog.TargetChain.Uint64() != nativeAsset.ChainId {
		ew.logger.Errorf("[%s] - Wrapped to Wrapped transfers currently not supported [%s] - [%d] for [%d]", eventLog.Raw.TxHash, nativeAsset.Asset, nativeAsset.ChainId, eventLog.TargetChain.Int64())
		return
	}

	recipientAccount := ""
	if eventLog.TargetChain.Uint64() == constants.HederaNetworkId {
		recipient, err := hedera.AccountIDFromBytes(eventLog.Receiver)
		if err != nil {
			ew.logger.Errorf("[%s] - Failed to parse account from bytes [%v]. Error: [%s].", eventLog.Raw.TxHash, eventLog.Receiver, err)
			return
		}
		recipientAccount = recipient.String()
	} else {
		recipientAccount = common.BytesToAddress(eventLog.Receiver).String()
	}

	blockTimestamp := ew.evmClient.GetBlockTimestamp(big.NewInt(int64(eventLog.Raw.BlockNumber)))
	tx, err := ew.evmClient.WaitForTransaction(eventLog.Raw.TxHash)
	if err != nil {
		ew.logger.Errorf("[%s] - Failed to get transaction receipt. Error: [%s]", eventLog.Raw.TxHash, err)
		return
	}
	originator, err := evm.OriginatorFromTx(tx)
	if err != nil {
		ew.logger.Errorf("[%s] - Failed to get originator. Error: [%s]", eventLog.Raw.TxHash, err)
		return
	}

	transfer := &transfer.Transfer{
		TransactionId: fmt.Sprintf("%s-%d", eventLog.Raw.TxHash, eventLog.Raw.Index),
		SourceChainId: sourceChainId,
		TargetChainId: eventLog.TargetChain.Uint64(),
		NativeChainId: nativeAsset.ChainId,
		SourceAsset:   eventLog.WrappedToken.String(),
		TargetAsset:   targetAsset,
		NativeAsset:   nativeAsset.Asset,
		Receiver:      recipientAccount,
		IsNft:         true,
		SerialNum:     eventLog.TokenId.Int64(),
		Originator:    originator,
		Timestamp:     time.Unix(int64(blockTimestamp), 0),
	}

	ew.logger.Infof("[%s] - New ERC-721Burn ERC-721 Event Log with TokenId [%d], Receiver Address [%s] has been found.",
		eventLog.Raw.TxHash.String(),
		eventLog.TokenId.Int64(),
		recipientAccount)

	currentBlockNumber := eventLog.Raw.BlockNumber

	if ew.validator && currentBlockNumber >= ew.targetBlock {
		if transfer.TargetChainId == constants.HederaNetworkId {
			q.Push(&queue.Message{Payload: transfer, Topic: constants.HederaNftTransfer})
		} else {
			ew.logger.Errorf("[%s] - NFT Transfer to TargetChain different than [%d]. Not supported.", transfer.TransactionId, constants.HederaNetworkId)
			return
		}
	} else {
		transfer.NetworkTimestamp = strconv.FormatUint(blockTimestamp, 10)
		if transfer.TargetChainId == constants.HederaNetworkId {
			q.Push(&queue.Message{Payload: transfer, Topic: constants.ReadOnlyHederaUnlockNftTransfer})
		} else {
			ew.logger.Errorf("[%s] - Read-only NFT Transfer to TargetChain different than [%d]. Not supported.", transfer.TransactionId, constants.HederaNetworkId)
			return
		}
	}
}

func (ew *Watcher) handleUnlockLog(eventLog *router.RouterUnlock) {
	ew.logger.Debugf("[%s] - New Unlock Event Log received.", eventLog.Raw.TxHash)

	if eventLog.Raw.Removed {
		ew.logger.Errorf("[%s] - Uncle block transaction was removed.", eventLog.Raw.TxHash)
		return
	}

	transactionId := string(eventLog.TransactionId)
	sourceChainId := eventLog.SourceChain.Uint64()
	targetChainId := ew.evmClient.GetChainID()
	oppositeToken := ew.assetsService.OppositeAsset(sourceChainId, targetChainId, eventLog.Token.String())

	metrics.SetUserGetHisTokens(sourceChainId, targetChainId, oppositeToken, transactionId, ew.prometheusService, ew.logger)
}

func (ew *Watcher) convertTargetAmount(sourceChainId, targetChainId uint64, sourceAsset, targetAsset string, amount *big.Int) (*big.Int, error) {
	sourceAssetInfo, exists := ew.assetsService.FungibleAssetInfo(sourceChainId, sourceAsset)
	if !exists {
		return nil, errors.New(fmt.Sprintf("Failed to retrieve fungible asset info of [%s].", sourceAsset))
	}

	targetAssetInfo, exists := ew.assetsService.FungibleAssetInfo(targetChainId, targetAsset)
	if !exists {
		return nil, errors.New(fmt.Sprintf("Failed to retrieve fungible asset info of [%s].", targetAsset))
	}

	targetAmount := decimal.TargetAmount(sourceAssetInfo.Decimals, targetAssetInfo.Decimals, amount)
	if targetAmount.Cmp(big.NewInt(0)) == 0 {
		return nil, errors.New(fmt.Sprintf("Insufficient amount provided: Event Amount [%s] and Target Amount [%s].", amount, targetAmount))
	}

	return targetAmount, nil
}

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

package cryptotransfer

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/limechain/hedera-eth-bridge-validator/app/process/payload"

	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/app/core/queue"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	qi "github.com/limechain/hedera-eth-bridge-validator/app/domain/queue"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/decimal"
	hederaHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/metrics"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/asset"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/limechain/hedera-eth-bridge-validator/app/helper/blacklist"
)

type Watcher struct {
	transfers           service.Transfers
	client              client.MirrorNode
	accountID           hedera.AccountID
	pollingInterval     time.Duration
	statusRepository    repository.Status
	targetTimestamp     int64
	logger              *log.Entry
	contractServices    map[uint64]service.Contracts
	assetsService       service.Assets
	validator           bool
	prometheusService   service.Prometheus
	pricingService      service.Pricing
	blacklistedAccounts []string
}

func NewWatcher(
	transfers service.Transfers,
	client client.MirrorNode,
	accountID string,
	pollingInterval time.Duration,
	repository repository.Status,
	startTimestamp int64,
	contractServices map[uint64]service.Contracts,
	assetsService service.Assets,
	validator bool,
	prometheusService service.Prometheus,
	pricingService service.Pricing,
	blacklistedAccounts []string,
) *Watcher {
	id, err := hedera.AccountIDFromString(accountID)
	if err != nil {
		log.Fatalf("Could not start Crypto Transfer Watcher for account [%s] - Error: [%s]", accountID, err)
	}

	targetTimestamp := time.Now().UnixNano()
	timeStamp := startTimestamp
	if startTimestamp == 0 {
		_, err := repository.Get(accountID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err := repository.Create(accountID, targetTimestamp)
				if err != nil {
					log.Fatalf("Failed to create Transfer Watcher timestamp. Error: [%s]", err)
				}
				log.Tracef("Created new Transfer Watcher timestamp [%s]", timestamp.ToHumanReadable(targetTimestamp))
			} else {
				log.Fatalf("Failed to fetch last Transfer Watcher timestamp. Error: [%s]", err)
			}
		}
	} else {
		err := repository.Update(accountID, timeStamp)
		if err != nil {
			log.Fatalf("Failed to update Transfer Watcher Status timestamp. Error [%s]", err)
		}
		targetTimestamp = timeStamp
		log.Tracef("Updated Transfer Watcher timestamp to [%s]", timestamp.ToHumanReadable(timeStamp))
	}
	instance := &Watcher{
		transfers:           transfers,
		client:              client,
		accountID:           id,
		pollingInterval:     pollingInterval,
		statusRepository:    repository,
		targetTimestamp:     targetTimestamp,
		logger:              config.GetLoggerFor(fmt.Sprintf("[%s] Transfer Watcher", accountID)),
		contractServices:    contractServices,
		assetsService:       assetsService,
		validator:           validator,
		pricingService:      pricingService,
		prometheusService:   prometheusService,
		blacklistedAccounts: blacklistedAccounts,
	}

	return instance

}

func (ctw Watcher) Watch(q qi.Queue) {
	if !ctw.client.AccountExists(ctw.accountID) {
		ctw.logger.Errorf("Could not start monitoring account [%s] - Account not found.", ctw.accountID.String())
		return
	}

	go ctw.beginWatching(q)
}

func (ctw Watcher) updateStatusTimestamp(ts int64) {
	err := ctw.statusRepository.Update(ctw.accountID.String(), ts)
	if err != nil {
		ctw.logger.Fatalf("Failed to update Transfer Watcher Status timestamp. Error [%s]", err)
	}
	ctw.logger.Tracef("Updated Transfer Watcher timestamp to [%s]", timestamp.ToHumanReadable(ts))
}

func (ctw Watcher) beginWatching(q qi.Queue) {
	milestoneTimestamp, err := ctw.statusRepository.Get(ctw.accountID.String())
	if err != nil {
		ctw.logger.Fatalf("Failed to retrieve Transfer Watcher Status timestamp. Error [%s]", err)
	}
	ctw.logger.Infof("Watching for Transfers after Timestamp [%s]", timestamp.ToHumanReadable(milestoneTimestamp))

	for {
		transactions, e := ctw.client.GetAccountCreditTransactionsAfterTimestamp(ctw.accountID, milestoneTimestamp)
		if e != nil {
			ctw.logger.Errorf("Suddenly stopped monitoring account. Error: [%s]", e)
			time.Sleep(ctw.pollingInterval * time.Second)
			go ctw.beginWatching(q)
			return
		}

		ctw.logger.Tracef("Polling found [%d] Transactions", len(transactions.Transactions))
		if len(transactions.Transactions) > 0 {
			for _, tx := range transactions.Transactions {
				go ctw.processTransaction(tx.TransactionID, q)
			}
			var err error
			milestoneTimestamp, err = timestamp.FromString(transactions.Transactions[len(transactions.Transactions)-1].ConsensusTimestamp)
			if err != nil {
				ctw.logger.Errorf("Unable to parse latest transfer timestamp. Error - [%s].", err)
				continue
			}

			ctw.updateStatusTimestamp(milestoneTimestamp)
		}
		time.Sleep(ctw.pollingInterval * time.Second)
	}
}

func (ctw Watcher) processTransaction(txID string, q qi.Queue) {
	ctw.logger.Infof("New Transaction with ID: [%s]", txID)

	// TX like: [HBAR -> WHBAR || HTS -> WHTS || WEVM -> EVM] (Hereda to EVM)
	tx, err := ctw.client.GetSuccessfulTransaction(txID)
	if err != nil {
		ctw.logger.Errorf("[%s] - Failed to get Transaction. Error: [%s]", txID, err)
		return
	}

	blackListError := blacklist.CheckTxForBlacklistedAccounts(ctw.blacklistedAccounts, tx)
	if blackListError != nil {
		ctw.logger.Errorf(blackListError.Error())
		return
	}

	parsedTransfer, err := tx.GetIncomingTransfer(ctw.accountID.String())
	if err != nil {
		ctw.logger.Errorf("[%s] - Could not extract incoming transfer. Error: [%s]", tx.TransactionID, err)
		return
	}
	sourceAsset := parsedTransfer.Asset
	checkResult := ctw.transfers.SanityCheckTransfer(tx)
	if checkResult.Err != nil {
		ctw.logger.Errorf("[%s] - Sanity check failed. Error: [%s]", tx.TransactionID, checkResult.Err)
		return
	}
	targetChainId := checkResult.ChainId

	nativeAsset := &asset.NativeAsset{
		ChainId: constants.HederaNetworkId,
		Asset:   sourceAsset,
	}

	targetChainAsset := ctw.assetsService.NativeToWrapped(sourceAsset, constants.HederaNetworkId, targetChainId)
	if targetChainAsset == "" {
		nativeAsset = ctw.assetsService.WrappedToNative(sourceAsset, constants.HederaNetworkId)
		if nativeAsset == nil {
			ctw.logger.Errorf("[%s] - Could not parse asset [%s] to its target chain correlation", tx.TransactionID, sourceAsset)
			return
		}
		targetChainAsset = nativeAsset.Asset
		if nativeAsset.ChainId != targetChainId {
			ctw.logger.Errorf("[%s] - Wrapped to Wrapped transfers currently not supported [%s] - [%d] for [%d]", tx.TransactionID, nativeAsset.Asset, nativeAsset.ChainId, targetChainId)
			return
		}
	}

	var transferMessage *payload.Transfer
	originator := hederaHelper.OriginatorFromTxId(tx.TransactionID)
	transferMessage, err = ctw.createFungiblePayload(tx.TransactionID, checkResult.EvmAddress, sourceAsset, *nativeAsset, parsedTransfer.Amount, targetChainId, targetChainAsset)

	if err != nil {
		ctw.logger.Errorf("[%s] - Failed to create payload. Error: [%s]", tx.TransactionID, err)
		return
	}

	transactionTimestamp, err := timestamp.FromString(tx.ConsensusTimestamp)
	if err != nil {
		ctw.logger.Errorf("[%s] - Failed to parse consensus timestamp [%s]. Error: [%s]", tx.TransactionID, tx.ConsensusTimestamp, err)
		return
	}

	transferMessage.Timestamp = time.Unix(0, transactionTimestamp)
	transferMessage.Originator = originator

	topic := ""
	if ctw.validator && transactionTimestamp > ctw.targetTimestamp {
		if nativeAsset.ChainId == constants.HederaNetworkId {
			topic = constants.HederaTransferMessageSubmission
		} else {
			topic = constants.HederaBurnMessageSubmission
		}
	} else {
		transferMessage.NetworkTimestamp = tx.ConsensusTimestamp
		if nativeAsset.ChainId == constants.HederaNetworkId {
			topic = constants.ReadOnlyHederaFeeTransfer
		} else {
			topic = constants.ReadOnlyHederaBurn
		}
	}

	q.Push(&queue.Message{Payload: transferMessage, Topic: topic})
}

func (ctw Watcher) createFungiblePayload(transactionID string, receiver string, sourceAsset string, asset asset.NativeAsset, amount int64, targetChainId uint64, targetChainAsset string) (*payload.Transfer, error) {
	nativeAsset := ctw.assetsService.FungibleNativeAsset(asset.ChainId, asset.Asset)

	sourceAssetInfo, exists := ctw.assetsService.FungibleAssetInfo(constants.HederaNetworkId, sourceAsset)
	if !exists {
		return nil, fmt.Errorf("failed to retrieve fungible asset info of [%s]", sourceAsset)
	}

	targetAssetInfo, exists := ctw.assetsService.FungibleAssetInfo(targetChainId, targetChainAsset)
	if !exists {
		return nil, fmt.Errorf("failed to retrieve fungible asset info of [%s]", targetChainAsset)
	}

	if (nativeAsset.ChainId == constants.HederaNetworkId) && (sourceAssetInfo.Decimals != targetAssetInfo.Decimals) {
		return nil, fmt.Errorf("decimals of source asset [%s] and target asset [%s] are not equal", sourceAsset, targetChainAsset)
	}

	targetAmount := decimal.TargetAmount(sourceAssetInfo.Decimals, targetAssetInfo.Decimals, big.NewInt(amount))
	if targetAmount.Cmp(big.NewInt(0)) == 0 {
		return nil, fmt.Errorf("insufficient amount provided: Amount [%d] and Target Amount [%s]", amount, targetAmount)
	}

	tokenPriceInfo, exist := ctw.pricingService.GetTokenPriceInfo(asset.ChainId, nativeAsset.Asset)
	if !exist {
		errMsg := fmt.Sprintf("[%s] - Couldn't get price info in USD for asset [%s]", transactionID, nativeAsset.Asset)
		return nil, errors.New(errMsg)
	}

	if targetAmount.Cmp(tokenPriceInfo.MinAmountWithFee) < 0 {
		return nil, fmt.Errorf("[%s] - Transfer Amount [%s] is less than Minimum Amount [%s]", transactionID, targetAmount, tokenPriceInfo.MinAmountWithFee)
	}

	return payload.New(
		transactionID,
		constants.HederaNetworkId,
		targetChainId,
		nativeAsset.ChainId,
		receiver,
		sourceAsset,
		targetChainAsset,
		nativeAsset.Asset,
		targetAmount.String()), nil
}

func (ctw Watcher) initSuccessRatePrometheusMetrics(tx transaction.Transaction, sourceChainId, targetChainId uint64, asset string) {
	if !ctw.prometheusService.GetIsMonitoringEnabled() {
		return
	}

	metrics.CreateMajorityReachedIfNotExists(sourceChainId, targetChainId, asset, tx.TransactionID, ctw.prometheusService, ctw.logger)
	if ctw.assetsService.IsNative(constants.HederaNetworkId, asset) && targetChainId != constants.HederaNetworkId {
		metrics.CreateFeeTransferredIfNotExists(sourceChainId, targetChainId, asset, tx.TransactionID, ctw.prometheusService, ctw.logger)
	}
	metrics.CreateUserGetHisTokensIfNotExists(sourceChainId, targetChainId, asset, tx.TransactionID, ctw.prometheusService, ctw.logger)
}

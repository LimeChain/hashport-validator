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

package transfers

import (
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"

	"github.com/hashgraph/hedera-sdk-go/v2"
	mirrorNodeTransaction "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	big_numbers "github.com/limechain/hedera-eth-bridge-validator/app/helper/big-numbers"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/decimal"
	hederaHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/memo"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/metrics"
	syncHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/sync"
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/schedule"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/payload"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/fee/distributor"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	logger             *log.Entry
	hederaNode         client.HederaNode
	mirrorNode         client.MirrorNode
	contractServices   map[uint64]service.Contracts
	transferRepository repository.Transfer
	scheduleRepository repository.Schedule
	feeRepository      repository.Fee
	distributor        service.Distributor
	feeService         service.Fee
	scheduledService   service.Scheduled
	messageService     service.Messages
	prometheusService  service.Prometheus
	assetsService      service.Assets
	topicID            hedera.TopicID
	bridgeAccountID    hedera.AccountID
}

func NewService(
	hederaNode client.HederaNode,
	mirrorNode client.MirrorNode,
	contractServices map[uint64]service.Contracts,
	transferRepository repository.Transfer,
	scheduleRepository repository.Schedule,
	feeRepository repository.Fee,
	feeService service.Fee,
	distributor service.Distributor,
	topicID string,
	bridgeAccount string,
	scheduledService service.Scheduled,
	messageService service.Messages,
	prometheusService service.Prometheus,
	assetsService service.Assets,
) *Service {
	tID, e := hedera.TopicIDFromString(topicID)
	if e != nil {
		log.Fatalf("Invalid monitoring Topic ID [%s] - Error: [%s]", topicID, e)
	}
	bridgeAccountID, e := hedera.AccountIDFromString(bridgeAccount)
	if e != nil {
		log.Fatalf("Invalid BridgeAccountID [%s] - Error: [%s]", bridgeAccount, e)
	}

	instance := &Service{
		logger:             config.GetLoggerFor(fmt.Sprintf("Transfers Service")),
		hederaNode:         hederaNode,
		mirrorNode:         mirrorNode,
		contractServices:   contractServices,
		transferRepository: transferRepository,
		scheduleRepository: scheduleRepository,
		feeRepository:      feeRepository,
		topicID:            tID,
		feeService:         feeService,
		distributor:        distributor,
		bridgeAccountID:    bridgeAccountID,
		scheduledService:   scheduledService,
		messageService:     messageService,
		prometheusService:  prometheusService,
		assetsService:      assetsService,
	}

	return instance
}

// SanityCheckTransfer performs validation on the memo and state proof for the transaction
func (ts *Service) SanityCheckTransfer(tx mirrorNodeTransaction.Transaction) model.SanityCheckResult {
	result := model.SanityCheckResult{}
	m, e := memo.Validate(tx.MemoBase64)
	if e != nil {
		result.Err = fmt.Errorf("[%s] - Could not parse transaction memo [%s]. Error: [%s]", tx.TransactionID, tx.MemoBase64, e)
		return result
	}

	memoArgs := strings.Split(m, "-")
	chainId, _ := strconv.ParseUint(memoArgs[0], 10, 64)
	result.ChainId = chainId
	evmAddress := memoArgs[1]
	result.EvmAddress = evmAddress
	if len(memoArgs) > 2 {
		nftId, e := hedera.NftIDFromString(memoArgs[2])
		if e != nil {
			result.Err = fmt.Errorf("[%s] - Could not parse NftId in transaction memo [%s]. Error: [%s]", tx.TransactionID, tx.MemoBase64, e)
			return result
		}
		result.NftId = &nftId
	}

	return result
}

// InitiateNewTransfer Stores the incoming transfer message into the Database aware of already processed transfers
func (ts *Service) InitiateNewTransfer(tm payload.Transfer) (*entity.Transfer, error) {
	dbTransaction, err := ts.transferRepository.GetByTransactionId(tm.TransactionId)
	if err != nil {
		ts.logger.Errorf("[%s] - Failed to get db record. Error [%s]", tm.TransactionId, err)
		return nil, err
	}

	if dbTransaction != nil {
		ts.logger.Infof("[%s] - Transaction already added", tm.TransactionId)
		return dbTransaction, err
	}

	ts.logger.Debugf("[%s] - Adding new Transaction Record", tm.TransactionId)
	tx, err := ts.transferRepository.Create(&tm)
	if err != nil {
		ts.logger.Errorf("[%s] - Failed to create a transaction record. Error [%s].", tm.TransactionId, err)
		return nil, err
	}
	return tx, nil
}

func (ts *Service) authMessageSubmissionCallbacks(txId string) (onSuccess, onRevert func()) {
	onSuccess = func() {
		ts.logger.Debugf("Authorisation Signature TX successfully executed for TX [%s]", txId)
	}

	onRevert = func() {
		ts.logger.Debugf("Authorisation Signature TX failed for TX ID [%s]", txId)
	}
	return onSuccess, onRevert
}

func (ts *Service) ProcessNativeTransfer(tm payload.Transfer) error {
	intAmount, err := strconv.ParseInt(tm.Amount, 10, 64)
	if err != nil {
		ts.logger.Errorf("[%s] - Failed to parse amount. Error: [%s]", tm.TransactionId, err)
		return err
	}

	fee, remainder := ts.feeService.CalculateFee(tm.NativeAsset, intAmount)
	validFee := ts.distributor.ValidAmount(fee)
	if validFee != fee {
		remainder += fee - validFee
	}

	go ts.processFeeTransfer(validFee, tm.SourceChainId, tm.TargetChainId, tm.TransactionId, tm.NativeAsset)

	wrappedAmount := strconv.FormatInt(remainder, 10)

	tm.Amount = wrappedAmount
	signatureMessage, err := ts.messageService.SignFungibleMessage(tm)
	if err != nil {
		return err
	}

	return ts.submitTopicMessageAndWaitForTransaction(tm.TransactionId, signatureMessage)
}

func (ts *Service) ProcessNativeNftTransfer(tm payload.Transfer) error {
	ts.logger.Infof("[%s] - Sending NFT to bridge account.", tm.TransactionId)
	status, wg, err := ts.transferNftToBridgeAccount(tm)
	if err != nil {
		return err
	}

	wg.Wait()
	if *status == syncHelper.DONE {
		ts.logger.Debugf("[%s] - Proceeding to process fee transfer to validators and mint wrapped NFT.", tm.TransactionId)
	} else {
		ts.logger.Errorf("[%s] - Scheduled Transaction for NFT transfer failed.", tm.TransactionId)
		return errors.New("failed-scheduled-nft-transfer")
	}

	feePerValidator := ts.distributor.ValidAmount(tm.Fee)
	go ts.processFeeTransfer(feePerValidator, tm.SourceChainId, tm.TargetChainId, tm.TransactionId, constants.Hbar)

	signatureMessage, err := ts.messageService.SignNftMessage(tm)
	if err != nil {
		return err
	}

	return ts.submitTopicMessageAndWaitForTransaction(tm.TransactionId, signatureMessage)
}

func (ts *Service) transferNftToBridgeAccount(tm payload.Transfer) (status *string, wg *sync.WaitGroup, err error) {
	status = new(string)
	wg = new(sync.WaitGroup)
	wg.Add(1)
	onExecutionSuccess, onExecutionFail := hederaHelper.ScheduledNftTxExecutionCallbacks(ts.transferRepository, ts.scheduleRepository, ts.logger, tm.TransactionId, true, status, schedule.TRANSFER, wg)
	onSuccess, onFail := hederaHelper.ScheduledNftTxMinedCallbacks(ts.transferRepository, ts.scheduleRepository, ts.logger, tm.TransactionId, status, wg)

	token, err := hedera.TokenIDFromString(tm.SourceAsset)
	if err != nil {
		ts.logger.Errorf("[%s] - Failed to parse token [%s]. Error [%s].", tm.TransactionId, tm.TargetAsset, err)
		return nil, nil, err
	}

	nftID := hedera.NftID{
		TokenID:      token,
		SerialNumber: tm.SerialNum,
	}

	sender, err := hedera.AccountIDFromString(tm.Originator)
	if err != nil {
		ts.logger.Errorf("[%s] - Failed to parse receiver [%s]. Error [%s].", tm.TransactionId, tm.Receiver, err)
		return nil, nil, err
	}

	ts.scheduledService.ExecuteScheduledNftTransferTransaction(tm.TransactionId, nftID, sender, ts.bridgeAccountID, true, onExecutionSuccess, onExecutionFail, onSuccess, onFail)
	return status, wg, err
}

func (ts *Service) ProcessWrappedTransfer(tm payload.Transfer) error {
	amount, err := big_numbers.ToBigInt(tm.Amount)
	if err != nil {
		return err
	}

	sourceAssetInfo, exists := ts.assetsService.FungibleAssetInfo(tm.SourceChainId, tm.SourceAsset)
	if !exists {
		return fmt.Errorf("Failed to retrieve fungible asset info of [%s].", tm.SourceAsset)
	}

	targetAssetInfo, exists := ts.assetsService.FungibleAssetInfo(tm.TargetChainId, tm.TargetAsset)
	if !exists {
		return fmt.Errorf("Failed to retrieve fungible asset info of [%s].", tm.TargetAsset)
	}

	// Convert the amount to the initial, so that the correct amount is being burned.
	targetAmount := decimal.TargetAmount(targetAssetInfo.Decimals, sourceAssetInfo.Decimals, amount)
	if targetAmount.Cmp(big.NewInt(0)) == 0 {
		return fmt.Errorf("Insufficient amount provided: Amount [%s] and Target Amount [%s].", amount, targetAmount)
	}

	status := make(chan string)
	onExecutionBurnSuccess, onExecutionBurnFail := ts.scheduledBurnTxExecutionCallbacks(tm.TransactionId, &status)
	onTokenBurnSuccess, onTokenBurnFail := ts.scheduledBurnTxMinedCallbacks(&status)
	ts.scheduledService.ExecuteScheduledBurnTransaction(tm.TransactionId, tm.SourceAsset, targetAmount.Int64(), &status, onExecutionBurnSuccess, onExecutionBurnFail, onTokenBurnSuccess, onTokenBurnFail)

statusBlocker:
	for {
		switch <-status {
		case syncHelper.DONE:
			ts.logger.Debugf("[%s] - Proceeding to sign and submit unlock permission messages.", tm.TransactionId)
			break statusBlocker
		case syncHelper.FAIL:
			ts.logger.Errorf("[%s] - Failed to await the execution of Scheduled Burn Transaction.", tm.TransactionId)
			return errors.New("failed-scheduled-burn")
		}
	}

	signatureMessage, err := ts.messageService.SignFungibleMessage(tm)
	if err != nil {
		return err
	}

	return ts.submitTopicMessageAndWaitForTransaction(tm.TransactionId, signatureMessage)
}

// TransferData returns from the database the given transfer, its signatures and
// calculates if its messages have reached super majority
func (ts *Service) TransferData(txId string) (interface{}, error) {
	t, err := ts.transferRepository.GetWithPreloads(txId)
	if err != nil {
		ts.logger.Errorf("[%s] - Failed to query Transfer with messages. Error: [%s].", txId, err)
		return nil, err
	}

	if t == nil {
		return nil, service.ErrNotFound
	}

	if t.TargetChainID == constants.HederaNetworkId {
		return service.TransferData{}, service.ErrBadRequestTransferTargetNetworkNoSignaturesRequired
	}

	if t != nil && t.NativeChainID == constants.HederaNetworkId && t.Fee == "" {
		return service.TransferData{}, service.ErrNotFound
	}

	transferData := service.TransferData{
		IsNft:         t.IsNft,
		Recipient:     t.Receiver,
		RouterAddress: ts.contractServices[t.TargetChainID].Address().String(),
		SourceChainId: t.SourceChainID,
		TargetChainId: t.TargetChainID,
		SourceAsset:   t.SourceAsset,
		NativeAsset:   t.NativeAsset,
		TargetAsset:   t.TargetAsset,
	}

	var signatures []string
	for _, m := range t.Messages {
		signatures = append(signatures, m.Signature)
	}

	bnSignaturesLength := big.NewInt(int64(len(t.Messages)))
	reachedMajority, err := ts.contractServices[t.TargetChainID].
		HasValidSignaturesLength(bnSignaturesLength)
	if err != nil {
		ts.logger.Errorf("[%s] - Failed to check has valid signatures length. Error [%s]", t.TransactionID, err)
		return nil, err
	}

	transferData.Signatures = signatures
	transferData.Majority = reachedMajority

	if !t.IsNft {
		signedAmount := t.Amount
		if t.NativeChainID == constants.HederaNetworkId {
			amount, err := strconv.ParseInt(t.Amount, 10, 64)
			if err != nil {
				ts.logger.Errorf("[%s] - Failed to parse transfer amount. Error [%s]", t.TransactionID, err)
				return nil, err
			}

			feeAmount, err := strconv.ParseInt(t.Fee, 10, 64)
			if err != nil {
				ts.logger.Errorf("[%s] - Failed to parse fee amount. Error [%s]", t.TransactionID, err)
				return nil, err
			}
			signedAmount = strconv.FormatInt(amount-feeAmount, 10)
		}
		return service.FungibleTransferData{
			TransferData: transferData,
			Amount:       signedAmount,
		}, nil
	}

	return service.NonFungibleTransferData{
		TransferData: transferData,
		TokenId:      t.SerialNumber,
		Metadata:     t.Metadata,
	}, nil
}

func (ts *Service) Paged(req *model.PagedRequest) (*model.Paged, error) {
	items, err := ts.transferRepository.Paged(req)
	if err != nil {
		ts.logger.Errorf("Failed to get paged transfers. Error: [%s]", err)
		return nil, err
	}

	res := make([]*model.Transfer, 0, len(items))
	for _, t := range items {
		res = append(res, t.ToDto())
	}

	count, err := ts.transferRepository.Count()
	if err != nil {
		ts.logger.Errorf("Failed to count transfers. Error: [%s]", err)
		return nil, err
	}

	return &model.Paged{
		Items:      res,
		TotalCount: count,
	}, nil
}

func (ts *Service) submitTopicMessageAndWaitForTransaction(transferID string, signatureMessageBytes []byte) error {
	messageTxId, err := ts.hederaNode.SubmitTopicConsensusMessage(
		ts.topicID,
		signatureMessageBytes)
	if err != nil {
		ts.logger.Errorf("[%s] - Failed to submit Signature Message to Topic. Error: [%s]", transferID, err)
		return err
	}

	// Attach update callbacks on Signature HCS Message
	ts.logger.Infof("[%s] - Submitted signature on Topic [%s]", transferID, ts.topicID)
	onSuccessfulAuthMessage, onFailedAuthMessage := ts.authMessageSubmissionCallbacks(transferID)
	ts.mirrorNode.WaitForTransaction(hederaHelper.ToMirrorNodeTransactionID(messageTxId.String()), onSuccessfulAuthMessage, onFailedAuthMessage)
	return nil
}

func (ts *Service) processFeeTransfer(totalFee int64, sourceChainId, targetChainId uint64, transferID string, nativeAsset string) {

	transfers, err := ts.distributor.CalculateMemberDistribution(totalFee)
	if err != nil {
		ts.logger.Errorf("[%s] Fee - Failed to Distribute to Members. Error: [%s].", transferID, err)
		return
	}

	splitTransfers := distributor.SplitAccountAmounts(transfers, model.Hedera{
		AccountID: ts.bridgeAccountID,
		Amount:    -totalFee,
	})

	err = ts.transferRepository.UpdateFee(transferID, strconv.FormatInt(totalFee, 10))
	if err != nil {
		ts.logger.Errorf("[%s] - Failed to update fee [%d]. Error [%s].", transferID, totalFee, err)
		return
	}

	var (
		feeOutParams *hederaHelper.FeeOutParams
	)

	if ts.prometheusService.GetIsMonitoringEnabled() {
		feeOutParams = hederaHelper.NewFeeOutParams(len(splitTransfers))
	}

	for _, splitTransfer := range splitTransfers {
		fee := -splitTransfer[len(splitTransfer)-1].Amount
		onExecutionSuccess, onExecutionFail := ts.scheduledFeeTxExecutionCallbacks(transferID, strconv.FormatInt(fee, 10))
		onSuccess, onFail := ts.scheduledFeeTxMinedCallbacks(feeOutParams, splitTransfer)

		ts.scheduledService.ExecuteScheduledTransferTransaction(transferID, nativeAsset, splitTransfer, onExecutionSuccess, onExecutionFail, onSuccess, onFail)
	}

	if ts.prometheusService.GetIsMonitoringEnabled() {
		go hederaHelper.AwaitMultipleScheduledTransactions(
			feeOutParams.OutParams,
			sourceChainId,
			targetChainId,
			nativeAsset,
			transferID,
			ts.onMinedFeeTransactionsSetMetrics,
		)
	}
}

func (ts *Service) onMinedFeeTransactionsSetMetrics(sourceChainId, targetChainId uint64, nativeAsset string, transferID string, isTransferSuccessful bool) {
	if sourceChainId != constants.HederaNetworkId || isTransferSuccessful == false || !ts.prometheusService.GetIsMonitoringEnabled() {
		return
	}

	metrics.SetFeeTransferred(sourceChainId, targetChainId, nativeAsset, transferID, ts.prometheusService, ts.logger)
}

func (ts *Service) scheduledBurnTxExecutionCallbacks(transferID string, blocker *chan string) (onExecutionSuccess func(transactionID string, scheduleID string), onExecutionFail func(transactionID string)) {
	onExecutionSuccess = func(transactionID, scheduleID string) {
		ts.logger.Debugf("[%s] - Updating db status to Submitted with TransactionID [%s].",
			transferID,
			transactionID)
		err := ts.scheduleRepository.Create(&entity.Schedule{
			TransactionID: transactionID,
			ScheduleID:    scheduleID,
			Operation:     schedule.BURN,
			Status:        status.Submitted,
			TransferID: sql.NullString{
				String: transferID,
				Valid:  true,
			},
		})
		if err != nil {
			*blocker <- syncHelper.FAIL
			ts.logger.Errorf(
				"[%s] - Failed to update submitted scheduled status with TransactionID [%s], ScheduleID [%s]. Error [%s].",
				transferID, transactionID, scheduleID, err)
			return
		}
	}

	onExecutionFail = func(transactionID string) {
		*blocker <- syncHelper.FAIL
		err := ts.scheduleRepository.Create(&entity.Schedule{
			TransactionID: transactionID,
			Operation:     schedule.BURN,
			Status:        status.Failed,
			TransferID: sql.NullString{
				String: transferID,
				Valid:  true,
			},
		})
		if err != nil {
			ts.logger.Errorf("[%s] - Failed to update status failed. Error [%s].", transferID, err)
			return
		}
	}

	return onExecutionSuccess, onExecutionFail
}

func (ts *Service) scheduledBurnTxMinedCallbacks(status *chan string) (onSuccess, onFail func(transactionID string)) {
	onSuccess = func(transactionID string) {
		ts.logger.Debugf("[%s] - Scheduled TX execution successful.", transactionID)

		err := ts.scheduleRepository.UpdateStatusCompleted(transactionID)
		if err != nil {
			ts.logger.Errorf("[%s] Schedule - Failed to update status completed. Error [%s].", transactionID, err)
			return
		}
		if err != nil {
			*status <- syncHelper.FAIL
			ts.logger.Errorf("[%s] - Failed to update scheduled burn status completed. Error [%s].", transactionID, err)
			return
		}
		*status <- syncHelper.DONE
	}

	onFail = func(transactionID string) {
		*status <- syncHelper.FAIL
		ts.logger.Debugf("[%s] - Scheduled TX execution has failed.", transactionID)
		err := ts.scheduleRepository.UpdateStatusFailed(transactionID)
		if err != nil {
			ts.logger.Errorf("[%s] Schedule - Failed to update status failed. Error [%s].", transactionID, err)
			return
		}

		if err != nil {
			ts.logger.Errorf("[%s] - Failed to update status signature failed. Error [%s].", transactionID, err)
			return
		}
	}

	return onSuccess, onFail
}

func (ts *Service) scheduledFeeTxExecutionCallbacks(transferID, feeAmount string) (onExecutionSuccess func(transactionID, scheduleID string), onExecutionFail func(transactionID string)) {
	onExecutionSuccess = func(transactionID, scheduleID string) {
		err := ts.scheduleRepository.Create(&entity.Schedule{
			TransactionID: transactionID,
			ScheduleID:    scheduleID,
			Operation:     schedule.TRANSFER,
			Status:        status.Submitted,
			TransferID: sql.NullString{
				String: transferID,
				Valid:  true,
			},
		})
		if err != nil {
			ts.logger.Errorf(
				"[%s] Fee - Failed to create Schedule Record [%s]. Error [%s].",
				transferID, transactionID, err)
			return
		}
		err = ts.feeRepository.Create(&entity.Fee{
			TransactionID: transactionID,
			ScheduleID:    scheduleID,
			Amount:        feeAmount,
			Status:        status.Submitted,
			TransferID: sql.NullString{
				String: transferID,
				Valid:  true,
			},
		})
		if err != nil {
			ts.logger.Errorf(
				"[%s] Fee - Failed to create Fee Record [%s]. Error [%s].",
				transferID, transactionID, err)
			return
		}
	}

	onExecutionFail = func(transactionID string) {
		err := ts.scheduleRepository.Create(&entity.Schedule{
			TransactionID: transactionID,
			Operation:     schedule.TRANSFER,
			Status:        status.Failed,
			TransferID: sql.NullString{
				String: transferID,
				Valid:  true,
			},
		})
		if err != nil {
			ts.logger.Errorf(
				"[%s] Fee - Failed to create failed Schedule Record [%s]. Error [%s].",
				transferID, transactionID, err)
			return
		}
		err = ts.feeRepository.Create(&entity.Fee{
			TransactionID: transactionID,
			Amount:        feeAmount,
			Status:        status.Failed,
			TransferID: sql.NullString{
				String: transferID,
				Valid:  true,
			},
		})
		if err != nil {
			ts.logger.Errorf("[%s] Fee - Failed to create failed record. Error [%s].", transferID, err)
			return
		}
	}

	return onExecutionSuccess, onExecutionFail
}

func (ts *Service) scheduledFeeTxMinedCallbacks(feeOutParams *hederaHelper.FeeOutParams, splitTransfer []model.Hedera) (onSuccess, onFail func(transactionID string)) {
	onSuccess = func(transactionID string) {
		ts.logger.Debugf("[%s] Fee - Scheduled TX execution successful.", transactionID)

		if ts.prometheusService.GetIsMonitoringEnabled() {
			result := true
			feeOutParams.HandleResultForAwaitedTransfer(&result, false, splitTransfer)
		}

		err := ts.scheduleRepository.UpdateStatusCompleted(transactionID)
		if err != nil {
			ts.logger.Errorf("[%s] Schedule - Failed to update status completed. Error [%s].", transactionID, err)
			return
		}

		err = ts.feeRepository.UpdateStatusCompleted(transactionID)
		if err != nil {
			ts.logger.Errorf("[%s] Fee - Failed to update status completed. Error [%s].", transactionID, err)
			return
		}
	}

	onFail = func(transactionID string) {
		ts.logger.Debugf("[%s] Fee - Scheduled TX execution has failed.", transactionID)
		result := false
		if ts.prometheusService.GetIsMonitoringEnabled() {
			feeOutParams.HandleResultForAwaitedTransfer(&result, false, splitTransfer)
		}

		err := ts.scheduleRepository.UpdateStatusFailed(transactionID)
		if err != nil {
			ts.logger.Errorf("[%s] Schedule - Failed to update status failed. Error [%s].", transactionID, err)
			return
		}

		err = ts.feeRepository.UpdateStatusFailed(transactionID)
		if err != nil {
			ts.logger.Errorf("[%s] Fee - Failed to update status failed. Error [%s].", transactionID, err)
			return
		}
	}
	return onSuccess, onFail
}

func (ts *Service) UpdateTransferStatusCompleted(transferID string) error {
	return ts.transferRepository.UpdateStatusCompleted(transferID)
}

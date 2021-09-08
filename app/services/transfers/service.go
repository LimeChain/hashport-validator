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

package transfers

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/hashgraph/hedera-state-proof-verifier-go/stateproof"
	mirror_node "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	hederahelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/memo"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/sync"
	auth_message "github.com/limechain/hedera-eth-bridge-validator/app/model/auth-message"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/message"
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/fee"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/schedule"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

type Service struct {
	logger             *log.Entry
	hederaNode         client.HederaNode
	mirrorNode         client.MirrorNode
	contractServices   map[int64]service.Contracts
	ethSigners         map[int64]service.Signer
	transferRepository repository.Transfer
	scheduleRepository repository.Schedule
	feeRepository      repository.Fee
	distributor        service.Distributor
	feeService         service.Fee
	scheduledService   service.Scheduled
	topicID            hedera.TopicID
	bridgeAccountID    hedera.AccountID
}

func NewService(
	hederaNode client.HederaNode,
	mirrorNode client.MirrorNode,
	contractServices map[int64]service.Contracts,
	signers map[int64]service.Signer,
	transferRepository repository.Transfer,
	scheduleRepository repository.Schedule,
	feeRepository repository.Fee,
	feeService service.Fee,
	distributor service.Distributor,
	topicID string,
	bridgeAccount string,
	scheduledService service.Scheduled,
) *Service {
	tID, e := hedera.TopicIDFromString(topicID)
	if e != nil {
		log.Fatalf("Invalid monitoring Topic ID [%s] - Error: [%s]", topicID, e)
	}
	bridgeAccountID, e := hedera.AccountIDFromString(bridgeAccount)
	if e != nil {
		log.Fatalf("Invalid BridgeAccountID [%s] - Error: [%s]", bridgeAccount, e)
	}

	return &Service{
		logger:             config.GetLoggerFor(fmt.Sprintf("Transfers Service")),
		hederaNode:         hederaNode,
		mirrorNode:         mirrorNode,
		contractServices:   contractServices,
		ethSigners:         signers,
		transferRepository: transferRepository,
		scheduleRepository: scheduleRepository,
		feeRepository:      feeRepository,
		topicID:            tID,
		feeService:         feeService,
		distributor:        distributor,
		bridgeAccountID:    bridgeAccountID,
		scheduledService:   scheduledService,
	}
}

// SanityCheck performs validation on the memo and state proof for the transaction
func (ts *Service) SanityCheckTransfer(tx mirror_node.Transaction) (int64, string, error) {
	m, e := memo.Validate(tx.MemoBase64)
	if e != nil {
		return 0, "", errors.New(fmt.Sprintf("[%s] - Could not parse transaction memo [%s]. Error: [%s]", tx.TransactionID, tx.MemoBase64, e))
	}

	stateProof, e := ts.mirrorNode.GetStateProof(tx.TransactionID)
	if e != nil {
		return 0, "", errors.New(fmt.Sprintf("Could not GET state proof. Error [%s]", e))
	}

	verified, e := stateproof.Verify(tx.TransactionID, stateProof)
	if e != nil {
		return 0, "", errors.New(fmt.Sprintf("State proof verification failed. Error [%s]", e))
	}

	if !verified {
		return 0, "", errors.New("invalid state proof")
	}

	memoArgs := strings.Split(m, "-")
	chainId, _ := strconv.ParseInt(memoArgs[0], 10, 64)
	evmAddress := memoArgs[1]

	return chainId, evmAddress, nil
}

// InitiateNewTransfer Stores the incoming transfer message into the Database aware of already processed transfers
func (ts *Service) InitiateNewTransfer(tm model.Transfer) (*entity.Transfer, error) {
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

//// SaveRecoveredTxn creates new Transaction record persisting the recovered Transfer TX
func (ts *Service) SaveRecoveredTxn(txId, amount, nativeAsset, wrappedAsset string, memo string) error {
	//// TODO: Add ChainID to the parameters and remove mockChainID
	//mockChainID := int64(80001)
	//err := ts.transferRepository.SaveRecoveredTxn(&model.Transfer{
	//	TransactionId: txId,
	//	RouterAddress: ts.contractServices[mockChainID].Address().String(),
	//	Receiver:      memo,
	//	Amount:        amount,
	//	NativeAsset:   nativeAsset,
	//	WrappedAsset:  wrappedAsset,
	//})
	//if err != nil {
	//	ts.logger.Errorf("[%s] - Something went wrong while saving new Recovered Transaction. Error [%s]", txId, err)
	//	return err
	//}
	//
	//ts.logger.Infof("Added new Transaction Record with Txn ID [%s]", txId)
	//return err
	return nil
}

func (ts *Service) authMessageSubmissionCallbacks(txId string) (onSuccess, onRevert func()) {
	onSuccess = func() {
		ts.logger.Debugf("Authorisation Signature TX successfully executed for TX [%s]", txId)
		err := ts.transferRepository.UpdateStatusSignatureMined(txId)
		if err != nil {
			ts.logger.Errorf("[%s] - Failed to update status signature mined. Error [%s].", txId, err)
			return
		}
	}

	onRevert = func() {
		ts.logger.Debugf("Authorisation Signature TX failed for TX ID [%s]", txId)
		err := ts.transferRepository.UpdateStatusSignatureFailed(txId)
		if err != nil {
			ts.logger.Errorf("[%s] - Failed to update status signature failed. Error [%s].", txId, err)
			return
		}
	}
	return onSuccess, onRevert
}

func (ts *Service) ProcessNativeTransfer(tm model.Transfer) error {
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

	go ts.processFeeTransfer(tm.TransactionId, validFee, tm.NativeAsset)

	wrappedAmount := strconv.FormatInt(remainder, 10)

	authMsgHash, err := auth_message.EncodeBytesFrom(tm.SourceChainId, tm.TargetChainId, tm.TransactionId, tm.TargetAsset, tm.Receiver, wrappedAmount)
	if err != nil {
		ts.logger.Errorf("[%s] - Failed to encode the authorisation signature. Error: [%s]", tm.TransactionId, err)
		return err
	}

	signatureBytes, err := ts.ethSigners[tm.TargetChainId].Sign(authMsgHash)
	if err != nil {
		ts.logger.Errorf("[%s] - Failed to sign the authorisation signature. Error: [%s]", tm.TransactionId, err)
		return err
	}
	signature := hex.EncodeToString(signatureBytes)

	signatureMessage := message.NewSignature(
		uint64(tm.SourceChainId),
		uint64(tm.TargetChainId),
		tm.TransactionId,
		tm.TargetAsset,
		tm.Receiver,
		wrappedAmount,
		signature)

	return ts.submitTopicMessageAndWaitForTransaction(signatureMessage)
}

func (ts *Service) ProcessWrappedTransfer(tm model.Transfer) error {
	intAmount, err := strconv.ParseInt(tm.Amount, 10, 64)
	if err != nil {
		ts.logger.Errorf("[%s] - Failed to parse amount. Error: [%s]", tm.TransactionId, err)
		return err
	}
	status := make(chan string)
	onExecutionBurnSuccess, onExecutionBurnFail := ts.scheduledBurnTxExecutionCallbacks(tm.TransactionId, &status)
	onTokenBurnSuccess, onTokenBurnFail := ts.scheduledBurnTxMinedCallbacks(&status)
	ts.scheduledService.ExecuteScheduledBurnTransaction(tm.TransactionId, tm.SourceAsset, intAmount, &status, onExecutionBurnSuccess, onExecutionBurnFail, onTokenBurnSuccess, onTokenBurnFail)

statusBlocker:
	for {
		switch <-status {
		case sync.DONE:
			ts.logger.Debugf("[%s] - Proceeding to sign and submit unlock permission messages.", tm.TransactionId)
			break statusBlocker
		case sync.FAIL:
			ts.logger.Errorf("[%s] - Failed to await the execution of Scheduled Burn Transaction.", tm.TransactionId)
			return errors.New("failed-scheduled-burn")
		}
	}

	authMsgHash, err := auth_message.EncodeBytesFrom(tm.SourceChainId, tm.TargetChainId, tm.TransactionId, tm.TargetAsset, tm.Receiver, tm.Amount)
	if err != nil {
		ts.logger.Errorf("[%s] - Failed to encode the authorisation signature. Error: [%s]", tm.TransactionId, err)
		return err
	}

	signatureBytes, err := ts.ethSigners[tm.TargetChainId].Sign(authMsgHash)
	if err != nil {
		ts.logger.Errorf("[%s] - Failed to sign the authorisation signature. Error: [%s]", tm.TransactionId, err)
		return err
	}
	signature := hex.EncodeToString(signatureBytes)

	signatureMessage := message.NewSignature(
		uint64(tm.SourceChainId),
		uint64(tm.TargetChainId),
		tm.TransactionId,
		tm.TargetAsset,
		tm.Receiver,
		tm.Amount,
		signature)

	return ts.submitTopicMessageAndWaitForTransaction(signatureMessage)
}

func (ts *Service) submitTopicMessageAndWaitForTransaction(signatureMessage *message.Message) error {
	sigMsgBytes, err := signatureMessage.ToBytes()
	if err != nil {
		ts.logger.Errorf("[%s] - Failed to encode Signature Message to bytes. Error [%s]", signatureMessage.TransferID, err)
		return err
	}

	messageTxId, err := ts.hederaNode.SubmitTopicConsensusMessage(
		ts.topicID,
		sigMsgBytes)
	if err != nil {
		ts.logger.Errorf("[%s] - Failed to submit Signature Message to Topic. Error: [%s]", signatureMessage.TransferID, err)
		return err
	}

	// Update Transfer Record
	err = ts.transferRepository.UpdateStatusSignatureSubmitted(signatureMessage.TransferID)
	if err != nil {
		ts.logger.Errorf("[%s] - Failed to update. Error [%s].", signatureMessage.TransferID, err)
		return err
	}

	// Attach update callbacks on Signature HCS Message
	ts.logger.Infof("[%s] - Submitted signature on Topic [%s]", signatureMessage.TransferID, ts.topicID)
	onSuccessfulAuthMessage, onFailedAuthMessage := ts.authMessageSubmissionCallbacks(signatureMessage.TransferID)
	ts.mirrorNode.WaitForTransaction(hederahelper.ToMirrorNodeTransactionID(messageTxId.String()), onSuccessfulAuthMessage, onFailedAuthMessage)
	return nil
}

func (ts *Service) processFeeTransfer(transferID string, feeAmount int64, nativeAsset string) {
	transfers, err := ts.distributor.CalculateMemberDistribution(feeAmount)
	if err != nil {
		ts.logger.Errorf("[%s] Fee - Failed to Distribute to Members. Error: [%s].", transferID, err)
		return
	}

	transfers = append(transfers,
		model.Hedera{
			AccountID: ts.bridgeAccountID,
			Amount:    -feeAmount,
		})

	onExecutionSuccess, onExecutionFail := ts.scheduledTxExecutionCallbacks(transferID, strconv.FormatInt(feeAmount, 10))
	onSuccess, onFail := ts.scheduledTxMinedCallbacks()

	ts.scheduledService.ExecuteScheduledTransferTransaction(transferID, nativeAsset, transfers, onExecutionSuccess, onExecutionFail, onSuccess, onFail)
}

func (ts *Service) scheduledBurnTxExecutionCallbacks(transferID string, status *chan string) (onExecutionSuccess func(transactionID string, scheduleID string), onExecutionFail func(transactionID string)) {
	onExecutionSuccess = func(transactionID, scheduleID string) {
		ts.logger.Debugf("[%s] - Updating db status to Submitted with TransactionID [%s].",
			transferID,
			transactionID)
		err := ts.scheduleRepository.Create(&entity.Schedule{
			TransactionID: transactionID,
			ScheduleID:    scheduleID,
			Operation:     schedule.BURN,
			Status:        fee.StatusSubmitted, // TODO: not fee
			TransferID: sql.NullString{
				String: transferID,
				Valid:  true,
			},
		})
		if err != nil {
			*status <- sync.FAIL
			ts.logger.Errorf(
				"[%s] - Failed to update submitted scheduled status with TransactionID [%s], ScheduleID [%s]. Error [%s].",
				transferID, transactionID, scheduleID, err)
			return
		}
	}

	onExecutionFail = func(transactionID string) {
		*status <- sync.FAIL
		err := ts.scheduleRepository.Create(&entity.Schedule{
			TransactionID: transactionID,
			Operation:     schedule.BURN,
			Status:        fee.StatusSubmitted, // TODO: not fee
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
			*status <- sync.FAIL
			ts.logger.Errorf("[%s] - Failed to update scheduled burn status completed. Error [%s].", transactionID, err)
			return
		}
		*status <- sync.DONE
	}

	onFail = func(transactionID string) {
		*status <- sync.FAIL
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

func (ts *Service) scheduledTxExecutionCallbacks(transferID, feeAmount string) (onExecutionSuccess func(transactionID, scheduleID string), onExecutionFail func(transactionID string)) {
	onExecutionSuccess = func(transactionID, scheduleID string) {
		err := ts.scheduleRepository.Create(&entity.Schedule{
			TransactionID: transactionID,
			ScheduleID:    scheduleID,
			Operation:     schedule.TRANSFER,
			Status:        fee.StatusSubmitted, // TODO: not fee
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
			Status:        fee.StatusSubmitted,
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
			Status:        fee.StatusSubmitted, // TODO: not fee
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
			Amount: feeAmount,
			Status: fee.StatusFailed,
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

func (ts *Service) scheduledTxMinedCallbacks() (onSuccess, onFail func(transactionID string)) {
	onSuccess = func(transactionID string) {
		ts.logger.Debugf("[%s] Fee - Scheduled TX execution successful.", transactionID)

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

// TransferData returns from the database the given transfer, its signatures and
// calculates if its messages have reached super majority
func (ts *Service) TransferData(txId string) (service.TransferData, error) {
	t, err := ts.transferRepository.GetWithPreloads(txId)
	if err != nil {
		ts.logger.Errorf("[%s] - Failed to query Transfer with messages. Error: [%s].", txId, err)
		return service.TransferData{}, err
	}

	if t == nil {
		return service.TransferData{}, service.ErrNotFound
	}

	// TODO: remove when fee check is not here
	if t != nil && t.HasFee && t.Fee.Amount == "" {
		return service.TransferData{}, service.ErrNotFound
	}

	signedAmount := t.Amount
	if t.HasFee {
		amount, err := strconv.ParseInt(t.Amount, 10, 64)
		if err != nil {
			ts.logger.Errorf("[%s] - Failed to parse transfer amount. Error [%s]", t.TransactionID, err)
			return service.TransferData{}, err
		}

		feeAmount, err := strconv.ParseInt(t.Fee.Amount, 10, 64)
		if err != nil {
			ts.logger.Errorf("[%s] - Failed to parse fee amount. Error [%s]", t.TransactionID, err)
			return service.TransferData{}, err
		}
		signedAmount = strconv.FormatInt(amount-feeAmount, 10)
	}

	var signatures []string
	for _, m := range t.Messages {
		signatures = append(signatures, m.Signature)
	}

	requiredSigCount := len(ts.contractServices[t.TargetChainID].GetMembers())/2 + 1
	reachedMajority := len(t.Messages) >= requiredSigCount

	return service.TransferData{
		Recipient:     t.Receiver,
		RouterAddress: ts.contractServices[t.TargetChainID].Address().String(),
		Amount:        signedAmount,
		SourceChainId: t.SourceChainID,
		TargetChainId: t.TargetChainID,
		SourceAsset:   t.SourceAsset,
		NativeAsset:   t.NativeAsset,
		TargetAsset:   t.TargetAsset,
		Signatures:    signatures,
		Majority:      reachedMajority,
	}, nil
}

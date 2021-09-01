package scheduled

import (
	"errors"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	hederahelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/sync"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	payerAccount     hedera.AccountID
	hederaNodeClient client.HederaNode
	mirrorNodeClient client.MirrorNode
	logger           *log.Entry
}

func New(
	payerAccount string,
	hederaNodeClient client.HederaNode,
	mirrorNodeClient client.MirrorNode) *Service {
	payer, err := hedera.AccountIDFromString(payerAccount)
	if err != nil {
		log.Fatalf("Invalid payer account: [%s].", payerAccount)
	}

	return &Service{
		payerAccount:     payer,
		hederaNodeClient: hederaNodeClient,
		mirrorNodeClient: mirrorNodeClient,
		logger:           config.GetLoggerFor("Scheduled Service"),
	}
}

// Execute submits a scheduled transaction and executes provided functions when necessary
func (s *Service) ExecuteScheduledTransferTransaction(
	id, nativeAsset string,
	transfers []transfer.Hedera,
	onExecutionSuccess func(transactionID, scheduleID string), onExecutionFail, onSuccess, onFail func(transactionID string)) {
	transactionResponse, err := s.executeScheduledTransfersTransaction(id, nativeAsset, transfers)
	if err != nil {
		s.logger.Errorf("[%s] - Failed to submit scheduled transfer transaction. Error [%s].", id, err)
		if transactionResponse != nil {
			onExecutionFail(hederahelper.ToMirrorNodeTransactionID(transactionResponse.TransactionID.String()))
		}
		return
	}
	err = s.createOrSignScheduledTransaction(transactionResponse, id, onExecutionSuccess, onExecutionFail, onSuccess, onFail)
	if err != nil {
		s.logger.Errorf("[%s] - Failed to create/sign scheduled transfer transaction. Error [%s].", id, err)
		return
	}
}

func (s *Service) executeScheduledTransfersTransaction(id, nativeAsset string, transfers []transfer.Hedera) (*hedera.TransactionResponse, error) {
	var tokenID hedera.TokenID
	var transactionResponse *hedera.TransactionResponse
	var err error

	if nativeAsset == constants.Hbar {
		transactionResponse, err = s.hederaNodeClient.
			SubmitScheduledHbarTransferTransaction(transfers, s.payerAccount, id)
	} else {
		tokenID, err = hedera.TokenIDFromString(nativeAsset)
		if err != nil {
			s.logger.Errorf("[%s] - Failed to parse native token [%s] to TokenID. Error [%s].", id, nativeAsset, err)
			return nil, err
		}
		transactionResponse, err = s.hederaNodeClient.
			SubmitScheduledTokenTransferTransaction(tokenID, transfers, s.payerAccount, id)
	}
	return transactionResponse, err
}

func (s *Service) ExecuteScheduledMintTransaction(id, asset string, amount int64, status *chan string, onExecutionSuccess func(transactionID, scheduleID string), onExecutionFail, onSuccess, onFail func(transactionID string)) {
	transactionResponse, err := s.executeScheduledTokenMintTransaction(id, asset, amount)
	if err != nil {
		s.logger.Errorf("[%s] - Failed to submit scheduled mint transaction. Error [%s].", id, err)
		if transactionResponse != nil {
			onExecutionFail(hederahelper.ToMirrorNodeTransactionID(transactionResponse.TransactionID.String()))
		}
		*status <- sync.FAIL
		return
	}

	err = s.createOrSignScheduledTransaction(transactionResponse, id, onExecutionSuccess, onExecutionFail, onSuccess, onFail)
	if err != nil {
		s.logger.Errorf("[%s] - Failed to create/sign scheduled mint transaction. Error [%s].", id, err)
		*status <- sync.FAIL
		return
	}
}

func (s *Service) ExecuteScheduledBurnTransaction(id, asset string, amount int64, status *chan string, onExecutionSuccess func(transactionID, scheduleID string), onExecutionFail, onSuccess, onFail func(transactionID string)) {
	transactionResponse, err := s.executeScheduledTokenBurnTransaction(id, asset, amount)
	if err != nil {
		s.logger.Errorf("[%s] - Failed to submit scheduled burn transaction. Error [%s].", id, err)
		if transactionResponse != nil {
			onExecutionFail(hederahelper.ToMirrorNodeTransactionID(transactionResponse.TransactionID.String()))
		}
		*status <- sync.FAIL
		return
	}

	err = s.createOrSignScheduledTransaction(transactionResponse, id, onExecutionSuccess, onExecutionFail, onSuccess, onFail)
	if err != nil {
		s.logger.Errorf("[%s] - Failed to create/sign scheduled burn transaction. Error [%s].", id, err)
		*status <- sync.FAIL
		return
	}
}

func (s *Service) executeScheduledTokenMintTransaction(id, asset string, amount int64) (*hedera.TransactionResponse, error) {
	var tokenID hedera.TokenID
	var transactionResponse *hedera.TransactionResponse
	var err error

	tokenID, err = hedera.TokenIDFromString(asset)
	if err != nil {
		s.logger.Errorf("[%s] - Failed to parse token [%s] to TokenID. Error [%s].", id, asset, err)
		return nil, err
	}

	transactionResponse, err = s.hederaNodeClient.
		SubmitScheduledTokenMintTransaction(tokenID, amount, s.payerAccount, id)

	return transactionResponse, err
}

func (s *Service) executeScheduledTokenBurnTransaction(id string, asset string, amount int64) (*hedera.TransactionResponse, error) {
	var tokenID hedera.TokenID
	var transactionResponse *hedera.TransactionResponse
	var err error

	tokenID, err = hedera.TokenIDFromString(asset)
	if err != nil {
		s.logger.Errorf("[%s] - Failed to parse token [%s] to TokenID. Error [%s].", id, asset, err)
		return nil, err
	}

	transactionResponse, err = s.hederaNodeClient.
		SubmitScheduledTokenBurnTransaction(tokenID, amount, s.payerAccount, id)

	return transactionResponse, err
}

func (s *Service) createOrSignScheduledTransaction(transactionResponse *hedera.TransactionResponse, id string, onExecutionSuccess func(transactionID string, scheduleID string), onExecutionFail, onSuccess, onFail func(transactionID string)) error {
	scheduledTxID := hederahelper.ToMirrorNodeTransactionID(transactionResponse.TransactionID.String())
	s.logger.Infof("[%s] - Successfully submitted scheduled transaction [%s].",
		id,
		scheduledTxID)

	txReceipt, err := transactionResponse.GetReceipt(s.hederaNodeClient.GetClient())
	if err != nil {
		s.logger.Errorf("[%s] - Failed to get transaction receipt for [%s]. %s", id, transactionResponse.TransactionID.String(), err)
		onExecutionFail(scheduledTxID)
		return err
	}

	switch txReceipt.Status {
	case hedera.StatusIdenticalScheduleAlreadyCreated:
		s.handleScheduleSign(id, *txReceipt.ScheduleID)
	case hedera.StatusSuccess:
	default:
		txID := hederahelper.ToMirrorNodeTransactionID(transactionResponse.TransactionID.String())
		s.logger.Errorf("[%s] - TX [%s] - Scheduled Transaction resolved with [%s].", id, txID, txReceipt.Status)

		onExecutionFail(txID)
		return errors.New(fmt.Sprintf("receipt-status: %s", txReceipt.Status))
	}

	transactionID := hederahelper.ToMirrorNodeTransactionID(txReceipt.ScheduledTransactionID.String())
	onExecutionSuccess(transactionID, txReceipt.ScheduleID.String())

	onMinedSuccess := func() {
		onSuccess(transactionID)
	}

	onMinedFail := func() {
		onFail(transactionID)
	}
	go s.mirrorNodeClient.WaitForScheduledTransaction(transactionID, onMinedSuccess, onMinedFail)
	return nil
}

func (s *Service) handleScheduleSign(id string, scheduleID hedera.ScheduleID) {
	s.logger.Debugf("[%s] - Scheduled transaction already created - Executing Scheduled Sign for [%s].", id, scheduleID)
	txResponse, err := s.hederaNodeClient.SubmitScheduleSign(scheduleID)
	if err != nil {
		s.logger.Errorf("[%s] - Failed to submit schedule sign [%s]. Error: [%s].", id, scheduleID, err)
		return
	}

	receipt, err := txResponse.GetReceipt(s.hederaNodeClient.GetClient())
	if err != nil {
		s.logger.Errorf("[%s] - Failed to get transaction receipt for schedule sign [%s]. Error: [%s].", id, scheduleID, err)
		return
	}

	switch receipt.Status {
	case hedera.StatusSuccess:
		s.logger.Debugf("[%s] - Successfully executed schedule sign for [%s].", id, scheduleID)
	case hedera.StatusScheduleAlreadyExecuted:
		s.logger.Debugf("[%s] - Scheduled Sign [%s] already executed.", id, scheduleID)
	default:
		s.logger.Errorf("[%s] - Schedule Sign [%s] failed with [%s].", id, scheduleID, receipt.Status)
	}
}

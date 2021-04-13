package burn_event

import (
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper"
	burn_event "github.com/limechain/hedera-eth-bridge-validator/app/model/burn-event"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	thresholdAccount hedera.AccountID
	payerAccount     hedera.AccountID
	repository       repository.BurnEvent
	hederaNodeClient client.HederaNode
	mirrorNodeClient client.MirrorNode
	logger           *log.Entry
}

func NewService(
	thresholdAccount string,
	payerAccount string,
	hederaNodeClient client.HederaNode,
	mirrorNodeClient client.MirrorNode,
	repository repository.BurnEvent) *Service {

	threshold, err := hedera.AccountIDFromString(thresholdAccount)
	if err != nil {
		log.Fatalf("Invalid bridge threshold account: [%s].", thresholdAccount)
	}

	payer, err := hedera.AccountIDFromString(payerAccount)
	if err != nil {
		log.Fatalf("Invalid payer account: [%s].", payerAccount)
	}

	return &Service{
		thresholdAccount: threshold,
		payerAccount:     payer,
		repository:       repository,
		hederaNodeClient: hederaNodeClient,
		mirrorNodeClient: mirrorNodeClient,
		logger:           config.GetLoggerFor("Burn Event Service"),
	}
}

func (s Service) ProcessEvent(event burn_event.BurnEvent) {
	err := s.repository.Create(event.TxHash, event.Amount, event.Recipient.String())
	if err != nil {
		s.logger.Errorf("[%s] - Failed to create a burn event record. Error [%s].", event.TxHash, err)
		return
	}

	var transactionResponse *hedera.TransactionResponse
	if event.NativeToken == "HBAR" {
		transactionResponse, err = s.hederaNodeClient.
			SubmitScheduledHbarTransferTransaction(event.Amount, event.Recipient, s.thresholdAccount, s.payerAccount, event.TxHash)
	} else {
		tokenID, err := hedera.TokenIDFromString(event.NativeToken)
		if err != nil {
			s.logger.Errorf("[%s] - failed to parse native token [%s] to TokenID. Error [%s].", event.TxHash, event.NativeToken, err)
			return
		}
		transactionResponse, err = s.hederaNodeClient.
			SubmitScheduledTokenTransferTransaction(event.Amount, tokenID, event.Recipient, s.thresholdAccount, s.payerAccount, event.TxHash)
	}
	if err != nil {
		s.logger.Errorf("[%s] - Failed to submit scheduled transaction. Error [%s].", event.TxHash, err)
		return
	}

	s.logger.Infof("[%s] - Successfully submitted scheduled transaction [%s] for [%s] to receive [%d] tinybars.",
		event.TxHash,
		transactionResponse.TransactionID, event.Recipient, event.Amount)

	txReceipt, err := transactionResponse.GetReceipt(s.hederaNodeClient.GetClient())
	if err != nil {
		s.logger.Errorf("[%s] - Failed to get transaction receipt for [%s]", event.TxHash, transactionResponse.TransactionID)
		return
	}

	switch txReceipt.Status {
	case hedera.StatusIdenticalScheduleAlreadyCreated:
		s.handleScheduleSign(event.TxHash, helper.ToMirrorNodeTransactionID(txReceipt.ScheduledTransactionID.String()), *txReceipt.ScheduleID)
	case hedera.StatusSuccess:
		transactionID := helper.ToMirrorNodeTransactionID(txReceipt.ScheduledTransactionID.String())
		s.logger.Infof("[%s] - Updating db status to Submitted with TransactionID [%s].", event.TxHash, transactionID)

		err := s.repository.UpdateStatusSubmitted(event.TxHash, txReceipt.ScheduleID.String(), transactionID)
		if err != nil {
			s.logger.Errorf(
				"[%s] - Failed to update submitted status with TransactionID [%s], ScheduleID [%s]. Error [%s].",
				event.TxHash, transactionID, txReceipt.ScheduleID, err)
			return
		}
	default:
		s.logger.Errorf("[%s] - TX [%s] - Scheduled Transaction resolved with [%s].", event.TxHash, transactionResponse.TransactionID, txReceipt.Status)

		err := s.repository.UpdateStatusFailed(transactionResponse.TransactionID.String())
		if err != nil {
			s.logger.Errorf("[%s] - Failed to update status failed. Error [%s].", transactionResponse.TransactionID.String(), err)
			return
		}
		return
	}

	transactionID := helper.ToMirrorNodeTransactionID(txReceipt.ScheduledTransactionID.String())

	onSuccess, onFail := s.scheduledTxExecutionCallbacks(transactionID)
	s.mirrorNodeClient.WaitForScheduledTransferTransaction(transactionID, onSuccess, onFail)
}

func (s *Service) handleScheduleSign(txHash, transactionID string, scheduleID hedera.ScheduleID) {
	s.logger.Debugf("[%s] - Scheduled transaction already created - Executing Scheduled Sign for [%s].", txHash, scheduleID)
	txResponse, err := s.hederaNodeClient.SubmitScheduleSign(scheduleID)
	if err != nil {
		s.logger.Errorf("[%s] - Failed to submit schedule sign [%s]. Error: [%s].", txHash, scheduleID, err)
		return
	}

	receipt, err := txResponse.GetReceipt(s.hederaNodeClient.GetClient())
	if err != nil {
		s.logger.Errorf("[%s] - Failed to get transaction receipt for schedule sign [%s]. Error: [%s].", txHash, scheduleID, err)
		return
	}

	if receipt.Status != hedera.StatusSuccess {
		s.logger.Errorf("[%s] - Schedule Sign [%s] failed with [%s].", txHash, scheduleID, receipt.Status)
	}
	s.logger.Infof("[%s] - Successfully executed schedule sign for [%s].", txHash, scheduleID)

	err = s.repository.UpdateStatusSubmitted(txHash, scheduleID.String(), transactionID)
	if err != nil {
		s.logger.Errorf(
			"[%s] - Failed to update submitted status with TransactionID [%s], ScheduleID [%s]. Error [%s].",
			txHash, transactionID, scheduleID, err)
		return
	}
}

func (s *Service) scheduledTxExecutionCallbacks(txId string) (onSuccess, onFail func()) {
	onSuccess = func() {
		s.logger.Debugf("[%s] - Scheduled TX execution successful.", txId)
		err := s.repository.UpdateStatusCompleted(txId)
		if err != nil {
			s.logger.Errorf("[%s] - Failed to update status completed. Error [%s].", txId, err)
			return
		}
	}

	onFail = func() {
		s.logger.Debugf("[%s] - Scheduled TX execution has failed.", txId)
		err := s.repository.UpdateStatusFailed(txId)
		if err != nil {
			s.logger.Errorf("[%s] - Failed to update status signature failed. Error [%s].", txId, err)
			return
		}
	}
	return onSuccess, onFail
}

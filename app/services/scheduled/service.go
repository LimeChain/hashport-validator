package scheduled

import (
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	hederahelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/hedera"
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
		logger:           config.GetLoggerFor("Burn Event Service"),
	}
}

// Execute submits a scheduled transaction and executes provided functions when necessary
func (s *Service) Execute(
	id, nativeAsset string,
	transfers []transfer.Hedera,
	onExecutionSuccess func(transactionID, scheduleID string), onExecutionFail, onSuccess, onFail func(transactionID string)) {
	transactionResponse, err := s.executeScheduledTransaction(id, nativeAsset, transfers)
	if err != nil {
		s.logger.Errorf("[%s] - Failed to submit scheduled transaction. Error [%s].", id, err)
		onExecutionFail(hederahelper.ToMirrorNodeTransactionID(transactionResponse.TransactionID.String()))
		return
	}

	scheduledTxID := hederahelper.ToMirrorNodeTransactionID(transactionResponse.TransactionID.String())
	s.logger.Infof("[%s] - Successfully submitted scheduled transaction [%s].",
		id,
		scheduledTxID)

	txReceipt, err := transactionResponse.GetReceipt(s.hederaNodeClient.GetClient())
	if err != nil {
		s.logger.Errorf("[%s] - Failed to get transaction receipt for [%s]", id, transactionResponse.TransactionID)
		onExecutionFail(scheduledTxID)
		return
	}

	switch txReceipt.Status {
	case hedera.StatusIdenticalScheduleAlreadyCreated:
		s.handleScheduleSign(id, *txReceipt.ScheduleID)
	case hedera.StatusSuccess:
		s.logger.Infof("[%s] - Updating db status to Submitted with TransactionID [%s].",
			id,
			hederahelper.ToMirrorNodeTransactionID(txReceipt.ScheduledTransactionID.String()))
	default:
		txID := hederahelper.ToMirrorNodeTransactionID(transactionResponse.TransactionID.String())
		s.logger.Errorf("[%s] - TX [%s] - Scheduled Transaction resolved with [%s].", id, txID, txReceipt.Status)

		onExecutionFail(txID)
		return
	}

	transactionID := hederahelper.ToMirrorNodeTransactionID(txReceipt.ScheduledTransactionID.String())
	onExecutionSuccess(transactionID, txReceipt.ScheduleID.String())

	onMinedSuccess := func() {
		onSuccess(transactionID)
	}

	onMinedFail := func() {
		onFail(transactionID)
	}

	s.mirrorNodeClient.WaitForScheduledTransferTransaction(transactionID, onMinedSuccess, onMinedFail)
}

func (s *Service) executeScheduledTransaction(id, nativeAsset string, transfers []transfer.Hedera) (*hedera.TransactionResponse, error) {
	var tokenID hedera.TokenID
	var transactionResponse *hedera.TransactionResponse
	var err error

	if nativeAsset == constants.Hbar {
		transactionResponse, err = s.hederaNodeClient.
			SubmitScheduledHbarTransferTransaction(transfers, s.payerAccount, id)
	} else {
		tokenID, err = hedera.TokenIDFromString(nativeAsset)
		if err != nil {
			s.logger.Errorf("[%s] - failed to parse native token [%s] to TokenID. Error [%s].", id, nativeAsset, err)
			return nil, err
		}
		transactionResponse, err = s.hederaNodeClient.
			SubmitScheduledTokenTransferTransaction(tokenID, transfers, s.payerAccount, id)
	}

	return transactionResponse, err
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

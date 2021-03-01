package recovery

import (
	"errors"
	"fmt"
	hederasdk "github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	processutils "github.com/limechain/hedera-eth-bridge-validator/app/helper/process"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
	tx "github.com/limechain/hedera-eth-bridge-validator/app/process/model/transaction"
	consensusmessage "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/consensus-message"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/process"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
	validatorproto "github.com/limechain/hedera-eth-bridge-validator/proto"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

type RecoveryService struct {
	transactionRepository   repositories.TransactionRepository
	topicStatusRepository   repositories.StatusRepository
	accountStatusRepository repositories.StatusRepository
	mirrorClient            *hedera.HederaMirrorClient
	nodeClient              *hedera.HederaNodeClient
	accountID               hederasdk.AccountID
	topicID                 hederasdk.TopicID
	configTimestamp         int64
	logger                  *log.Entry
	processingService       *process.ProcessingService
}

func NewRecoveryService(
	processingService *process.ProcessingService,
	transactionRepository repositories.TransactionRepository,
	topicStatusRepository repositories.StatusRepository,
	accountStatusRepository repositories.StatusRepository,
	mirrorClient *hedera.HederaMirrorClient,
	nodeClient *hedera.HederaNodeClient,
	accountID hederasdk.AccountID,
	topicID hederasdk.TopicID,
) *RecoveryService {
	return &RecoveryService{
		processingService:       processingService,
		transactionRepository:   transactionRepository,
		topicStatusRepository:   topicStatusRepository,
		accountStatusRepository: accountStatusRepository,
		mirrorClient:            mirrorClient,
		nodeClient:              nodeClient,
		accountID:               accountID,
		topicID:                 topicID,
		logger:                  config.GetLoggerFor(fmt.Sprintf("Recovery Service")),
	}
}

func (rs *RecoveryService) Recover() (int64, error) {
	log.Infof("Crypto Transfer Recovery for Account [%s]", rs.accountID.String())
	now, err := rs.cryptoTransferRecovery()
	if err != nil {
		rs.logger.Errorf("Error - could not finish crypto transfer recovery process: [%s]", err)
		return 0, err
	}
	log.Infof("[SUCCESSFUL] Crypto Transfer Recovery for Account [%s]", rs.accountID.String())

	log.Infof("Consensus Message Recovery for Topic [%s]", rs.topicID.String())
	now, err = rs.consensusMessageRecovery(now)
	if err != nil {
		rs.logger.Errorf("Error - could not finish consensus message recovery process: [%s]", err)
		return 0, err
	}
	log.Infof("[SUCCESSFUL] Consensus Message Recovery for Topic [%s]", rs.topicID.String())

	// TODO Handle unprocessed TXs
	// 1. Get all Skipped TX
	// 2. Get all message records for the set of TX IDs (from the Skipped TX records)
	// 3. Group messages and TX IDs into a map (TX ID->Messages)
	// 4. Go through all TX ID -> Messages. If current validator node haven't submitted a signature message -> sign and submit signature message to topic

	return now, nil
}

func (rs *RecoveryService) cryptoTransferRecovery() (int64, error) {
	now := time.Now().UnixNano()
	result, err := rs.mirrorClient.GetSuccessfulAccountCreditTransactionsAfterDate(rs.accountID, rs.getStartTimestampFor(rs.accountStatusRepository, rs.accountID.String()))
	if err != nil {
		return 0, err
	}

	rs.logger.Debugf("Found [%d] unprocessed transactions", len(result.Transactions))
	for _, tr := range result.Transactions {
		if ifRecent(tr, now) {
			break
		}

		memoInfo, err := processutils.DecodeMemo(tr.MemoBase64)
		if err != nil {
			rs.logger.Errorf("Could not decode memo for Transaction with ID [%s] - Error: [%s]", tr.TransactionID, err)
			continue
		}

		rs.logger.Debugf("Adding a transaction with ID [%s] unprocessed transactions with status [%s]", tr.TransactionID, transaction.StatusSkipped)

		err = rs.transactionRepository.Skip(&proto.CryptoTransferMessage{
			TransactionId: tr.TransactionID,
			EthAddress:    memoInfo.EthAddress,
			Amount:        strconv.Itoa(int(processutils.ExtractAmount(tr, rs.accountID))),
			Fee:           memoInfo.Fee,
		})

		if err != nil {
			return 0, err
		}
	}

	return now, nil
}

// TODO -> have blocking channel in order for the recovery to complete before starting the node
func (rs *RecoveryService) consensusMessageRecovery(now int64) (int64, error) {
	_, err := hederasdk.NewTopicMessageQuery().
		SetStartTime(time.Unix(0, rs.getStartTimestampFor(rs.topicStatusRepository, rs.topicID.String()))).
		SetEndTime(time.Unix(0, now)).
		SetTopicID(rs.topicID).
		Subscribe(
			rs.nodeClient.GetClient(),
			func(response hederasdk.TopicMessage) {
				m, err := consensusmessage.PrepareMessage(response.Contents, response.ConsensusTimestamp.UnixNano())
				if err != nil {
					return
				}
				switch m.Type {
				case validatorproto.TopicSubmissionType_EthSignature:
					_, _, err = rs.processingService.ValidateAndSaveSignature(m)
				case validatorproto.TopicSubmissionType_EthTransaction:
					err = rs.checkStatusAndUpdate(m.GetTopicEthTransactionMessage())
				default:
					err = errors.New(fmt.Sprintf("Error - invalid topic submission message type [%s]", m.Type))
				}

				if err != nil {
					rs.logger.Errorf("Error - could not handle recovery payload: [%s]", err)
					return
				}
			},
		)

	if err != nil {
		rs.logger.Errorf("Error - could not retrieve messages for recovery: [%s]", err)
		return 0, err
	}

	return now, nil
}

func (rs *RecoveryService) getStartTimestampFor(repository repositories.StatusRepository, address string) int64 {
	timestamp, err := repository.GetLastFetchedTimestamp(address)
	if err == nil {
		return timestamp
	}

	if rs.configTimestamp > 0 {
		return rs.configTimestamp
	}

	return time.Now().UnixNano()
}

func (rs *RecoveryService) checkStatusAndUpdate(m *validatorproto.TopicEthTransactionMessage) error {
	err := rs.transactionRepository.UpdateStatusEthTxSubmitted(m.TransactionId, m.EthTxHash)
	if err != nil {
		rs.logger.Errorf("Failed to update status to [%s] of transaction with TransactionID [%s]. Error [%s].", transaction.StatusEthTxSubmitted, m.TransactionId, err)
		return err
	}

	go rs.processingService.AcknowledgeTransactionSuccess(m)
	return nil
}

func ifRecent(tr tx.HederaTransaction, now int64) bool {
	consensusTimestampParams := strings.Split(tr.ConsensusTimestamp, ".")
	microseconds, _ := strconv.ParseInt(consensusTimestampParams[0], 10, 64)
	nanoseconds, _ := strconv.ParseInt(consensusTimestampParams[1], 10, 64)
	ct := microseconds*1000 + nanoseconds
	if ct > now {
		return true
	}
	return false
}

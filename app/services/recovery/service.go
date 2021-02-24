package recovery

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	hederasdk "github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	ethhelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/ethereum"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/process"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
	tx "github.com/limechain/hedera-eth-bridge-validator/app/process/model/transaction"
	consensusmessage "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/consensus-message"
	cryptotransfer "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/crypto-transfer"
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
	messageRepository       repositories.MessageRepository
	mirrorClient            *hedera.HederaMirrorClient
	ethereumClient          *ethereum.EthereumClient
	nodeClient              *hedera.HederaNodeClient
	accountID               hederasdk.AccountID
	topicID                 hederasdk.TopicID
	configTimestamp         int64
	operatorsEthAddresses   []string
	logger                  *log.Entry
}

func NewRecoveryService(
	transactionRepository repositories.TransactionRepository,
	topicStatusRepository repositories.StatusRepository,
	accountStatusRepository repositories.StatusRepository,
	messageRepository repositories.MessageRepository,
	mirrorClient *hedera.HederaMirrorClient,
	ethereumClient *ethereum.EthereumClient,
	nodeClient *hedera.HederaNodeClient,
	accountID hederasdk.AccountID,
	topicID hederasdk.TopicID,
) *RecoveryService {
	return &RecoveryService{
		transactionRepository:   transactionRepository,
		topicStatusRepository:   topicStatusRepository,
		accountStatusRepository: accountStatusRepository,
		messageRepository:       messageRepository,
		mirrorClient:            mirrorClient,
		ethereumClient:          ethereumClient,
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

	return now, nil
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

func (rs *RecoveryService) cryptoTransferRecovery() (int64, error) {
	now := time.Now().UnixNano()
	result, err := rs.mirrorClient.GetSuccessfulAccountCreditTransactionsAfterDate(rs.accountID, rs.getStartTimestampFor(rs.accountStatusRepository, rs.accountID.String()))
	if err != nil {
		// TODO: Log error properly
		return 0, err
	}

	rs.logger.Debugf("Found [%d] unprocessed transactions", len(result.Transactions))
	for _, tr := range result.Transactions {
		if ifRecent(tr, now) {
			break
		}

		memoInfo, err := cryptotransfer.DecodeMemo(tr.MemoBase64)
		if err != nil {
			rs.logger.Errorf("Could not decode memo for Transaction with ID [%s] - Error: [%s]", tr.TransactionID, err)
			continue
		}

		rs.logger.Debugf("Adding a transaction with ID [%s] unprocessed transactions with status [%s]", tr.TransactionID, transaction.StatusSkipped)

		err = rs.transactionRepository.Skip(&proto.CryptoTransferMessage{
			TransactionId: tr.TransactionID,
			EthAddress:    memoInfo.EthAddress,
			Amount:        strconv.Itoa(int(cryptotransfer.ExtractAmount(tr, rs.accountID))),
			Fee:           memoInfo.Fee,
		})

		if err != nil {
			return 0, err
		}
	}

	return now, nil
}

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
					err = rs.validateAndSaveSignature(m)
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

func (rs *RecoveryService) validateAndSaveSignature(msg *validatorproto.TopicSubmissionMessage) error {
	m := msg.GetTopicSignatureMessage()
	ctm := &validatorproto.CryptoTransferMessage{
		TransactionId: m.TransactionId,
		EthAddress:    m.EthAddress,
		Amount:        m.Amount,
		Fee:           m.Fee,
	}

	rs.logger.Debugf("Signature for TX ID [%s] was received", m.TransactionId)

	encodedData, err := ethhelper.EncodeData(ctm)
	if err != nil {
		rs.logger.Errorf("Failed to encode data for TransactionID [%s]. Error [%s].", ctm.TransactionId, err)
	}

	hash := crypto.Keccak256(encodedData)
	hexHash := hex.EncodeToString(hash)

	decodedSig, ethSig, err := ethhelper.DecodeSignature(m.GetSignature())
	m.Signature = ethSig
	if err != nil {
		return errors.New(fmt.Sprintf("[%s] - Failed to decode signature. - [%s]", m.TransactionId, err))
	}

	exists, err := process.AlreadyExists(rs.messageRepository, m, ethSig, hexHash)
	if err != nil {
		return err
	}
	if exists {
		return errors.New(fmt.Sprintf("Duplicated Transaction Id and Signature - [%s]-[%s]", m.TransactionId, m.Signature))
	}

	key, err := crypto.Ecrecover(hash, decodedSig)
	if err != nil {
		return errors.New(fmt.Sprintf("[%s] - Failed to recover public key. Hash - [%s] - [%s]", m.TransactionId, hexHash, err))
	}

	pubKey, err := crypto.UnmarshalPubkey(key)
	if err != nil {
		return errors.New(fmt.Sprintf("[%s] - Failed to unmarshal public key. - [%s]", m.TransactionId, err))
	}

	address := crypto.PubkeyToAddress(*pubKey)

	if process.IsValidAddress(address.String(), rs.operatorsEthAddresses) {
		return errors.New(fmt.Sprintf("[%s] - Address is not valid - [%s]", m.TransactionId, address.String()))
	}

	err = rs.messageRepository.Create(&message.TransactionMessage{
		TransactionId:        m.TransactionId,
		EthAddress:           m.EthAddress,
		Amount:               m.Amount,
		Fee:                  m.Fee,
		Signature:            ethSig,
		Hash:                 hexHash,
		SignerAddress:        address.String(),
		TransactionTimestamp: msg.TransactionTimestamp,
	})
	if err != nil {
		return errors.New(fmt.Sprintf("Could not add Transaction Message with Transaction Id and Signature - [%s]-[%s] - [%s]", m.TransactionId, ethSig, err))
	}

	rs.logger.Debugf("Verified and saved signature for TX ID [%s]", m.TransactionId)
	return nil
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

	go process.AcknowledgeTransactionSuccess(m, rs.logger, rs.ethereumClient, rs.transactionRepository)
	return nil
}

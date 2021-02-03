package consensusmessage

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/protobuf/proto"
	"github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum"
	hederaClient "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	ethhelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/ethereum"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/model/ethsubmission"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/scheduler"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/signer/eth"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	validatorproto "github.com/limechain/hedera-eth-bridge-validator/proto"
	"github.com/limechain/hedera-watcher-sdk/queue"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"strings"
)

type ConsensusMessageHandler struct {
	ethereumClient        *ethereum.EthereumClient
	hederaNodeClient      *hederaClient.HederaNodeClient
	operatorsEthAddresses []string
	messageRepository     repositories.MessageRepository
	transactionRepository repositories.TransactionRepository
	scheduler             *scheduler.Scheduler
	signer                *eth.Signer
	topicID               hedera.TopicID
	logger                *log.Entry
}

func NewConsensusMessageHandler(
	configuration config.ConsensusMessageHandler,
	messageRepository repositories.MessageRepository,
	transactionRepository repositories.TransactionRepository,
	ethereumClient *ethereum.EthereumClient,
	hederaNodeClient *hederaClient.HederaNodeClient,
	scheduler *scheduler.Scheduler,
	signer *eth.Signer,
) *ConsensusMessageHandler {
	topicID, err := hedera.TopicIDFromString(configuration.TopicId)
	if err != nil {
		log.Fatalf("Invalid topic id: [%v]", configuration.TopicId)
	}

	return &ConsensusMessageHandler{
		messageRepository:     messageRepository,
		transactionRepository: transactionRepository,
		operatorsEthAddresses: configuration.Addresses,
		hederaNodeClient:      hederaNodeClient,
		ethereumClient:        ethereumClient,
		topicID:               topicID,
		scheduler:             scheduler,
		signer:                signer,
		logger:                config.GetLoggerFor(fmt.Sprintf("Topic [%s] Handler", topicID.String())),
	}
}

func (cmh ConsensusMessageHandler) Recover(queue *queue.Queue) {
	cmh.logger.Println("Recovery method not implemented yet.")
}

func (cmh ConsensusMessageHandler) Handle(payload []byte) {
	m := &validatorproto.TopicSubmissionMessage{}
	err := proto.Unmarshal(payload, m)
	if err != nil {
		log.Errorf("Error could not unmarshal payload. Error [%s].", err)
	}

	switch m.Type {
	case validatorproto.TopicSubmissionType_EthSignature:
		err = cmh.handleSignatureMessage(m)
	case validatorproto.TopicSubmissionType_EthTransaction:
		err = cmh.handleEthTxMessage(m.GetTopicEthTransactionMessage())
	default:
		err = errors.New(fmt.Sprintf("Error - invalid topic submission message type [%s]", m.Type))
	}

	if err != nil {
		cmh.logger.Errorf("Error - could not handle payload: [%s]", err)
		return
	}
}

func (cmh ConsensusMessageHandler) handleEthTxMessage(m *validatorproto.TopicEthTransactionMessage) error {
	err := cmh.transactionRepository.UpdateStatusEthTxSubmitted(m.TransactionId)
	if err != nil {
		cmh.logger.Errorf("Failed to update status to [ETH_TX_SUBMITTED] of transaction with TransactionID [%s]. Error [%s].", m.TransactionId, err)
		return err
	}

	isSuccessful, err := cmh.ethereumClient.WaitForTransactionSuccess(common.HexToHash(m.Hash))
	if err != nil {
		cmh.logger.Errorf("Failed await a transaction with Id [%s] and Hash [%s]. Error [%s].", m.TransactionId, m.Hash, err)
		return err
	}

	if !isSuccessful {
		cmh.transactionRepository.UpdateStatusEthTxReverted(m.TransactionId)
		cmh.logger.Errorf("Failed to update status to [ETH_TX_REVERTED] of transaction with TransactionID [%s]. Error [%s].", m.TransactionId, err)
		return err
	}

	err = cmh.transactionRepository.UpdateStatusCompleted(m.TransactionId)
	if err != nil {
		cmh.logger.Errorf("Failed to update status to [COMPLETED] of transaction with TransactionID [%s]. Error [%s].", m.TransactionId, err)
		return err
	}

	return cmh.scheduler.Cancel(m.TransactionId)
}

func (cmh ConsensusMessageHandler) handleSignatureMessage(msg *validatorproto.TopicSubmissionMessage) error {
	m := msg.GetTopicSignatureMessage()
	ctm := &validatorproto.CryptoTransferMessage{
		TransactionId: m.TransactionId,
		EthAddress:    m.EthAddress,
		Amount:        m.Amount,
		Fee:           m.Fee,
	}

	cmh.logger.Debugf("Signature for TX ID [%s] was received", m.TransactionId)

	encodedData, err := ethhelper.EncodeData(ctm)
	if err != nil {
		cmh.logger.Errorf("Failed to encode data for TransactionID [%s]. Error [%s].", ctm.TransactionId, err)
	}

	hash := crypto.Keccak256(encodedData)
	hexHash := hex.EncodeToString(hash)

	decodedSig, ethSig, err := ethhelper.DecodeSignature(m.GetSignature())
	m.Signature = ethSig
	if err != nil {
		return errors.New(fmt.Sprintf("[%s] - Failed to decode signature. - [%s]", m.TransactionId, err))
	}

	exists, err := cmh.alreadyExists(m, ethSig, hexHash)
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

	if !cmh.isValidAddress(address.String()) {
		return errors.New(fmt.Sprintf("[%s] - Address is not valid - [%s]", m.TransactionId, address.String()))
	}

	err = cmh.messageRepository.Create(&message.TransactionMessage{
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

	cmh.logger.Debugf("Verified and saved signature for TX ID [%s]", m.TransactionId)

	txSignatures, err := cmh.messageRepository.GetTransactions(m.TransactionId, hexHash)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not retrieve transaction messages for Transaction ID [%s]. Error [%s]", m.TransactionId))
	}

	if cmh.enoughSignaturesCollected(txSignatures, m.TransactionId) {
		submission := &ethsubmission.Submission{
			TransactOps:           cmh.signer.NewKeyTransactor(),
			CryptoTransferMessage: ctm,
			Messages:              txSignatures,
		}
		err := cmh.scheduler.Schedule(m.TransactionId, *submission)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cmh ConsensusMessageHandler) alreadyExists(m *validatorproto.TopicEthSignatureMessage, ethSig, hexHash string) (bool, error) {
	_, err := cmh.messageRepository.GetTransaction(m.TransactionId, ethSig, hexHash)
	notFound := errors.Is(err, gorm.ErrRecordNotFound)

	if err != nil && !notFound {
		return false, errors.New(fmt.Sprintf("Failed to retrieve messages for TxId [%s], with signature [%s]. - [%s]", m.TransactionId, m.Signature, err))
	}
	return !notFound, nil
}

func (cmh ConsensusMessageHandler) enoughSignaturesCollected(txSignatures []message.TransactionMessage, transactionId string) bool {
	requiredSigCount := len(cmh.operatorsEthAddresses)/2 + 1
	cmh.logger.Infof("Collected [%d/%d] Signatures for TX ID [%s] ", len(txSignatures), len(cmh.operatorsEthAddresses), transactionId)
	return len(txSignatures) >= requiredSigCount
}

func (cmh ConsensusMessageHandler) isValidAddress(key string) bool {
	for _, k := range cmh.operatorsEthAddresses {
		if strings.ToLower(k) == strings.ToLower(key) {
			return true
		}
	}
	return false
}

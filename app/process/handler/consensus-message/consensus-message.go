package consensusmessage

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/protobuf/proto"
	"github.com/hashgraph/hedera-sdk-go"
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
	hederaNodeClient      *hederaClient.HederaNodeClient
	operatorsEthAddresses []string
	repository            repositories.MessageRepository
	scheduler             *scheduler.Scheduler
	signer                *eth.Signer
	topicID               hedera.TopicID
	logger                *log.Entry
}

func NewConsensusMessageHandler(
	configuration config.ConsensusMessageHandler,
	r repositories.MessageRepository,
	hederaNodeClient *hederaClient.HederaNodeClient,
	scheduler *scheduler.Scheduler,
	signer *eth.Signer,
) *ConsensusMessageHandler {
	topicID, err := hedera.TopicIDFromString(configuration.TopicId)
	if err != nil {
		log.Fatal("Invalid topic id: [%v]", configuration.TopicId)
	}

	return &ConsensusMessageHandler{
		repository:            r,
		operatorsEthAddresses: configuration.Addresses,
		hederaNodeClient:      hederaNodeClient,
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
	// TODO: verify authenticity of transaction hash

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

	err = cmh.repository.Create(&message.TransactionMessage{
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

	txMessages, err := cmh.repository.GetTransactions(m.TransactionId, hexHash)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not retrieve transaction messages for Transaction ID [%s]. Error [%s]", m.TransactionId, err))
	}

	if cmh.enoughSignaturesCollected(txMessages, m.TransactionId) {
		cmh.logger.Debugf("Enough signatures have been collected for Transaction ID [%s].", m.TransactionId)

		slot, isFound := cmh.computeExecutionSlot(txMessages)
		if !isFound {
			cmh.logger.Debugf("Operator [%s] has not been found amongst the signatures collected for Transaction ID [%s].", cmh.signer.Address(), m.TransactionId)
			return nil
		}

		submission := &ethsubmission.Submission{
			CryptoTransferMessage: ctm,
			Messages:              txMessages,
			Slot:                  slot,
			TransactOps:           cmh.signer.NewKeyTransactor(),
		}

		err := cmh.scheduler.Schedule(m.TransactionId, *submission)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cmh ConsensusMessageHandler) alreadyExists(m *validatorproto.TopicEthSignatureMessage, ethSig, hexHash string) (bool, error) {
	_, err := cmh.repository.GetTransaction(m.TransactionId, ethSig, hexHash)
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

// computeExecutionSlot - computes the slot order in which the TX will execute
// Important! Transaction messages ARE expected to be sorted by ascending Timestamp
func (cmh ConsensusMessageHandler) computeExecutionSlot(messages []message.TransactionMessage) (slot int64, isFound bool) {
	for i := 0; i < len(messages); i++ {
		if strings.ToLower(messages[i].SignerAddress) == strings.ToLower(cmh.signer.Address()) {
			return int64(i), true
		}
	}

	return -1, false
}

func (cmh ConsensusMessageHandler) isValidAddress(key string) bool {
	for _, k := range cmh.operatorsEthAddresses {
		if strings.ToLower(k) == strings.ToLower(key) {
			return true
		}
	}
	return false
}

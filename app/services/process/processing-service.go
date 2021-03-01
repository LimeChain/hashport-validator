package process

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	ethhelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/ethereum"
	processutils "github.com/limechain/hedera-eth-bridge-validator/app/helper/process"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	validatorproto "github.com/limechain/hedera-eth-bridge-validator/proto"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ProcessingService struct {
	logger                *log.Entry
	ethereumClient        *ethereum.EthereumClient
	transactionRepository repositories.TransactionRepository
	messageRepository     repositories.MessageRepository
	operatorsEthAddresses []string
}

func NewProcessingService(ethereumClient *ethereum.EthereumClient,
	transactionRepository repositories.TransactionRepository,
	messageRepository repositories.MessageRepository,
	operatorsEthAddresses []string) *ProcessingService {
	return &ProcessingService{
		messageRepository:     messageRepository,
		transactionRepository: transactionRepository,
		ethereumClient:        ethereumClient,
		operatorsEthAddresses: operatorsEthAddresses,
		logger:                config.GetLoggerFor(fmt.Sprintf("Processing Service")),
	}
}

func (ps *ProcessingService) AcknowledgeTransactionSuccess(m *validatorproto.TopicEthTransactionMessage) {
	ps.logger.Infof("Waiting for Transaction with ID [%s] to be mined.", m.TransactionId)

	isSuccessful, err := ps.ethereumClient.WaitForTransactionSuccess(common.HexToHash(m.EthTxHash))
	if err != nil {
		ps.logger.Errorf("Failed to await TX ID [%s] with ETH TX [%s] to be mined. Error [%s].", m.TransactionId, m.Hash, err)
		return
	}

	if !isSuccessful {
		ps.logger.Infof("Transaction with ID [%s] was reverted. Updating status to [%s].", m.TransactionId, transaction.StatusEthTxReverted)
		err = ps.transactionRepository.UpdateStatusEthTxReverted(m.TransactionId)
		if err != nil {
			ps.logger.Errorf("Failed to update status to [%s] of transaction with TransactionID [%s]. Error [%s].", transaction.StatusEthTxReverted, m.TransactionId, err)
			return
		}
	} else {
		ps.logger.Infof("Transaction with ID [%s] was successfully mined. Updating status to [%s].", m.TransactionId, transaction.StatusCompleted)
		err = ps.transactionRepository.UpdateStatusCompleted(m.TransactionId)
		if err != nil {
			ps.logger.Errorf("Failed to update status to [%s] of transaction with TransactionID [%s]. Error [%s].", transaction.StatusCompleted, m.TransactionId, err)
			return
		}
	}
}

func (ps *ProcessingService) AlreadyExists(m *validatorproto.TopicEthSignatureMessage, ethSig, hexHash string) (bool, error) {
	_, err := ps.messageRepository.GetTransaction(m.TransactionId, ethSig, hexHash)
	notFound := errors.Is(err, gorm.ErrRecordNotFound)

	if err != nil && !notFound {
		return false, errors.New(fmt.Sprintf("Failed to retrieve messages for TxId [%s], with signature [%s]. - [%s]", m.TransactionId, m.Signature, err))
	}
	return !notFound, nil
}

func (ps *ProcessingService) ValidateAndSaveSignature(msg *validatorproto.TopicSubmissionMessage) (string, *validatorproto.CryptoTransferMessage, error) {
	m := msg.GetTopicSignatureMessage()
	ctm := &validatorproto.CryptoTransferMessage{
		TransactionId: m.TransactionId,
		EthAddress:    m.EthAddress,
		Amount:        m.Amount,
		Fee:           m.Fee,
	}

	ps.logger.Debugf("Signature for TX ID [%s] was received", m.TransactionId)

	encodedData, err := ethhelper.EncodeData(ctm)
	if err != nil {
		ps.logger.Errorf("Failed to encode data for TransactionID [%s]. Error [%s].", ctm.TransactionId, err)
	}

	hash := crypto.Keccak256(encodedData)
	hexHash := hex.EncodeToString(hash)

	decodedSig, ethSig, err := ethhelper.DecodeSignature(m.GetSignature())
	m.Signature = ethSig
	if err != nil {
		return "", nil, errors.New(fmt.Sprintf("[%s] - Failed to decode signature. - [%s]", m.TransactionId, err))
	}

	exists, err := ps.AlreadyExists(m, ethSig, hexHash)
	if err != nil {
		return "", nil, err
	}
	if exists {
		return "", nil, errors.New(fmt.Sprintf("Duplicated Transaction Id and Signature - [%s]-[%s]", m.TransactionId, m.Signature))
	}

	key, err := crypto.Ecrecover(hash, decodedSig)
	if err != nil {
		return "", nil, errors.New(fmt.Sprintf("[%s] - Failed to recover public key. Hash - [%s] - [%s]", m.TransactionId, hexHash, err))
	}

	pubKey, err := crypto.UnmarshalPubkey(key)
	if err != nil {
		return "", nil, errors.New(fmt.Sprintf("[%s] - Failed to unmarshal public key. - [%s]", m.TransactionId, err))
	}

	address := crypto.PubkeyToAddress(*pubKey)

	if processutils.IsValidAddress(address.String(), ps.operatorsEthAddresses) {
		return "", nil, errors.New(fmt.Sprintf("[%s] - Address is not valid - [%s]", m.TransactionId, address.String()))
	}

	err = ps.messageRepository.Create(&message.TransactionMessage{
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
		return "", nil, errors.New(fmt.Sprintf("Could not add Transaction Message with Transaction Id and Signature - [%s]-[%s] - [%s]", m.TransactionId, ethSig, err))
	}

	ps.logger.Debugf("Verified and saved signature for TX ID [%s]", m.TransactionId)
	return hexHash, ctm, nil
}

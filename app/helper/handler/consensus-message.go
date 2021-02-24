package handler

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
	validatorproto "github.com/limechain/hedera-eth-bridge-validator/proto"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"strings"
)

func AcknowledgeTransactionSuccess(m *validatorproto.TopicEthTransactionMessage, logger *log.Entry, ethereumClient *ethereum.EthereumClient, transactionRepository repositories.TransactionRepository) {
	logger.Infof("Waiting for Transaction with ID [%s] to be mined.", m.TransactionId)

	isSuccessful, err := ethereumClient.WaitForTransactionSuccess(common.HexToHash(m.EthTxHash))
	if err != nil {
		logger.Errorf("Failed to await TX ID [%s] with ETH TX [%s] to be mined. Error [%s].", m.TransactionId, m.Hash, err)
		return
	}

	if !isSuccessful {
		logger.Infof("Transaction with ID [%s] was reverted. Updating status to [%s].", m.TransactionId, transaction.StatusEthTxReverted)
		err = transactionRepository.UpdateStatusEthTxReverted(m.TransactionId)
		if err != nil {
			logger.Errorf("Failed to update status to [%s] of transaction with TransactionID [%s]. Error [%s].", transaction.StatusEthTxReverted, m.TransactionId, err)
			return
		}
	} else {
		logger.Infof("Transaction with ID [%s] was successfully mined. Updating status to [%s].", m.TransactionId, transaction.StatusCompleted)
		err = transactionRepository.UpdateStatusCompleted(m.TransactionId)
		if err != nil {
			logger.Errorf("Failed to update status to [%s] of transaction with TransactionID [%s]. Error [%s].", transaction.StatusCompleted, m.TransactionId, err)
			return
		}
	}
}

func IsValidAddress(key string, operatorsEthAddresses []string) bool {
	for _, k := range operatorsEthAddresses {
		if strings.ToLower(k) == strings.ToLower(key) {
			return true
		}
	}
	return false
}

func AlreadyExists(messageRepository repositories.MessageRepository, m *validatorproto.TopicEthSignatureMessage, ethSig, hexHash string) (bool, error) {
	_, err := messageRepository.GetTransaction(m.TransactionId, ethSig, hexHash)
	notFound := errors.Is(err, gorm.ErrRecordNotFound)

	if err != nil && !notFound {
		return false, errors.New(fmt.Sprintf("Failed to retrieve messages for TxId [%s], with signature [%s]. - [%s]", m.TransactionId, m.Signature, err))
	}
	return !notFound, nil
}

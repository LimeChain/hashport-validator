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

package bridge

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/services"
	"github.com/limechain/hedera-eth-bridge-validator/app/encoding"
	"github.com/limechain/hedera-eth-bridge-validator/app/encoding/auth-message"
	"github.com/limechain/hedera-eth-bridge-validator/app/encoding/memo"
	ethhelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/ethereum"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"

	"github.com/limechain/hedera-eth-bridge-validator/app/domain/clients"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/fees"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/signer/eth"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	validatorproto "github.com/limechain/hedera-eth-bridge-validator/proto"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	logger                *log.Entry
	transactionRepository repositories.Transaction
	MessageRepository     repositories.Message
	feeCalculator         *fees.Calculator
	ethSigner             *eth.Signer
	topicID               hedera.TopicID
	clients               clients.Clients
	contractsService      services.Contracts
}

func NewService(
	clients clients.Clients,
	transactionRepository repositories.Transaction,
	messageRepository repositories.Message,
	contractsService services.Contracts,
	feeCalculator *fees.Calculator,
	ethSigner *eth.Signer,
	topicID string,
) *Service {
	tID, e := hedera.TopicIDFromString(topicID)
	if e != nil {
		panic(fmt.Sprintf("Invalid monitoring Topic ID [%s] - Error: [%s]", topicID, e))
	}

	return &Service{
		MessageRepository:     messageRepository,
		transactionRepository: transactionRepository,
		logger:                config.GetLoggerFor(fmt.Sprintf("Bridge Service")),
		feeCalculator:         feeCalculator,
		ethSigner:             ethSigner,
		contractsService:      contractsService,
		topicID:               tID,
		clients:               clients,
	}
}

// SanityCheck performs validation on the memo and state proof for the transaction
func (bs *Service) SanityCheck(tx mirror_node.Transaction) (*memo.Memo, error) {
	m, e := memo.FromBase64String(tx.MemoBase64)
	if e != nil {
		return nil, errors.New(fmt.Sprintf("Could not parse transaction memo. Error: [%s]", e))
	}

	// TODO
	//stateProof, e = bs.clients.MirrorNode.GetStateProof(tx.TransactionID)
	//if e != nil {
	//	return nil, errors.New(fmt.Sprintf("Could not GET state proof. Error [%s]", e))
	//}

	//verified, e := proof.Verify(tx.TransactionID, stateProof)
	//if e != nil {
	//	return nil, errors.New(fmt.Sprintf("State proof verification failed. Error [%s]"), e))
	//}
	//
	//if !verified {
	//	return nil, errors.New("State proof not valid")
	//}

	return m, nil
}

// InitiateNewTransfer Stores the incoming transfer message into the Database aware of already processed transactions
func (bs *Service) InitiateNewTransfer(tm encoding.TransferMessage) (*transaction.Transaction, error) {
	dbTransaction, err := bs.transactionRepository.GetByTransactionId(tm.TransactionId)
	if err != nil {
		bs.logger.Errorf("Failed to get record with TransactionID [%s]. Error [%s]", tm.TransactionId, err)
		return nil, err
	}

	if dbTransaction != nil {
		bs.logger.Infof("Transaction with ID [%s] already added", tm.TransactionId)
		return dbTransaction, err
	}

	bs.logger.Debugf("Adding new Transaction Record with Txn ID [%s]", tm.TransactionId)
	tx, err := bs.transactionRepository.Create(tm.TransferMessage)
	if err != nil {
		bs.logger.Errorf("Failed to create a transaction record for TransactionID [%s]. Error [%s].", tm.TransactionId, err)
		return nil, err
	}
	return tx, nil
}

// SaveRecoveredTxn creates new Transaction record persisting the recovered Transfer TXn
func (bs *Service) SaveRecoveredTxn(txId, amount string, m memo.Memo) error {
	err := bs.transactionRepository.SaveRecoveredTxn(&validatorproto.TransferMessage{
		TransactionId: txId,
		EthAddress:    m.EthereumAddress,
		Amount:        amount,
		Fee:           m.TxReimbursementFee,
		GasPriceGwei:  m.GasPriceGwei,
	})
	if err != nil {
		bs.logger.Errorf("Something went wrong while saving new Recovered Transaction with ID [%s]. Err: [%s]", txId, err)
		return err
	}

	bs.logger.Infof("Added new Transaction Record with Txn ID [%s]", txId)
	return err
}

// VerifyFee verifies that the provided TX reimbursement fee is enough using the
// Fee Calculator and updates the Transaction Record to Insufficient Fee if necessary
func (bs *Service) VerifyFee(tm encoding.TransferMessage) error {
	isSufficient, err := bs.feeCalculator.ValidateExecutionFee(tm.Fee, tm.Amount, tm.GasPriceGwei)
	if !isSufficient {
		bs.logger.Errorf("Fee validation for TX ID [%s] failed. Provided tx reimbursement fee is invalid/insufficient. Error [%s]", tm.TransactionId, err)
		err = bs.transactionRepository.UpdateStatusInsufficientFee(tm.TransactionId)
		if err != nil {
			bs.logger.Errorf("Failed to update status to [%s] of transaction with TransactionID [%s]. Error [%s].", transaction.StatusInsufficientFee, tm.TransactionId, err)
			return err
		}
		bs.logger.Debugf("TX with ID [%s] was updated to [%s]. Provided fee [%s]", tm.TransactionId, transaction.StatusInsufficientFee, tm.Fee)
		return err
	}
	return nil
}

// ProcessTransfer processes the transfer message by signing the required
// authorisation signature submitting it into the required HCS Topic
func (bs *Service) ProcessTransfer(tm encoding.TransferMessage) error {
	authMsgHash, err := auth_message.FromTransferMessage(tm)
	if err != nil {
		bs.logger.Errorf("Failed to encode the authorisation signature for TX ID [%s]. Error: %s", tm.TransactionId, err)
		return err
	}
	signatureBytes, err := bs.ethSigner.Sign(authMsgHash)
	if err != nil {
		bs.logger.Errorf("Failed to sign the authorisation signature for TX ID [%s]. Error: %s", tm.TransactionId, err)
		return err
	}
	signature := hex.EncodeToString(signatureBytes)
	signatureMessage := encoding.NewSignatureMessage(
		tm.TransactionId,
		tm.EthAddress,
		tm.Amount,
		tm.Fee,
		signature)

	sigMsgBytes, err := signatureMessage.ToBytes()
	messageTxId, err := bs.clients.HederaNode.SubmitTopicConsensusMessage(
		bs.topicID,
		sigMsgBytes)
	if err != nil {
		bs.logger.Errorf("Failed to submit Signature Message to Topic for TX [%s]. Error: %s", tm.TransactionId, err)
	}
	bs.logger.Infof("Submitted signature for TX ID [%s] on Topic [%s]", tm.TransactionId, bs.topicID)

	err = bs.transactionRepository.UpdateStatusSignatureSubmitted(tm.TransactionId, messageTxId.String(), signature)
	if err != nil {
		bs.logger.Errorf("Failed to update Status for TX [%s]. Error [%s].", tm.TransactionId, err)
		return err
	}

	onSuccessfulAuthMessage, onFailedAuthMessage := bs.authMessageSubmissionCallbacks(tm.TransactionId)
	bs.clients.MirrorNode.WaitForTransaction(messageTxId.String(), onSuccessfulAuthMessage, onFailedAuthMessage)

	bs.logger.Infof("Successfully processed Transfer with ID [%s]", tm.TransactionId)
	return nil
}

func (bs *Service) authMessageSubmissionCallbacks(txId string) (func(), func()) {
	onSuccessfulAuthMessage := func() {
		bs.logger.Infof("Publish Authorisation signature TX successfully executed for TX [%s]", txId)
		err := bs.transactionRepository.UpdateStatusCompleted(txId)
		if err != nil {
			bs.logger.Errorf("Failed to update status for TX [%s]. Error [%s].", txId, err)
			return
		}
	}

	onFailedAuthMessage := func() {
		bs.logger.Infof("Publish Authorisation signature TX failed for TX ID [%s]", txId)
		err := bs.transactionRepository.UpdateStatusEthTxReverted(txId)
		if err != nil {
			bs.logger.Errorf("Failed to update status for TX [%s]. Error [%s].", txId, err)
			return
		}
	}
	return onSuccessfulAuthMessage, onFailedAuthMessage
}

// ProcessSignature processes the signature message, verifying and updating all necessary fields in the DB
func (bs *Service) ProcessSignature(tm encoding.TopicMessage) error {
	// Parse incoming message
	topicMessage := tm.GetTopicSignatureMessage()
	authMsgBytes, err := auth_message.FromTopicMessage(topicMessage)
	if err != nil {
		bs.logger.Errorf("Failed to encode the authorisation signature for TX ID [%s]. Error: %s", topicMessage.TransactionId, err)
	}

	// Prepare Signature
	signatureBytes, signatureHex, err := ethhelper.DecodeSignature(topicMessage.GetSignature())
	if err != nil {
		bs.logger.Errorf("[%s] - Decoding Signature [%s] for TX failed. Err: %s", topicMessage.TransactionId, topicMessage.GetSignature())
		return err
	}
	authMessageStr := hex.EncodeToString(authMsgBytes)

	// Check for duplicated signature
	exists, err := bs.MessageRepository.Exist(topicMessage.TransactionId, signatureHex, authMessageStr)
	if err != nil {
		bs.logger.Errorf("An error occurred while getting TX [%s] from DB. Error: %s", topicMessage.TransactionId, err)
	}
	if exists {
		bs.logger.Errorf("[%s] - Signature [%s] already received for TX ID [%s]", topicMessage.TransactionId, topicMessage.GetSignature())
		return err
	}

	// Verify Signature
	address, err := bs.verifySignature(err, authMsgBytes, signatureBytes, topicMessage, authMessageStr)
	if err != nil {
		return err
	}

	// Persist in DB
	err = bs.MessageRepository.Create(&message.TransactionMessage{
		TransactionId:        topicMessage.TransactionId,
		EthAddress:           topicMessage.EthAddress,
		Amount:               topicMessage.Amount,
		Fee:                  topicMessage.Fee,
		Signature:            signatureHex,
		Hash:                 authMessageStr,
		SignerAddress:        address.String(),
		TransactionTimestamp: tm.TransactionTimestamp,
	})
	if err != nil {
		bs.logger.Errorf("[%s] - Failed to save Transaction Message in DB with Signature [%s]. Error: %s", topicMessage.TransactionId, signatureHex, err)
		return err
	}

	bs.logger.Infof("[%s] - Successfully processed Signature Message [%s]", topicMessage.TransactionId, signatureHex)
	return nil
}

func (bs *Service) verifySignature(err error, authMsgBytes []byte, signatureBytes []byte, topicMessage *validatorproto.TopicEthSignatureMessage, authMessageStr string) (common.Address, error) {
	publicKey, err := crypto.Ecrecover(authMsgBytes, signatureBytes)
	if err != nil {
		bs.logger.Errorf("[%s] - Failed to recover public key. Hash [%s]. Error: %s", topicMessage.TransactionId, authMessageStr, err)
		return common.Address{}, err
	}
	unmarshalledPublicKey, err := crypto.UnmarshalPubkey(publicKey)
	if err != nil {
		bs.logger.Errorf("[%s] - Failed to unmarshall public key. Error: %s", topicMessage.TransactionId, err)
		return common.Address{}, err
	}
	address := crypto.PubkeyToAddress(*unmarshalledPublicKey)
	if !bs.contractsService.IsMember(address.String()) {
		bs.logger.Errorf("[%s] - Received Signature [%s] is not signed by Bridge member", topicMessage.TransactionId, authMessageStr)
		return common.Address{}, errors.New(fmt.Sprintf("signer is not bridge member"))
	}
	return address, nil
}

// TODO ->

//func (bs *Service) AcknowledgeTransactionSuccess(m *validatorproto.TopicEthTransactionMessage) {
//	bs.logger.Infof("Waiting for Transaction with ID [%s] to be mined.", m.TransactionId)
//
//	isSuccessful, err := bs.clients.Ethereum.WaitForTransactionSuccess(common.HexToHash(m.EthTxHash))
//	if err != nil {
//		bs.logger.Errorf("Failed to await TX ID [%s] with ETH TX [%s] to be mined. Error [%s].", m.TransactionId, m.Hash, err)
//		return
//	}
//
//	if !isSuccessful {
//		bs.logger.Infof("Transaction with ID [%s] was reverted. Updating status to [%s].", m.TransactionId, transaction.StatusEthTxReverted)
//		err = bs.transactionRepository.UpdateStatusEthTxReverted(m.TransactionId)
//		if err != nil {
//			bs.logger.Errorf("Failed to update status to [%s] of transaction with TransactionID [%s]. Error [%s].", transaction.StatusEthTxReverted, m.TransactionId, err)
//			return
//		}
//	} else {
//		bs.logger.Infof("Transaction with ID [%s] was successfully mined. Updating status to [%s].", m.TransactionId, transaction.StatusCompleted)
//		err = bs.transactionRepository.UpdateStatusCompleted(m.TransactionId)
//		if err != nil {
//			bs.logger.Errorf("Failed to update status to [%s] of transaction with TransactionID [%s]. Error [%s].", transaction.StatusCompleted, m.TransactionId, err)
//			return
//		}
//	}
//}

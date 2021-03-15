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

package transfers

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go"
	mirror_node "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/encoding"
	auth_message "github.com/limechain/hedera-eth-bridge-validator/app/encoding/auth-message"
	"github.com/limechain/hedera-eth-bridge-validator/app/encoding/memo"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	validatorproto "github.com/limechain/hedera-eth-bridge-validator/proto"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	logger                *log.Entry
	hederaNode            client.HederaNode
	mirrorNode            client.MirrorNode
	fees                  service.Fees
	ethSigner             service.Signer
	transactionRepository repository.Transaction
	topicID               hedera.TopicID
}

func NewService(
	hederaNode client.HederaNode,
	mirrorNode client.MirrorNode,
	fees service.Fees,
	signer service.Signer,
	transactionRepository repository.Transaction,
	topicID string,
) *Service {
	tID, e := hedera.TopicIDFromString(topicID)
	if e != nil {
		panic(fmt.Sprintf("Invalid monitoring Topic ID [%s] - Error: [%s]", topicID, e))
	}

	return &Service{
		logger:                config.GetLoggerFor(fmt.Sprintf("Transfers Service")),
		hederaNode:            hederaNode,
		mirrorNode:            mirrorNode,
		fees:                  fees,
		ethSigner:             signer,
		transactionRepository: transactionRepository,
		topicID:               tID,
	}
}

// SanityCheck performs validation on the memo and state proof for the transaction
func (bs *Service) SanityCheckTransfer(tx mirror_node.Transaction) (*memo.Memo, error) {
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

	bs.logger.Debugf("Adding new Transaction Record TX ID [%s]", tm.TransactionId)
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
	isSufficient, err := bs.fees.ValidateExecutionFee(tm.Fee, tm.Amount, tm.GasPriceGwei)
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
	authMsgHash, err := auth_message.EncodeBytesFrom(tm.TransactionId, tm.EthAddress, tm.Amount, tm.Fee)
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
	if err != nil {
		bs.logger.Error("Failed to encode Signature Message to bytes for TX [%s]. Error %s", err, tm.TransactionId)
		return err
	}
	messageTxId, err := bs.hederaNode.SubmitTopicConsensusMessage(
		bs.topicID,
		sigMsgBytes)
	if err != nil {
		bs.logger.Errorf("Failed to submit Signature Message to Topic for TX [%s]. Error: %s", tm.TransactionId, err)
		return err
	}

	// Update Transaction Record
	tx, err := bs.transactionRepository.GetByTransactionId(tm.TransactionId)
	if err != nil {
		bs.logger.Errorf("Failed to get TX [%s] from DB", tm.TransactionId)
		return err
	}

	tx.Signature = signature
	tx.SignatureMsgTxId = messageTxId.String()
	tx.Status = transaction.StatusInProgress
	tx.SignatureMsgStatus = transaction.StatusSignatureSubmitted
	err = bs.transactionRepository.Save(tx)
	if err != nil {
		bs.logger.Errorf("Failed to update TX [%s]. Error [%s].", tm.TransactionId, err)
		return err
	}

	// Attach update callbacks on Signature HCS Message
	bs.logger.Infof("Submitted signature for TX ID [%s] on Topic [%s]", tm.TransactionId, bs.topicID)
	onSuccessfulAuthMessage, onFailedAuthMessage := bs.authMessageSubmissionCallbacks(tm.TransactionId)
	bs.mirrorNode.WaitForTransaction(messageTxId.String(), onSuccessfulAuthMessage, onFailedAuthMessage)
	return nil
}

func (bs *Service) authMessageSubmissionCallbacks(txId string) (func(), func()) {
	onSuccessfulAuthMessage := func() {
		bs.logger.Debugf("Authorisation Signature TX successfully executed for TX [%s]", txId)
		err := bs.transactionRepository.UpdateStatusSignatureMined(txId)
		if err != nil {
			bs.logger.Errorf("Failed to update status for TX [%s]. Error [%s].", txId, err)
			return
		}
	}

	onFailedAuthMessage := func() {
		bs.logger.Debugf("Authorisation Signature TX failed for TX ID [%s]", txId)
		err := bs.transactionRepository.UpdateStatusSignatureFailed(txId)
		if err != nil {
			bs.logger.Errorf("Failed to update status for TX [%s]. Error [%s].", txId, err)
			return
		}
	}
	return onSuccessfulAuthMessage, onFailedAuthMessage
}

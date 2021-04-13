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
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/hashgraph/hedera-state-proof-verifier-go/stateproof"
	mirror_node "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper"
	auth_message "github.com/limechain/hedera-eth-bridge-validator/app/model/auth-message"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/memo"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/message"
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	logger             *log.Entry
	hederaNode         client.HederaNode
	mirrorNode         client.MirrorNode
	contractsService   service.Contracts
	fees               service.Fees
	ethSigner          service.Signer
	transferRepository repository.Transfer
	topicID            hedera.TopicID
}

func NewService(
	hederaNode client.HederaNode,
	mirrorNode client.MirrorNode,
	contractsService service.Contracts,
	fees service.Fees,
	signer service.Signer,
	transferRepository repository.Transfer,
	topicID string,
) *Service {
	tID, e := hedera.TopicIDFromString(topicID)
	if e != nil {
		panic(fmt.Sprintf("Invalid monitoring Topic ID [%s] - Error: [%s]", topicID, e))
	}

	return &Service{
		logger:             config.GetLoggerFor(fmt.Sprintf("Transfers Service")),
		hederaNode:         hederaNode,
		mirrorNode:         mirrorNode,
		contractsService:   contractsService,
		fees:               fees,
		ethSigner:          signer,
		transferRepository: transferRepository,
		topicID:            tID,
	}
}

// SanityCheck performs validation on the memo and state proof for the transaction
func (ts *Service) SanityCheckTransfer(tx mirror_node.Transaction) (*memo.Memo, error) {
	m, e := memo.FromBase64String(tx.MemoBase64)
	if e != nil {
		return nil, errors.New(fmt.Sprintf("[%s] - Could not parse transaction memo [%s]. Error: [%s]", tx.TransactionID, tx.MemoBase64, e))
	}

	stateProof, e := ts.mirrorNode.GetStateProof(tx.TransactionID)
	if e != nil {
		return nil, errors.New(fmt.Sprintf("Could not GET state proof. Error [%s]", e))
	}

	verified, e := stateproof.Verify(tx.TransactionID, stateProof)
	if e != nil {
		return nil, errors.New(fmt.Sprintf("State proof verification failed. Error [%s]", e))
	}

	if !verified {
		return nil, errors.New("invalid state proof")
	}

	return m, nil
}

// InitiateNewTransfer Stores the incoming transfer message into the Database aware of already processed transfers
func (ts *Service) InitiateNewTransfer(tm model.Transfer) (*entity.Transfer, error) {
	dbTransaction, err := ts.transferRepository.GetByTransactionId(tm.TransactionId)
	if err != nil {
		ts.logger.Errorf("[%s] - Failed to get db record. Error [%s]", tm.TransactionId, err)
		return nil, err
	}

	if dbTransaction != nil {
		ts.logger.Infof("[%s] - Transaction already added", tm.TransactionId)
		return dbTransaction, err
	}

	ts.logger.Debugf("[%s] - Adding new Transaction Record", tm.TransactionId)
	tx, err := ts.transferRepository.Create(&tm)
	if err != nil {
		ts.logger.Errorf("[%s] - Failed to create a transaction record. Error [%s].", tm.TransactionId, err)
		return nil, err
	}
	return tx, nil
}

// SaveRecoveredTxn creates new Transaction record persisting the recovered Transfer TXn
func (ts *Service) SaveRecoveredTxn(txId, amount, nativeToken, wrappedToken string, m memo.Memo) error {
	err := ts.transferRepository.SaveRecoveredTxn(&model.Transfer{
		TransactionId:         txId,
		Receiver:              m.EthereumAddress,
		Amount:                amount,
		TxReimbursement:       m.TxReimbursementFee,
		GasPrice:              m.GasPrice,
		NativeToken:           nativeToken,
		WrappedToken:          wrappedToken,
		ExecuteEthTransaction: m.ExecuteEthTransaction,
	})
	if err != nil {
		ts.logger.Errorf("[%s] - Something went wrong while saving new Recovered Transaction. Error [%s]", txId, err)
		return err
	}

	ts.logger.Infof("Added new Transaction Record with Txn ID [%s]", txId)
	return err
}

// VerifyFee verifies that the provided TX reimbursement fee is enough using the
// Fee Calculator and updates the Transaction Record to Insufficient Fee if necessary
func (ts *Service) VerifyFee(tm model.Transfer) error {
	isSufficient, err := ts.fees.ValidateExecutionFee(tm.TxReimbursement, tm.Amount, tm.GasPrice)
	if !isSufficient {
		ts.logger.Errorf("[%s] - Fee validation failed. Provided tx reimbursement fee is invalid/insufficient. Error [%s].", tm.TransactionId, err)
		if err := ts.transferRepository.UpdateStatusInsufficientFee(tm.TransactionId); err != nil {
			ts.logger.Errorf("Failed to update status to [%s] of transaction with TransactionID [%s]. Error [%s].", transfer.StatusInsufficientFee, tm.TransactionId, err)
			return err
		}

		ts.logger.Debugf("[%s] - was updated to [%s]. Provided TxReimbursement [%s].", tm.TransactionId, transfer.StatusInsufficientFee, tm.TxReimbursement)
		return err
	}
	return nil
}

func (ts *Service) authMessageSubmissionCallbacks(txId string) (onSuccess, onRevert func()) {
	onSuccess = func() {
		ts.logger.Debugf("Authorisation Signature TX successfully executed for TX [%s]", txId)
		err := ts.transferRepository.UpdateStatusSignatureMined(txId)
		if err != nil {
			ts.logger.Errorf("[%s] - Failed to update status signature mined. Error [%s].", txId, err)
			return
		}
	}

	onRevert = func() {
		ts.logger.Debugf("Authorisation Signature TX failed for TX ID [%s]", txId)
		err := ts.transferRepository.UpdateStatusSignatureFailed(txId)
		if err != nil {
			ts.logger.Errorf("[%s] - Failed to update status signature failed. Error [%s].", txId, err)
			return
		}
	}
	return onSuccess, onRevert
}

func (ts *Service) ProcessTransfer(tm model.Transfer) error {
	gasPriceWeiBn, err := helper.ToBigInt(tm.GasPrice)
	if err != nil {
		ts.logger.Errorf("[%s] - Failed to parse Gas Price Wei to bigint [%s]. Error [%s].", tm.TransactionId, tm.GasPrice, err)
		return err
	}

	var authMsgHash []byte
	if tm.ExecuteEthTransaction {
		authMsgHash, err = auth_message.EncodeBytesForMintWithReimbursement(tm.TransactionId, tm.WrappedToken, tm.Receiver, tm.Amount, tm.TxReimbursement, gasPriceWeiBn.String())
	} else {
		authMsgHash, err = auth_message.EncodeBytesForMint(tm.TransactionId, tm.WrappedToken, tm.Receiver, tm.Amount)
	}

	if err != nil {
		ts.logger.Errorf("[%s] - Failed to encode the authorisation signature. Error: [%s]", tm.TransactionId, err)
		return err
	}

	signatureBytes, err := ts.ethSigner.Sign(authMsgHash)
	if err != nil {
		ts.logger.Errorf("[%s] - Failed to sign the authorisation signature. Error: [%s]", tm.TransactionId, err)
		return err
	}
	signature := hex.EncodeToString(signatureBytes)

	signatureMessage := message.NewSignature(
		tm.TransactionId,
		tm.Receiver,
		tm.Amount,
		tm.TxReimbursement,
		tm.GasPrice,
		signature,
		tm.WrappedToken)

	tsm := signatureMessage.GetTopicSignatureMessage()
	sigMsgBytes, err := signatureMessage.ToBytes()
	if err != nil {
		ts.logger.Errorf("[%s] - Failed to encode Signature Message to bytes. Error [%s]", tsm.TransferID, err)
		return err
	}

	messageTxId, err := ts.hederaNode.SubmitTopicConsensusMessage(
		ts.topicID,
		sigMsgBytes)
	if err != nil {
		ts.logger.Errorf("[%s] - Failed to submit Signature Message to Topic. Error: [%s]", tsm.TransferID, err)
		return err
	}

	// Update Transfer Record
	err = ts.transferRepository.UpdateStatusSignatureSubmitted(tsm.TransferID)
	if err != nil {
		ts.logger.Errorf("[%s] - Failed to update. Error [%s].", tsm.TransferID, err)
		return err
	}

	// Attach update callbacks on Signature HCS Message
	ts.logger.Infof("[%s] - Submitted signature on Topic [%s]", tsm.TransferID, ts.topicID)
	onSuccessfulAuthMessage, onFailedAuthMessage := ts.authMessageSubmissionCallbacks(tsm.TransferID)
	ts.mirrorNode.WaitForTransaction(messageTxId.String(), onSuccessfulAuthMessage, onFailedAuthMessage)
	return nil
}

// TransferData returns from the database the given transfer, its signatures and
// calculates if its messages have reached super majority
func (ts *Service) TransferData(txId string) (service.TransferData, error) {
	t, err := ts.transferRepository.GetWithMessages(txId)
	if err != nil {
		ts.logger.Errorf("[%s] - Failed to query Transfer with messages. Error: [%s].", txId, err)
		return service.TransferData{}, err
	}

	var signatures []string
	for _, m := range t.Messages {
		signatures = append(signatures, m.Signature)
	}

	requiredSigCount := len(ts.contractsService.GetMembers())/2 + 1
	reachedMajority := len(t.Messages) >= requiredSigCount

	return service.TransferData{
		Recipient:    t.Receiver,
		Amount:       t.Amount,
		Fee:          t.TxReimbursement,
		NativeToken:  t.NativeToken,
		WrappedToken: t.WrappedToken,
		Signatures:   signatures,
		Majority:     reachedMajority,
		GasPrice:     t.GasPrice,
	}, nil
}

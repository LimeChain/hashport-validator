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

package signatures

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/encoding"
	"github.com/limechain/hedera-eth-bridge-validator/app/encoding/auth-message"
	ethhelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/ethereum"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"strings"

	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	validatorproto "github.com/limechain/hedera-eth-bridge-validator/proto"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	ethSigner             service.Signer
	contractsService      service.Contracts
	transactionRepository repository.Transaction
	messageRepository     repository.Message
	topicID               hedera.TopicID
	logger                *log.Entry
}

func NewService(
	ethSigner service.Signer,
	contractsService service.Contracts,
	transactionRepository repository.Transaction,
	messageRepository repository.Message,
	topicID string,
) *Service {
	tID, e := hedera.TopicIDFromString(topicID)
	if e != nil {
		panic(fmt.Sprintf("Invalid monitoring Topic ID [%s] - Error: [%s]", topicID, e))
	}

	return &Service{
		ethSigner:             ethSigner,
		contractsService:      contractsService,
		messageRepository:     messageRepository,
		transactionRepository: transactionRepository,
		logger:                config.GetLoggerFor(fmt.Sprintf("Transfers Service")),
		topicID:               tID,
	}
}

// SanityCheckSignature performs validation on the topic message metadata.
// Validates it against the Transaction Record metadata from DB
func (bs *Service) SanityCheckSignature(tm encoding.TopicMessage) (bool, error) {
	topicMessage := tm.GetTopicSignatureMessage()
	t, err := bs.transactionRepository.GetByTransactionId(topicMessage.TransactionId)
	if err != nil {
		bs.logger.Errorf("Failed to retrieve Transaction Record for TX ID [%s]. Error: %s", topicMessage.TransactionId, err)
		return false, err
	}
	match := t.EthAddress == topicMessage.EthAddress &&
		t.Amount == topicMessage.Amount &&
		t.Fee == topicMessage.Fee
	return match, nil
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
	exists, err := bs.messageRepository.Exist(topicMessage.TransactionId, signatureHex, authMessageStr)
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
	err = bs.messageRepository.Create(&message.TransactionMessage{
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

// ScheduleForSubmission computes the execution slot and schedules the Eth TX for submission
func (bs *Service) ScheduleForSubmission(txId string) error {
	//signatureMessages, err := bs.messageRepository.GetMessagesFor(txId)
	//if err != nil {
	//	bs.logger.Errorf("Failed to query all Signature Messages for TX [%s]. Error: %s", txId, err)
	//	return err
	//}
	//
	//slot, isFound := bs.computeExecutionSlot(signatureMessages)
	//if !isFound {
	//	bs.logger.Debugf("TX [%s] - Operator [%s] has not been found as signer amongst the signatures collected.", txId, bs.services.EthSigner.Address())
	//	return nil
	//}
	//
	//amount := signatureMessages[0].Amount
	//fee := signatureMessages[0].Fee
	//ethAddress := signatureMessages[0].EthAddress
	//signatures, err := getSignatures(signatureMessages)
	//if err != nil {
	//	return err
	//}
	//
	//// TODO
	//submitEthTx := func() error {
	//
	//	tx, err = bs.services.Contracts.SubmitSignatures(bs.services.EthSigner.NewKeyTransactor(), submission.TransferMessage, signatures)
	//	if err != nil {
	//		bs.logger.Errorf("Failed to execute Scheduled TX for [%s]. Error [%s].", id, err)
	//		return err
	//	}
	//	ethTxHashString := ethTx.Hash().String()
	//
	//	s.logger.Infof("Executed Scheduled TX [%s], Eth TX Hash [%s].", id, ethTxHashString)
	//	tx, err := s.submitEthTxTopicMessage(id, submission, ethTxHashString)
	//	if err != nil {
	//		s.logger.Errorf("Failed to submit topic consensus eth tx message for TX [%s], TX Hash [%s]. Error [%s].", id, ethTxHashString, err)
	//		return
	//	}
	//	s.logger.Infof("Submitted Eth TX Hash [%s] for TX [%s] at HCS Transaction ID [%s]", ethTxHashString, id, tx.String())
	//
	//	success, err := s.waitForEthTxMined(ethTx.Hash())
	//	if err != nil {
	//		s.logger.Errorf("Waiting for execution for TX [%s] and Hash [%s] failed. Error [%s].", id, ethTxHashString, err)
	//		return
	//	}
	//
	//	if success {
	//		s.logger.Infof("Successful execution of TX [%s] with TX Hash [%s].", id, ethTxHashString)
	//	} else {
	//		s.logger.Warnf("Execution for TX [%s] with TX Hash [%s] was not successful.", id, ethTxHashString)
	//	}
	//}
	//
	//err = bs.services.Scheduler.Schedule(txId, signatureMessages[0].TransactionTimestamp, slot, submitEthTx)
	//if err != nil {
	//	return err
	//}
	return nil
}

func getSignatures(messages []message.TransactionMessage) ([][]byte, error) {
	var signatures [][]byte

	for _, msg := range messages {
		signature, err := hex.DecodeString(msg.Signature)
		if err != nil {
			return nil, err
		}
		signatures = append(signatures, signature)
	}

	return signatures, nil
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
		return common.Address{}, errors.New(fmt.Sprintf("signer is not signatures member"))
	}
	return address, nil
}

// computeExecutionSlot - computes the slot order in which the TX will execute
// Important! Transaction messages ARE expected to be sorted by ascending Timestamp
func (bs *Service) computeExecutionSlot(messages []message.TransactionMessage) (slot int64, isFound bool) {
	for i := 0; i < len(messages); i++ {
		if strings.ToLower(messages[i].SignerAddress) == strings.ToLower(bs.ethSigner.Address()) {
			return int64(i), true
		}
	}

	return -1, false
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

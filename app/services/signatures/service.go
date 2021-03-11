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
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/encoding"
	"github.com/limechain/hedera-eth-bridge-validator/app/encoding/auth-message"
	ethhelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/ethereum"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"strings"

	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	ethSigner             service.Signer
	contractsService      service.Contracts
	scheduler             service.Scheduler
	transactionRepository repository.Transaction
	messageRepository     repository.Message
	topicID               hedera.TopicID
	hederaClient          client.HederaNode
	mirrorClient          client.MirrorNode
	ethClient             client.Ethereum
	logger                *log.Entry
}

func NewService(
	ethSigner service.Signer,
	contractsService service.Contracts,
	scheduler service.Scheduler,
	transactionRepository repository.Transaction,
	messageRepository repository.Message,
	hederaClient client.HederaNode,
	mirrorClient client.MirrorNode,
	ethClient client.Ethereum,
	topicID string,
) *Service {
	tID, e := hedera.TopicIDFromString(topicID)
	if e != nil {
		panic(fmt.Sprintf("Invalid monitoring Topic ID [%s] - Error: [%s]", topicID, e))
	}

	return &Service{
		ethSigner:             ethSigner,
		contractsService:      contractsService,
		scheduler:             scheduler,
		messageRepository:     messageRepository,
		transactionRepository: transactionRepository,
		logger:                config.GetLoggerFor(fmt.Sprintf("Signatures Service")),
		topicID:               tID,
		hederaClient:          hederaClient,
		mirrorClient:          mirrorClient,
		ethClient:             ethClient,
	}
}

// SanityCheckSignature performs validation on the topic message metadata.
// Validates it against the Transaction Record metadata from DB
func (ss *Service) SanityCheckSignature(tm encoding.TopicMessage) (bool, error) {
	topicMessage := tm.GetTopicSignatureMessage()
	t, err := ss.transactionRepository.GetByTransactionId(topicMessage.TransactionId)
	if err != nil {
		ss.logger.Errorf("Failed to retrieve Transaction Record for TX ID [%s]. Error: %s", topicMessage.TransactionId, err)
		return false, err
	}
	match := t.EthAddress == topicMessage.EthAddress &&
		t.Amount == topicMessage.Amount &&
		t.Fee == topicMessage.Fee
	return match, nil
}

// ProcessSignature processes the signature message, verifying and updating all necessary fields in the DB
func (ss *Service) ProcessSignature(tm encoding.TopicMessage) error {
	// Parse incoming message
	topicMessage := tm.GetTopicSignatureMessage()
	authMsgBytes, err := auth_message.FromTopicMessage(topicMessage)
	if err != nil {
		ss.logger.Errorf("Failed to encode the authorisation signature for TX ID [%s]. Error: %s", topicMessage.TransactionId, err)
	}

	// Prepare Signature
	signatureBytes, signatureHex, err := ethhelper.DecodeSignature(topicMessage.GetSignature())
	if err != nil {
		ss.logger.Errorf("[%s] - Decoding Signature [%s] for TX failed. Err: %s", topicMessage.TransactionId, topicMessage.GetSignature(), err)
		return err
	}
	authMessageStr := hex.EncodeToString(authMsgBytes)

	// Check for duplicated signature
	exists, err := ss.messageRepository.Exist(topicMessage.TransactionId, signatureHex, authMessageStr)
	if err != nil {
		ss.logger.Errorf("An error occurred while getting TX [%s] from DB. Error: %s", topicMessage.TransactionId, err)
	}
	if exists {
		ss.logger.Errorf("Signature already received for TX [%s]", topicMessage.TransactionId)
		return err
	}

	// Verify Signature
	address, err := ss.verifySignature(err, authMsgBytes, signatureBytes, topicMessage.TransactionId, authMessageStr)
	if err != nil {
		return err
	}

	// Persist in DB
	err = ss.messageRepository.Create(&message.TransactionMessage{
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
		ss.logger.Errorf("[%s] - Failed to save Transaction Message in DB with Signature [%s]. Error: %s", topicMessage.TransactionId, signatureHex, err)
		return err
	}

	ss.logger.Infof("[%s] - Successfully processed Signature Message [%s]", topicMessage.TransactionId, signatureHex)
	return nil
}

// ScheduleForSubmission computes the execution slot and schedules the Eth TX for submission
func (ss *Service) ScheduleForSubmission(txId string) error {
	signatureMessages, err := ss.messageRepository.GetMessagesFor(txId)
	if err != nil {
		ss.logger.Errorf("Failed to query all Signature Messages for TX [%s]. Error: %s", txId, err)
		return err
	}

	slot, isFound := ss.computeExecutionSlot(signatureMessages)
	if !isFound {
		ss.logger.Debugf("TX [%s] - Operator [%s] has not been found as signer amongst the signatures collected.", txId, ss.ethSigner.Address())
		return nil
	}

	amount := signatureMessages[0].Amount
	fee := signatureMessages[0].Fee
	ethAddress := signatureMessages[0].EthAddress
	messageHash := signatureMessages[0].Hash
	signatures, err := getSignatures(signatureMessages)
	if err != nil {
		return err
	}

	ethereumMintTask := ss.prepareEthereumMintTask(txId, ethAddress, amount, fee, signatures, messageHash)
	err = ss.scheduler.Schedule(txId, signatureMessages[0].TransactionTimestamp, slot, ethereumMintTask)
	if err != nil {
		return err
	}
	return nil
}

// prepareEthereumMintTask returns the function to be executed for processing the
// Ethereum Mint transaction and HCS topic message with the ethereum TX hash after that
func (ss *Service) prepareEthereumMintTask(txId string, ethAddress string, amount string, fee string, signatures [][]byte, messageHash string) func() {
	ethereumMintTask := func() {
		// Submit and monitor Ethereum TX
		ethTx, err := ss.contractsService.SubmitSignatures(ss.ethSigner.NewKeyTransactor(), txId, ethAddress, amount, fee, signatures)
		if err != nil {
			ss.logger.Errorf("Failed to Submit Signatures for TX [%s]. Error: %s", txId, err)
			return
		}
		err = ss.transactionRepository.UpdateStatusEthTxSubmitted(txId, ethTx.Hash().String())
		if err != nil {
			ss.logger.Errorf("Failed to update status for TX [%s]", txId)
			return
		}
		ss.logger.Infof("Submitted Ethereum Mint TX [%s] for TX [%s]", ethTx.Hash().String(), txId)

		onEthTxSuccess, onEthTxRevert := ss.ethTxCallbacks(txId, ethTx.Hash().String())
		ss.ethClient.WaitForTransaction(ethTx.Hash(), onEthTxSuccess, onEthTxRevert)

		// Submit and monitor HCS Message for Ethereum TX Hash
		hcsTx, err := ss.submitEthTxTopicMessage(txId, messageHash, ethTx.Hash().String())
		if err != nil {
			ss.logger.Errorf("Failed to submit Ethereum TX Hash to Bridge Topic for TX [%s]. Error %s", txId, err)
			return
		}
		ss.logger.Infof("Submitted Ethereum TX Hash [%s] for TX [%s] to HCS. Transaction ID [%s]", ethTx.Hash().String(), txId, hcsTx.String())
		onHcsMessageSuccess, onHcsMessageFail := ss.hcsTxCallbacks(txId)
		ss.mirrorClient.WaitForTransaction(hcsTx.String(), onHcsMessageSuccess, onHcsMessageFail)

		ss.logger.Infof("Successfully processed Ethereum Minting for TX [%s]", txId)
	}
	return ethereumMintTask
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

func (ss *Service) verifySignature(err error, authMsgBytes []byte, signatureBytes []byte, txId, authMessageStr string) (common.Address, error) {
	publicKey, err := crypto.Ecrecover(authMsgBytes, signatureBytes)
	if err != nil {
		ss.logger.Errorf("[%s] - Failed to recover public key. Hash [%s]. Error: %s", txId, authMessageStr, err)
		return common.Address{}, err
	}
	unmarshalledPublicKey, err := crypto.UnmarshalPubkey(publicKey)
	if err != nil {
		ss.logger.Errorf("[%s] - Failed to unmarshall public key. Error: %s", txId, err)
		return common.Address{}, err
	}
	address := crypto.PubkeyToAddress(*unmarshalledPublicKey)
	if !ss.contractsService.IsMember(address.String()) {
		ss.logger.Errorf("[%s] - Received Signature [%s] is not signed by Bridge member", txId, authMessageStr)
		return common.Address{}, errors.New(fmt.Sprintf("signer is not signatures member"))
	}
	return address, nil
}

// computeExecutionSlot - computes the slot order in which the TX will execute
// Important! Transaction messages ARE expected to be sorted by ascending Timestamp
func (ss *Service) computeExecutionSlot(messages []message.TransactionMessage) (slot int64, isFound bool) {
	for i := 0; i < len(messages); i++ {
		if strings.ToLower(messages[i].SignerAddress) == strings.ToLower(ss.ethSigner.Address()) {
			return int64(i), true
		}
	}

	return -1, false
}

func (ss *Service) submitEthTxTopicMessage(txId, messageHash, ethereumTxHash string) (*hedera.TransactionID, error) {
	ethTxHashMessage := encoding.NewEthereumHashMessage(txId, messageHash, ethereumTxHash)
	ethTxHashBytes, err := ethTxHashMessage.ToBytes()
	if err != nil {
		ss.logger.Errorf("Failed to encode Eth TX Hash Message to bytes for TX [%s]. Error: %s", txId, err)
		return nil, err
	}

	return ss.hederaClient.SubmitTopicConsensusMessage(ss.topicID, ethTxHashBytes)
}

func (ss *Service) ethTxCallbacks(txId, hash string) (func(), func()) {
	onSuccess := func() {
		ss.logger.Infof("Ethereum TX [%s] for TX [%s] was successfully mined", hash, txId)
		err := ss.transactionRepository.UpdateStatusEthTxMined(txId)
		if err != nil {
			ss.logger.Errorf("Failed to update status for TX [%s]. Error [%s].", txId, err)
			return
		}
	}

	onRevert := func() {
		ss.logger.Infof("Ethereum TX [%s] for TX [%s] reverted", hash, txId)
		err := ss.transactionRepository.UpdateStatusSignatureFailed(txId)
		if err != nil {
			ss.logger.Errorf("Failed to update status for TX [%s]. Error [%s].", txId, err)
			return
		}
	}
	return onSuccess, onRevert
}

func (ss *Service) hcsTxCallbacks(txId string) (func(), func()) {
	onSuccess := func() {
		ss.logger.Infof("Ethereum TX Hash message was successfully mined for TX [%s]", txId)
		err := ss.transactionRepository.UpdateStatusEthTxMsgMined(txId)
		if err != nil {
			ss.logger.Errorf("Failed to update status for TX [%s]. Error [%s].", txId, err)
			return
		}
	}

	onFailure := func() {
		ss.logger.Infof("Ethereum TX Hash message failed for TX ID [%s]", txId)
		err := ss.transactionRepository.UpdateStatusEthTxMsgFailed(txId)
		if err != nil {
			ss.logger.Errorf("Failed to update status for TX [%s]. Error [%s].", txId, err)
			return
		}
	}
	return onSuccess, onFailure
}

// TODO ->

//func (bs *Service) AcknowledgeTransactionSuccess(m *validatorproto.TopicEthTransactionMessage) {
//	bs.logger.Infof("Waiting for Transaction with ID [%s] to be mined.", m.TransactionId)
//
//	isSuccessful, err := bs.clients.Ethereum.WaitForTransaction(common.HexToHash(m.EthTxHash))
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

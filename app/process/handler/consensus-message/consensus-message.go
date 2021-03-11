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

package consensusmessage

import (
	"errors"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/encoding"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	validatorproto "github.com/limechain/hedera-eth-bridge-validator/proto"
	"github.com/limechain/hedera-watcher-sdk/queue"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	messageRepository repository.Message
	contractsService  service.Contracts
	signaturesService service.Signatures
	logger            *log.Entry
}

func NewHandler(
	configuration config.ConsensusMessageHandler,
	messageRepository repository.Message,
	contractsService service.Contracts,
	signaturesService service.Signatures,
) *Handler {
	topicID, err := hedera.TopicIDFromString(configuration.TopicId)
	if err != nil {
		log.Fatalf("Invalid topic id: [%v]", configuration.TopicId)
	}

	return &Handler{
		messageRepository: messageRepository,
		contractsService:  contractsService,
		signaturesService: signaturesService,
		logger:            config.GetLoggerFor(fmt.Sprintf("Topic [%s] Handler", topicID.String())),
	}
}

func (cmh Handler) Recover(queue *queue.Queue) {
}

func (cmh Handler) Handle(payload []byte) {
	m, err := encoding.NewTopicMessageFromBytes(payload)
	if err != nil {
		log.Errorf("Error could not unmarshal payload. Error [%s].", err)
		return
	}

	switch m.Type {
	case validatorproto.TopicMessageType_EthSignature:
		cmh.handleSignatureMessage(*m)
	case validatorproto.TopicMessageType_EthTransaction:
		//err = cmh.handleEthTxMessage(m.GetTopicEthTransactionMessage())
	default:
		err = errors.New(fmt.Sprintf("Error - invalid topic submission message type [%s]", m.Type))
	}

	if err != nil {
		cmh.logger.Errorf("Error - could not handle payload: [%s]", err)
		return
	}
}

//func (cmh Handler) handleEthTxMessage(m *validatorproto.TopicEthTransactionMessage) error {
//	isValid, err := cmh.verifyEthTxAuthenticity(m)
//	if err != nil {
//		cmh.logger.Errorf("[%s] - ETH TX [%s] - Error while trying to verify TX authenticity.", m.TransactionId, m.EthTxHash)
//		return err
//	}
//
//	if !isValid {
//		cmh.logger.Infof("[%s] - Eth TX [%s] - Invalid authenticity.", m.TransactionId, m.EthTxHash)
//		return nil
//	}
//
//	err = cmh.transactionRepository.UpdateStatusEthTxSubmitted(m.TransactionId, m.EthTxHash)
//	if err != nil {
//		cmh.logger.Errorf("Failed to update status to [%s] of transaction with TransactionID [%s]. Error [%s].", transaction.StatusEthTxSubmitted, m.TransactionId, err)
//		return err
//	}
//
//	go cmh.transfersService.AcknowledgeTransactionSuccess(m)
//
//	return cmh.scheduler.Cancel(m.TransactionId)
//}

//func (cmh Handler) verifyEthTxAuthenticity(m *validatorproto.TopicEthTransactionMessage) (bool, error) {
//	tx, _, err := cmh.ethereumClient.GetClient().TransactionByHash(context.Background(), common.HexToHash(m.EthTxHash))
//	if err != nil {
//		cmh.logger.Warnf("[%s] - Failed to get eth transaction by hash [%s]. Error [%s].", m.TransactionId, m.EthTxHash, err)
//		return false, err
//	}
//
//	if strings.ToLower(tx.To().String()) != strings.ToLower(cmh.contractsService.GetBridgeContractAddress().String()) {
//		cmh.logger.Debugf("[%s] - ETH TX [%s] - Failed authenticity - Different To Address [%s].", m.TransactionId, m.EthTxHash, tx.To().String())
//		return false, nil
//	}
//
//	txMessage, signatures, err := ethhelper.DecodeBridgeMintFunction(tx.Data())
//	if err != nil {
//		return false, err
//	}
//
//	if txMessage.TransactionId != m.TransactionId {
//		cmh.logger.Debugf("[%s] - ETH TX [%s] - Different txn id [%s].", m.TransactionId, m.EthTxHash, txMessage.TransactionId)
//		return false, nil
//	}
//
//	dbTx, err := cmh.transactionRepository.GetByTransactionId(m.TransactionId)
//	if err != nil {
//		return false, err
//	}
//	if dbTx == nil {
//		cmh.logger.Debugf("[%s] - ETH TX [%s] - Transaction not found in database.", m.TransactionId, m.EthTxHash)
//		return false, nil
//	}
//
//	if dbTx.Amount != txMessage.Amount ||
//		dbTx.EthAddress != txMessage.EthAddress ||
//		dbTx.Fee != txMessage.Fee {
//		cmh.logger.Debugf("[%s] - ETH TX [%s] - Invalid arguments.", m.TransactionId, m.EthTxHash)
//		return false, nil
//	}
//
//	encodedData, err := ethhelper.EncodeData(txMessage)
//	if err != nil {
//		return false, err
//	}
//	hash := ethhelper.KeccakData(encodedData)
//
//	checkedAddresses := make(map[string]bool)
//	for _, signature := range signatures {
//		address, err := ethhelper.GetAddressBySignature(hash, signature)
//		if err != nil {
//			return false, err
//		}
//		if checkedAddresses[address] {
//			return false, err
//		}
//
//		if !cmh.contractsService.IsMember(address) {
//			cmh.logger.Debugf("[%s] - ETH TX [%s] - Invalid operator process - [%s].", m.TransactionId, m.EthTxHash, address)
//			return false, nil
//		}
//		checkedAddresses[address] = true
//	}
//
//	return true, nil
//}
//
//func (cmh Handler) acknowledgeTransactionSuccess(m *validatorproto.TopicEthTransactionMessage) {
//	cmh.logger.Infof("Waiting for Transaction with ID [%s] to be mined.", m.TransactionId)
//
//	isSuccessful, err := cmh.ethereumClient.WaitForTransactionSuccess(common.HexToHash(m.EthTxHash))
//	if err != nil {
//		cmh.logger.Errorf("Failed to await TX ID [%s] with ETH TX [%s] to be mined. Error [%s].", m.TransactionId, m.Hash, err)
//		return
//	}
//
//	if !isSuccessful {
//		cmh.logger.Infof("Transaction with ID [%s] was reverted. Updating status to [%s].", m.TransactionId, transaction.StatusEthTxReverted)
//		err = cmh.transactionRepository.UpdateStatusEthTxReverted(m.TransactionId)
//		if err != nil {
//			cmh.logger.Errorf("Failed to update status to [%s] of transaction with TransactionID [%s]. Error [%s].", transaction.StatusEthTxReverted, m.TransactionId, err)
//			return
//		}
//	} else {
//		cmh.logger.Infof("Transaction with ID [%s] was successfully mined. Updating status to [%s].", m.TransactionId, transaction.StatusCompleted)
//		err = cmh.transactionRepository.UpdateStatusCompleted(m.TransactionId)
//		if err != nil {
//			cmh.logger.Errorf("Failed to update status to [%s] of transaction with TransactionID [%s]. Error [%s].", transaction.StatusCompleted, m.TransactionId, err)
//			return
//		}
//	}
//}

// handleSignatureMessage is the main component responsible for the processing of new incoming Signature Messages
func (cmh Handler) handleSignatureMessage(tm encoding.TopicMessage) {
	tsm := tm.GetTopicSignatureMessage()
	valid, err := cmh.signaturesService.SanityCheckSignature(tm)
	if err != nil {
		cmh.logger.Errorf("Failed to perform sanity check on incoming signature [%s] for TX [%s]", tsm.GetSignature(), tsm.TransactionId)
		return
	}
	if !valid {
		cmh.logger.Errorf("Incoming signature for TX [%s] is invalid", tsm.GetTransactionId())
		return
	}

	err = cmh.signaturesService.ProcessSignature(tm)
	if err != nil {
		cmh.logger.Errorf("Could not process Signature [%s] for TX [%s]", tsm.GetSignature(), tsm.TransactionId)
		return
	}

	majorityReached, err := cmh.hasReachedMajority(tsm.TransactionId)
	if err != nil {
		cmh.logger.Errorf("Could not determine whether majority was reached for TX [%s]", tsm.TransactionId)
		return
	}

	if majorityReached {
		cmh.logger.Debugf("TX [%s] - Enough signatures have been collected.", tsm.TransactionId)
		err = cmh.signaturesService.ScheduleForSubmission(tsm.TransactionId)
		if err != nil {
			cmh.logger.Errorf("Could not schedule TX [%s] for submission", tsm.TransactionId)
		}
	}
}

func (cmh *Handler) hasReachedMajority(txId string) (bool, error) {
	signatureMessages, err := cmh.messageRepository.GetMessagesFor(txId)
	if err != nil {
		cmh.logger.Errorf("Failed to query all Signature Messages for TX [%s]. Error: %s", txId, err)
		return false, err
	}
	requiredSigCount := len(cmh.contractsService.GetMembers())/2 + 1
	cmh.logger.Infof("Collected [%d/%d] Signatures for TX ID [%s] ", len(signatureMessages), len(cmh.contractsService.GetMembers()), txId)
	return len(signatureMessages) >= requiredSigCount, nil
}

//
//func (s *Scheduler) execute(submission ethsubmission.Submission) (*types.Transaction, error) {
//	signatures, err := getSignatures(submission.Messages)
//	if err != nil {
//		return nil, err
//	}
//	return s.contractsService.SubmitSignatures(submission.TransactOps, submission.TransferMessage, signatures)
//}
//
//func (s *Scheduler) submitEthTxTopicMessage(id string, submission ethsubmission.Submission, ethTxHash string) (*hedera.TransactionID, error) {
//	ethTxMsg := &protomsg.TopicEthTransactionMessage{
//		TransactionId: id,
//		Hash:          submission.Messages[0].Hash,
//		EthTxHash:     ethTxHash,
//	}
//
//	msg := &protomsg.TopicMessage{
//		Type: protomsg.TopicMessageType_EthTransaction,
//		Message: &protomsg.TopicMessage_TopicEthTransactionMessage{
//			TopicEthTransactionMessage: ethTxMsg}}
//
//	msgBytes, err := proto.Marshal(msg)
//	if err != nil {
//		s.logger.Errorf("Failed to marshal protobuf TX [%s], TX Hash [%s]. Error [%s].", id, ethTxHash, err)
//	}
//
//	// TODO refactor such that the "waitForEthTxMined" is performed inside the signatures service and scheduler only reacts to that
//	return s.hederaClient.SubmitTopicConsensusMessage(s.topicID, msgBytes)
//}
//
//// TODO
//func (s *Scheduler) waitForEthTxMined(ethTx common.Hash) (bool, error) {
//	//return s.contractsService.Client.WaitForTransactionSuccess(ethTx)
//	return true, nil
//}
//

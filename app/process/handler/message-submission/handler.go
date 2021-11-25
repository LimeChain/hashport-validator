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

package message_submission

import (
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	hederahelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/hedera"
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

// Handler is transfers event handler
type Handler struct {
	hederaNode         client.HederaNode
	mirrorNode         client.MirrorNode
	transfersService   service.Transfers
	transferRepository repository.Transfer
	topicID            hedera.TopicID
	messageService     service.Messages
	logger             *log.Entry
}

func NewHandler(
	hederaNode client.HederaNode,
	mirrorNode client.MirrorNode,
	transfersService service.Transfers,
	transferRepository repository.Transfer,
	messageService service.Messages,
	topicId string,
) *Handler {
	topicID, err := hedera.TopicIDFromString(topicId)
	if err != nil {
		log.Fatalf("Invalid topic id: [%v]", topicId)
	}

	return &Handler{
		hederaNode:         hederaNode,
		mirrorNode:         mirrorNode,
		logger:             config.GetLoggerFor("Topic Message Submission Handler"),
		transfersService:   transfersService,
		transferRepository: transferRepository,
		messageService:     messageService,
		topicID:            topicID,
	}
}

func (smh Handler) Handle(payload interface{}) {
	transferMsg, ok := payload.(*model.Transfer)
	if !ok {
		smh.logger.Errorf("Could not cast payload [%s]", payload)
		return
	}

	transactionRecord, err := smh.transfersService.InitiateNewTransfer(*transferMsg)
	if err != nil {
		smh.logger.Errorf("[%s] - Error occurred while initiating processing. Error: [%s]", transferMsg.TransactionId, err)
		return
	}

	if transactionRecord.Status != status.Initial {
		smh.logger.Debugf("[%s] - Previously added with status [%s]. Skipping further execution.", transactionRecord.TransactionID, transactionRecord.Status)
		return
	}

	err = smh.submitMessage(transferMsg)
	if err != nil {
		smh.logger.Errorf("[%s] - Processing failed. Error: [%s]", transferMsg.TransactionId, err)
		return
	}
}

func (smh Handler) submitMessage(tm *model.Transfer) error {
	signatureMessageBytes, err := smh.messageService.SignFungibleMessage(*tm)
	if err != nil {
		return err
	}

	messageTxId, err := smh.hederaNode.SubmitTopicConsensusMessage(
		smh.topicID,
		signatureMessageBytes)
	if err != nil {
		smh.logger.Errorf("[%s] - Failed to submit Signature Message to Topic. Error: [%s]", tm.TransactionId, err)
		return err
	}

	// Attach update callbacks on Signature HCS Message
	smh.logger.Infof("[%s] - Submitted signature on Topic [%s]", tm.TransactionId, smh.topicID)
	onSuccessfulAuthMessage, onFailedAuthMessage := smh.authMessageSubmissionCallbacks(tm.TransactionId)
	smh.mirrorNode.WaitForTransaction(hederahelper.ToMirrorNodeTransactionID(messageTxId.String()), onSuccessfulAuthMessage, onFailedAuthMessage)
	return nil
}

func (smh Handler) authMessageSubmissionCallbacks(txId string) (onSuccess, onRevert func()) {
	onSuccess = func() {
		smh.logger.Debugf("Authorisation Signature TX successfully executed for TX [%s]", txId)
	}

	onRevert = func() {
		smh.logger.Debugf("Authorisation Signature TX failed for TX ID [%s]", txId)
	}
	return onSuccess, onRevert
}

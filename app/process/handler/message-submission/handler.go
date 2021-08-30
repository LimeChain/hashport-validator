package message_submission

import (
	"encoding/hex"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	hederahelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/hedera"
	auth_message "github.com/limechain/hedera-eth-bridge-validator/app/model/auth-message"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/message"
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

// Handler is transfers event handler
type Handler struct {
	hederaNode         client.HederaNode
	mirrorNode         client.MirrorNode
	ethSigners         map[int64]service.Signer
	transfersService   service.Transfers
	transferRepository repository.Transfer
	topicID            hedera.TopicID
	logger             *log.Entry
}

func NewHandler(
	hederaNode client.HederaNode,
	mirrorNode client.MirrorNode,
	ethSigners map[int64]service.Signer,
	transfersService service.Transfers,
	transferRepository repository.Transfer,
	topicId string,
) *Handler {
	topicID, err := hedera.TopicIDFromString(topicId)
	if err != nil {
		log.Fatalf("Invalid topic id: [%v]", topicId)
	}

	return &Handler{
		hederaNode:         hederaNode,
		mirrorNode:         mirrorNode,
		ethSigners:         ethSigners,
		logger:             config.GetLoggerFor("Topic Message Submission Handler"),
		transfersService:   transfersService,
		transferRepository: transferRepository,
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

	if transactionRecord.Status != transfer.StatusInitial {
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
	authMsgHash, err := auth_message.EncodeBytesFrom(tm.SourceChainId, tm.TargetChainId, tm.TransactionId, tm.TargetAsset, tm.Receiver, tm.Amount)
	if err != nil {
		smh.logger.Errorf("[%s] - Failed to encode the authorisation signature. Error: [%s]", tm.TransactionId, err)
		return err
	}

	signatureBytes, err := smh.ethSigners[tm.TargetChainId].Sign(authMsgHash)
	if err != nil {
		smh.logger.Errorf("[%s] - Failed to sign the authorisation signature. Error: [%s]", tm.TransactionId, err)
		return err
	}
	signature := hex.EncodeToString(signatureBytes)

	signatureMessage := message.NewSignature(
		uint64(tm.SourceChainId),
		uint64(tm.TargetChainId),
		tm.TransactionId,
		tm.TargetAsset,
		tm.Receiver,
		tm.Amount,
		signature)

	sigMsgBytes, err := signatureMessage.ToBytes()
	if err != nil {
		smh.logger.Errorf("[%s] - Failed to encode Signature Message to bytes. Error [%s]", signatureMessage.TransferID, err)
		return err
	}

	messageTxId, err := smh.hederaNode.SubmitTopicConsensusMessage(
		smh.topicID,
		sigMsgBytes)
	if err != nil {
		smh.logger.Errorf("[%s] - Failed to submit Signature Message to Topic. Error: [%s]", signatureMessage.TransferID, err)
		return err
	}

	// Update Transfer Record
	err = smh.transferRepository.UpdateStatusSignatureSubmitted(signatureMessage.TransferID)
	if err != nil {
		smh.logger.Errorf("[%s] - Failed to update. Error [%s].", signatureMessage.TransferID, err)
		return err
	}

	// Attach update callbacks on Signature HCS Message
	smh.logger.Infof("[%s] - Submitted signature on Topic [%s]", signatureMessage.TransferID, smh.topicID)
	onSuccessfulAuthMessage, onFailedAuthMessage := smh.authMessageSubmissionCallbacks(signatureMessage.TransferID)
	smh.mirrorNode.WaitForTransaction(hederahelper.ToMirrorNodeTransactionID(messageTxId.String()), onSuccessfulAuthMessage, onFailedAuthMessage)
	return nil
}

func (smh Handler) authMessageSubmissionCallbacks(txId string) (onSuccess, onRevert func()) {
	onSuccess = func() {
		smh.logger.Debugf("Authorisation Signature TX successfully executed for TX [%s]", txId)
		err := smh.transferRepository.UpdateStatusSignatureMined(txId)
		if err != nil {
			smh.logger.Errorf("[%s] - Failed to update status signature mined. Error [%s].", txId, err)
			return
		}
	}

	onRevert = func() {
		smh.logger.Debugf("Authorisation Signature TX failed for TX ID [%s]", txId)
		err := smh.transferRepository.UpdateStatusSignatureFailed(txId)
		if err != nil {
			smh.logger.Errorf("[%s] - Failed to update status signature failed. Error [%s].", txId, err)
			return
		}
	}
	return onSuccess, onRevert
}

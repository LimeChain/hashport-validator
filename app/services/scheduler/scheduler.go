package scheduler

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/golang/protobuf/proto"
	"github.com/hashgraph/hedera-sdk-go"
	hederaClient "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/model/ethsubmission"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/model/scheduler"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/ethereum/bridge"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	protomsg "github.com/limechain/hedera-eth-bridge-validator/proto"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Scheduler struct {
	topicID         hedera.TopicID
	logger          *log.Entry
	tasks           *sync.Map
	operator        string
	executionWindow int64
	contractService *bridge.BridgeContractService
	hederaClient    *hederaClient.HederaNodeClient
}

// Schedule - Schedules new Transaction for execution at the right leader elected slot
func (s *Scheduler) Schedule(id string, submission ethsubmission.Submission) error {
	// Important! Transaction messages ARE expected to be sorted by ascending Timestamp
	et := s.computeExecutionTime(submission.Messages[0].TransactionTimestamp, submission.Slot)

	executeIn := time.Until(et)
	timer := time.NewTimer(executeIn)

	storedValue, alreadyExisted := s.tasks.LoadOrStore(id, &scheduler.Storage{
		Executed: false,
		Timer:    timer,
	})

	if alreadyExisted {
		s.logger.Infof("TX with ID [%s] already scheduled for execution/executed.", id)
		return nil
	}

	go func() {
		<-timer.C
		storedValue.(*scheduler.Storage).Executed = true

		ethTx, err := s.execute(submission)
		if err != nil {
			s.logger.Errorf("Failed to execute Scheduled TX for [%s]. Error [%s].", submission.CryptoTransferMessage.TransactionId, err)
			return
		}
		ethTxHashString := ethTx.Hash().String()

		s.logger.Infof("Executed Scheduled TX [%s], Eth TX Hash [%s].", id, ethTxHashString)
		tx, err := s.submitEthTxTopicMessage(id, submission, ethTxHashString)
		if err != nil {
			s.logger.Errorf("Failed to submit topic consensus eth tx message for TX [%s], TX Hash [%s]. Error [%s].", id, ethTxHashString, err)
			return
		}
		s.logger.Infof("Submitted Eth TX Hash [%s] for TX [%s] at HCS Transaction ID [%s]", ethTxHashString, id, tx.String())

		success, err := s.waitForEthTxMined(ethTx.Hash())
		if err != nil {
			s.logger.Errorf("Waiting for execution for TX [%s] and Hash [%s] failed. Error [%s].", id, ethTxHashString, err)
			return
		}

		if success {
			s.logger.Infof("Successful execution of TX [%s] with TX Hash [%s].", id, ethTxHashString)
		} else {
			s.logger.Warn("Execution for TX [%s] with TX Hash [%s] was not successful.", id, ethTxHashString)
		}
	}()

	s.logger.Infof("Scheduled new TX with ID [%s] for execution in [%s]", id, executeIn)

	return nil
}

// Cancel - Removes and cancels an already scheduled Transaction
func (s *Scheduler) Cancel(id string) error {
	t, exists := s.tasks.Load(id)
	if !exists {
		s.logger.Warnf("Scheduled transaction execution for [%s] not found.", id)
		return nil
	}

	storage := t.(*scheduler.Storage)

	if !storage.Executed {
		storage.Timer.Stop()
		s.logger.Infof("Cancelled scheduled execution for TX [%s].", id)
	} else {
		s.logger.Infof("TX [%s] was already broadcast/executed.", id)
	}

	return nil
}

// NewScheduler - Creates new instance of Scheduler
func NewScheduler(
	topicId string,
	operator string,
	executionWindow int64,
	contractService *bridge.BridgeContractService,
	hederaClient *hederaClient.HederaNodeClient,
) *Scheduler {
	topicID, err := hedera.TopicIDFromString(topicId)
	if err != nil {
		log.Fatal("Invalid topic id: [%v]", topicID)
	}

	return &Scheduler{
		logger:          config.GetLoggerFor("Scheduler"),
		tasks:           new(sync.Map),
		operator:        operator,
		executionWindow: executionWindow,
		contractService: contractService,
		hederaClient:    hederaClient,
		topicID:         topicID,
	}
}

// computeExecutionTime - computes the time at which the TX must be executed based on message timestamp and slot provided
func (s *Scheduler) computeExecutionTime(messageTimestamp int64, slot int64) time.Time {
	executionTimeNanos := messageTimestamp + timestamp.ToNanos(slot*s.executionWindow)

	return time.Unix(0, executionTimeNanos)
}

func (s *Scheduler) execute(submission ethsubmission.Submission) (*types.Transaction, error) {
	signatures, err := getSignatures(submission.Messages)
	if err != nil {
		return nil, err
	}
	return s.contractService.SubmitSignatures(submission.TransactOps, submission.CryptoTransferMessage, signatures)
}

func (s *Scheduler) submitEthTxTopicMessage(id string, submission ethsubmission.Submission, ethTxHash string) (*hedera.TransactionID, error) {
	ethTxMsg := &protomsg.TopicEthTransactionMessage{
		TransactionId: id,
		Hash:          submission.Messages[0].Hash,
		EthTxHash:     ethTxHash,
	}

	msg := &protomsg.TopicSubmissionMessage{
		Type: protomsg.TopicSubmissionType_EthTransaction,
		Message: &protomsg.TopicSubmissionMessage_TopicEthTransactionMessage{
			TopicEthTransactionMessage: ethTxMsg}}

	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		s.logger.Errorf("Failed to marshal protobuf TX [%s], TX Hash [%s]. Error [%s].", id, ethTxHash)
	}

	return s.hederaClient.SubmitTopicConsensusMessage(s.topicID, msgBytes)
}

func (s *Scheduler) waitForEthTxMined(ethTx common.Hash) (bool, error) {
	return s.contractService.Client.WaitForTransactionSuccess(ethTx)
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

package scheduler

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/golang/protobuf/proto"
	"github.com/hashgraph/hedera-sdk-go"
	hederaClient "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/process"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/model/ethsubmission"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/ethereum/bridge"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	protomsg "github.com/limechain/hedera-eth-bridge-validator/proto"
	log "github.com/sirupsen/logrus"
	"sort"
	"strings"
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
	_, exists := s.tasks.Load(id)
	if exists {
		return errors.New(fmt.Sprintf("Transaction with ID [%s] already scheduled for execution.", id))
	}

	et, err := s.computeExecutionTime(submission.Messages)
	if err != nil {
		return err
	}

	executeIn := et.Sub(time.Now())
	timer := time.NewTimer(executeIn)
	s.tasks.Store(id, timer)
	go func() {
		<-timer.C

		s.tasks.Delete(id)

		ethTx, err := s.execute(submission)
		if err != nil {
			s.logger.Errorf("Failed to execute Scheduled TX for [%s]. Error [%s].", submission.CryptoTransferMessage.TransactionId, err)
			return
		}
		ethTxHashString := ethTx.Hash().String()

		s.logger.Infof("Executed Scheduled TX [%s], TX Hash [%s].", id, ethTxHashString)
		tx, err := s.submitEthTxTopicMessage(id, submission, ethTxHashString)
		if err != nil {
			log.Errorf("Failed to submit topic consensus eth tx message for TX [%s], TX Hash [%s]. Error [%s].", id, ethTxHashString, err)
			return
		}
		log.Infof("Submitted topic consensus eth tx message for TX [%s], Tx Hash [%s] at Transaction ID [%s].", id, ethTxHashString, tx.String())

		s.waitForEthTxSuccess(id, ethTx.Hash())
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
	s.tasks.Delete(id)

	timer := t.(*time.Timer)
	timer.Stop()

	s.logger.Infof("Cancelled scheduled execution for TX [%s]", id)
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

func (s *Scheduler) computeExecutionTime(messages []message.TransactionMessage) (time.Time, error) {
	sort.Sort(message.ByTimestamp(messages))
	slot, err := s.computeExecutionSlot(messages)
	if err != nil {
		return time.Unix(0, 0), err
	}

	firstSignatureTimestamp := messages[0].TransactionTimestamp
	executionTime := time.Unix(0, firstSignatureTimestamp).Unix() + (int64(slot) * s.executionWindow)

	return time.Unix(executionTime, 0), nil
}

func (s *Scheduler) computeExecutionSlot(messages []message.TransactionMessage) (int, error) {
	for i := 0; i < len(messages); i++ {
		if strings.ToLower(messages[i].SignerAddress) == strings.ToLower(s.operator) {
			return i, nil
		}
	}

	return -1, errors.New(fmt.Sprintf("Operator is not amongst the potential leaders - [%v]", s.operator))
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
		Type: process.EthTransactionMessage,
		Message: &protomsg.TopicSubmissionMessage_TopicEthTransactionMessage{
			TopicEthTransactionMessage: ethTxMsg}}

	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		s.logger.Errorf("Failed to marshal protobuf TX [%s], TX Hash [%s]. Error [%s].", id, ethTxHash)
	}

	return s.hederaClient.SubmitTopicConsensusMessage(s.topicID, msgBytes)
}

func (s *Scheduler) waitForEthTxSuccess(id string, ethTx common.Hash) {
	ethTxString := ethTx.String()
	success, err := s.contractService.Client.WaitForTransactionSuccess(ethTx)
	if err != nil {
		s.logger.Errorf("Waiting for execution for TX [%s] and Hash [%s] failed. Error [%s].", id, ethTxString, err)
		return
	}

	if success {
		s.logger.Infof("Successful execution of TX [%s] with TX Hash [%s].", id, ethTxString)
	} else {
		s.logger.Warn("Execution for TX [%s] with TX Hash [%s] was not successful.", id, ethTxString)
	}
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

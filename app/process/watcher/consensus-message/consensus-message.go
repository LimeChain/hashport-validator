package consensusmessage

import (
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hashgraph/hedera-sdk-go"
	hederaClient "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	"github.com/limechain/hedera-eth-bridge-validator/app/process"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/publisher"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	validatorproto "github.com/limechain/hedera-eth-bridge-validator/proto"
	"github.com/limechain/hedera-watcher-sdk/queue"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

type ConsensusTopicWatcher struct {
	nodeClient       *hederaClient.HederaNodeClient
	mirrorClient     *hederaClient.HederaMirrorClient
	topicID          hedera.TopicID
	typeMessage      string
	maxRetries       int
	statusRepository repositories.StatusRepository
	startTimestamp   int64
	started          bool
	logger           *log.Entry
}

func NewConsensusTopicWatcher(nodeClient *hederaClient.HederaNodeClient, mirrorClient *hederaClient.HederaMirrorClient, topicID hedera.TopicID, repository repositories.StatusRepository, maxRetries int, startTimestamp int64) *ConsensusTopicWatcher {
	return &ConsensusTopicWatcher{
		nodeClient:       nodeClient,
		mirrorClient:     mirrorClient,
		topicID:          topicID,
		typeMessage:      process.HCSMessageType,
		statusRepository: repository,
		maxRetries:       maxRetries,
		startTimestamp:   startTimestamp,
		started:          false,
		logger:           config.GetLoggerFor(fmt.Sprintf("Topic [%s] Watcher", topicID.String())),
	}
}

func (ctw ConsensusTopicWatcher) Watch(q *queue.Queue) {
	go ctw.subscribeToTopic(q)
}

func (ctw ConsensusTopicWatcher) getTimestamp(q *queue.Queue) int64 {
	topicAddress := ctw.topicID.String()
	milestoneTimestamp := ctw.startTimestamp
	var err error

	if !ctw.started {
		if milestoneTimestamp > 0 {
			return milestoneTimestamp
		}

		ctw.logger.Warn("Starting Timestamp was empty, proceeding to get [timestamp] from database.")
		milestoneTimestamp, err := ctw.statusRepository.GetLastFetchedTimestamp(topicAddress)
		if err == nil && milestoneTimestamp > 0 {
			return milestoneTimestamp
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			ctw.logger.Fatal(err)
		}

		ctw.logger.Warn("Database Timestamp was empty, proceeding with [timestamp] from current moment.")
		milestoneTimestamp = time.Now().UnixNano()
		e := ctw.statusRepository.CreateTimestamp(topicAddress, milestoneTimestamp)
		if e != nil {
			ctw.logger.Fatal(e)
		}
		return milestoneTimestamp
	}

	milestoneTimestamp, err = ctw.statusRepository.GetLastFetchedTimestamp(topicAddress)
	if err != nil {
		ctw.logger.Warnf("Database Timestamp was empty. Restarting. Error - [%s]", err)
		ctw.started = false
		ctw.restart(q)
	}

	return milestoneTimestamp
}

func (ctw ConsensusTopicWatcher) processMessage(message []byte, timestamp int64, q *queue.Queue) {
	msg := &validatorproto.TopicSubmissionMessage{}
	err := proto.Unmarshal(message, msg)
	if err != nil {
		ctw.logger.Errorf("Could not unmarshal message - [%s]. Skipping the processing of this message -  [%s]", message, err)
		return
	}
	msg.TransactionTimestamp = timestamp

	publisher.Publish(msg, ctw.typeMessage, ctw.topicID, q)
	err = ctw.statusRepository.UpdateLastFetchedTimestamp(ctw.topicID.String(), timestamp)
	if err != nil {
		ctw.logger.Errorf("Could not update last fetched timestamp - [%d]", timestamp)
	}
}

func (ctw ConsensusTopicWatcher) subscribeToTopic(q *queue.Queue) {
	ctw.logger.Info("Starting watcher")
	milestoneTimestamp := ctw.getTimestamp(q)
	if milestoneTimestamp == 0 {
		ctw.logger.Fatalf("Could not start Consensus Message Watcher for topic [%s] - Could not generate a milestone timestamp.", ctw.topicID)
	}

	ctw.started = true

	_, err := hedera.NewTopicMessageQuery().
		SetStartTime(time.Unix(0, milestoneTimestamp)).
		SetTopicID(ctw.topicID).
		Subscribe(
			ctw.nodeClient.GetClient(),
			func(response hedera.TopicMessage) {
				ctw.logger.Debugf("Consensus Topic [%s] - Message incoming: [%s] - Contents: [%s]", response.ConsensusTimestamp, ctw.topicID, response.Contents)
				ctw.processMessage(response.Contents, response.ConsensusTimestamp.UnixNano(), q)
			},
		)

	if err != nil {
		ctw.logger.Errorf("Did not subscribe to [%s].", ctw.topicID)
		return
	}
	ctw.logger.Infof("Subscribed to [%s] successfully.", ctw.topicID)
}

func (ctw ConsensusTopicWatcher) restart(q *queue.Queue) {
	if ctw.maxRetries > 0 {
		ctw.maxRetries--
		ctw.logger.Infof("Watcher is trying to reconnect")
		go ctw.Watch(q)
		return
	}
	ctw.logger.Errorf("Watcher failed: [Too many retries]")
}

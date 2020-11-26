package consensusmessage

import (
	b64 "encoding/base64"
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/hashgraph/hedera-sdk-go"
	hederaClient "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	"github.com/limechain/hedera-eth-bridge-validator/app/process"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/publisher"
	validatorproto "github.com/limechain/hedera-eth-bridge-validator/proto"
	"github.com/limechain/hedera-watcher-sdk/queue"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"strconv"
	"time"
)

type ConsensusTopicWatcher struct {
	nodeClient       *hederaClient.HederaNodeClient
	mirrorClient     *hederaClient.HederaMirrorClient
	topicID          hedera.TopicID
	typeMessage      string
	maxRetries       int
	statusRepository repositories.StatusRepository
	startTimestamp   string
	started          bool
}

func NewConsensusTopicWatcher(nodeClient *hederaClient.HederaNodeClient, mirrorClient *hederaClient.HederaMirrorClient, topicID hedera.TopicID, repository repositories.StatusRepository, maxRetries int, startTimestamp string) *ConsensusTopicWatcher {
	return &ConsensusTopicWatcher{
		nodeClient:       nodeClient,
		mirrorClient:     mirrorClient,
		topicID:          topicID,
		typeMessage:      process.HCSMessageType,
		statusRepository: repository,
		maxRetries:       maxRetries,
		startTimestamp:   startTimestamp,
		started:          false,
	}
}

func (ctw ConsensusTopicWatcher) Watch(q *queue.Queue) {
	go ctw.subscribeToTopic(q)
}

func (ctw ConsensusTopicWatcher) getTimestamp(q *queue.Queue) string {
	topicAddress := ctw.topicID.String()
	milestoneTimestamp := ctw.startTimestamp
	var err error

	if !ctw.started {
		if milestoneTimestamp != "" {
			return milestoneTimestamp
		}

		log.Warnf("[%s] Starting Timestamp was empty, proceeding to get [timestamp] from database.\n", topicAddress)
		milestoneTimestamp, err := ctw.statusRepository.GetLastFetchedTimestamp(topicAddress)
		if milestoneTimestamp != "" {
			return milestoneTimestamp
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Fatal(err)
		}

		log.Warnf("[%s] Database Timestamp was empty, proceeding with [timestamp] from current moment.\n", topicAddress)
		milestoneTimestamp = strconv.FormatInt(time.Now().Unix(), 10)
		e := ctw.statusRepository.CreateTimestamp(topicAddress, milestoneTimestamp)
		if e != nil {
			log.Fatal(e)
		}
		return milestoneTimestamp
	}

	milestoneTimestamp, err = ctw.statusRepository.GetLastFetchedTimestamp(topicAddress)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Warnf("[%s] Database Timestamp was empty. Restarting.\n", topicAddress)
		ctw.started = false
		ctw.restart(q)
	}

	return milestoneTimestamp
}

func (ctw ConsensusTopicWatcher) processMessage(message []byte, timestamp string, q *queue.Queue) {
	msg := &validatorproto.TopicSignatureMessage{}
	err := proto.Unmarshal(message, msg)
	if err != nil {
		log.Errorf("Could not unmarshal message - [%s]. Skipping the processing of this message -  [%s]", message, err)
		return
	}

	publisher.Publish(msg, ctw.typeMessage, ctw.topicID, q)
	err = ctw.statusRepository.UpdateLastFetchedTimestamp(ctw.topicID.String(), timestamp)
	if err != nil {
		log.Errorf("Could not update last fetched timestamp - [%s]", timestamp)
	}
}

func (ctw ConsensusTopicWatcher) subscribeToTopic(q *queue.Queue) {
	log.Infof("Starting Consensus Message Watcher for topic [%s]\n", ctw.topicID)
	milestoneTimestamp := ctw.getTimestamp(q)
	if milestoneTimestamp == "" {
		log.Fatalf("Could not start Consensus Message Watcher for topic [%s] - Could not generate a milestone timestamp.\n", ctw.topicID)
	}

	log.Infof("Started Consensus Message Watcher for topic [%s]\n", ctw.topicID)
	unprocessedMessages, err := ctw.mirrorClient.GetUnprocessedMessagesAfterTimestamp(ctw.topicID, milestoneTimestamp)
	if err != nil {
		log.Errorf("Could not get unprocessed messages after timestamp [%s]", milestoneTimestamp)
		log.Fatal(err)
	}
	ctw.started = true
	log.Printf("Found [%v] unprocessed messages. Processing now\n", len(unprocessedMessages.Messages))

	for _, u := range unprocessedMessages.Messages {
		decodedMessage, err := b64.StdEncoding.DecodeString(u.Message)
		if err != nil {
			log.Errorf("Could not decode message - [%s]. Skipping the processing of this message", u.Message)
			continue
		}

		ctw.processMessage(decodedMessage, u.ConsensusTimestamp, q)
	}

	_, err = hedera.NewTopicMessageQuery().
		SetTopicID(ctw.topicID).
		Subscribe(
			ctw.nodeClient.GetClient(),
			func(response hedera.TopicMessage) {
				log.Infof("Consensus Topic [%s] - Message incoming: [%s]", response.ConsensusTimestamp, ctw.topicID, response.Contents)
				ctw.processMessage(response.Contents, strconv.FormatInt(response.ConsensusTimestamp.Unix(), 10), q)
			})

	if err != nil {
		log.Errorf("Did not subscribe to [%s].", ctw.topicID)
		return
	}
	log.Infof("Subscribed to [%s] successfully.", ctw.topicID)
}

func (ctw ConsensusTopicWatcher) restart(q *queue.Queue) {
	if ctw.maxRetries > 0 {
		ctw.maxRetries--
		log.Infof("Consensus Topic [%s] - Watcher is trying to reconnect\n", ctw.topicID)
		go ctw.Watch(q)
		return
	}
	log.Errorf("Consensus Topic [%s] - Watcher failed: [Too many retries]\n", ctw.topicID)
}

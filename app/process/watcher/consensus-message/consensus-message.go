package consensusmessage

import (
	"errors"
	"github.com/hashgraph/hedera-sdk-go"
	hederaClient "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/publisher"
	"github.com/limechain/hedera-watcher-sdk/queue"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"strconv"
	"time"
)

type ConsensusTopicWatcher struct {
	client           *hederaClient.HederaMirrorClient
	topicID          hedera.ConsensusTopicID
	typeMessage      string
	maxRetries       int
	statusRepository repositories.StatusRepository
	startTimestamp   string
	started          bool
}

func NewConsensusTopicWatcher(client *hederaClient.HederaMirrorClient, topicID hedera.ConsensusTopicID, repository repositories.StatusRepository, maxRetries int, startTimestamp string) *ConsensusTopicWatcher {
	return &ConsensusTopicWatcher{
		client:           client,
		topicID:          topicID,
		typeMessage:      "HCS_TOPIC_MSG",
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

func (ctw ConsensusTopicWatcher) subscribeToTopic(q *queue.Queue) {
	log.Infof("Starting Consensus Message Watcher for topic [%s]\n", ctw.topicID)
	milestoneTimestamp := ctw.getTimestamp(q)
	if milestoneTimestamp == "" {
		log.Fatalf("Could not start Consensus Message Watcher for topic [%s] - Could not generate a milestone timestamp.\n", ctw.topicID)
	}

	log.Infof("Started Consensus Message Watcher for topic [%s]\n", ctw.topicID)
	unprocessedMessages, err := ctw.client.GetUnprocessedMessagesAfterTimestamp(ctw.topicID, milestoneTimestamp)
	if err != nil {
		log.Fatal(err)
	}
	ctw.started = true
	for _, u := range unprocessedMessages.Messages {
		publisher.Publish(u, ctw.typeMessage, ctw.topicID, q)
		err := ctw.statusRepository.UpdateLastFetchedTimestamp(ctw.topicID.String(), u.ConsensusTimestamp)
		if err != nil {
			log.Fatal(err)
		}
	}

	_, err = hedera.NewMirrorConsensusTopicQuery().
		SetTopicID(ctw.topicID).
		Subscribe(
			*ctw.client.GetMirrorClient(),
			func(response hedera.MirrorConsensusTopicResponse) {
				log.Infof("Consensus Topic [%s] - Message incoming: [%s]", response.ConsensusTimestamp, ctw.topicID, response.Message)
				publisher.Publish(response, ctw.typeMessage, ctw.topicID, q)
				err := ctw.statusRepository.UpdateLastFetchedTimestamp(ctw.topicID.String(), strconv.FormatInt(response.ConsensusTimestamp.Unix(), 10))
				if err != nil {
					log.Fatal(err)
				}
			},
			func(err error) {
				log.Errorf("Consensus Topic [%s] - Error incoming: [%s]", ctw.topicID, err)
				time.Sleep(10 * time.Second)
				ctw.restart(q)
			},
		)

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

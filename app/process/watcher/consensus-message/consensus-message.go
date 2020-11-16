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
	client           *hederaClient.HederaClient
	topicID          hedera.ConsensusTopicID
	typeMessage      string
	maxRetries       int
	statusRepository repositories.StatusRepository
	startTimestamp   string
	started          bool
}

func NewConsensusTopicWatcher(client *hederaClient.HederaClient, topicID hedera.ConsensusTopicID, repository repositories.StatusRepository, maxRetries int, startTimestamp string) *ConsensusTopicWatcher {
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

func (ctw ConsensusTopicWatcher) retrieveTimestamp() string {
	if ctw.startTimestamp == "" {
		now := time.Now()
		milestoneTimestamp := strconv.FormatInt(now.Unix(), 10)
		log.Infof("Proceeding to monitor from current moment [%s]\n", now.String())
		return milestoneTimestamp
	}
	return ctw.startTimestamp
}

func (ctw ConsensusTopicWatcher) subscribeToTopic(q *queue.Queue) {
	var err error
	milestoneTimestamp := ctw.startTimestamp

	if !ctw.started {
		log.Warnln("Starting Timestamp was empty, proceeding to get [timestamp] from database.")
		if milestoneTimestamp == "" {
			milestoneTimestamp, err = ctw.statusRepository.GetLastFetchedTimestamp(ctw.topicID.String())
			if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warnln("Database Timestamp was empty, proceeding with [timestamp] from current moment.")
				milestoneTimestamp = strconv.FormatInt(time.Now().Unix(), 10)
				e := ctw.statusRepository.CreateTimestamp(ctw.topicID.String(), milestoneTimestamp)
				if e != nil {
					log.Fatal(e)
				}
			}
		}
	} else {
		milestoneTimestamp, err = ctw.statusRepository.GetLastFetchedTimestamp(ctw.topicID.String())
		if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warnln("Database Timestamp was empty. Restarting.")
			ctw.started = false
			ctw.subscribeToTopic(q)
		}
	}

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

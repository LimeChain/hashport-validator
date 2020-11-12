package consensusmessage

import (
	"github.com/hashgraph/hedera-sdk-go"
	hederaClient "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/publisher"
	"github.com/limechain/hedera-watcher-sdk/queue"
	log "github.com/sirupsen/logrus"
	"time"
)

type ConsensusTopicWatcher struct {
	client      *hederaClient.HederaClient
	topicID     hedera.ConsensusTopicID
	typeMessage string
	maxRetries  int
}

func NewConsensusTopicWatcher(client *hederaClient.HederaClient, topicID hedera.ConsensusTopicID, maxRetries int) *ConsensusTopicWatcher {
	return &ConsensusTopicWatcher{
		client:      client,
		topicID:     topicID,
		typeMessage: "HCS_TOPIC_MSG",
		maxRetries:  maxRetries,
	}
}

func (ctw ConsensusTopicWatcher) Watch(q *queue.Queue) {
	go ctw.subscribeToTopic(q)
}

func (ctw ConsensusTopicWatcher) subscribeToTopic(q *queue.Queue) {
	_, e := hedera.NewMirrorConsensusTopicQuery().
		SetTopicID(ctw.topicID).
		Subscribe(
			*ctw.client.GetMirrorClient(),
			func(response hedera.MirrorConsensusTopicResponse) {
				log.Infof("Consensus Topic [%s] - Message incoming: [%s]", response.ConsensusTimestamp, ctw.topicID, response.Message)
				publisher.Publish(response, ctw.typeMessage, ctw.topicID, q)
			},
			func(err error) {
				log.Errorf("Consensus Topic [%s] - Error incoming: [%s]", ctw.topicID, err)
				time.Sleep(10 * time.Second)
				ctw.restart(q)
			},
		)

	if e != nil {
		log.Infof("Did not subscribe to [%s].", ctw.topicID)
		return
	}
	log.Infof("Subscribed to [%s] successfully.", ctw.topicID)
}

func (ctw ConsensusTopicWatcher) restart(q *queue.Queue) {
	if ctw.maxRetries > 0 {
		ctw.maxRetries--
		log.Printf("Consensus Topic [%s] - Watcher is trying to reconnect\n", ctw.topicID)
		go ctw.Watch(q)
		return
	}
	log.Errorf("Consensus Topic [%s] - Watcher failed: [Too many retries]\n", ctw.topicID)
}

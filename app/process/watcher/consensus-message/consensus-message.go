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
	go ctw.subscribeToTopic(ctw.topicID, ctw.typeMessage, q)
}

func (ctw ConsensusTopicWatcher) subscribeToTopic(topicId hedera.ConsensusTopicID, typeMessage string, q *queue.Queue) {
	_, e := hedera.NewMirrorConsensusTopicQuery().
		SetTopicID(topicId).
		Subscribe(
			*ctw.client.GetMirrorClient(),
			func(response hedera.MirrorConsensusTopicResponse) {
				log.Infof("[%s] - Topic [%s] - Response incoming: [%s]", response.ConsensusTimestamp, topicId, response.Message)
				publisher.Publish(response, typeMessage, topicId, q)
			},
			func(err error) {
				log.Errorf("Error incoming: [%s]", err)
				time.Sleep(10 * time.Second)
				if ctw.maxRetries > 0 {
					ctw.maxRetries--
					log.Printf("Topic [%s] - Watcher is trying to reconnect\n", ctw.topicID)
					go ctw.subscribeToTopic(topicId, typeMessage, q)
					return
				}
				log.Errorf("Topic [%s] - Watcher failed: [Too many retries]\n", ctw.topicID)
			},
		)

	if e != nil {
		log.Infof("Did not subscribe to [%s].", topicId)
		return
	}
	log.Infof("Subscribed to [%s] successfully.", topicId)
}

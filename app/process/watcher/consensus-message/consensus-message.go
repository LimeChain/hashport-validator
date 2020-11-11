package consensusmessage

import (
	hederasdk "github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/publisher"
	"github.com/limechain/hedera-watcher-sdk/queue"
	"log"
)

type ConsensusTopicWatcher struct {
	client      *hedera.Client
	topicID     hederasdk.ConsensusTopicID
	typeMessage string
}

func (ctw ConsensusTopicWatcher) Watch(q *queue.Queue) {
	ctw.subscribeToTopic(ctw.topicID, ctw.typeMessage, q)
}

func (ctw ConsensusTopicWatcher) subscribeToTopic(topicId hederasdk.ConsensusTopicID, typeMessage string, q *queue.Queue) {
	_, e := hederasdk.NewMirrorConsensusTopicQuery().
		SetTopicID(topicId).
		Subscribe(
			*ctw.client.GetMirror(),
			func(response hederasdk.MirrorConsensusTopicResponse) {
				log.Printf("[%s] - Topic [%s] - Response incoming: [%s]", response.ConsensusTimestamp, topicId, response.Message)
				publisher.Publish(response, typeMessage, topicId, q)
			},
			func(err error) {
				log.Printf("Error incoming: [%s]", err)
			},
		)

	if e != nil {
		log.Printf("Did not subscribe to [%s].", topicId)
		return
	}
	log.Printf("Subscribed to [%s] successfully.", topicId)
}

func NewConsensusTopicWatcher(client *hedera.Client, topicID hederasdk.ConsensusTopicID) *ConsensusTopicWatcher {
	return &ConsensusTopicWatcher{
		client:      client,
		topicID:     topicID,
		typeMessage: "HCS_TOPIC_MSG",
	}
}

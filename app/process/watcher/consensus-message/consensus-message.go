package consensus_message

import (
	hederasdk "github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/proceed"
	"github.com/limechain/hedera-watcher-sdk/queue"
	"log"
)

type ConsensusTopicWatcher struct {
	client      hederasdk.MirrorClient
	topicID     hederasdk.ConsensusTopicID
	typeMessage string
}

func (ctw ConsensusTopicWatcher) Watch(q *queue.Queue) {
	subscribeToTopic(ctw.client, ctw.topicID, ctw.typeMessage, q)
}

func NewConsensusTopicWatcher(client hederasdk.MirrorClient, topicID hederasdk.ConsensusTopicID) *ConsensusTopicWatcher {
	return &ConsensusTopicWatcher{
		client:      client,
		topicID:     topicID,
		typeMessage: "HCS_TOPIC_MSG",
	}
}

func subscribeToTopic(client hederasdk.MirrorClient, topicId hederasdk.ConsensusTopicID, typeMessage string, q *queue.Queue) {
	_, e := hederasdk.NewMirrorConsensusTopicQuery().
		SetTopicID(topicId).
		Subscribe(
			client,
			func(response hederasdk.MirrorConsensusTopicResponse) {
				log.Printf("[%s] - Topic [%s] - Response incoming: [%s]", response.ConsensusTimestamp, topicId, response.Message)
				proceed.Proceed(response, typeMessage, topicId, q)
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

package consensus_message

import (
	"encoding/json"
	hederasdk "github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-watcher-sdk/queue"
	"github.com/limechain/hedera-watcher-sdk/types"
	"log"
)

type ConsensusTopicWatcher struct {
	TopicID hederasdk.ConsensusTopicID
}

func (ctw ConsensusTopicWatcher) Watch(q *queue.Queue) {
	subscribeToTopic(ctw.TopicID, q)
}

func subscribeToTopic(topicId hederasdk.ConsensusTopicID, q *queue.Queue) {
	client, e := hederasdk.NewMirrorClient(config.LoadConfig().Hedera.MirrorNode.Client)
	if e != nil {
		log.Printf("Did not subscribe to [%s].", topicId)
		return
	}
	_, e = hederasdk.NewMirrorConsensusTopicQuery().
		SetTopicID(topicId).
		Subscribe(
			client,
			func(response hederasdk.MirrorConsensusTopicResponse) {
				log.Printf("[%s] - Topic [%s] - Response incoming: [%s]", response.ConsensusTimestamp, topicId, response.Message)

				message, e := json.Marshal(response)
				if e != nil {
					log.Printf("Failed marshalling response from topic [%s]\n", topicId)
				}

				q.Push(&types.Message{
					Payload: message,
					Type:    "HCS_TOPIC_MSG",
				})
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

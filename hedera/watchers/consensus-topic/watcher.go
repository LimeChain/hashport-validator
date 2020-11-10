package consensus_topic

import (
	"Event-Listener/hedera/config"
	hederasdk "github.com/hashgraph/hedera-sdk-go"
	"log"
)

type ConsensusTopicWatcher struct {
	TopicID hederasdk.ConsensusTopicID
}

func (ctw ConsensusTopicWatcher) Watch( /* TODO: add SDK queue as a parameter */ ) {
	subscribeToTopic(ctw.TopicID /* TODO: add SDK queue as a parameter */)
}

func subscribeToTopic(topicId hederasdk.ConsensusTopicID /* TODO: add SDK queue as a parameter */) {
	client, e := hederasdk.NewMirrorClient(config.MirrorNodeAPIAddress)
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
				// TODO: Push response to SDK queue
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

package consensus_topic

import (
	"Event-Listener/hedera/config"
	"fmt"
	hederasdk "github.com/hashgraph/hedera-sdk-go"
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
		fmt.Printf("Did not subscribe to [%s].", topicId)
		return
	}

	_, e = hederasdk.NewMirrorConsensusTopicQuery().
		SetTopicID(topicId).
		Subscribe(
			client,
			func(response hederasdk.MirrorConsensusTopicResponse) {
				fmt.Printf("[%s] - Topic [%s] - Response incoming: [%s]", response.ConsensusTimestamp, topicId, response.Message)
				// TODO: Push response to SDK queue
			},
			func(err error) {
				fmt.Printf("Error incoming: [%s]", err)
			},
		)

	if e != nil {
		fmt.Printf("Did not subscribe to [%s].", topicId)
		return
	}
	fmt.Printf("Subscribed to [%s] successfully.", topicId)
}

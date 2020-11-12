package consensusmessage

import (
	"github.com/hashgraph/hedera-sdk-go"
	hederaClient "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/publisher"
	"github.com/limechain/hedera-watcher-sdk/queue"
	log "github.com/sirupsen/logrus"
)

type ConsensusTopicWatcher struct {
	client      *hederaClient.HederaClient
	topicID     hedera.ConsensusTopicID
	typeMessage string
}

func (ctw ConsensusTopicWatcher) Watch(q *queue.Queue) {
	ctw.subscribeToTopic(ctw.topicID, ctw.typeMessage, q)
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
			},
		)

	if e != nil {
		log.Infof("Did not subscribe to [%s].", topicId)
		return
	}
	log.Infof("Subscribed to [%s] successfully.", topicId)
}

func NewConsensusTopicWatcher(client *hederaClient.HederaClient, topicID hedera.ConsensusTopicID) *ConsensusTopicWatcher {
	return &ConsensusTopicWatcher{
		client:      client,
		topicID:     topicID,
		typeMessage: "HCS_TOPIC_MSG",
	}
}

package handler

import (
	"fmt"
	hederasdk "github.com/hashgraph/hedera-sdk-go"
)

func TopicCreation(client *hederasdk.Client) (hederasdk.TransactionReceipt, error) {
	txId, _ := hederasdk.NewConsensusTopicCreateTransaction().Execute(client)
	return txId.GetReceipt(client)
}

func SubscribeToTopic(mirrorClient hederasdk.MirrorClient, topicId hederasdk.ConsensusTopicID) {
	_, e := hederasdk.NewMirrorConsensusTopicQuery().
		SetTopicID(topicId).
		Subscribe(
			mirrorClient,
			func(response hederasdk.MirrorConsensusTopicResponse) {
				// TODO: Handle incoming message
				fmt.Printf("[%s] - Topic [%s] - Response incoming: [%s]", response.ConsensusTimestamp, topicId, response.Message)
				handle(response)
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

func handle(hederasdk.MirrorConsensusTopicResponse) {

}

func SubmitMessageToTopic(client *hederasdk.Client, topic hederasdk.ConsensusTopicID, message string) (hederasdk.TransactionReceipt, error) {
	receipt, _ := hederasdk.NewConsensusMessageSubmitTransaction().
		SetTopicID(topic).
		SetMessage([]byte(message)).
		Execute(client)
	return receipt.GetReceipt(client)
}

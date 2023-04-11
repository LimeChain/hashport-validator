package update_config

import (
	"github.com/hashgraph/hedera-sdk-go/v2"
	"time"
)

func CreateNewTopicFroxenTx(client *hedera.Client, content []byte, topicIdParsed hedera.TopicID, executor hedera.AccountID, nodeAccount hedera.AccountID, additionTime time.Duration) []byte {
	transactionID := hedera.NewTransactionIDWithValidStart(executor, time.Now().Add(additionTime))
	frozenTx, err := hedera.NewTopicMessageSubmitTransaction().
		SetTopicID(topicIdParsed).
		SetMessage(content).
		SetMaxChunks(100).
		SetTransactionID(transactionID).
		SetNodeAccountIDs([]hedera.AccountID{nodeAccount}).
		FreezeWith(client)
	if err != nil {
		panic(err)
	}

	bytes, err := frozenTx.ToBytes()
	if err != nil {
		panic(err)
	}

	return bytes
}

package publisher

import (
	"encoding/json"
	"github.com/limechain/hedera-watcher-sdk/queue"
	"github.com/limechain/hedera-watcher-sdk/types"
	log "github.com/sirupsen/logrus"
)

func Publish(m interface{}, typeMessage string, id interface{}, q *queue.Queue) {
	message, e := json.Marshal(m)
	if e != nil {
		log.Fatalf("[%s] - Failed marshalling response - ID: [%s]\n", typeMessage, id)
	}

	q.Push(&types.Message{
		Payload: message,
		Type:    typeMessage,
	})
}

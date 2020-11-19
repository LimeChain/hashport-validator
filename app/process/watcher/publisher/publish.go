package publisher

import (
	"github.com/golang/protobuf/proto"
	"github.com/limechain/hedera-watcher-sdk/queue"
	"github.com/limechain/hedera-watcher-sdk/types"
	log "github.com/sirupsen/logrus"
)

func Publish(m proto.Message, typeMessage string, id interface{}, q *queue.Queue) {
	message, e := proto.Marshal(m)
	if e != nil {
		log.Fatalf("[%s] - Failed marshalling response - ID: [%s]\n", typeMessage, id)
	}
	log.Println(m)
	q.Push(&types.Message{
		Payload: message,
		Type:    typeMessage,
	})
}

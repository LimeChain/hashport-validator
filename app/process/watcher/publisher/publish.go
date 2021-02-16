/*
 * Copyright 2021 LimeChain Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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

	q.Push(&types.Message{
		Payload: message,
		Type:    typeMessage,
	})
}

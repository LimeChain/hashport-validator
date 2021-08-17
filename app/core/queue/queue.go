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

package queue

import "math/big"

type Message struct {
	Payload interface{}
	ChainId *big.Int
}

// Queue is a wrapper of a go channel, particularly to restrict actions on the channel itself
type Queue struct {
	channel chan *Message
}

// Push pushes a message to the channel
func (q *Queue) Push(message *Message) {
	q.channel <- message
}

func (q *Queue) Channel() chan *Message {
	return q.channel
}

func NewQueue() *Queue {
	ch := make(chan *Message)
	return &Queue{channel: ch}
}

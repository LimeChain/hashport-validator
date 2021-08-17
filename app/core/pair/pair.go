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

package pair

import (
	q "github.com/limechain/hedera-eth-bridge-validator/app/core/queue"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/queue"
	"math/big"
)

type Watcher interface {
	Watch(queue queue.Queue)
}

type Handler interface {
	Handle(interface{})
}

// Pair represents a pair of a watcher and handlers, to which the watcher pushes messages
// which the handlers process
type Pair struct {
	queue    queue.Queue
	watcher  Watcher
	handlers map[*big.Int]Handler
}

// Listen begins the actions of the handlers and the watcher
func (p *Pair) Listen() {
	p.handle()
	p.watch()
}

// handle subscribes to the channel messages, processing them synchronously
func (p *Pair) handle() {
	go func() {
		for messages := range p.queue.Channel() {
			go p.handlers[messages.ChainId].Handle(messages.Payload)
		}
	}()
}

// watch initializes the Watcher's Watch
func (p *Pair) watch() {
	go p.watcher.Watch(p.queue)
}

func NewPair(watcher Watcher, handler map[*big.Int]Handler) *Pair {
	return &Pair{
		watcher:  watcher,
		handlers: handler,
		queue:    q.NewQueue(),
	}
}

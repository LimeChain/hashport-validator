/*
 * Copyright 2022 LimeChain Ltd.
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

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

var (
	q   *Queue   = nil
	msg *Message = nil
)

func setup() {
	q = NewQueue()
	msg = &Message{
		Payload: nil,
		Topic:   "topic",
	}
}

func Test_NewQueue(t *testing.T) {
	setup()
	assert.NotNil(t, q)
}

func Test_Push(t *testing.T) {
	setup()

	received := 0
	var wg sync.WaitGroup
	go func() {
		select {
		case <-q.Channel():
			received++
			wg.Done()
		}
	}()

	wg.Add(1)
	q.Push(msg)
	wg.Wait()
	assert.Equal(t, received, 1)
}

func Test_Channel(t *testing.T) {
	setup()

	assert.NotNil(t, q.Channel())
}

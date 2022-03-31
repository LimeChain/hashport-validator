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

package server

import (
	q "github.com/limechain/hedera-eth-bridge-validator/app/core/queue"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/queue"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	server        *Server
	queueInstance queue.Queue
	handlerTopic  = constants.TopicMessageSubmission
	port          = ":8000"
)

func Test_NewServer(t *testing.T) {
	setup()

	actualServer := NewServer()

	assert.Equal(t, server.logger, actualServer.logger)
	assert.Equal(t, server.handlers, actualServer.handlers)
	assert.Equal(t, server.watchers, actualServer.watchers)
}

func Test_AddWatcher(t *testing.T) {
	setup()

	server.AddWatcher(mocks.MWatcher)

	assert.Len(t, server.watchers, 1)
	assert.Equal(t, server.watchers[0], mocks.MWatcher)
}

func Test_AddHandler(t *testing.T) {
	setup()

	server.AddHandler(handlerTopic, mocks.MHandler)

	assert.Len(t, server.handlers, 1)
	assert.Equal(t, server.handlers[handlerTopic], mocks.MHandler)
}

func setup() {
	mocks.Setup()
	queueInstance = q.NewQueue()

	server = &Server{
		logger:   config.GetLoggerFor("Server"),
		handlers: make(map[string]Handler),
		queue:    queueInstance,
	}
}

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
	"github.com/go-chi/chi"
	q "github.com/limechain/hedera-eth-bridge-validator/app/core/queue"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/queue"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type Watcher interface {
	Watch(queue queue.Queue)
}

type Handler interface {
	Handle(interface{})
}

type Server struct {
	logger   *log.Entry
	watchers []Watcher
	handlers map[string]Handler
	queue    queue.Queue
}

func NewServer() *Server {
	return &Server{
		logger:   config.GetLoggerFor("Server"),
		handlers: make(map[string]Handler),
		queue:    q.NewQueue(),
	}
}

func (s *Server) AddWatcher(watcher Watcher) {
	s.watchers = append(s.watchers, watcher)
}

func (s *Server) AddHandler(topic string, handler Handler) {
	s.handlers[topic] = handler
}

// Run starts every handler and watcher, serving the chi.Mux on a given port
func (s *Server) Run(chi *chi.Mux, port string) {
	go func() {
		for message := range s.queue.Channel() {
			go s.handlers[message.Topic].Handle(message.Payload)
		}
	}()

	for _, watcher := range s.watchers {
		go watcher.Watch(s.queue)
	}
	s.logger.Infof("Listening on port [%s]", port)
	s.logger.Fatal(http.ListenAndServe(port, chi))
}

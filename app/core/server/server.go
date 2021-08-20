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

package server

import (
	"github.com/go-chi/chi"
	"github.com/limechain/hedera-eth-bridge-validator/app/core/pair"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type Server struct {
	logger *log.Entry
	pairs  []*pair.Pair
}

func NewServer() *Server {
	return &Server{
		logger: config.GetLoggerFor("Server"),
	}
}

func (s *Server) AddPair(watcher pair.Watcher, handlers map[int64]pair.Handler) {
	s.pairs = append(s.pairs, pair.NewPair(watcher, handlers))
}

// Run starts every pair's Listen and serves the chi.Mux on a given port
func (s *Server) Run(chi *chi.Mux, port string) {
	for _, p := range s.pairs {
		p.Listen()
	}
	s.logger.Infof("Listening on port [%s]", port)
	s.logger.Fatal(http.ListenAndServe(port, chi))
}

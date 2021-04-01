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

func (s *Server) AddPair(watcher pair.Watcher, handler pair.Handler) {
	s.pairs = append(s.pairs, pair.NewPair(watcher, handler))
}

func (s *Server) Run(chi *chi.Mux, port string) {
	for _, p := range s.pairs {
		p.Listen()
	}
	s.logger.Infof("Listening on port [%s]", port)
	s.logger.Fatal(http.ListenAndServe(port, chi))
}

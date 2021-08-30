package burn_message

import (
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	logger *log.Entry
}

func NewHandler() *Handler {
	return &Handler{
		logger: config.GetLoggerFor("Hedera Burn and Topic Message Handler"),
	}
}

func (mhh Handler) Handle(payload interface{}) {
	// TODO:
}

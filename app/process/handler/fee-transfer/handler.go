package fee_transfer

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	burn_event "github.com/limechain/hedera-eth-bridge-validator/app/model/burn-event"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	burnService service.BurnEvent
	logger      *log.Entry
}

func NewHandler(burnService service.BurnEvent) *Handler {
	return &Handler{
		burnService: burnService,
		logger:      config.GetLoggerFor("Hedera Fee and Schedule Transfer Handler"),
	}
}

func (fth Handler) Handle(payload interface{}) {
	event, ok := payload.(*burn_event.BurnEvent)
	if !ok {
		fth.logger.Errorf("Could not cast payload [%s]", payload)
		return
	}
	fth.burnService.ProcessEvent(*event)
}

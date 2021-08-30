package mint_hts

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	lock_event "github.com/limechain/hedera-eth-bridge-validator/app/model/lock-event"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	lockService service.LockEvent
	logger      *log.Entry
}

func NewHandler(lockService service.LockEvent) *Handler {
	return &Handler{
		lockService: lockService,
		logger:      config.GetLoggerFor("Hedera Mint and Transfer Handler"),
	}
}

func (mhh Handler) Handle(payload interface{}) {
	event, ok := payload.(*lock_event.LockEvent)
	if !ok {
		mhh.logger.Errorf("Could not cast payload [%s]", payload)
		return
	}
	mhh.lockService.ProcessEvent(*event)
}

package fee_message

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

// Handler is transfers event handler
type Handler struct {
	transfersService service.Transfers
	logger           *log.Entry
}

func NewHandler(transfersService service.Transfers) *Handler {
	return &Handler{
		logger:           config.GetLoggerFor("Hedera Transfer and Topic Submission Handler"),
		transfersService: transfersService,
	}
}

func (fmh Handler) Handle(payload interface{}) {
	transferMsg, ok := payload.(*model.Transfer)
	if !ok {
		fmh.logger.Errorf("Could not cast payload [%s]", payload)
		return
	}

	transactionRecord, err := fmh.transfersService.InitiateNewTransfer(*transferMsg)
	if err != nil {
		fmh.logger.Errorf("[%s] - Error occurred while initiating processing. Error: [%s]", transferMsg.TransactionId, err)
		return
	}

	if transactionRecord.Status != transfer.StatusInitial {
		fmh.logger.Debugf("[%s] - Previously added with status [%s]. Skipping further execution.", transactionRecord.TransactionID, transactionRecord.Status)
		return
	}

	err = fmh.transfersService.ProcessTransfer(*transferMsg)
	if err != nil {
		fmh.logger.Errorf("[%s] - Processing failed. Error: [%s]", transferMsg.TransactionId, err)
		return
	}
}

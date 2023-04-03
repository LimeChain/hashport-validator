package transfer_reset

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/metrics"
	transferModel "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/response"
	"github.com/limechain/hedera-eth-bridge-validator/config"
)

var (
	Route  = "/transfer-reset"
	logger = config.GetLoggerFor(fmt.Sprintf("Router [%s]", Route))
)

// POST: .../transfer-reset
func getTransfer(transferService service.Transfers, prometheusService service.Prometheus) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// transferID := chi.URLParam(r, "id")
		req := new(transferModel.TransferReset)
		err := json.NewDecoder(r.Body).Decode(req)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.ErrorResponse(err))
			return
		}

		transactionId := req.TransactionId
		sourceChainId := req.SourceChainId
		targetChainId := req.TargetChainId
		oppositeToken := req.OppositeToken

		metrics.SetUserGetHisTokens(sourceChainId, targetChainId, oppositeToken, transactionId, prometheusService, logger)
		transferService.UpdateTransferStatusCompleted(transactionId)
	}
}

func NewRouter(transferService service.Transfers, prometheusService service.Prometheus) chi.Router {
	r := chi.NewRouter()
	r.Post("/", getTransfer(transferService, prometheusService))
	return r
}

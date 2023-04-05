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

func NewRouter(transferService service.Transfers, prometheusService service.Prometheus) chi.Router {
	r := chi.NewRouter()
	r.Post("/", transferReset(transferService, prometheusService))
	return r
}

// POST: .../transfer-reset
func transferReset(transferService service.Transfers, prometheusService service.Prometheus) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(transferModel.TransferReset)
		err := json.NewDecoder(r.Body).Decode(req)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.ErrorResponse(err))
			return
		}

		metrics.SetUserGetHisTokens(req.SourceChainId, req.TargetChainId, req.TargetToken, req.TransactionId, prometheusService, logger)
		transferService.UpdateTransferStatusCompleted(req.TransactionId)

		render.Status(r, http.StatusOK)
		render.PlainText(w, r, "OK")
	}
}

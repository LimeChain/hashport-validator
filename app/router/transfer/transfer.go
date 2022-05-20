package transfer

import (
	"encoding/json"
	"fmt"
	"net/http"

	transferModel "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"

	httpHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/response"
	"github.com/limechain/hedera-eth-bridge-validator/config"
)

var (
	Route  = "/transfers"
	logger = config.GetLoggerFor(fmt.Sprintf("Router [%s]", Route))
)

// GET: .../transfers/:id
func getTransfer(transfersService service.Transfers) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		transferID := chi.URLParam(r, "id")

		transferData, err := transfersService.TransferData(transferID)
		if err != nil {
			logger.Errorf("Router resolved with an error. Error [%s].", err)
			httpHelper.WriteErrorResponse(w, r, err)
			return
		}

		render.JSON(w, r, transferData)
	}
}

// POST: .../history
func history(transferService service.Transfers) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(transferModel.PagedRequest)
		err := json.NewDecoder(r.Body).Decode(req)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.ErrorResponse(err))
		}

		res, err := transferService.Paged(req)
		if err != nil {
			logger.Errorf("Router resolved with an error. Error [%v]", err)
			httpHelper.WriteErrorResponse(w, r, err)
			return
		}

		render.JSON(w, r, res)
	}
}

func NewRouter(service service.Transfers) chi.Router {
	r := chi.NewRouter()
	r.Get("/{id}", getTransfer(service))
	r.Post("/history", history(service))
	return r
}

package transfer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/limechain/hedera-eth-bridge-validator/constants"

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

const maxHistoryPageSize = 50

func NewRouter(service service.Transfers) chi.Router {
	r := chi.NewRouter()
	r.Get("/{id}", getTransfer(service))
	r.Post("/history", history(service))
	return r
}

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
			return
		}
		if req.Page <= 0 {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.ErrorResponse(fmt.Errorf("page must be greater than 0")))
			return
		}
		if req.PageSize <= 0 {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.ErrorResponse(fmt.Errorf("page size must be greater than 0")))
			return
		}
		if req.PageSize > maxHistoryPageSize {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.ErrorResponse(fmt.Errorf("maximum page size is %d", maxHistoryPageSize)))
			return
		}
		if t := req.Filter.TransactionId; strings.Contains(t, "0x") {
			if s := t[2:]; len(s) != constants.TransactionHashLength {
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, response.ErrorResponse(fmt.Errorf("invalid tx hash length")))
				return
			}
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

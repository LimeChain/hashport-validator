package transfer

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/response"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"net/http"
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
			switch err {
			case service.ErrNotFound:
				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, response.ErrorResponse(err))
			default:
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, response.ErrorResponse(response.ErrorInternalServerError))
			}

			return
		}

		render.JSON(w, r, transferData)
	}
}

func NewRouter(service service.Transfers) chi.Router {
	r := chi.NewRouter()
	r.Get("/{id}", getTransfer(service))
	return r
}

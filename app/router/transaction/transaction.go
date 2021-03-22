package transaction

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
	Route  = "/transaction"
	logger = config.GetLoggerFor(fmt.Sprintf("Router [%s]", Route))
)

// GET: .../transaction/:id
func getTransaction(messageService service.Messages) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		transactionId := chi.URLParam(r, "id")

		transactionData, err := messageService.TransactionData(transactionId)
		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ErrorResponse(response.ErrorInternalServerError))

			logger.Errorf("Router resolved with an error. Error [%s].", err)
			return
		}

		render.JSON(w, r, transactionData)
	}
}

func NewRouter(messageService service.Messages) chi.Router {
	r := chi.NewRouter()
	r.Get("/{id}", getTransaction(messageService))
	return r
}

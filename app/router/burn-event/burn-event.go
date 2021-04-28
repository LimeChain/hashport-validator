package burn_event

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
	Route  = "/events"
	logger = config.GetLoggerFor(fmt.Sprintf("Router [%s]", Route))
)

// GET: .../events/:id/tx
func getTxID(burnService service.BurnEvent) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		eventID := chi.URLParam(r, "id")

		txID, err := burnService.ScheduledTxID(eventID)
		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ErrorResponse(response.ErrorInternalServerError))

			logger.Errorf("Router resolved with an error. Error [%s].", err)
			return
		}

		render.JSON(w, r, txID)
	}
}

func NewRouter(service service.BurnEvent) chi.Router {
	r := chi.NewRouter()
	r.Get("/{id}/tx", getTxID(service))
	return r
}

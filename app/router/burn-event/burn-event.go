package burn_event

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	httpHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/http"
	"github.com/limechain/hedera-eth-bridge-validator/config"
)

var (
	Route  = "/events"
	logger = config.GetLoggerFor(fmt.Sprintf("Router [%s]", Route))
)

// GET: .../events/:id/tx
func getTxID(burnService service.BurnEvent) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		eventID := chi.URLParam(r, "id")

		txID, err := burnService.TransactionID(eventID)
		if err != nil {
			logger.Errorf("Router resolved with an error. Error [%s].", err)
			httpHelper.WriteErrorResponse(w, r, err)
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

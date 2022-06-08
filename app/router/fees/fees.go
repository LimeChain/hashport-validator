package fees

import (
	"errors"
	"net/http"

	"github.com/limechain/hedera-eth-bridge-validator/app/router/response"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
)

const Route = "/fees"

func NewRouter(pricingService service.Pricing) http.Handler {
	r := chi.NewRouter()
	r.Get("/nft", feesNftResponse(pricingService))
	return r
}

func feesNftResponse(pricingService service.Pricing) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		res := pricingService.NftFees()
		if len(res) == 0 {
			err := errors.New("router resolved with an error. Error [No NFT fees records]")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ErrorResponse(err))
			return
		}

		render.JSON(w, r, res)
	}
}

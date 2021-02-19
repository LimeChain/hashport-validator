package metadata

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/response"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/fees"
	"net/http"
)

var (
	ErrorInternalServerError = errors.New("SOMETHING_WENT_WRONG")
)

const (
	GasPriceGweiParam = "gasPriceGwei"
	HBARCurrency      = "HBAR"
)

var metadataRoute = fmt.Sprintf("/metadata/{%s}", GasPriceGweiParam)

func getMetadata(calculator *fees.FeeCalculator) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		gasPriceGwei := chi.URLParam(r, GasPriceGweiParam)

		txFee, err := calculator.GetEstimatedTxFee(gasPriceGwei)
		if err != nil {
			if errors.Is(fees.InvalidGasPrice, err) {
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, response.ErrInvalidRequest(err))

			} else {
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, response.ErrInvalidRequest(ErrorInternalServerError))
			}
			return
		}

		render.JSON(w, r, &response.MetadataResponse{
			TransactionFee:         txFee,
			TransactionFeeCurrency: HBARCurrency,
			GasPriceGwei:           gasPriceGwei,
		})
	}
}

func NewMetadataRouter(feeCalculator *fees.FeeCalculator) http.Handler {
	r := chi.NewRouter()
	r.Get(metadataRoute, getMetadata(feeCalculator))

	return r
}

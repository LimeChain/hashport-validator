package metadata

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/response"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/fees"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"net/http"
)

const (
	GasPriceGweiParam = "gasPriceGwei"
	HBARCurrency      = "HBAR"
)

var (
	ErrorInternalServerError = errors.New("SOMETHING_WENT_WRONG")
)

var (
	metadataRoute = "/metadata" //
	logger        = config.GetLoggerFor(fmt.Sprintf("Router [%s]", metadataRoute))
)

// /metadata?gasPriceGwei=${gasPriceGwei}
func getMetadata(calculator *fees.FeeCalculator) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		gasPriceGwei := r.URL.Query().Get(GasPriceGweiParam)

		txFee, err := calculator.GetEstimatedTxFee(gasPriceGwei)
		if err != nil {
			if errors.Is(fees.InvalidGasPrice, err) {
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, response.ErrorResponse(err))

				logger.Debugf("Invalid provided value: [%s].", gasPriceGwei)
			} else {
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, response.ErrorResponse(ErrorInternalServerError))

				logger.Errorf("Router resolved with an error. Error [%s].", err)
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

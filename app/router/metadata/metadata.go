/*
 * Copyright 2021 LimeChain Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
	MetadataRoute = "/metadata"
	logger        = config.GetLoggerFor(fmt.Sprintf("Router [%s]", MetadataRoute))
)

// GET: .../metadata?gasPriceGwei=${gasPriceGwei}
func getMetadata(calculator *fees.Calculator) func(w http.ResponseWriter, r *http.Request) {
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

func NewRouter(feeCalculator *fees.Calculator) chi.Router {
	r := chi.NewRouter()
	r.Get("/", getMetadata(feeCalculator))
	return r
}

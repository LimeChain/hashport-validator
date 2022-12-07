/*
 * Copyright 2022 LimeChain Ltd.
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

package fees

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/limechain/hedera-eth-bridge-validator/constants"

	"github.com/limechain/hedera-eth-bridge-validator/app/router/response"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
)

const Route = "/fees"

func NewRouter(pricingService service.Pricing, feeService service.Fee, feePolicyHandler service.FeePolicyHandler) http.Handler {
	r := chi.NewRouter()
	r.Get("/nft", feesNftResponse(pricingService))
	r.Get("/calculate-for", calculateForResponse(feeService, feePolicyHandler))
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

func calculateForResponse(feeService service.Fee, feePolicyHandler service.FeePolicyHandler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		targetChainIdStr := r.URL.Query().Get("targetChain")
		account := r.URL.Query().Get("account")
		token := r.URL.Query().Get("token")
		amountStr := r.URL.Query().Get("amount")

		targetChainId, err := strconv.ParseUint(targetChainIdStr, 10, 64)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.ErrorResponse(err))
			return
		}

		// Validator knows only hedera fees. Fees from other networks are handled from bridge.yml. Check config.NewBridge
		if targetChainId == constants.HederaNetworkId {
			render.Status(r, http.StatusBadRequest)
			return
		}

		amount, err := strconv.ParseInt(amountStr, 10, 64)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.ErrorResponse(err))
			return
		}

		feeAmount, exist := feePolicyHandler.FeeAmountFor(targetChainId, account, token, amount)

		if !exist {
			feeAmount, _ = feeService.CalculateFee(token, amount)
		}

		render.JSON(w, r, feeAmount)
	}
}

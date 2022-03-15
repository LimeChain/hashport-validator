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

package min_amounts

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/response"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"net/http"
)

var (
	Route  = "/min-amounts"
	logger = config.GetLoggerFor(fmt.Sprintf("Router [%s]", Route))
)

// Router for min amounts
func NewRouter(pricingService service.Pricing) http.Handler {
	r := chi.NewRouter()
	r.Get("/", minAmountsResponse(pricingService))
	return r
}

// GET: .../min-amounts
func minAmountsResponse(pricingService service.Pricing) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		minAmounts := pricingService.GetMinAmountsForAPI()
		if len(minAmounts) == 0 {
			err := errors.New("Router resolved with an error. Error [No min amount records].")
			logger.Errorf(err.Error())
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ErrorResponse(err))
			return
		}

		render.JSON(w, r, minAmounts)
	}
}

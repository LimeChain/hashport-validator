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

package utils

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	httpHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/http"
	"github.com/limechain/hedera-eth-bridge-validator/config"
)

var (
	Route  = "/utils"
	logger = config.GetLoggerFor(fmt.Sprintf("Router [%s]", Route))
)

func NewRouter(utilsSvc service.Utils) chi.Router {
	r := chi.NewRouter()
	r.Get("/convert-evm-hash-to-bridge-tx-id/{evmHash}/{chainId}", convertEvmTxHashToBridgeTxId(utilsSvc))
	return r
}

func convertEvmTxHashToBridgeTxId(utilsSvc service.Utils) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		evmTxId := chi.URLParam(r, "evmHash")
		chainIdStr := chi.URLParam(r, "chainId")
		chainId, err := strconv.ParseUint(chainIdStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		res, err := utilsSvc.ConvertEvmHashToBridgeTxId(evmTxId, chainId)
		if err != nil {
			logger.Errorf("Router resolved with an error. Error [%s].", err)
			httpHelper.WriteErrorResponse(w, r, err)
			return
		}

		render.JSON(w, r, res)
	}
}

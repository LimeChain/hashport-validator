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
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/response"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"net/http"
	"strconv"
)

var (
	Route  = "/utils"
	logger = config.GetLoggerFor(fmt.Sprintf("Router [%s]", Route))
)

func NewRouter(utilsSvc service.Utils) chi.Router {
	r := chi.NewRouter()
	r.Get("/convert-evm-tx-id-to-hedera-tx-id/{evmTxId}/{chainId}", convertEvmTxToHederaTx(utilsSvc))
	return r
}

func convertEvmTxToHederaTx(utilsSvc service.Utils) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		evmTxId := chi.URLParam(r, "evmTxId")
		chainIdStr := chi.URLParam(r, "chainId")
		chainId, err := strconv.ParseUint(chainIdStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		res, err := utilsSvc.ConvertEvmTxIdToHederaTxId(evmTxId, chainId)
		if err != nil {
			logger.Errorf("Router resolved with an error. Error [%s].", err)
			switch err {
			case service.ErrNotFound:
				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, response.ErrorResponse(err))
			default:
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, response.ErrorResponse(response.ErrorInternalServerError))
			}
			return
		}

		render.JSON(w, r, res)
	}
}

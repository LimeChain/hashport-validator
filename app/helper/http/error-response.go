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

package http

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/response"
)

func WriteErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	switch err {
	case service.ErrBadRequestTransferTargetNetworkNoSignaturesRequired:
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, response.ErrorResponse(err))
	case service.ErrNotFound:
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, response.ErrorResponse(err))
	default:
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, response.ErrorResponse(response.ErrorInternalServerError))
	}
}

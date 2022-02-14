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

package config_bridge

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"net/http"
)

var (
	Route = "/config/bridge"
)

//Router for bridge config
func NewRouter(bridgeConfig parser.Bridge) http.Handler {
	r := chi.NewRouter()
	r.Get("/", configBridgeResponse(bridgeConfig))
	return r
}

// GET: .../config/bridge
func configBridgeResponse(bridgeConfig parser.Bridge) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, r, bridgeConfig)
	}
}

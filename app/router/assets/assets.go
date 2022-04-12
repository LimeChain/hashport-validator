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

package assets

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/asset"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"net/http"
)

var (
	Route = "/assets"
)

type NetworkAssets struct {
	Fungible    map[string]asset.FungibleAssetInfo    `json:"fungible"`
	NonFungible map[string]asset.NonFungibleAssetInfo `json:"nonFungible"`
}

// Router for assets
func NewRouter(assetsService service.Assets) http.Handler {
	r := chi.NewRouter()
	r.Get("/", assetsResponse(assetsService))
	return r
}

// GET: .../assets
func assetsResponse(assetsService service.Assets) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		responseContent := generateResponseContent(assetsService)

		render.JSON(w, r, responseContent)
	}
}

func generateResponseContent(assetsService service.Assets) map[uint64]NetworkAssets {
	response := make(map[uint64]NetworkAssets)

	fungibleNetworkAssets := assetsService.FungibleNetworkAssets()
	nonFungibleNetworkAssets := assetsService.NonFungibleNetworkAssets()
	for networkId := range constants.NetworksById {
		response[networkId] = NetworkAssets{
			Fungible:    map[string]asset.FungibleAssetInfo{},
			NonFungible: map[string]asset.NonFungibleAssetInfo{},
		}

		// Fungible
		for _, assetAddress := range fungibleNetworkAssets[networkId] {
			fungibleAssetInfo, exist := assetsService.FungibleAssetInfo(networkId, assetAddress)
			if exist {
				response[networkId].Fungible[assetAddress] = fungibleAssetInfo
			}
		}

		// Non-Fungible
		for _, assetAddress := range nonFungibleNetworkAssets[networkId] {
			nonFungibleAssetInfo, exist := assetsService.NonFungibleAssetInfo(networkId, assetAddress)
			if exist {
				response[networkId].NonFungible[assetAddress] = nonFungibleAssetInfo
			}
		}
	}
	return response
}

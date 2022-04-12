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
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"net/http"
)

var (
	Route = "/assets"
)

type fungibleBridgeDetails struct {
	*asset.FungibleAssetInfo
	MinAmount string            `json:"minAmount"`
	Networks  map[uint64]string `json:"networks"`
}

type nonFungibleBridgeDetails struct {
	*asset.NonFungibleAssetInfo
	Networks map[uint64]string `json:"networks"`
}

type networkAssets struct {
	Fungible    map[string]fungibleBridgeDetails    `json:"fungible"`
	NonFungible map[string]nonFungibleBridgeDetails `json:"nonFungible"`
}

// Router for assets
func NewRouter(bridgeConfig parser.Bridge, assetsService service.Assets, pricingService service.Pricing) http.Handler {
	r := chi.NewRouter()
	r.Get("/", assetsResponse(bridgeConfig, assetsService, pricingService))
	return r
}

// GET: .../assets
func assetsResponse(bridgeConfig parser.Bridge, assetsService service.Assets, pricingService service.Pricing) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		responseContent := generateResponseContent(bridgeConfig, assetsService, pricingService)
		render.JSON(w, r, responseContent)
	}
}

func generateResponseContent(bridgeConfig parser.Bridge, assetsService service.Assets, pricingService service.Pricing) map[uint64]networkAssets {
	response := make(map[uint64]networkAssets)

	fungibleNetworkAssets := assetsService.FungibleNetworkAssets()
	nonFungibleNetworkAssets := assetsService.NonFungibleNetworkAssets()
	for networkId := range constants.NetworksById {
		response[networkId] = networkAssets{
			Fungible:    map[string]fungibleBridgeDetails{},
			NonFungible: map[string]nonFungibleBridgeDetails{},
		}

		// Fungible
		for _, assetAddress := range fungibleNetworkAssets[networkId] {
			fungibleAssetInfo, existInfo := assetsService.FungibleAssetInfo(networkId, assetAddress)
			minAmount, existMinAmount := pricingService.GetTokenPriceInfo(networkId, assetAddress)
			if existInfo && existMinAmount {
				fungibleAssetDetails := fungibleBridgeDetails{
					FungibleAssetInfo: &fungibleAssetInfo,
					MinAmount:         minAmount.MinAmountWithFee.String(),
					Networks:          bridgeConfig.Networks[networkId].Tokens.Fungible[assetAddress].Networks,
				}
				response[networkId].Fungible[assetAddress] = fungibleAssetDetails
			}
		}

		// Non-Fungible
		for _, assetAddress := range nonFungibleNetworkAssets[networkId] {
			nonFungibleAssetInfo, exist := assetsService.NonFungibleAssetInfo(networkId, assetAddress)
			if exist {
				fungibleAssetDetails := nonFungibleBridgeDetails{
					NonFungibleAssetInfo: &nonFungibleAssetInfo,
					Networks:             bridgeConfig.Networks[networkId].Tokens.Nft[assetAddress].Networks,
				}
				response[networkId].NonFungible[assetAddress] = fungibleAssetDetails
			}
		}
	}

	return response
}

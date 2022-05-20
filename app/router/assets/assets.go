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
	Route        = "/assets"
	BridgeConfig *parser.Bridge
)

type fungibleBridgeDetails struct {
	*asset.FungibleAssetInfo
	FeePercentage feePercentageInfo `json:"feePercentage"`
	MinAmount     string            `json:"minAmount"`
	Networks      map[uint64]string `json:"networks"`
	ReserveAmount string            `json:"reserveAmount"`
}

type feePercentageInfo struct {
	Amount        int64 `json:"amount"`
	MaxPercentage int64 `json:"maxPercentage"`
}

type nonFungibleBridgeDetails struct {
	*asset.NonFungibleAssetInfo
	Fee           int64             `json:"fee"`
	Networks      map[uint64]string `json:"networks"`
	ReserveAmount string            `json:"reserveAmount"`
}

type networkAssets struct {
	Fungible    map[string]fungibleBridgeDetails    `json:"fungible"`
	NonFungible map[string]nonFungibleBridgeDetails `json:"nonFungible"`
}

// Router for assets
func NewRouter(bridgeCfg *parser.Bridge, assetsService service.Assets, pricingService service.Pricing) http.Handler {
	BridgeConfig = bridgeCfg
	r := chi.NewRouter()
	r.Get("/", assetsResponse(assetsService, pricingService))
	return r
}

// GET: .../assets
func assetsResponse(assetsService service.Assets, pricingService service.Pricing) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		responseContent := generateResponseContent(assetsService, pricingService)
		render.JSON(w, r, responseContent)
	}
}

func generateResponseContent(assetsService service.Assets, pricingService service.Pricing) map[uint64]networkAssets {
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
				bridgeTokenInfo := BridgeConfig.Networks[networkId].Tokens.Fungible[assetAddress]
				var nativeAsset *asset.NativeAsset
				if !fungibleAssetInfo.IsNative {
					nativeAsset = assetsService.WrappedToNative(assetAddress, networkId)
				} else {
					nativeAsset = assetsService.FungibleNativeAsset(networkId, assetAddress)
				}
				feePercentage := nativeAsset.FeePercentage

				fungibleAssetDetails := fungibleBridgeDetails{
					FungibleAssetInfo: fungibleAssetInfo,
					FeePercentage:     feePercentageInfo{feePercentage, constants.FeeMaxPercentage},
					MinAmount:         minAmount.MinAmountWithFee.String(),
					Networks:          bridgeTokenInfo.Networks,
					ReserveAmount:     fungibleAssetInfo.ReserveAmount.String(),
				}
				response[networkId].Fungible[assetAddress] = fungibleAssetDetails
			}
		}

		// Non-Fungible
		for _, assetAddress := range nonFungibleNetworkAssets[networkId] {
			nonFungibleAssetInfo, exist := assetsService.NonFungibleAssetInfo(networkId, assetAddress)
			if exist {
				var nativeAddress string
				if nonFungibleAssetInfo.IsNative {
					nativeAddress = assetAddress
				} else {
					nativeAsset := assetsService.WrappedToNative(assetAddress, networkId)
					nativeAddress = nativeAsset.Asset
				}

				bridgeTokenInfo := BridgeConfig.Networks[networkId].Tokens.Nft[nativeAddress]
				nonFungibleAssetDetails := nonFungibleBridgeDetails{
					NonFungibleAssetInfo: nonFungibleAssetInfo,
					Fee:                  bridgeTokenInfo.Fee,
					Networks:             bridgeTokenInfo.Networks,
					ReserveAmount:        nonFungibleAssetInfo.ReserveAmount.String(),
				}
				response[networkId].NonFungible[assetAddress] = nonFungibleAssetDetails
			}
		}
	}

	return response
}

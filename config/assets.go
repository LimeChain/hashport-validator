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

package config

import (
	big_numbers "github.com/limechain/hedera-eth-bridge-validator/app/helper/big-numbers"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	log "github.com/sirupsen/logrus"
	"math/big"
	"strconv"
)

type Assets struct {
	nativeToWrapped map[int64]map[string]map[int64]string
	wrappedToNative map[int64]map[string]*NativeAsset
	// A mapping, storing all fungible tokens per network
	fungibleNetworkAssets map[int64][]string
	// A mapping, storing all fungible native assets per network
	fungibleNativeAssets map[int64]map[string]*NativeAsset
	// A mapping, storing all wrapped tokens per native network
	// to avoid a collision the key is a concatenation of wrappedChainId_wrappedAsset
	wrappedToNativeNetwork map[string]int64
}

type NativeAsset struct {
	MinAmount *big.Int
	ChainId   int64
	Asset     string
}

func (a Assets) NativeToWrapped(nativeAsset string, nativeChainId, targetChainId int64) string {
	return a.nativeToWrapped[nativeChainId][nativeAsset][targetChainId]
}

func (a Assets) WrappedToNative(wrappedAsset string, wrappedChainId int64) *NativeAsset {
	return a.wrappedToNative[wrappedChainId][wrappedAsset]
}

func (a Assets) NetworkAssets(id int64) []string {
	return a.fungibleNetworkAssets[id]
}

func (a Assets) GetFungibleNetworkAssets() map[int64][]string {
	return a.fungibleNetworkAssets
}

func (a Assets) FungibleNativeAsset(id int64, asset string) *NativeAsset {
	return a.fungibleNativeAssets[id][asset]
}

func (a Assets) IsNative(networkId int64, asset string) bool {
	_, isNative := a.nativeToWrapped[networkId][asset]
	return isNative
}

func (a Assets) GetOppositeAsset(sourceChainId uint64, targetChainId uint64, asset string) string {
	sourceChainIdCasted, targetChainIdCasted := int64(sourceChainId), int64(targetChainId)

	nativeAssetForTargetChain := a.WrappedToNative(asset, sourceChainIdCasted)
	if nativeAssetForTargetChain != nil {
		return nativeAssetForTargetChain.Asset
	}

	nativeAssetForSourceChain := a.WrappedToNative(asset, targetChainIdCasted)
	if nativeAssetForSourceChain != nil {
		return nativeAssetForSourceChain.Asset
	}

	if a.IsNative(sourceChainIdCasted, asset) {
		return a.NativeToWrapped(asset, sourceChainIdCasted, targetChainIdCasted)
	} else {
		return a.NativeToWrapped(asset, targetChainIdCasted, sourceChainIdCasted)
	}

}

func (a Assets) GetWrappedToNativeNetwork(wrappedAsset string) int64 {
	return a.wrappedToNativeNetwork[wrappedAsset]
}

func LoadAssets(networks map[int64]*parser.Network) Assets {
	nativeToWrapped := make(map[int64]map[string]map[int64]string)
	wrappedToNative := make(map[int64]map[string]*NativeAsset)
	fungibleNetworkAssets := make(map[int64][]string)
	fungibleNativeAssets := make(map[int64]map[string]*NativeAsset)
	wrappedToNativeNetwork := make(map[string]int64)

	for nativeChainId, network := range networks {
		if nativeToWrapped[nativeChainId] == nil {
			nativeToWrapped[nativeChainId] = make(map[string]map[int64]string)
		}
		if fungibleNativeAssets[nativeChainId] == nil {
			fungibleNativeAssets[nativeChainId] = make(map[string]*NativeAsset)
		}

		for nativeAsset, nativeAssetMapping := range network.Tokens {
			if nativeToWrapped[nativeChainId][nativeAsset] == nil {
				nativeToWrapped[nativeChainId][nativeAsset] = make(map[int64]string)
			}

			minAmount, err := parseAmount(nativeAssetMapping.MinAmount)
			if err != nil {
				log.Fatalf("Failed to parse min amount [%s]. Error: [%s]", nativeAssetMapping.MinAmount, err)
			}
			asset := &NativeAsset{
				MinAmount: minAmount,
				ChainId:   nativeChainId,
				Asset:     nativeAsset,
			}
			fungibleNativeAssets[nativeChainId][nativeAsset] = asset

			fungibleNetworkAssets[nativeChainId] = append(fungibleNetworkAssets[nativeChainId], nativeAsset)
			for wrappedChainId, wrappedAsset := range nativeAssetMapping.Networks {
				nativeToWrapped[nativeChainId][nativeAsset][wrappedChainId] = wrappedAsset

				uniqueKey := strconv.FormatInt(wrappedChainId, 10) + "_" + wrappedAsset
				wrappedToNativeNetwork[uniqueKey] = nativeChainId

				if wrappedToNative[wrappedChainId] == nil {
					wrappedToNative[wrappedChainId] = make(map[string]*NativeAsset)
				}
				fungibleNetworkAssets[wrappedChainId] = append(fungibleNetworkAssets[wrappedChainId], wrappedAsset)
				wrappedToNative[wrappedChainId][wrappedAsset] = asset
			}
		}
	}

	return Assets{
		nativeToWrapped:        nativeToWrapped,
		wrappedToNative:        wrappedToNative,
		fungibleNativeAssets:   fungibleNativeAssets,
		fungibleNetworkAssets:  fungibleNetworkAssets,
		wrappedToNativeNetwork: wrappedToNativeNetwork,
	}
}

func parseAmount(amount string) (*big.Int, error) {
	if amount == "" {
		return big.NewInt(0), nil
	}

	return big_numbers.ToBigInt(amount)
}

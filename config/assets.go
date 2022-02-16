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

package config

import (
	"github.com/ethereum/go-ethereum/common"
	big_numbers "github.com/limechain/hedera-eth-bridge-validator/app/helper/big-numbers"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	log "github.com/sirupsen/logrus"
	"math/big"
	"regexp"
)

type Assets struct {
	// A mapping, storing all networks' native tokens and their corresponding wrapped tokens
	nativeToWrapped map[uint64]map[string]map[uint64]string
	// A mapping, storing all networks' wrapped tokens and their corresponding native asset
	wrappedToNative map[uint64]map[string]*NativeAsset
	// A mapping, storing all fungible tokens per network
	fungibleNetworkAssets map[uint64][]string
	// A mapping, storing all fungible native assets per network
	fungibleNativeAssets map[uint64]map[string]*NativeAsset
}

type NativeAsset struct {
	MinAmount *big.Int
	ChainId   uint64
	Asset     string
}

func (a Assets) GetFungibleNetworkAssets() map[uint64][]string {
	return a.fungibleNetworkAssets
}

func (a Assets) WrappedFromNative(nativeChainId uint64, nativeAsset string) map[uint64]string {
	return a.nativeToWrapped[nativeChainId][nativeAsset]
}

func (a Assets) GetNativeToWrapped() map[uint64]map[string]map[uint64]string {
	return a.nativeToWrapped
}

func (a Assets) NativeToWrapped(nativeAsset string, nativeChainId, targetChainId uint64) string {
	return a.nativeToWrapped[nativeChainId][nativeAsset][targetChainId]
}

func (a Assets) WrappedToNative(wrappedAsset string, wrappedChainId uint64) *NativeAsset {
	return a.wrappedToNative[wrappedChainId][wrappedAsset]
}

func (a Assets) FungibleNetworkAssets(id uint64) []string {
	return a.fungibleNetworkAssets[id]
}

func (a Assets) FungibleNativeAsset(id uint64, asset string) *NativeAsset {
	return a.fungibleNativeAssets[id][asset]
}

func (a Assets) IsNative(networkId uint64, asset string) bool {
	_, isNative := a.nativeToWrapped[networkId][asset]
	return isNative
}

func (a Assets) GetOppositeAsset(sourceChainId uint64, targetChainId uint64, asset string) string {
	sourceChainIdCasted, targetChainIdCasted := sourceChainId, targetChainId

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

func LoadAssets(networks map[uint64]*parser.Network) Assets {
	nativeToWrapped := make(map[uint64]map[string]map[uint64]string)
	wrappedToNative := make(map[uint64]map[string]*NativeAsset)
	fungibleNetworkAssets := make(map[uint64][]string)
	fungibleNativeAssets := make(map[uint64]map[string]*NativeAsset)

	re, _ := regexp.Compile(constants.EvmCompatibleAddressPattern)

	for nativeChainId, network := range networks {
		if nativeToWrapped[nativeChainId] == nil {
			nativeToWrapped[nativeChainId] = make(map[string]map[uint64]string)
		}
		if fungibleNativeAssets[nativeChainId] == nil {
			fungibleNativeAssets[nativeChainId] = make(map[string]*NativeAsset)
		}

		for nativeAsset, nativeAssetMapping := range network.Tokens.Fungible {
			if nativeChainId != constants.HederaNetworkId {
				nativeAsset = common.HexToAddress(nativeAsset).String()
			}

			if nativeToWrapped[nativeChainId][nativeAsset] == nil {
				nativeToWrapped[nativeChainId][nativeAsset] = make(map[uint64]string)
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
				if isMatch := re.MatchString(wrappedAsset); isMatch {
					wrappedAsset = common.HexToAddress(wrappedAsset).String()
				}

				nativeToWrapped[nativeChainId][nativeAsset][wrappedChainId] = wrappedAsset

				if wrappedToNative[wrappedChainId] == nil {
					wrappedToNative[wrappedChainId] = make(map[string]*NativeAsset)
				}
				fungibleNetworkAssets[wrappedChainId] = append(fungibleNetworkAssets[wrappedChainId], wrappedAsset)
				wrappedToNative[wrappedChainId][wrappedAsset] = asset
			}
		}

		for nativeAsset, nativeAssetMapping := range network.Tokens.Nft {
			if nativeChainId != constants.HederaNetworkId {
				nativeAsset = common.HexToAddress(nativeAsset).String()
			}

			if nativeToWrapped[nativeChainId][nativeAsset] == nil {
				nativeToWrapped[nativeChainId][nativeAsset] = make(map[uint64]string)
			}
			for wrappedChainId, wrappedAsset := range nativeAssetMapping.Networks {
				if isMatch := re.MatchString(wrappedAsset); isMatch {
					wrappedAsset = common.HexToAddress(wrappedAsset).String()
				}

				nativeToWrapped[nativeChainId][nativeAsset][wrappedChainId] = wrappedAsset
				if wrappedToNative[wrappedChainId] == nil {
					wrappedToNative[wrappedChainId] = make(map[string]*NativeAsset)
				}
				wrappedToNative[wrappedChainId][wrappedAsset] = &NativeAsset{
					ChainId: nativeChainId,
					Asset:   nativeAsset,
				}
			}
		}
	}

	return Assets{
		nativeToWrapped:       nativeToWrapped,
		wrappedToNative:       wrappedToNative,
		fungibleNativeAssets:  fungibleNativeAssets,
		fungibleNetworkAssets: fungibleNetworkAssets,
	}
}

func parseAmount(amount string) (*big.Int, error) {
	if amount == "" {
		return big.NewInt(0), nil
	}

	return big_numbers.ToBigInt(amount)
}

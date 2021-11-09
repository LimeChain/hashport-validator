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

import "github.com/limechain/hedera-eth-bridge-validator/config/parser"

type Assets struct {
	nativeToWrapped map[int64]map[string]map[int64]string
	wrappedToNative map[int64]map[string]*NativeAsset
	networkAssets   map[int64][]string
}

type NativeAsset struct {
	ChainId int64
	Asset   string
}

func (a Assets) NativeToWrapped(nativeAsset string, nativeChainId, targetChainId int64) string {
	return a.nativeToWrapped[nativeChainId][nativeAsset][targetChainId]
}

func (a Assets) WrappedToNative(wrappedAsset string, wrappedChainId int64) *NativeAsset {
	return a.wrappedToNative[wrappedChainId][wrappedAsset]
}

func (a Assets) NetworkAssets(id int64) []string {
	return a.networkAssets[id]
}

func LoadAssets(networks map[int64]*parser.Network) Assets {
	nativeToWrapped := make(map[int64]map[string]map[int64]string)
	wrappedToNative := make(map[int64]map[string]*NativeAsset)
	networkAssets := make(map[int64][]string)
	for nativeChainId, network := range networks {
		if nativeToWrapped[nativeChainId] == nil {
			nativeToWrapped[nativeChainId] = make(map[string]map[int64]string)
		}
		for nativeAsset, nativeAssetMapping := range network.Tokens {
			if nativeToWrapped[nativeChainId][nativeAsset] == nil {
				nativeToWrapped[nativeChainId][nativeAsset] = make(map[int64]string)
			}
			networkAssets[nativeChainId] = append(networkAssets[nativeChainId], nativeAsset)
			for wrappedChainId, wrappedAsset := range nativeAssetMapping.Networks {
				nativeToWrapped[nativeChainId][nativeAsset][wrappedChainId] = wrappedAsset

				if wrappedToNative[wrappedChainId] == nil {
					wrappedToNative[wrappedChainId] = make(map[string]*NativeAsset)
				}
				networkAssets[wrappedChainId] = append(networkAssets[wrappedChainId], wrappedAsset)
				wrappedToNative[wrappedChainId][wrappedAsset] = &NativeAsset{
					ChainId: nativeChainId,
					Asset:   nativeAsset,
				}
			}
		}
	}

	return Assets{
		nativeToWrapped: nativeToWrapped,
		wrappedToNative: wrappedToNative,
		networkAssets:   networkAssets,
	}
}

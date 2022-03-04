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
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/wtoken"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	assetModel "github.com/limechain/hedera-eth-bridge-validator/app/model/asset"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strconv"
)

type Service struct {
	// A mapping, storing all networks' native tokens and their corresponding wrapped tokens
	nativeToWrapped map[uint64]map[string]map[uint64]string
	// A mapping, storing all networks' wrapped tokens and their corresponding native asset
	wrappedToNative map[uint64]map[string]*assetModel.NativeAsset
	// A mapping, storing all fungible tokens per network
	fungibleNetworkAssets map[uint64][]string
	// A mapping, storing all fungible native assets per network
	fungibleNativeAssets map[uint64]map[string]*assetModel.NativeAsset
	// A mapping, storing name and symbol for asset per network
	fungibleAssetInfos map[uint64]map[string]assetModel.FungibleAssetInfo

	logger *log.Entry
}

func (a *Service) GetFungibleNetworkAssets() map[uint64][]string {
	return a.fungibleNetworkAssets
}

func (a *Service) GetNativeToWrappedAssets() map[uint64]map[string]map[uint64]string {
	return a.nativeToWrapped
}

func (a *Service) WrappedFromNative(nativeChainId uint64, nativeAsset string) map[uint64]string {
	return a.nativeToWrapped[nativeChainId][nativeAsset]
}

func (a *Service) NativeToWrapped(nativeAsset string, nativeChainId, targetChainId uint64) string {
	return a.nativeToWrapped[nativeChainId][nativeAsset][targetChainId]
}

func (a *Service) WrappedToNative(wrappedAsset string, wrappedChainId uint64) *assetModel.NativeAsset {
	return a.wrappedToNative[wrappedChainId][wrappedAsset]
}

func (a *Service) FungibleNetworkAssets(id uint64) []string {
	return a.fungibleNetworkAssets[id]
}

func (a *Service) FungibleNativeAsset(id uint64, asset string) *assetModel.NativeAsset {
	return a.fungibleNativeAssets[id][asset]
}

func (a *Service) IsNative(networkId uint64, asset string) bool {
	_, isNative := a.nativeToWrapped[networkId][asset]
	return isNative
}

func (a *Service) GetOppositeAsset(sourceChainId uint64, targetChainId uint64, asset string) string {
	nativeAssetForTargetChain := a.WrappedToNative(asset, sourceChainId)
	if nativeAssetForTargetChain != nil {
		return nativeAssetForTargetChain.Asset
	}

	nativeAssetForSourceChain := a.WrappedToNative(asset, targetChainId)
	if nativeAssetForSourceChain != nil {
		return nativeAssetForSourceChain.Asset
	}

	if a.IsNative(sourceChainId, asset) {
		return a.NativeToWrapped(asset, sourceChainId, targetChainId)
	}

	return a.NativeToWrapped(asset, targetChainId, sourceChainId)
}

func (a *Service) GetFungibleAssetInfo(networkId uint64, assetAddressOrId string) (assetInfo assetModel.FungibleAssetInfo, exist bool) {
	assetInfo, exist = a.fungibleAssetInfos[networkId][assetAddressOrId]

	return assetInfo, exist
}

func (a *Service) getEVMFungibleAssetInfo(networkId uint64, assetAddress string, EVMClients map[uint64]client.EVM) (assetInfo assetModel.FungibleAssetInfo, err error) {
	evm := EVMClients[networkId].GetClient()

	evmAssetInstance, err := wtoken.NewWtoken(common.HexToAddress(assetAddress), evm)
	if err != nil {
		a.logger.Errorf("EVM with networkId [%d] for Asset [%s], and method NewWtoken - Error: [%s]", networkId, assetAddress, err)
		return assetInfo, err
	}

	name, err := evmAssetInstance.Name(&bind.CallOpts{})
	if err != nil {
		a.logger.Errorf("EVM with networkId [%d] for Asset [%s], and method Name - Error: [%s]", networkId, assetAddress, err)
		return assetInfo, err
	}
	assetInfo.Name = name

	symbol, err := evmAssetInstance.Symbol(&bind.CallOpts{})
	if err != nil {
		a.logger.Errorf("EVM with networkId [%d] for Asset [%s], and method Symbol - Error: [%s]", networkId, assetAddress, err)
		return assetInfo, err
	}
	assetInfo.Symbol = symbol

	decimals, err := evmAssetInstance.Decimals(&bind.CallOpts{})
	if err != nil {
		a.logger.Errorf("EVM with networkId [%d] for Asset [%s], and method Decimals - Error: [%s]", networkId, assetAddress, err)
		return assetInfo, err
	}
	assetInfo.Decimals = decimals

	return assetInfo, err
}

func (a *Service) getHederaFungibleAssetInfo(assetId string, mirrorNode client.MirrorNode) (assetInfo assetModel.FungibleAssetInfo, err error) {
	if assetId == constants.Hbar {
		assetInfo.Name = constants.Hbar
		assetInfo.Symbol = constants.Hbar
		return assetInfo, err
	}

	assetInfoResponse, e := mirrorNode.GetToken(assetId)
	if e != nil {
		a.logger.Errorf("Hedera Mirror Node method GetToken for Asset [%s] - Error: [%s]", assetId, e)
	} else {
		assetInfo.Name = assetInfoResponse.Name
		assetInfo.Symbol = assetInfoResponse.Symbol
		parsedDecimals, _ := strconv.Atoi(assetInfoResponse.Decimals)
		assetInfo.Decimals = uint8(parsedDecimals)
	}

	return assetInfo, err
}

func (a *Service) loadFungibleAssetInfos(networks map[uint64]*parser.Network, mirrorNode client.MirrorNode, EVMClients map[uint64]client.EVM) {
	a.fungibleAssetInfos = make(map[uint64]map[string]assetModel.FungibleAssetInfo)

	if len(EVMClients) == 0 {
		return
	}

	re, _ := regexp.Compile(constants.EvmCompatibleAddressPattern)

	for nativeChainId, networkInfo := range networks {
		if _, exist := a.fungibleAssetInfos[nativeChainId]; !exist {
			a.fungibleAssetInfos[nativeChainId] = make(map[string]assetModel.FungibleAssetInfo)
		}

		for nativeAsset, nativeAssetMapping := range networkInfo.Tokens.Fungible {

			var (
				assetInfo assetModel.FungibleAssetInfo
				err       error
			)

			if nativeChainId == constants.HederaNetworkId { // Hedera
				assetInfo, err = a.getHederaFungibleAssetInfo(nativeAsset, mirrorNode)
				if nativeAsset == constants.Hbar {
					assetInfo.Decimals = constants.HederaDefaultDecimals
				}

				if err != nil {
					a.logger.Fatalf("Failed to load Hedera Fungible Asset Info. Error [%v]", err)
				}
			} else { // EVM
				nativeAsset = common.HexToAddress(nativeAsset).String()
				assetInfo, err = a.getEVMFungibleAssetInfo(nativeChainId, nativeAsset, EVMClients)
				if err != nil {
					a.logger.Fatalf("Failed to load EVM NetworkId [%v] Fungible Asset Info. Error [%v]", nativeChainId, err)
				}
			}

			a.fungibleAssetInfos[nativeChainId][nativeAsset] = assetInfo

			for wrappedChainId, wrappedAsset := range nativeAssetMapping.Networks {
				if _, exist := a.fungibleAssetInfos[wrappedChainId]; !exist {
					a.fungibleAssetInfos[wrappedChainId] = make(map[string]assetModel.FungibleAssetInfo)
				}

				if isMatch := re.MatchString(wrappedAsset); isMatch {
					wrappedAsset = common.HexToAddress(wrappedAsset).String()
				}
				if wrappedChainId == constants.HederaNetworkId { // Hedera
					assetInfo, err = a.getHederaFungibleAssetInfo(wrappedAsset, mirrorNode)
					if err != nil {
						a.logger.Fatalf("Failed to load Hedera Fungible Asset Info. Error [%v]", err)
					}
				} else { // EVM
					wrappedAsset = common.HexToAddress(wrappedAsset).String()
					assetInfo, err = a.getEVMFungibleAssetInfo(wrappedChainId, wrappedAsset, EVMClients)
					if err != nil {
						a.logger.Fatalf("Failed to load EVM NetworkId [%v] Fungible Asset Info. Error [%v]", wrappedChainId, err)
					}
				}

				a.fungibleAssetInfos[wrappedChainId][wrappedAsset] = assetInfo
			}
		}
	}
}

func NewService(networks map[uint64]*parser.Network, HederaFeePercentages map[string]int64, routerClients map[uint64]*router.Router, mirrorNode client.MirrorNode, EVMClients map[uint64]client.EVM) *Service {
	nativeToWrapped := make(map[uint64]map[string]map[uint64]string)
	wrappedToNative := make(map[uint64]map[string]*assetModel.NativeAsset)
	fungibleNetworkAssets := make(map[uint64][]string)
	fungibleNativeAssets := make(map[uint64]map[string]*assetModel.NativeAsset)

	re, _ := regexp.Compile(constants.EvmCompatibleAddressPattern)

	for nativeChainId, network := range networks {
		if nativeToWrapped[nativeChainId] == nil {
			nativeToWrapped[nativeChainId] = make(map[string]map[uint64]string)
		}
		if fungibleNativeAssets[nativeChainId] == nil {
			fungibleNativeAssets[nativeChainId] = make(map[string]*assetModel.NativeAsset)
		}

		for nativeAsset, nativeAssetMapping := range network.Tokens.Fungible {
			if nativeChainId != constants.HederaNetworkId {
				nativeAsset = common.HexToAddress(nativeAsset).String()
			}

			if nativeToWrapped[nativeChainId][nativeAsset] == nil {
				nativeToWrapped[nativeChainId][nativeAsset] = make(map[uint64]string)
			}

			minAmount, err := parseAmount(nativeAssetMapping.MinFeeAmountInUsd)
			if err != nil {
				log.Fatalf("Failed to parse min amount [%s]. Error: [%s]", nativeAssetMapping.MinFeeAmountInUsd, err)
			}
			var feePercentage int64
			if nativeChainId == constants.HederaNetworkId {
				feePercentage = HederaFeePercentages[nativeAsset]
			} else {
				routerClient, exist := routerClients[nativeChainId]
				if exist {
					tokenFeeData, err := routerClient.TokenFeeData(&bind.CallOpts{}, common.HexToAddress(nativeAsset))
					if err != nil {
						log.Fatalf("Failed to get fee persentage from router contact for asset [%s]. Error: [%s]", nativeAsset, err)
					}
					feePercentage = tokenFeeData.ServiceFeePercentage.Int64()
				}
			}

			asset := &assetModel.NativeAsset{
				MinFeeAmountInUsd: minAmount,
				ChainId:           nativeChainId,
				Asset:             nativeAsset,
				FeePercentage:     feePercentage,
			}
			fungibleNativeAssets[nativeChainId][nativeAsset] = asset

			fungibleNetworkAssets[nativeChainId] = append(fungibleNetworkAssets[nativeChainId], nativeAsset)
			for wrappedChainId, wrappedAsset := range nativeAssetMapping.Networks {
				if isMatch := re.MatchString(wrappedAsset); isMatch {
					wrappedAsset = common.HexToAddress(wrappedAsset).String()
				}

				nativeToWrapped[nativeChainId][nativeAsset][wrappedChainId] = wrappedAsset

				if wrappedToNative[wrappedChainId] == nil {
					wrappedToNative[wrappedChainId] = make(map[string]*assetModel.NativeAsset)
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
					wrappedToNative[wrappedChainId] = make(map[string]*assetModel.NativeAsset)
				}
				wrappedToNative[wrappedChainId][wrappedAsset] = &assetModel.NativeAsset{
					ChainId: nativeChainId,
					Asset:   nativeAsset,
				}
			}
		}
	}

	instance := &Service{
		nativeToWrapped:       nativeToWrapped,
		wrappedToNative:       wrappedToNative,
		fungibleNativeAssets:  fungibleNativeAssets,
		fungibleNetworkAssets: fungibleNetworkAssets,
		logger:                config.GetLoggerFor("Service"),
	}
	instance.loadFungibleAssetInfos(networks, mirrorNode, EVMClients)

	return instance
}

func parseAmount(amount string) (result *decimal.Decimal, err error) {
	if amount == "" {
		return result, nil
	}
	newResult, err := decimal.NewFromString(amount)

	return &newResult, err
}

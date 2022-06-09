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
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gookit/event"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/token"
	client "github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	decimalHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/decimal"
	assetModel "github.com/limechain/hedera-eth-bridge-validator/app/model/asset"
	bridge_config_event "github.com/limechain/hedera-eth-bridge-validator/app/model/bridge-config-event"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	log "github.com/sirupsen/logrus"
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
	// A mapping, storing name and symbol for fungible asset per network
	fungibleAssetInfos map[uint64]map[string]*assetModel.FungibleAssetInfo
	// A mapping, storing all non-fungible tokens per network
	nonFungibleNetworkAssets map[uint64][]string
	// A mapping, storing name and symbol for non-fungible asset per network
	nonFungibleAssetInfos map[uint64]map[string]*assetModel.NonFungibleAssetInfo

	bridgeAccountId string
	logger          *log.Entry
}

func (a *Service) FungibleNetworkAssets() map[uint64][]string {
	return a.fungibleNetworkAssets
}

func (a *Service) NonFungibleNetworkAssets() map[uint64][]string {
	return a.nonFungibleNetworkAssets
}

func (a *Service) NativeToWrappedAssets() map[uint64]map[string]map[uint64]string {
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

func (a *Service) FungibleNetworkAssetsByChainId(chainId uint64) []string {
	return a.fungibleNetworkAssets[chainId]
}

func (a *Service) FungibleNativeAsset(nativeChainId uint64, nativeAssetAddress string) *assetModel.NativeAsset {
	return a.fungibleNativeAssets[nativeChainId][nativeAssetAddress]
}

func (a *Service) IsNative(networkId uint64, asset string) bool {
	_, isNative := a.nativeToWrapped[networkId][asset]
	return isNative
}

func (a *Service) OppositeAsset(chainOne uint64, chainTwo uint64, asset string) string {
	nativeAssetForTargetChain := a.WrappedToNative(asset, chainOne)
	if nativeAssetForTargetChain != nil {
		return nativeAssetForTargetChain.Asset
	}

	nativeAssetForSourceChain := a.WrappedToNative(asset, chainTwo)
	if nativeAssetForSourceChain != nil {
		return nativeAssetForSourceChain.Asset
	}

	if a.IsNative(chainOne, asset) {
		return a.NativeToWrapped(asset, chainOne, chainTwo)
	}

	return a.NativeToWrapped(asset, chainTwo, chainOne)
}

func (a *Service) FungibleAssetInfo(networkId uint64, assetAddressOrId string) (assetInfo *assetModel.FungibleAssetInfo, exist bool) {
	assetInfo, exist = a.fungibleAssetInfos[networkId][assetAddressOrId]

	return assetInfo, exist
}

func (a *Service) NonFungibleAssetInfo(networkId uint64, assetAddressOrId string) (assetInfo *assetModel.NonFungibleAssetInfo, exist bool) {
	assetInfo, exist = a.nonFungibleAssetInfos[networkId][assetAddressOrId]

	return assetInfo, exist
}

func (a *Service) FetchEvmFungibleReserveAmount(
	networkId uint64,
	assetAddress string,
	isNative bool,
	evmTokenClient client.EvmFungibleToken,
	routerContractAddress string,
) (inLowestDenomination *big.Int, err error) {
	if isNative {
		inLowestDenomination, err = evmTokenClient.BalanceOf(&bind.CallOpts{}, common.HexToAddress(routerContractAddress))
		if err != nil {
			a.logger.Errorf("EVM with networkId [%d] for asset [%s], and method BalanceOf - Error: [%s]", networkId, assetAddress, err)
			return nil, err
		}
	} else {
		inLowestDenomination, err = evmTokenClient.TotalSupply(&bind.CallOpts{})
		if err != nil {
			a.logger.Errorf("EVM with networkId [%d] for asset [%s], and method TotalSupply - Error: [%s]", networkId, assetAddress, err)
			return nil, err
		}
	}
	return inLowestDenomination, err
}

func (a *Service) FetchEvmNonFungibleReserveAmount(
	networkId uint64,
	assetAddress string,
	isNative bool,
	evmTokenClient client.EvmNft,
	routerContractAddress string,
) (inLowestDenomination *big.Int, err error) {
	if isNative {
		inLowestDenomination, err = evmTokenClient.BalanceOf(&bind.CallOpts{}, common.HexToAddress(routerContractAddress))
		if err != nil {
			a.logger.Errorf("EVM with networkId [%d] for asset [%s], and method BalanceOf - Error: [%s]", networkId, assetAddress, err)
			return nil, err
		}
	} else {
		// TODO: Remove the line below and uncomment the next one when we update the NFTs to extend ERC721Enumerable
		inLowestDenomination = big.NewInt(0)
		//inLowestDenomination, err = evmTokenClient.TotalSupply(&bind.CallOpts{})
		//if err != nil {
		//	a.logger.Errorf("EVM with networkId [%d] for asset [%s], and method TotalSupply - Error: [%s]", networkId, assetAddress, err)
		//	return nil, err
		//}
	}
	return inLowestDenomination, err
}

func (a *Service) FetchHederaTokenReserveAmount(
	assetId string,
	mirrorNode client.MirrorNode,
	isNative bool,
	hederaTokenBalances map[string]int,
) (reserveAmount *big.Int, err error) {

	if assetId == constants.Hbar {
		bridgeAccount, err := mirrorNode.GetAccount(a.bridgeAccountId)
		if err != nil {
			a.logger.Errorf("Hedera Mirror Node for Account ID [%s] method GetAccount - Error: [%s]", a.bridgeAccountId, err)
			return nil, err
		}

		return big.NewInt(int64(bridgeAccount.Balance.Balance)), nil
	}

	assetInfoResponse, err := mirrorNode.GetToken(assetId)
	if err != nil {
		a.logger.Errorf("Hedera Mirror Node method GetToken for Asset [%s] - Error: [%s]", assetId, err)
	} else {
		reserveAmount, err = a.getHederaTokenReserveAmount(assetId, isNative, hederaTokenBalances, assetInfoResponse)
	}

	return reserveAmount, err
}

func (a *Service) fetchEvmFungibleAssetInfo(
	networkId uint64,
	assetAddress string,
	evmTokenClient client.EvmFungibleToken,
	isNative bool,
	routerContractAddress string,
) (assetInfo *assetModel.FungibleAssetInfo, err error) {
	assetInfo = &assetModel.FungibleAssetInfo{}
	name, err := evmTokenClient.Name(&bind.CallOpts{})
	if err != nil {
		a.logger.Errorf("Failed to get Name for Asset [%s] for EVM with networkId [%d]  - Error: [%s]", assetAddress, networkId, err)
		return assetInfo, err
	}
	assetInfo.Name = name

	symbol, err := evmTokenClient.Symbol(&bind.CallOpts{})
	if err != nil {
		a.logger.Errorf("EVM with networkId [%d] for Asset [%s], and method Symbol - Error: [%s]", networkId, assetAddress, err)
		return assetInfo, err
	}
	assetInfo.Symbol = symbol

	decimals, err := evmTokenClient.Decimals(&bind.CallOpts{})
	if err != nil {
		a.logger.Errorf("EVM with networkId [%d] for Asset [%s], and method Decimals - Error: [%s]", networkId, assetAddress, err)
		return assetInfo, err
	}
	assetInfo.Decimals = decimals
	assetInfo.ReserveAmount, err = a.FetchEvmFungibleReserveAmount(networkId, assetAddress, isNative, evmTokenClient, routerContractAddress)

	return assetInfo, err
}

func (a *Service) fetchEvmNonFungibleAssetInfo(
	networkId uint64,
	assetAddress string,
	evmTokenClients map[uint64]map[string]client.EvmNft,
	isNative bool,
	routerContractAddress string,
) (assetInfo *assetModel.NonFungibleAssetInfo, err error) {
	assetInfo = &assetModel.NonFungibleAssetInfo{}
	evmTokenClient := evmTokenClients[networkId][assetAddress]
	name, err := evmTokenClient.Name(&bind.CallOpts{})
	if err != nil {
		a.logger.Errorf("Failed to get Name for Asset [%s] for EVM with networkId [%d]  - Error: [%s]", assetAddress, networkId, err)
		return assetInfo, err
	}
	assetInfo.Name = name

	symbol, err := evmTokenClient.Symbol(&bind.CallOpts{})
	if err != nil {
		a.logger.Errorf("EVM with networkId [%d] for Asset [%s], and method Symbol - Error: [%s]", networkId, assetAddress, err)
		return assetInfo, err
	}
	assetInfo.Symbol = symbol

	assetInfo.ReserveAmount, err = a.FetchEvmNonFungibleReserveAmount(networkId, assetAddress, isNative, evmTokenClient, routerContractAddress)

	return assetInfo, err
}

func (a *Service) fetchHederaFungibleAssetInfo(
	assetId string,
	mirrorNode client.MirrorNode,
	isNative bool,
	hederaTokenBalances map[string]int,
) (assetInfo *assetModel.FungibleAssetInfo, err error) {
	assetInfo = &assetModel.FungibleAssetInfo{}
	if assetId == constants.Hbar {
		assetInfo.Name = constants.Hbar
		assetInfo.Symbol = constants.Hbar
		assetInfo.Decimals = constants.HederaDefaultDecimals
		assetInfo.ReserveAmount = big.NewInt(int64(hederaTokenBalances[constants.Hbar]))

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
		assetInfo.ReserveAmount, err = a.getHederaTokenReserveAmount(assetId, isNative, hederaTokenBalances, assetInfoResponse)
	}

	return assetInfo, err
}

func (a *Service) loadFungibleAssetInfos(
	networks map[uint64]*parser.Network,
	mirrorNode client.MirrorNode,
	evmTokenClients map[uint64]map[string]client.EvmFungibleToken,
	hederaTokenBalances map[string]int,
) {
	a.fungibleAssetInfos = make(map[uint64]map[string]*assetModel.FungibleAssetInfo)

	for nativeChainId, networkInfo := range networks {
		if _, ok := a.fungibleAssetInfos[nativeChainId]; !ok {
			a.fungibleAssetInfos[nativeChainId] = make(map[string]*assetModel.FungibleAssetInfo)
		}

		for nativeAsset, nativeAssetMapping := range networkInfo.Tokens.Fungible {
			assetInfo, nativeAsset, err := a.fetchFungibleAssetInfo(nativeChainId, nativeAsset, mirrorNode, evmTokenClients, true, hederaTokenBalances, networks[nativeChainId].RouterContractAddress)
			if err != nil {
				a.logger.Fatal(err)
			}
			a.fungibleAssetInfos[nativeChainId][nativeAsset] = assetInfo

			for wrappedChainId, wrappedAsset := range nativeAssetMapping.Networks {
				if _, ok := a.fungibleAssetInfos[wrappedChainId]; !ok {
					a.fungibleAssetInfos[wrappedChainId] = make(map[string]*assetModel.FungibleAssetInfo)
				}
				assetInfo, wrappedAsset, err = a.fetchFungibleAssetInfo(wrappedChainId, wrappedAsset, mirrorNode, evmTokenClients, false, hederaTokenBalances, networks[wrappedChainId].RouterContractAddress)
				if err != nil {
					a.logger.Fatal(err)
				}
				a.fungibleAssetInfos[wrappedChainId][wrappedAsset] = assetInfo
			}
		}
	}
}

func (a *Service) fetchHederaNonFungibleAssetInfo(
	assetId string,
	mirrorNode client.MirrorNode,
	isNative bool,
	hederaTokenBalances map[string]int,
) (assetInfo *assetModel.NonFungibleAssetInfo, err error) {

	assetInfo = &assetModel.NonFungibleAssetInfo{}
	assetInfoResponse, e := mirrorNode.GetToken(assetId)
	if e != nil {
		a.logger.Errorf("Hedera Mirror Node method GetToken for Asset [%s] - Error: [%s]", assetId, e)
	} else {
		assetInfo.Name = assetInfoResponse.Name
		assetInfo.Symbol = assetInfoResponse.Symbol
		assetInfo.ReserveAmount, err = a.getHederaTokenReserveAmount(assetId, isNative, hederaTokenBalances, assetInfoResponse)
	}

	return assetInfo, err
}

func (a *Service) getHederaTokenReserveAmount(
	assetId string,
	isNative bool,
	hederaTokenBalances map[string]int,
	assetInfoResponse *token.TokenResponse,
) (*big.Int, error) {
	if isNative {
		return big.NewInt(int64(hederaTokenBalances[assetId])), nil
	}

	reserveAmount, ok := new(big.Int).SetString(assetInfoResponse.TotalSupply, 10)
	if !ok {
		err := errors.New(fmt.Sprintf(`"Hedera asset [%s] total supply SetString - Error": [%s].`, assetId, assetInfoResponse.TotalSupply))
		a.logger.Errorf(err.Error())
		return nil, err
	}

	return reserveAmount, nil
}

func (a *Service) fetchFungibleAssetInfo(
	chainId uint64,
	assetAddress string,
	mirrorNode client.MirrorNode,
	evmTokenClients map[uint64]map[string]client.EvmFungibleToken,
	isNative bool,
	hederaTokenBalances map[string]int,
	routerContractAddress string,
) (*assetModel.FungibleAssetInfo, string, error) {
	var (
		err       error
		assetInfo *assetModel.FungibleAssetInfo
	)

	if chainId == constants.HederaNetworkId { // Hedera
		assetInfo, err = a.fetchHederaFungibleAssetInfo(assetAddress, mirrorNode, isNative, hederaTokenBalances)
		if err != nil {
			err = errors.New(fmt.Sprintf("Failed to load Hedera Fungible Asset Info. Error [%v]", err))
			return assetInfo, assetAddress, err
		}
	} else { // EVM
		re := regexp.MustCompile(constants.EvmCompatibleAddressPattern)
		if isMatch := re.MatchString(assetAddress); isMatch {
			assetAddress = common.HexToAddress(assetAddress).String()
		}
		assetAddress = common.HexToAddress(assetAddress).String()
		evmTokenClient := evmTokenClients[chainId][assetAddress]
		assetInfo, err = a.fetchEvmFungibleAssetInfo(chainId, assetAddress, evmTokenClient, isNative, routerContractAddress)
		if err != nil {
			err = errors.New(fmt.Sprintf("Failed to load EVM NetworkId [%v] Fungible Asset Info. Error [%v]", chainId, err))
			return assetInfo, assetAddress, err
		}
	}

	assetInfo.IsNative = isNative

	return assetInfo, assetAddress, err
}

func (a *Service) loadNonFungibleAssetInfos(
	networks map[uint64]*parser.Network,
	mirrorNode client.MirrorNode,
	evmTokenClients map[uint64]map[string]client.EvmNft,
	hederaTokenBalances map[string]int,
) {
	a.nonFungibleAssetInfos = make(map[uint64]map[string]*assetModel.NonFungibleAssetInfo)

	for nativeChainId, networkInfo := range networks {
		if len(networkInfo.Tokens.Nft) == 0 {
			continue
		}

		if _, ok := a.nonFungibleAssetInfos[nativeChainId]; !ok {
			a.nonFungibleAssetInfos[nativeChainId] = make(map[string]*assetModel.NonFungibleAssetInfo)
		}

		for nativeAsset, nativeAssetMapping := range networkInfo.Tokens.Nft {
			assetInfo, nativeAsset, err := a.fetchNonFungibleAssetInfo(nativeChainId, nativeAsset, mirrorNode, evmTokenClients, true, hederaTokenBalances, networks[nativeChainId].RouterContractAddress)
			if err != nil {
				a.logger.Fatal(err)
			}
			a.nonFungibleAssetInfos[nativeChainId][nativeAsset] = assetInfo

			for wrappedChainId, wrappedAsset := range nativeAssetMapping.Networks {
				if _, ok := a.nonFungibleAssetInfos[wrappedChainId]; !ok {
					a.nonFungibleAssetInfos[wrappedChainId] = make(map[string]*assetModel.NonFungibleAssetInfo)
				}
				assetInfo, wrappedAsset, err = a.fetchNonFungibleAssetInfo(wrappedChainId, wrappedAsset, mirrorNode, evmTokenClients, false, hederaTokenBalances, networks[wrappedChainId].RouterContractAddress)
				if err != nil {
					a.logger.Fatal(err)
				}
				a.nonFungibleAssetInfos[wrappedChainId][wrappedAsset] = assetInfo
			}
		}
	}
}

func (a *Service) fetchNonFungibleAssetInfo(
	chainId uint64,
	assetAddress string,
	mirrorNode client.MirrorNode,
	evmTokenClients map[uint64]map[string]client.EvmNft,
	isNative bool,
	hederaTokenBalances map[string]int,
	routerContractAddress string,
) (*assetModel.NonFungibleAssetInfo, string, error) {
	var (
		err       error
		assetInfo *assetModel.NonFungibleAssetInfo
	)

	if chainId == constants.HederaNetworkId { // Hedera
		assetInfo, err = a.fetchHederaNonFungibleAssetInfo(assetAddress, mirrorNode, isNative, hederaTokenBalances)
		if err != nil {
			err = errors.New(fmt.Sprintf("Failed to load Hedera Non-Fungible Asset Info. Error [%v]", err))
			return assetInfo, assetAddress, err
		}
	} else { // EVM
		re := regexp.MustCompile(constants.EvmCompatibleAddressPattern)
		if isMatch := re.MatchString(assetAddress); isMatch {
			assetAddress = common.HexToAddress(assetAddress).String()
		}
		assetAddress = common.HexToAddress(assetAddress).String()
		assetInfo, err = a.fetchEvmNonFungibleAssetInfo(chainId, assetAddress, evmTokenClients, isNative, routerContractAddress)
		if err != nil {
			err = errors.New(fmt.Sprintf("Failed to load EVM NetworkId [%v] Non-Fungible Asset Info. Error [%v]", chainId, err))
			return assetInfo, assetAddress, err
		}
	}
	assetInfo.IsNative = isNative

	return assetInfo, assetAddress, err
}

func NewService(
	networks map[uint64]*parser.Network,
	bridgeAccountId string,
	hederaFeePercentages map[string]int64,
	routerClients map[uint64]client.DiamondRouter,
	mirrorNode client.MirrorNode,
	evmTokenClients map[uint64]map[string]client.EvmFungibleToken,
	evmNftClients map[uint64]map[string]client.EvmNft,
) *Service {
	instance := initialize(
		networks,
		bridgeAccountId,
		hederaFeePercentages,
		routerClients,
		mirrorNode,
		evmTokenClients,
		evmNftClients,
	)

	event.On(constants.EventBridgeConfigUpdate, event.ListenerFunc(func(e event.Event) error {
		return bridgeCfgUpdateEventHandler(e, mirrorNode, routerClients, instance)
	}), constants.AssetServicePriority)

	return instance
}

func initialize(networks map[uint64]*parser.Network, bridgeAccountId string, HederaFeePercentages map[string]int64, routerClients map[uint64]client.DiamondRouter, mirrorNode client.MirrorNode, evmTokenClients map[uint64]map[string]client.EvmFungibleToken, evmNftClients map[uint64]map[string]client.EvmNft) *Service {
	nativeToWrapped := make(map[uint64]map[string]map[uint64]string)
	wrappedToNative := make(map[uint64]map[string]*assetModel.NativeAsset)
	fungibleNetworkAssets := make(map[uint64][]string)
	nonFungibleNetworkAssets := make(map[uint64][]string)
	fungibleNativeAssets := make(map[uint64]map[string]*assetModel.NativeAsset)

	re := regexp.MustCompile(constants.EvmCompatibleAddressPattern)

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

			minAmount, err := decimalHelper.ParseAmount(nativeAssetMapping.MinFeeAmountInUsd)
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

			nonFungibleNetworkAssets[nativeChainId] = append(nonFungibleNetworkAssets[nativeChainId], nativeAsset)

			for wrappedChainId, wrappedAsset := range nativeAssetMapping.Networks {
				if isMatch := re.MatchString(wrappedAsset); isMatch {
					wrappedAsset = common.HexToAddress(wrappedAsset).String()
				}

				nativeToWrapped[nativeChainId][nativeAsset][wrappedChainId] = wrappedAsset
				if wrappedToNative[wrappedChainId] == nil {
					wrappedToNative[wrappedChainId] = make(map[string]*assetModel.NativeAsset)
				}

				nonFungibleNetworkAssets[wrappedChainId] = append(nonFungibleNetworkAssets[wrappedChainId], wrappedAsset)
				wrappedToNative[wrappedChainId][wrappedAsset] = &assetModel.NativeAsset{
					ChainId: nativeChainId,
					Asset:   nativeAsset,
				}
			}
		}
	}
	logger := config.GetLoggerFor("Assets Service")

	instance := &Service{
		nativeToWrapped:          nativeToWrapped,
		wrappedToNative:          wrappedToNative,
		fungibleNativeAssets:     fungibleNativeAssets,
		fungibleNetworkAssets:    fungibleNetworkAssets,
		nonFungibleNetworkAssets: nonFungibleNetworkAssets,
		bridgeAccountId:          bridgeAccountId,
		logger:                   logger,
	}

	bridgeAccount, e := mirrorNode.GetAccount(bridgeAccountId)
	if e != nil {
		logger.Fatalf("Hedera Mirror Node for Account ID [%s] method GetAccount - Error: [%s]", bridgeAccountId, e)
		return nil
	}
	hederaTokenBalances := bridgeAccount.Balance.GetAccountTokenBalancesByAddress()
	instance.loadFungibleAssetInfos(networks, mirrorNode, evmTokenClients, hederaTokenBalances)
	instance.loadNonFungibleAssetInfos(networks, mirrorNode, evmNftClients, hederaTokenBalances)

	return instance
}

func bridgeCfgUpdateEventHandler(e event.Event, mirrorNode client.MirrorNode, routerClients map[uint64]client.DiamondRouter, instance *Service) error {
	params, ok := e.Get(constants.BridgeConfigUpdateEventParamsKey).(*bridge_config_event.Params)
	if !ok {
		errMsg := fmt.Sprintf("failed to cast params from event [%s]", constants.EventBridgeConfigUpdate)
		log.Errorf(errMsg)
		return errors.New(errMsg)
	}

	newInstance := initialize(
		params.ParsedBridge.Networks,
		params.ParsedBridge.Networks[constants.HederaNetworkId].BridgeAccount,
		params.Bridge.Hedera.FeePercentages,
		routerClients,
		mirrorNode,
		params.EvmFungibleTokenClients,
		params.EvmNFTClients,
	)
	copyFields(newInstance, instance)
	params.Bridge.LoadStaticMinAmountsForWrappedFungibleTokens(*params.ParsedBridge, instance)

	return nil
}

func copyFields(from *Service, to *Service) {
	to.nativeToWrapped = from.nativeToWrapped
	to.wrappedToNative = from.wrappedToNative
	to.fungibleNetworkAssets = from.fungibleNetworkAssets
	to.fungibleNativeAssets = from.fungibleNativeAssets
	to.fungibleAssetInfos = from.fungibleAssetInfos
	to.nonFungibleNetworkAssets = from.nonFungibleNetworkAssets
	to.nonFungibleAssetInfos = from.nonFungibleAssetInfos
	to.bridgeAccountId = from.bridgeAccountId
}

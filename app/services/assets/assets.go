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
	"github.com/shopspring/decimal"
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

	bridgeAccountId string
	logger          *log.Entry
}

func (a *Service) FungibleNetworkAssets() map[uint64][]string {
	return a.fungibleNetworkAssets
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

func (a *Service) FetchEvmRegularReserveAmount(
	networkId uint64,
	assetAddress string,
	isNative bool,
	evmTokenClient client.EvmRegularToken,
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
	evmTokenClient client.EvmRegularToken,
	isNative bool,
	routerContractAddress string,
) (assetInfo *assetModel.FungibleAssetInfo, err error) {
	if evmTokenClient == nil {
		a.logger.Fatalf("Evm Token Client is missing for network [%d] and asset [%s]", networkId, assetAddress)
	}
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
	assetInfo.ReserveAmount, err = a.FetchEvmRegularReserveAmount(networkId, assetAddress, isNative, evmTokenClient, routerContractAddress)

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

	assetInfoResponse, err := mirrorNode.GetToken(assetId)
	if err != nil {
		a.logger.Errorf("Hedera Mirror Node method GetToken for Asset [%s] - Error: [%s]", assetId, err)
		return nil, err
	}

	assetInfo.Name = assetInfoResponse.Name
	assetInfo.Symbol = assetInfoResponse.Symbol
	parsedDecimals, _ := strconv.Atoi(assetInfoResponse.Decimals)
	assetInfo.Decimals = uint8(parsedDecimals)
	assetInfo.ReserveAmount, _ = a.getHederaTokenReserveAmount(assetId, isNative, hederaTokenBalances, assetInfoResponse)

	return assetInfo, nil
}

func (a *Service) loadFungibleAssetInfos(
	regularTokens map[string]*parser.RegularToken,
	evmNetworks map[uint64]*parser.EVMNetwork,
	mirrorNode client.MirrorNode,
	evmTokenClients map[uint64]map[string]client.EvmRegularToken,
	hederaTokenBalances map[string]int,
) {
	a.fungibleAssetInfos = make(map[uint64]map[string]*assetModel.FungibleAssetInfo)

	for tokenName, tokenInfo := range regularTokens {
		nativeChainId := tokenInfo.NativeChain
		nativeAsset := tokenName
		if tokenInfo.Address != nil && *tokenInfo.Address != "" {
			nativeAsset = *tokenInfo.Address
		}

		if _, ok := a.fungibleAssetInfos[nativeChainId]; !ok {
			a.fungibleAssetInfos[nativeChainId] = make(map[string]*assetModel.FungibleAssetInfo)
		}

		assetInfo, nativeAsset, err := a.fetchFungibleAssetInfo(nativeChainId, nativeAsset, mirrorNode, evmTokenClients, true, hederaTokenBalances, a.evmRouterAddr(evmNetworks, nativeChainId))
		if err != nil {
			a.logger.Fatal(err)
		}
		a.fungibleAssetInfos[nativeChainId][nativeAsset] = assetInfo

		for wrappedChainId, wrappedAsset := range tokenInfo.AddressesPerNetwork {
			if _, ok := a.fungibleAssetInfos[wrappedChainId]; !ok {
				a.fungibleAssetInfos[wrappedChainId] = make(map[string]*assetModel.FungibleAssetInfo)
			}
			var wrappedInfo *assetModel.FungibleAssetInfo
			wrappedInfo, wrappedAsset, err = a.fetchFungibleAssetInfo(wrappedChainId, wrappedAsset, mirrorNode, evmTokenClients, false, hederaTokenBalances, a.evmRouterAddr(evmNetworks, wrappedChainId))
			if err != nil {
				a.logger.Fatal(err)
			}
			a.fungibleAssetInfos[wrappedChainId][wrappedAsset] = wrappedInfo
		}
	}
}

func (a *Service) evmRouterAddr(evmNetworks map[uint64]*parser.EVMNetwork, chainId uint64) string {
	if net, ok := evmNetworks[chainId]; ok {
		return net.RouterContractAddress
	}
	return ""
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
		err := fmt.Errorf(`"Hedera asset [%s] total supply SetString - Error": [%s].`, assetId, assetInfoResponse.TotalSupply)
		a.logger.Errorf(err.Error())
		return nil, err
	}

	return reserveAmount, nil
}

func (a *Service) fetchFungibleAssetInfo(
	chainId uint64,
	assetAddress string,
	mirrorNode client.MirrorNode,
	evmTokenClients map[uint64]map[string]client.EvmRegularToken,
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
			err = fmt.Errorf("Failed to load Hedera Fungible Asset Info. Error [%v]", err)
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
			err = fmt.Errorf("Failed to load EVM NetworkId [%v] Fungible Asset Info. Error [%v]", chainId, err)
			return assetInfo, assetAddress, err
		}
	}

	assetInfo.IsNative = isNative

	return assetInfo, assetAddress, err
}

func NewService(
	regularTokens map[string]*parser.RegularToken,
	evmNetworks map[uint64]*parser.EVMNetwork,
	bridgeAccountId string,
	hederaFeePercentages map[string]int64,
	routerClients map[uint64]client.DiamondRouter,
	mirrorNode client.MirrorNode,
	evmTokenClients map[uint64]map[string]client.EvmRegularToken,
) *Service {
	instance := initialize(
		regularTokens,
		evmNetworks,
		bridgeAccountId,
		hederaFeePercentages,
		routerClients,
		mirrorNode,
		evmTokenClients,
	)

	event.On(constants.EventBridgeConfigUpdate, event.ListenerFunc(func(e event.Event) error {
		return bridgeCfgUpdateEventHandler(e, mirrorNode, instance)
	}), constants.AssetServicePriority)

	return instance
}

func initialize(regularTokens map[string]*parser.RegularToken, evmNetworks map[uint64]*parser.EVMNetwork, bridgeAccountId string, HederaFeePercentages map[string]int64, routerClients map[uint64]client.DiamondRouter, mirrorNode client.MirrorNode, evmTokenClients map[uint64]map[string]client.EvmRegularToken) *Service {
	nativeToWrapped := make(map[uint64]map[string]map[uint64]string)
	wrappedToNative := make(map[uint64]map[string]*assetModel.NativeAsset)
	fungibleNetworkAssets := make(map[uint64][]string)
	fungibleNativeAssets := make(map[uint64]map[string]*assetModel.NativeAsset)

	re := regexp.MustCompile(constants.EvmCompatibleAddressPattern)

	for tokenName, tokenInfo := range regularTokens {
		nativeChainId := tokenInfo.NativeChain
		nativeAsset := tokenName
		if tokenInfo.Address != nil && *tokenInfo.Address != "" {
			nativeAsset = *tokenInfo.Address
		}
		if nativeChainId != constants.HederaNetworkId {
			nativeAsset = common.HexToAddress(nativeAsset).String()
		}

		if nativeToWrapped[nativeChainId] == nil {
			nativeToWrapped[nativeChainId] = make(map[string]map[uint64]string)
		}
		if fungibleNativeAssets[nativeChainId] == nil {
			fungibleNativeAssets[nativeChainId] = make(map[string]*assetModel.NativeAsset)
		}
		if nativeToWrapped[nativeChainId][nativeAsset] == nil {
			nativeToWrapped[nativeChainId][nativeAsset] = make(map[uint64]string)
		}

		var minAmount *decimal.Decimal
		if tokenInfo.MinFeeAmountInUsd != nil {
			minAmount = tokenInfo.MinFeeAmountInUsd
		} else {
			var err error
			minAmount, err = decimalHelper.ParseAmount("")
			if err != nil {
				log.Fatalf("Failed to parse min amount. Error: [%s]", err)
			}
		}
		var feePercentage int64
		if nativeChainId == constants.HederaNetworkId {
			feePercentage = HederaFeePercentages[nativeAsset]
		} else {
			routerClient, exist := routerClients[nativeChainId]
			if exist {
				tokenFeeData, err := routerClient.TokenFeeData(&bind.CallOpts{}, common.HexToAddress(nativeAsset))
				if err != nil {
					log.Fatalf("Failed to get fee percentage from router contract for asset [%s]. Error: [%s]", nativeAsset, err)
				}
				feePercentage = tokenFeeData.ServiceFeePercentage.Int64()
			}
		}

		asset := &assetModel.NativeAsset{
			MinFeeAmountInUsd: minAmount,
			ChainId:           nativeChainId,
			Asset:             nativeAsset,
			FeePercentage:     feePercentage,
			ReleaseTimestamp:  tokenInfo.ReleaseTimestamp,
		}
		fungibleNativeAssets[nativeChainId][nativeAsset] = asset

		fungibleNetworkAssets[nativeChainId] = append(fungibleNetworkAssets[nativeChainId], nativeAsset)
		for wrappedChainId, wrappedAsset := range tokenInfo.AddressesPerNetwork {
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
	logger := config.GetLoggerFor("Assets Service")

	instance := &Service{
		nativeToWrapped:          nativeToWrapped,
		wrappedToNative:          wrappedToNative,
		fungibleNativeAssets:     fungibleNativeAssets,
		fungibleNetworkAssets:    fungibleNetworkAssets,
		bridgeAccountId:          bridgeAccountId,
		logger:                   logger,
	}

	bridgeAccount, e := mirrorNode.GetAccount(bridgeAccountId)
	if e != nil {
		logger.Fatalf("Hedera Mirror Node for Account ID [%s] method GetAccount - Error: [%s]", bridgeAccountId, e)
		return nil
	}
	hederaTokenBalances := bridgeAccount.Balance.GetAccountTokenBalancesByAddress()
	instance.loadFungibleAssetInfos(regularTokens, evmNetworks, mirrorNode, evmTokenClients, hederaTokenBalances)

	return instance
}

func bridgeCfgUpdateEventHandler(e event.Event, mirrorNode client.MirrorNode, instance *Service) error {
	params, ok := e.Get(constants.BridgeConfigUpdateEventParamsKey).(*bridge_config_event.Params)
	if !ok {
		errMsg := fmt.Sprintf("failed to cast params from event [%s]", constants.EventBridgeConfigUpdate)
		log.Errorf(errMsg)
		return errors.New(errMsg)
	}

	newInstance := initialize(
		params.ParsedBridge.RegularTokens,
		params.ParsedBridge.Networks.EVM,
		params.Bridge.Hedera.BridgeAccount,
		params.Bridge.Hedera.FeePercentages,
		params.RouterClients,
		mirrorNode,
		params.EvmRegularTokenClients,
	)
	*instance = *newInstance
	params.Bridge.LoadStaticMinAmountsForWrappedFungibleTokens(*params.ParsedBridge, instance)

	return nil
}

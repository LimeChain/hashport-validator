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
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/account"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/token"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	testConstants "github.com/limechain/hedera-eth-bridge-validator/test/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/helper"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	testClient "github.com/limechain/hedera-eth-bridge-validator/test/mocks/client"
	"github.com/stretchr/testify/assert"
	"math/big"
	"sort"
	"strconv"
	"testing"
)

var (
	serviceInstance         *Service
	routerClients           = make(map[uint64]client.DiamondRouter)
	evmClients              = make(map[uint64]client.EVM)
	evmCoreClients          = make(map[uint64]client.Core)
	evmFungibleTokenClients = make(map[uint64]map[string]client.EvmFungibleToken)
	evmNftClients           = make(map[uint64]map[string]client.EvmNft)
	hederaPercentages       = make(map[string]int64)
	hederaAccount           = account.AccountsResponse{
		Account: testConstants.BridgeAccountId,
		Balance: account.Balance{
			Balance:   int(testConstants.ReserveAmount),
			Timestamp: "",
			Tokens:    []account.AccountToken{},
		},
	}
	hederaTokenBalances = make(map[string]int)
)

func Test_New(t *testing.T) {
	setup()
	setupClientMocks()

	actualService := NewService(testConstants.ParserBridge.Networks, testConstants.ParserBridge.Networks[0].BridgeAccount, hederaPercentages, routerClients, mocks.MHederaMirrorClient, evmFungibleTokenClients, evmNftClients)

	assert.Equal(t, serviceInstance.nativeToWrapped, actualService.nativeToWrapped)
	assert.Equal(t, serviceInstance.wrappedToNative, actualService.wrappedToNative)
	assert.Equal(t, serviceInstance.fungibleNativeAssets, actualService.fungibleNativeAssets)
	assert.Equal(t, serviceInstance.fungibleAssetInfos, actualService.fungibleAssetInfos)
	assert.Equal(t, serviceInstance.nonFungibleAssetInfos, actualService.nonFungibleAssetInfos)

	for networkId := range testConstants.Networks {
		// Fungible
		sort.Strings(serviceInstance.fungibleNetworkAssets[networkId])
		sort.Strings(actualService.fungibleNetworkAssets[networkId])
		assert.Equal(t, serviceInstance.fungibleNetworkAssets[networkId], actualService.fungibleNetworkAssets[networkId])

		// Non-Fungible
		sort.Strings(serviceInstance.nonFungibleNetworkAssets[networkId])
		sort.Strings(actualService.nonFungibleNetworkAssets[networkId])
		assert.Equal(t, serviceInstance.nonFungibleNetworkAssets[networkId], actualService.nonFungibleNetworkAssets[networkId])
	}

}

func Test_IsNative(t *testing.T) {
	setup()

	actual := serviceInstance.IsNative(0, constants.Hbar)
	assert.Equal(t, true, actual)

	actual = serviceInstance.IsNative(0, testConstants.NetworkHederaFungibleWrappedTokenForNetworkPolygon)
	assert.Equal(t, false, actual)
}

func Test_OppositeAsset(t *testing.T) {
	setup()

	actual := serviceInstance.OppositeAsset(testConstants.PolygonNetworkId, constants.HederaNetworkId, testConstants.NetworkPolygonFungibleWrappedTokenForNetworkHedera)
	expected := constants.Hbar

	assert.Equal(t, expected, actual)

	actual = serviceInstance.OppositeAsset(constants.HederaNetworkId, testConstants.PolygonNetworkId, constants.Hbar)
	expected = testConstants.NetworkPolygonFungibleWrappedTokenForNetworkHedera

	assert.Equal(t, expected, actual)

	actual = serviceInstance.OppositeAsset(testConstants.PolygonNetworkId, constants.HederaNetworkId, constants.Hbar)
	expected = testConstants.NetworkPolygonFungibleWrappedTokenForNetworkHedera

	assert.Equal(t, expected, actual)

	actual = serviceInstance.OppositeAsset(constants.HederaNetworkId, testConstants.PolygonNetworkId, testConstants.NetworkPolygonFungibleWrappedTokenForNetworkHedera)
	expected = constants.Hbar

	assert.Equal(t, expected, actual)

}

func Test_NativeToWrapped(t *testing.T) {
	setup()

	actual := serviceInstance.NativeToWrapped(constants.Hbar, constants.HederaNetworkId, testConstants.PolygonNetworkId)
	expected := testConstants.NetworkPolygonFungibleWrappedTokenForNetworkHedera

	assert.Equal(t, expected, actual)
}

func Test_WrappedToNative(t *testing.T) {
	setup()

	actual := serviceInstance.WrappedToNative(testConstants.NetworkPolygonFungibleWrappedTokenForNetworkHedera, testConstants.PolygonNetworkId)
	expected := constants.Hbar

	assert.NotNil(t, actual)
	assert.Equal(t, expected, actual.Asset)
}

func Test_FungibleNetworkAssets(t *testing.T) {
	setup()

	actual := serviceInstance.FungibleNetworkAssets()
	expected := testConstants.FungibleNetworkAssets

	assert.NotNil(t, actual)
	assert.Equal(t, expected, actual)
}

func Test_NonFungibleNetworkAssets(t *testing.T) {
	setup()

	actual := serviceInstance.NonFungibleNetworkAssets()
	expected := testConstants.NonFungibleNetworkAssets

	assert.NotNil(t, actual)
	assert.Equal(t, expected, actual)
}

func Test_NativeToWrappedAssets(t *testing.T) {
	setup()

	actual := serviceInstance.NativeToWrappedAssets()
	expected := testConstants.NativeToWrapped

	assert.NotNil(t, actual)
	assert.Equal(t, expected, actual)
}

func Test_WrappedFromNative(t *testing.T) {
	setup()

	actual := serviceInstance.WrappedFromNative(constants.HederaNetworkId, testConstants.NetworkHederaFungibleNativeToken)
	expected := testConstants.NativeToWrapped[constants.HederaNetworkId][testConstants.NetworkHederaFungibleNativeToken]

	assert.NotNil(t, actual)
	assert.Equal(t, expected, actual)
}

func Test_FungibleNetworkAssetsByChainId(t *testing.T) {
	setup()

	actual := serviceInstance.FungibleNetworkAssetsByChainId(constants.HederaNetworkId)
	expected := testConstants.FungibleNetworkAssets[constants.HederaNetworkId]

	assert.NotNil(t, actual)
	assert.Equal(t, expected, actual)
}

func Test_FungibleNativeAsset(t *testing.T) {
	setup()

	actual := serviceInstance.FungibleNativeAsset(constants.HederaNetworkId, constants.Hbar)
	expected := testConstants.NetworkHederaFungibleNativeAsset

	assert.NotNil(t, actual)
	assert.Equal(t, expected, actual)
}

func Test_FungibleAssetInfo(t *testing.T) {
	setup()

	actual, exists := serviceInstance.FungibleAssetInfo(constants.HederaNetworkId, constants.Hbar)
	expected := testConstants.NetworkHederaFungibleNativeTokenFungibleAssetInfo

	assert.NotNil(t, actual)
	assert.True(t, exists)
	assert.Equal(t, expected, actual)
}

func Test_NonFungibleAssetInfo(t *testing.T) {
	setup()

	actual, exists := serviceInstance.NonFungibleAssetInfo(constants.HederaNetworkId, testConstants.NetworkHederaNonFungibleNativeToken)
	expected := testConstants.NetworkHederaNonFungibleNativeTokenNonFungibleAssetInfo

	assert.NotNil(t, actual)
	assert.True(t, exists)
	assert.Equal(t, expected, actual)
}

func Test_FetchEvmFungibleReserveAmount_Native(t *testing.T) {
	setup()
	setupClientMocks()

	routerContractAddress := "router"
	asset := testConstants.NetworkEthereumFungibleNativeToken
	tokenClient := evmFungibleTokenClients[testConstants.EthereumNetworkId][asset]
	expectedReserveAmount := testConstants.ReserveAmountBigInt
	tokenClient.(*testClient.MockEvmFungibleToken).On("BalanceOf", &bind.CallOpts{}, common.HexToAddress(routerContractAddress)).Return(expectedReserveAmount, nil)

	actual, err := serviceInstance.FetchEvmFungibleReserveAmount(testConstants.EthereumNetworkId, asset, true, tokenClient, routerContractAddress)

	assert.Nil(t, err)
	assert.Equal(t, expectedReserveAmount, actual)
}

func Test_FetchEvmFungibleReserveAmount_Wrapped(t *testing.T) {
	setup()
	setupClientMocks()

	routerContractAddress := "router"
	asset := testConstants.NetworkEthereumFungibleWrappedTokenForNetworkHedera
	tokenClient := evmFungibleTokenClients[testConstants.EthereumNetworkId][asset]
	expectedReserveAmount := testConstants.ReserveAmountBigInt
	tokenClient.(*testClient.MockEvmFungibleToken).On("TotalSupply", &bind.CallOpts{}).Return(expectedReserveAmount, nil)

	actual, err := serviceInstance.FetchEvmFungibleReserveAmount(testConstants.EthereumNetworkId, asset, false, tokenClient, routerContractAddress)

	assert.Nil(t, err)
	assert.Equal(t, expectedReserveAmount, actual)
}

func Test_FetchEvmNonFungibleReserveAmount_Native(t *testing.T) {
	setup()
	setupClientMocks()

	routerContractAddress := "router"
	asset := testConstants.NetworkPolygonWrappedNonFungibleTokenForHedera
	tokenClient := evmNftClients[testConstants.PolygonNetworkId][asset]
	expectedReserveAmount := testConstants.ReserveAmountBigInt
	tokenClient.(*testClient.MockEvmNonFungibleToken).On("BalanceOf", &bind.CallOpts{}, common.HexToAddress(routerContractAddress)).Return(expectedReserveAmount, nil)

	actual, err := serviceInstance.FetchEvmNonFungibleReserveAmount(testConstants.PolygonNetworkId, asset, true, tokenClient, routerContractAddress)

	assert.Nil(t, err)
	assert.Equal(t, expectedReserveAmount, actual)
}

func Test_FetchEvmNonFungibleReserveAmount_Wrapped(t *testing.T) {
	setup()
	setupClientMocks()

	routerContractAddress := "router"
	asset := testConstants.NetworkPolygonWrappedNonFungibleTokenForHedera
	tokenClient := evmNftClients[testConstants.PolygonNetworkId][asset]
	expectedReserveAmount := testConstants.ReserveAmountWrappedNFTBigInt

	actual, err := serviceInstance.FetchEvmNonFungibleReserveAmount(testConstants.PolygonNetworkId, asset, false, tokenClient, routerContractAddress)

	assert.Nil(t, err)
	assert.Equal(t, expectedReserveAmount, actual)
}

func Test_FetchHederaTokenReserveAmount_Native(t *testing.T) {
	setup()
	setupClientMocks()

	asset := constants.Hbar
	expectedReserveAmount := testConstants.ReserveAmountBigInt
	actual, err := serviceInstance.FetchHederaTokenReserveAmount(asset, mocks.MHederaMirrorClient, true, hederaTokenBalances)

	assert.Nil(t, err)
	assert.Equal(t, expectedReserveAmount, actual)
}

func Test_FetchHederaTokenReserveAmount_Wrapped(t *testing.T) {
	setup()
	setupClientMocks()

	asset := constants.Hbar
	expectedReserveAmount := testConstants.ReserveAmountBigInt
	mocks.MHederaMirrorClient.On("GetToken", asset).Return(&token.TokenResponse{
		TotalSupply: expectedReserveAmount.String(),
	})
	actual, err := serviceInstance.FetchHederaTokenReserveAmount(asset, mocks.MHederaMirrorClient, false, hederaTokenBalances)

	assert.Nil(t, err)
	assert.Equal(t, expectedReserveAmount, actual)
}

func setup() {
	mocks.Setup()
	helper.SetupNetworks()

	serviceInstance = &Service{
		nativeToWrapped:          testConstants.NativeToWrapped,
		wrappedToNative:          testConstants.WrappedToNative,
		fungibleNativeAssets:     testConstants.FungibleNativeAssets,
		fungibleNetworkAssets:    testConstants.FungibleNetworkAssets,
		fungibleAssetInfos:       testConstants.FungibleAssetInfos,
		nonFungibleNetworkAssets: testConstants.NonFungibleNetworkAssets,
		nonFungibleAssetInfos:    testConstants.NonFungibleAssetInfos,
		bridgeAccountId:          testConstants.BridgeAccountId,
		logger:                   config.GetLoggerFor("Assets Service"),
	}
}

func setupClientMocks() {
	for networkId, networkInfo := range testConstants.Networks {
		if networkId != constants.HederaNetworkId {
			evmFungibleTokenClients[networkId] = make(map[string]client.EvmFungibleToken)
			evmNftClients[networkId] = make(map[string]client.EvmNft)
			evmClients[networkId] = &testClient.MockEVM{}
			evmCoreClients[networkId] = &testClient.MockEVMCore{}
			evmClients[networkId].(*testClient.MockEVM).On("GetClient").Return(evmCoreClients[networkId])
			routerClients[networkId] = new(testClient.MockDiamondRouter)
		}

		// FUNGIBLE //
		fungibleAssets := testConstants.FungibleNetworkAssets[networkId]
		for _, asset := range fungibleAssets {
			hederaPercentages[asset] = testConstants.FeePercentage
			assetInfo := testConstants.FungibleAssetInfos[networkId][asset]
			isNative := assetInfo.IsNative
			if networkId == constants.HederaNetworkId {
				hederaTokenBalances[asset] = int(testConstants.ReserveAmount)
				hederaAccount.Balance.Tokens = append(hederaAccount.Balance.Tokens, account.AccountToken{
					TokenID: asset,
					Balance: int(testConstants.ReserveAmount),
				})
				// Hedera
				tokenResponse := token.TokenResponse{
					TokenID:     asset,
					Name:        asset,
					Symbol:      asset,
					TotalSupply: testConstants.ReserveAmountStr,
					Decimals:    strconv.Itoa(int(constants.HederaDefaultDecimals)),
				}
				mocks.MHederaMirrorClient.On("GetToken", asset).Return(&tokenResponse, nil)
				continue
			}

			// EVM
			evmFungibleTokenClients[networkId][asset] = new(testClient.MockEvmFungibleToken)
			evmFungibleTokenClients[networkId][asset].(*testClient.MockEvmFungibleToken).On("Name", &bind.CallOpts{}).Return(asset, nil)
			evmFungibleTokenClients[networkId][asset].(*testClient.MockEvmFungibleToken).On("Symbol", &bind.CallOpts{}).Return(asset, nil)
			evmFungibleTokenClients[networkId][asset].(*testClient.MockEvmFungibleToken).On("Decimals", &bind.CallOpts{}).Return(constants.EvmDefaultDecimals, nil)
			if isNative {
				evmFungibleTokenClients[networkId][asset].(*testClient.MockEvmFungibleToken).On("BalanceOf", &bind.CallOpts{}, common.HexToAddress(networkInfo.RouterContractAddress)).Return(testConstants.ReserveAmountBigInt, nil)
			} else {
				evmFungibleTokenClients[networkId][asset].(*testClient.MockEvmFungibleToken).On("TotalSupply", &bind.CallOpts{}).Return(testConstants.ReserveAmountBigInt, nil)
			}
			tokenFeeDataResult := struct {
				ServiceFeePercentage *big.Int
				FeesAccrued          *big.Int
				PreviousAccrued      *big.Int
				Accumulator          *big.Int
			}{
				ServiceFeePercentage: big.NewInt(testConstants.FeePercentage),
				FeesAccrued:          big.NewInt(0),
				PreviousAccrued:      big.NewInt(0),
				Accumulator:          big.NewInt(0),
			}
			routerClients[networkId].(*testClient.MockDiamondRouter).On("TokenFeeData", &bind.CallOpts{}, common.HexToAddress(asset)).Return(tokenFeeDataResult, nil)
		}

		// NON-FUNGIBLE //
		nonFungibleAssets := testConstants.NonFungibleNetworkAssets[networkId]
		for _, asset := range nonFungibleAssets {
			assetInfo := testConstants.NonFungibleAssetInfos[networkId][asset]
			isNative := assetInfo.IsNative
			hederaPercentages[asset] = testConstants.FeePercentage
			if networkId == constants.HederaNetworkId {
				hederaTokenBalances[asset] = int(testConstants.ReserveAmount)
				hederaAccount.Balance.Tokens = append(hederaAccount.Balance.Tokens, account.AccountToken{
					TokenID: asset,
					Balance: int(testConstants.ReserveAmount),
				})

				// Hedera
				tokenResponse := token.TokenResponse{
					TokenID:     asset,
					Name:        asset,
					Symbol:      asset,
					TotalSupply: testConstants.ReserveAmountStr,
					Decimals:    "0",
				}
				mocks.MHederaMirrorClient.On("GetToken", asset).Return(&tokenResponse, nil)
				continue
			}

			// EVM
			evmNftClients[networkId][asset] = new(testClient.MockEvmNonFungibleToken)
			evmNftClients[networkId][asset].(*testClient.MockEvmNonFungibleToken).On("Name", &bind.CallOpts{}).Return(asset, nil)
			evmNftClients[networkId][asset].(*testClient.MockEvmNonFungibleToken).On("Symbol", &bind.CallOpts{}).Return(asset, nil)
			tokenFeeDataResult := struct {
				ServiceFeePercentage *big.Int
				FeesAccrued          *big.Int
				PreviousAccrued      *big.Int
				Accumulator          *big.Int
			}{big.NewInt(testConstants.FeePercentage), big.NewInt(0), big.NewInt(0), big.NewInt(0)}
			routerClients[networkId].(*testClient.MockDiamondRouter).On("TokenFeeData", &bind.CallOpts{}, common.HexToAddress(asset)).Return(tokenFeeDataResult, nil)
			if isNative {
				evmNftClients[networkId][asset].(*testClient.MockEvmNonFungibleToken).On("BalanceOf", &bind.CallOpts{}, common.HexToAddress(networkInfo.RouterContractAddress)).Return(testConstants.ReserveAmountBigInt, nil)
			} /* else { // TODO: Uncomment the line below when we update the NFTs to extend ERC721Enumerable
				evmNftClients[networkId][asset].(*testClient.MockEvmFungibleToken).On("TotalSupply", &bind.CallOpts{}).Return(testConstants.ReserveAmountBigInt, nil)
			} */
		}

	}

	mocks.MHederaMirrorClient.On("GetAccount", testConstants.BridgeAccountId).Return(&hederaAccount, nil)
}

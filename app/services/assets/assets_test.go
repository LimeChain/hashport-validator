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
	evmNFTClients           = make(map[uint64]map[string]client.EvmNft)
	hederaPercentages       = make(map[string]int64)
)

func Test_New(t *testing.T) {
	setup(true)

	actualService := NewService(testConstants.ParserBridge.Networks, hederaPercentages, routerClients, mocks.MHederaMirrorClient, evmFungibleTokenClients, evmNFTClients)

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
	setup(false)

	actual := serviceInstance.IsNative(0, constants.Hbar)
	assert.Equal(t, true, actual)

	actual = serviceInstance.IsNative(0, testConstants.NetworkHederaFungibleWrappedTokenForNetworkPolygon)
	assert.Equal(t, false, actual)
}

func Test_OppositeAsset(t *testing.T) {
	setup(false)

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
	setup(false)

	actual := serviceInstance.NativeToWrapped(constants.Hbar, constants.HederaNetworkId, testConstants.PolygonNetworkId)
	expected := testConstants.NetworkPolygonFungibleWrappedTokenForNetworkHedera

	assert.Equal(t, expected, actual)
}

func Test_WrappedToNative(t *testing.T) {
	setup(false)

	actual := serviceInstance.WrappedToNative(testConstants.NetworkPolygonFungibleWrappedTokenForNetworkHedera, testConstants.PolygonNetworkId)
	expected := constants.Hbar

	assert.NotNil(t, actual)
	assert.Equal(t, expected, actual.Asset)
}

func Test_FungibleNetworkAssets(t *testing.T) {
	setup(false)

	actual := serviceInstance.FungibleNetworkAssets()
	expected := testConstants.FungibleNetworkAssets

	assert.NotNil(t, actual)
	assert.Equal(t, expected, actual)
}

func Test_NonFungibleNetworkAssets(t *testing.T) {
	setup(false)

	actual := serviceInstance.NonFungibleNetworkAssets()
	expected := testConstants.NonFungibleNetworkAssets

	assert.NotNil(t, actual)
	assert.Equal(t, expected, actual)
}

func Test_NativeToWrappedAssets(t *testing.T) {
	setup(false)

	actual := serviceInstance.NativeToWrappedAssets()
	expected := testConstants.NativeToWrapped

	assert.NotNil(t, actual)
	assert.Equal(t, expected, actual)
}

func Test_WrappedFromNative(t *testing.T) {
	setup(false)

	actual := serviceInstance.WrappedFromNative(constants.HederaNetworkId, testConstants.NetworkHederaFungibleNativeToken)
	expected := testConstants.NativeToWrapped[constants.HederaNetworkId][testConstants.NetworkHederaFungibleNativeToken]

	assert.NotNil(t, actual)
	assert.Equal(t, expected, actual)
}

func Test_FungibleNetworkAssetsByChainId(t *testing.T) {
	setup(false)

	actual := serviceInstance.FungibleNetworkAssetsByChainId(constants.HederaNetworkId)
	expected := testConstants.FungibleNetworkAssets[constants.HederaNetworkId]

	assert.NotNil(t, actual)
	assert.Equal(t, expected, actual)
}

func Test_FungibleNativeAsset(t *testing.T) {
	setup(false)

	actual := serviceInstance.FungibleNativeAsset(constants.HederaNetworkId, constants.Hbar)
	expected := testConstants.NetworkHederaFungibleNativeAsset

	assert.NotNil(t, actual)
	assert.Equal(t, expected, actual)
}

func Test_FungibleAssetInfo(t *testing.T) {
	setup(false)

	actual, exists := serviceInstance.FungibleAssetInfo(constants.HederaNetworkId, constants.Hbar)
	expected := testConstants.NetworkHederaFungibleNativeTokenFungibleAssetInfo

	assert.NotNil(t, actual)
	assert.True(t, exists)
	assert.Equal(t, expected, actual)
}

func Test_NonFungibleAssetInfo(t *testing.T) {
	setup(false)

	actual, exists := serviceInstance.NonFungibleAssetInfo(constants.HederaNetworkId, testConstants.NetworkHederaNonFungibleNativeToken)
	expected := testConstants.NetworkHederaNonFungibleNativeTokenNonFungibleAssetInfo

	assert.NotNil(t, actual)
	assert.True(t, exists)
	assert.Equal(t, expected, actual)
}

func setup(withClientMocks bool) {
	mocks.Setup()
	helper.SetupNetworks()

	if withClientMocks {
		setupClientMocks()
	}

	serviceInstance = &Service{
		nativeToWrapped:          testConstants.NativeToWrapped,
		wrappedToNative:          testConstants.WrappedToNative,
		fungibleNativeAssets:     testConstants.FungibleNativeAssets,
		fungibleNetworkAssets:    testConstants.FungibleNetworkAssets,
		fungibleAssetInfos:       testConstants.FungibleAssetInfos,
		nonFungibleNetworkAssets: testConstants.NonFungibleNetworkAssets,
		nonFungibleAssetInfos:    testConstants.NonFungibleAssetInfos,
		logger:                   config.GetLoggerFor("Assets Service"),
	}
}

func setupClientMocks() {
	for networkId := range testConstants.Networks {
		if networkId != constants.HederaNetworkId {
			evmFungibleTokenClients[networkId] = make(map[string]client.EvmFungibleToken)
			evmNFTClients[networkId] = make(map[string]client.EvmNft)
			evmClients[networkId] = &testClient.MockEVM{}
			evmCoreClients[networkId] = &testClient.MockEVMCore{}
			evmClients[networkId].(*testClient.MockEVM).On("GetClient").Return(evmCoreClients[networkId])
			routerClients[networkId] = new(testClient.MockDiamondRouter)
		}

		fungibleAssets := testConstants.FungibleNetworkAssets[networkId]
		for _, asset := range fungibleAssets {
			hederaPercentages[asset] = testConstants.FeePercentage
			if networkId == constants.HederaNetworkId {
				// Hedera
				tokenResponse := token.TokenResponse{
					TokenID:     asset,
					Name:        asset,
					Symbol:      asset,
					TotalSupply: "100",
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

		nonFungibleAssets := testConstants.NonFungibleNetworkAssets[networkId]
		for _, asset := range nonFungibleAssets {
			hederaPercentages[asset] = testConstants.FeePercentage
			if networkId == constants.HederaNetworkId {
				// Hedera
				tokenResponse := token.TokenResponse{
					TokenID:     asset,
					Name:        asset,
					Symbol:      asset,
					TotalSupply: "100",
					Decimals:    "0",
				}
				mocks.MHederaMirrorClient.On("GetToken", asset).Return(&tokenResponse, nil)
				continue
			}

			// EVM
			evmNFTClients[networkId][asset] = new(testClient.MockEvmNonFungibleToken)
			evmNFTClients[networkId][asset].(*testClient.MockEvmNonFungibleToken).On("Name", &bind.CallOpts{}).Return(asset, nil)
			evmNFTClients[networkId][asset].(*testClient.MockEvmNonFungibleToken).On("Symbol", &bind.CallOpts{}).Return(asset, nil)
			tokenFeeDataResult := struct {
				ServiceFeePercentage *big.Int
				FeesAccrued          *big.Int
				PreviousAccrued      *big.Int
				Accumulator          *big.Int
			}{big.NewInt(testConstants.FeePercentage), big.NewInt(0), big.NewInt(0), big.NewInt(0)}
			routerClients[networkId].(*testClient.MockDiamondRouter).On("TokenFeeData", &bind.CallOpts{}, common.HexToAddress(asset)).Return(tokenFeeDataResult, nil)
		}

	}
}

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
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model"
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
	hederaFeePercentages = make(map[string]int64)
	nilBigInt            *big.Int
	routerClients        = make(map[uint64]*router.Router)
	evmClients           = make(map[uint64]client.EVM)
	evmCoreClients       = make(map[uint64]client.Core)
	evmTokenClients      = make(map[uint64]map[string]client.EVMToken)
	serviceInstance      *Service
	nullAddress          = common.HexToAddress("0x0000000000000000000000000000000000000000")
	hederaPercentage     = int64(0)
	hederaPercentages    = make(map[string]int64)
)

func Test_New(t *testing.T) {
	setup()

	for networkId := range testConstants.Networks {
		if networkId != constants.HederaNetworkId {
			evmTokenClients[networkId] = make(map[string]client.EVMToken)
			evmClients[networkId] = &testClient.MockEVM{}
			evmCoreClients[networkId] = &testClient.MockEVMCore{}
			evmClients[networkId].(*testClient.MockEVM).On("GetClient").Return(evmCoreClients[networkId])
		}

		fungibleNetworkAssets := testConstants.FungibleNetworkAssets[networkId]

		for _, asset := range fungibleNetworkAssets {
			hederaPercentages[asset] = hederaPercentage
			if networkId == constants.HederaNetworkId {
				tokenResponse := model.TokenResponse{
					TokenID:     asset,
					Name:        asset,
					Symbol:      asset,
					TotalSupply: "100",
					Decimals:    strconv.Itoa(int(constants.HederaDefaultDecimals)),
				}
				mocks.MHederaMirrorClient.On("GetToken", asset).Return(&tokenResponse, nil)
				continue
			}

			evmTokenClients[networkId][asset] = new(testClient.MockEVMToken)
			evmTokenClients[networkId][asset].(*testClient.MockEVMToken).On("Name", &bind.CallOpts{}).Return(asset, nil)
			evmTokenClients[networkId][asset].(*testClient.MockEVMToken).On("Symbol", &bind.CallOpts{}).Return(asset, nil)
			evmTokenClients[networkId][asset].(*testClient.MockEVMToken).On("Decimals", &bind.CallOpts{}).Return(constants.EvmDefaultDecimals, nil)
		}
	}

	actualService := NewService(testConstants.Networks, hederaPercentages, routerClients, mocks.MHederaMirrorClient, evmTokenClients)
	assert.Equal(t, serviceInstance.nativeToWrapped, actualService.nativeToWrapped)
	assert.Equal(t, serviceInstance.wrappedToNative, actualService.wrappedToNative)
	assert.Equal(t, serviceInstance.fungibleNativeAssets, actualService.fungibleNativeAssets)
	assert.Equal(t, serviceInstance.fungibleAssetInfos, actualService.fungibleAssetInfos)
	for networkId := range testConstants.Networks {
		sort.Strings(serviceInstance.fungibleNetworkAssets[networkId])
		sort.Strings(actualService.fungibleNetworkAssets[networkId])
		assert.Equal(t, serviceInstance.fungibleNetworkAssets[networkId], actualService.fungibleNetworkAssets[networkId])
	}

}

func Test_IsNative(t *testing.T) {
	setup()

	actual := serviceInstance.IsNative(0, constants.Hbar)
	assert.Equal(t, true, actual)

	actual = serviceInstance.IsNative(0, nullAddress.String())
	assert.Equal(t, false, actual)
}

func Test_OppositeAsset(t *testing.T) {
	setup()

	actual := serviceInstance.OppositeAsset(33, 0, testConstants.NetworkPolygonFungibleWrappedTokenForNetworkHedera)
	expected := constants.Hbar

	assert.Equal(t, expected, actual)

	actual = serviceInstance.OppositeAsset(0, 33, constants.Hbar)
	expected = testConstants.NetworkPolygonFungibleWrappedTokenForNetworkHedera

	assert.Equal(t, expected, actual)

	actual = serviceInstance.OppositeAsset(33, 0, constants.Hbar)
	expected = testConstants.NetworkPolygonFungibleWrappedTokenForNetworkHedera

	assert.Equal(t, expected, actual)

	actual = serviceInstance.OppositeAsset(0, 33, testConstants.NetworkPolygonFungibleWrappedTokenForNetworkHedera)
	expected = constants.Hbar

	assert.Equal(t, expected, actual)

}

func Test_NativeToWrapped(t *testing.T) {
	setup()

	actual := serviceInstance.NativeToWrapped(constants.Hbar, 0, 33)
	expected := testConstants.NetworkPolygonFungibleWrappedTokenForNetworkHedera

	assert.Equal(t, expected, actual)
}

func Test_WrappedToNative(t *testing.T) {
	setup()

	actual := serviceInstance.WrappedToNative(testConstants.NetworkPolygonFungibleWrappedTokenForNetworkHedera, 33)
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

func setup() {
	mocks.Setup()
	helper.SetupNetworks()

	serviceInstance = &Service{
		nativeToWrapped:       testConstants.NativeToWrapped,
		wrappedToNative:       testConstants.WrappedToNative,
		fungibleNativeAssets:  testConstants.FungibleNativeAssets,
		fungibleNetworkAssets: testConstants.FungibleNetworkAssets,
		fungibleAssetInfos:    testConstants.FungibleAssetInfos,
		logger:                config.GetLoggerFor("Assets Service"),
	}
}

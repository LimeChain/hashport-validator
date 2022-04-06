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

package pricing

import (
	"errors"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/asset"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/pricing"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	testConstants "github.com/limechain/hedera-eth-bridge-validator/test/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/helper"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks/client"
	test_config "github.com/limechain/hedera-eth-bridge-validator/test/test-config"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"math/big"
	"sync"
	"testing"
)

var (
	serviceInstance       *Service
	coinGeckoClient       *client.MockPricingClient
	coinMarketCapClient   *client.MockPricingClient
	tokenPriceInfoMutex   *sync.RWMutex
	minAmountsForApiMutex *sync.RWMutex
)

func Test_New(t *testing.T) {
	setup(true, true)

	actualService := NewService(test_config.TestConfig.Bridge, mocks.MAssetsService, mocks.MHederaMirrorClient, coinGeckoClient, coinMarketCapClient)

	assert.Equal(t, serviceInstance, actualService)
}

func Test_New_ShouldPanic(t *testing.T) {
	setup(false, false)

	coinGeckoClient = new(client.MockPricingClient)
	coinMarketCapClient = new(client.MockPricingClient)
	mocks.MAssetsService.On("FungibleAssetInfo", constants.HederaNetworkId, testConstants.NetworkHederaFungibleNativeToken).Return(testConstants.NetworkHederaFungibleNativeTokenFungibleAssetInfo, false)
	mocks.MAssetsService.On("FungibleNativeAsset", constants.HederaNetworkId, testConstants.NetworkHederaFungibleNativeToken).Return(testConstants.NetworkHederaFungibleNativeAsset)
	mocks.MAssetsService.On("FungibleAssetInfo", testConstants.EthereumNetworkId, testConstants.NetworkEthereumFungibleNativeToken).Return(testConstants.NetworkEthereumFungibleNativeTokenFungibleAssetInfo, false)
	mocks.MAssetsService.On("FungibleNativeAsset", testConstants.EthereumNetworkId, testConstants.NetworkEthereumFungibleNativeToken).Return(testConstants.NetworkEthereumFungibleNativeAsset)
	mocks.MHederaMirrorClient.On("GetHBARUsdPrice").Return(decimal.Decimal{}, errors.New("failed to get HBAR USD price"))
	coinGeckoClient.On("GetUsdPrices", test_config.TestConfig.Bridge.CoinGeckoIds).Return(make(map[uint64]map[string]decimal.Decimal), errors.New("failed to get USD prices"))
	coinMarketCapClient.On("GetUsdPrices", test_config.TestConfig.Bridge.CoinMarketCapIds).Return(make(map[uint64]map[string]decimal.Decimal), errors.New("failed to get USD prices"))

	assert.Panics(t, func() {
		NewService(test_config.TestConfig.Bridge, mocks.MAssetsService, mocks.MHederaMirrorClient, coinGeckoClient, coinMarketCapClient)
	})
}

func Test_FetchAndUpdateUsdPrices(t *testing.T) {
	setup(true, false)

	err := serviceInstance.FetchAndUpdateUsdPrices()

	assert.Nil(t, err)

	assert.Equal(t, len(constants.NetworksById), len(serviceInstance.tokensPriceInfo))
	assert.Equal(t, len(testConstants.TokenPriceInfos[constants.HederaNetworkId]), len(serviceInstance.tokensPriceInfo[constants.HederaNetworkId]))
	assert.Equal(t, len(testConstants.TokenPriceInfos[testConstants.EthereumNetworkId]), len(serviceInstance.tokensPriceInfo[testConstants.EthereumNetworkId]))
	assert.Equal(t, len(testConstants.TokenPriceInfos[testConstants.PolygonNetworkId]), len(serviceInstance.tokensPriceInfo[testConstants.PolygonNetworkId]))
	assert.Equal(t, testConstants.HbarPriceInUsd, serviceInstance.tokensPriceInfo[constants.HederaNetworkId][constants.Hbar].UsdPrice)
	assert.Equal(t, testConstants.HbarMinAmountWithFee, serviceInstance.tokensPriceInfo[constants.HederaNetworkId][constants.Hbar].MinAmountWithFee)
	assert.Equal(t, testConstants.EthereumNativeTokenPriceInUsd, serviceInstance.tokensPriceInfo[testConstants.EthereumNetworkId][testConstants.NetworkEthereumFungibleNativeToken].UsdPrice)
	assert.Equal(t, testConstants.EthereumNativeTokenMinAmountWithFee, serviceInstance.tokensPriceInfo[testConstants.EthereumNetworkId][testConstants.NetworkEthereumFungibleNativeToken].MinAmountWithFee)

	assert.Equal(t, len(constants.NetworksById), len(serviceInstance.minAmountsForApi))
	assert.Equal(t, testConstants.HbarMinAmountWithFee.String(), serviceInstance.minAmountsForApi[constants.HederaNetworkId][constants.Hbar])
	assert.Equal(t, testConstants.EthereumNativeTokenMinAmountWithFee.String(), serviceInstance.minAmountsForApi[testConstants.EthereumNetworkId][testConstants.NetworkEthereumFungibleNativeToken])
}

func Test_GetTokenPriceInfo(t *testing.T) {
	setup(true, true)

	tokenPriceInfo, exists := serviceInstance.GetTokenPriceInfo(constants.HederaNetworkId, constants.Hbar)

	assert.True(t, exists)
	assert.Equal(t, testConstants.HbarPriceInUsd, tokenPriceInfo.UsdPrice)
	assert.Equal(t, testConstants.HbarMinAmountWithFee, tokenPriceInfo.MinAmountWithFee)
}

func Test_GetTokenPriceInfo_NotExist(t *testing.T) {
	setup(true, true)

	_, exists := serviceInstance.GetTokenPriceInfo(92919929912, "")

	assert.False(t, exists)
}

func Test_GetMinAmountsForAPI(t *testing.T) {
	setup(true, true)

	minAmountsForApi := serviceInstance.GetMinAmountsForAPI()

	assert.Equal(t, len(constants.NetworksById), len(minAmountsForApi))
	assert.Equal(t, testConstants.HbarMinAmountWithFee.String(), serviceInstance.minAmountsForApi[constants.HederaNetworkId][constants.Hbar])
	assert.Equal(t, testConstants.EthereumNativeTokenMinAmountWithFee.String(), serviceInstance.minAmountsForApi[testConstants.EthereumNetworkId][testConstants.NetworkEthereumFungibleNativeToken])
}

func Test_calculateMinAmountWithFee(t *testing.T) {
	setup(true, true)

	minAmountWithFee, err := serviceInstance.calculateMinAmountWithFee(testConstants.NetworkHederaFungibleNativeAsset, constants.HederaDefaultDecimals, testConstants.HbarPriceInUsd)
	expectedMinAmountFee := testConstants.HbarMinAmountWithFee

	assert.Nil(t, err)
	assert.Equal(t, expectedMinAmountFee, minAmountWithFee)
}

func Test_calculateMinAmountWithFee_WithZeroMinFeeAmount(t *testing.T) {
	setup(true, true)
	minFeeAmountInUsd := decimal.NewFromFloat(0)

	nativeAsset := &asset.NativeAsset{
		MinFeeAmountInUsd: &minFeeAmountInUsd,
		ChainId:           constants.HederaNetworkId,
		Asset:             testConstants.NetworkHederaFungibleNativeToken,
		FeePercentage:     testConstants.FeePercentage,
	}
	minAmountWithFee, err := serviceInstance.calculateMinAmountWithFee(nativeAsset, constants.HederaDefaultDecimals, testConstants.HbarPriceInUsd)
	expectedMinAmountFee := big.NewInt(0)

	assert.Nil(t, err)
	assert.Equal(t, expectedMinAmountFee, minAmountWithFee)
}

func setup(setupMocks bool, setTokenPriceInfosAndMinAmounts bool) {
	mocks.Setup()
	helper.SetupNetworks()

	coinGeckoClient = new(client.MockPricingClient)
	coinMarketCapClient = new(client.MockPricingClient)
	tokenPriceInfoMutex = new(sync.RWMutex)
	minAmountsForApiMutex = new(sync.RWMutex)

	if setupMocks {
		mocks.MAssetsService.On("FungibleAssetInfo", constants.HederaNetworkId, testConstants.NetworkHederaFungibleNativeToken).Return(testConstants.NetworkHederaFungibleNativeTokenFungibleAssetInfo, true)
		mocks.MAssetsService.On("FungibleNativeAsset", constants.HederaNetworkId, testConstants.NetworkHederaFungibleNativeToken).Return(testConstants.NetworkHederaFungibleNativeAsset)
		mocks.MAssetsService.On("FungibleAssetInfo", testConstants.EthereumNetworkId, testConstants.NetworkEthereumFungibleNativeToken).Return(testConstants.NetworkEthereumFungibleNativeTokenFungibleAssetInfo, true)
		mocks.MAssetsService.On("FungibleNativeAsset", testConstants.EthereumNetworkId, testConstants.NetworkEthereumFungibleNativeToken).Return(testConstants.NetworkEthereumFungibleNativeAsset)
		mocks.MHederaMirrorClient.On("GetHBARUsdPrice").Return(testConstants.HbarPriceInUsd, nil)
		coinGeckoClient.On("GetUsdPrices", test_config.TestConfig.Bridge.CoinGeckoIds).Return(testConstants.UsdPrices, nil)
		mocks.MAssetsService.On("NativeToWrapped", testConstants.NetworkEthereumFungibleNativeToken, testConstants.EthereumNetworkId, constants.HederaNetworkId).Return("")
		mocks.MAssetsService.On("NativeToWrapped", testConstants.NetworkHederaFungibleNativeToken, constants.HederaNetworkId, testConstants.EthereumNetworkId).Return(testConstants.NetworkEthereumFungibleWrappedTokenForNetworkHedera)
		mocks.MAssetsService.On("NativeToWrapped", testConstants.NetworkEthereumFungibleNativeToken, testConstants.EthereumNetworkId, testConstants.PolygonNetworkId).Return(testConstants.NetworkPolygonFungibleWrappedTokenForNetworkEthereum)
		mocks.MAssetsService.On("FungibleAssetInfo", testConstants.PolygonNetworkId, testConstants.NetworkPolygonFungibleWrappedTokenForNetworkEthereum).Return(testConstants.NetworkPolygonFungibleWrappedTokenForNetworkEthereumFungibleAssetInfo, true)
		mocks.MAssetsService.On("FungibleAssetInfo", testConstants.EthereumNetworkId, testConstants.NetworkEthereumFungibleWrappedTokenForNetworkHedera).Return(testConstants.NetworkEthereumFungibleWrappedTokenForNetworkHederaFungibleAssetInfo, true)
		mocks.MAssetsService.On("NativeToWrapped", constants.Hbar, constants.HederaNetworkId, testConstants.PolygonNetworkId).Return(testConstants.NetworkPolygonFungibleWrappedTokenForNetworkHedera)
		mocks.MAssetsService.On("FungibleAssetInfo", testConstants.PolygonNetworkId, testConstants.NetworkPolygonFungibleWrappedTokenForNetworkHedera).Return(testConstants.NetworkPolygonFungibleWrappedTokenForNetworkHederaFungibleAssetInfo, true)
	}

	var (
		tokensPriceInfo  map[uint64]map[string]pricing.TokenPriceInfo
		minAmountsForApi map[uint64]map[string]string
	)

	if setTokenPriceInfosAndMinAmounts {
		tokensPriceInfo = testConstants.TokenPriceInfos
		minAmountsForApi = testConstants.MinAmountsForApi
	} else {
		tokensPriceInfo = make(map[uint64]map[string]pricing.TokenPriceInfo)
		minAmountsForApi = make(map[uint64]map[string]string)
		for networkId := range constants.NetworksById {
			tokensPriceInfo[networkId] = make(map[string]pricing.TokenPriceInfo)
			minAmountsForApi[networkId] = make(map[string]string)
		}
	}

	serviceInstance = &Service{
		assetsService:         mocks.MAssetsService,
		mirrorNodeClient:      mocks.MHederaMirrorClient,
		coinGeckoClient:       coinGeckoClient,
		coinMarketCapClient:   coinMarketCapClient,
		tokenPriceInfoMutex:   tokenPriceInfoMutex,
		minAmountsForApiMutex: minAmountsForApiMutex,
		coinMarketCapIds:      test_config.TestConfig.Bridge.CoinMarketCapIds,
		coinGeckoIds:          test_config.TestConfig.Bridge.CoinGeckoIds,
		tokensPriceInfo:       tokensPriceInfo,
		minAmountsForApi:      minAmountsForApi,
		hbarFungibleAssetInfo: testConstants.NetworkHederaFungibleNativeTokenFungibleAssetInfo,
		hbarNativeAsset:       testConstants.NetworkHederaFungibleNativeAsset,
		logger:                config.GetLoggerFor("Pricing Service"),
	}

	serviceInstance.loadStaticMinAmounts(test_config.TestConfig.Bridge, tokensPriceInfo, minAmountsForApi)
}

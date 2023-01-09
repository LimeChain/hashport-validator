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
	"math/big"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	validatorClient "github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	decimalHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/decimal"

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
)

var (
	serviceInstance       *Service
	coinGeckoClient       *client.MockPricingClient
	coinMarketCapClient   *client.MockPricingClient
	tokenPriceInfoMutex   *sync.RWMutex
	minAmountsForApiMutex *sync.RWMutex
	nftFeesForApiMutex    *sync.RWMutex
	diamondRouters        map[uint64]validatorClient.DiamondRouter
)

func Test_New(t *testing.T) {
	setup(true, true)

	actualService := NewService(test_config.TestConfig.Bridge, mocks.MAssetsService, diamondRouters, mocks.MHederaMirrorClient, coinGeckoClient, coinMarketCapClient)

	// reset fields
	serviceInstance.hederaNftDynamicFees = nil

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
		NewService(test_config.TestConfig.Bridge, mocks.MAssetsService, nil, mocks.MHederaMirrorClient, coinGeckoClient, coinMarketCapClient)
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

func Test_PriceFetchingServiceDown(t *testing.T) {
	setup(true, false)

	// Check updatePricesWithoutHbar() function
	FetchResults := serviceInstance.fetchUsdPricesFromAPIs()

	// Use cached price
	FetchResults.AllPrices[1]["0xb083879B1e10C8476802016CB12cd2F25a896691"] = decimal.NewFromFloat(0)
	serviceInstance.tokensPriceInfo[1]["0xb083879B1e10C8476802016CB12cd2F25a896691"] = pricing.TokenPriceInfo{
		UsdPrice:         decimal.NewFromFloat(100),
		MinAmountWithFee: big.NewInt(1250000000000000),
	}
	err := serviceInstance.updatePricesWithoutHbar(FetchResults.AllPrices)
	assert.Nil(t, err)

	// Use DefaultMinAmount (min_amount from yaml)
	FetchResults.AllPrices[1]["0xb083879B1e10C8476802016CB12cd2F25a896691"] = decimal.NewFromFloat(0)
	serviceInstance.tokensPriceInfo[1]["0xb083879B1e10C8476802016CB12cd2F25a896691"] = pricing.TokenPriceInfo{
		UsdPrice:         decimal.NewFromFloat(0),
		MinAmountWithFee: big.NewInt(1250000000000000),
		DefaultMinAmount: big.NewInt(1250000000000000),
	}
	err = serviceInstance.updatePricesWithoutHbar(FetchResults.AllPrices)
	assert.Nil(t, err)

	// Throw if no cached price and no DefaultMinAmount (min_amount from yaml) is set
	FetchResults.AllPrices[1]["0xb083879B1e10C8476802016CB12cd2F25a896691"] = decimal.NewFromFloat(0)
	serviceInstance.tokensPriceInfo[1]["0xb083879B1e10C8476802016CB12cd2F25a896691"] = pricing.TokenPriceInfo{
		UsdPrice:         decimal.NewFromFloat(0),
		MinAmountWithFee: big.NewInt(1250000000000000),
		DefaultMinAmount: big.NewInt(0),
	}
	err = serviceInstance.updatePricesWithoutHbar(FetchResults.AllPrices)
	assert.Error(t, err)

	// Check updateHbarPrice() function
	FetchResults2 := fetchResults{
		HbarPrice: decimal.NewFromFloat(0),
		AllPrices: FetchResults.AllPrices,
	}

	// Use cached price
	serviceInstance.tokensPriceInfo[296]["HBAR"] = pricing.TokenPriceInfo{
		UsdPrice:         decimal.NewFromFloat(0.2),
		MinAmountWithFee: big.NewInt(5000000000),
	}
	err = serviceInstance.updateHbarPrice(FetchResults2)
	assert.Nil(t, err)

	// Throw if no fetched price and no cache price
	FetchResults2.HbarPrice = decimal.NewFromFloat(0)
	serviceInstance.tokensPriceInfo[296]["HBAR"] = pricing.TokenPriceInfo{
		UsdPrice:         decimal.NewFromFloat(0),
		MinAmountWithFee: big.NewInt(5000000000),
	}

	err = serviceInstance.updateHbarPrice(FetchResults2)
	assert.Error(t, err)

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

func Test_updatePricesWithoutHbar_NonExistingAddress(t *testing.T) {
	setup(true, true)
	var nilFungibleAssetInfo *asset.FungibleAssetInfo
	mocks.MAssetsService.On("FungibleAssetInfo", constants.HederaNetworkId, "nonExistingAddress").Return(nilFungibleAssetInfo, false)
	pricesByNetworkIdAndAddress := testConstants.UsdPrices
	pricesByNetworkIdAndAddress[constants.HederaNetworkId] = map[string]decimal.Decimal{
		"nonExistingAddress": {},
	}

	err := serviceInstance.updatePricesWithoutHbar(pricesByNetworkIdAndAddress)

	assert.Nil(t, err)
}
func Test_GetHederaNftFee(t *testing.T) {
	setup(true, true)

	priceInUsd := decimal.NewFromInt(400)
	expectedFee := decimalHelper.ToLowestDenomination(testConstants.HederaNftDynamicFees[testConstants.NetworkHederaNonFungibleNativeToken].Div(priceInUsd), serviceInstance.hbarFungibleAssetInfo.Decimals).Int64()

	serviceInstance.updateHederaNftDynamicFeesBasedOnHbar(priceInUsd, serviceInstance.hbarFungibleAssetInfo.Decimals)

	fee, ok := serviceInstance.GetHederaNftFee(testConstants.NetworkHederaNonFungibleNativeToken)
	assert.Equal(t, expectedFee, fee)
	assert.True(t, ok)
}

func Test_GetHederaNftPrevFee_ShouldExists(t *testing.T) {
	setup(true, true)

	priceInUsd := decimal.NewFromInt(400)
	expectedPrevFee := testConstants.HederaNftFees[testConstants.NetworkHederaNonFungibleNativeToken]
	expectedFee := decimalHelper.ToLowestDenomination(testConstants.HederaNftDynamicFees[testConstants.NetworkHederaNonFungibleNativeToken].Div(priceInUsd), serviceInstance.hbarFungibleAssetInfo.Decimals).Int64()

	fee, ok := serviceInstance.GetHederaNftFee(testConstants.NetworkHederaNonFungibleNativeToken)
	assert.Equal(t, testConstants.HederaNftFees[testConstants.NetworkHederaNonFungibleNativeToken], fee)
	assert.True(t, ok)

	serviceInstance.updateHederaNftDynamicFeesBasedOnHbar(priceInUsd, serviceInstance.hbarFungibleAssetInfo.Decimals)

	fee, ok = serviceInstance.GetHederaNftFee(testConstants.NetworkHederaNonFungibleNativeToken)
	assert.Equal(t, expectedFee, fee)
	assert.True(t, ok)

	prevFee, ok := serviceInstance.GetHederaNftPrevFee(testConstants.NetworkHederaNonFungibleNativeToken)
	assert.Equal(t, expectedPrevFee, prevFee)
	assert.True(t, ok)
}

func Test_GetHederaNftPrevFee_ShouldNotExists(t *testing.T) {
	setup(true, true)

	prevFee, ok := serviceInstance.GetHederaNftPrevFee(testConstants.NetworkHederaNonFungibleNativeToken)
	assert.Equal(t, int64(0), prevFee)
	assert.False(t, ok)
}

func setup(setupMocks bool, setTokenPriceInfosAndMinAmounts bool) {
	mocks.Setup()
	helper.SetupNetworks()

	coinGeckoClient = new(client.MockPricingClient)
	coinMarketCapClient = new(client.MockPricingClient)
	tokenPriceInfoMutex = new(sync.RWMutex)
	minAmountsForApiMutex = new(sync.RWMutex)
	nftFeesForApiMutex = new(sync.RWMutex)
	diamondRouters = map[uint64]validatorClient.DiamondRouter{
		testConstants.PolygonNetworkId:  mocks.MDiamondRouter,
		testConstants.EthereumNetworkId: mocks.MDiamondRouter,
	}

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
		mocks.MAssetsService.On("NonFungibleNetworkAssets").Return(testConstants.NonFungibleNetworkAssets)
		mocks.MAssetsService.On("NonFungibleAssetInfo", testConstants.PolygonNetworkId, testConstants.NetworkPolygonWrappedNonFungibleTokenForHedera).Return(testConstants.NetworkPolygonWrappedNonFungibleTokenForHederaNonFungibleAssetInfo, true)
		mocks.MAssetsService.On("NonFungibleAssetInfo", testConstants.EthereumNetworkId, testConstants.NetworkEthereumNFTWrappedTokenForNetworkHedera).Return(testConstants.NetworkEthereumFungibleWrappedTokenForNetworkHederaFungibleAssetInfo, true)
		mocks.MAssetsService.On("NonFungibleAssetInfo", constants.HederaNetworkId, testConstants.NetworkHederaNonFungibleNativeToken).Return(testConstants.NetworkHederaNonFungibleNativeTokenNonFungibleAssetInfo, true)
		mocks.MDiamondRouter.On("Erc721Payment", &bind.CallOpts{}, common.HexToAddress(testConstants.NetworkPolygonWrappedNonFungibleTokenForHedera)).Return(common.HexToAddress(testConstants.NftFeesForApi[testConstants.PolygonNetworkId][testConstants.NetworkPolygonWrappedNonFungibleTokenForHedera].PaymentToken), nil)
		mocks.MDiamondRouter.On("Erc721Fee", &bind.CallOpts{}, common.HexToAddress(testConstants.NetworkPolygonWrappedNonFungibleTokenForHedera)).Return(big.NewInt(testConstants.NftFeesForApi[testConstants.PolygonNetworkId][testConstants.NetworkPolygonWrappedNonFungibleTokenForHedera].Fee.IntPart()), nil)
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
		nftFeesForApiMutex:    nftFeesForApiMutex,
		coinMarketCapIds:      test_config.TestConfig.Bridge.CoinMarketCapIds,
		coinGeckoIds:          test_config.TestConfig.Bridge.CoinGeckoIds,
		tokensPriceInfo:       tokensPriceInfo,
		minAmountsForApi:      minAmountsForApi,
		hbarFungibleAssetInfo: testConstants.NetworkHederaFungibleNativeTokenFungibleAssetInfo,
		hbarNativeAsset:       testConstants.NetworkHederaFungibleNativeAsset,
		hederaNftDynamicFees:  testConstants.HederaNftDynamicFees,
		hederaNftFees:         testConstants.HederaNftFees,
		hederaNftPrevFees:     make(map[string]int64),
		diamondRouters:        diamondRouters,
		nftFeesForApi:         testConstants.NftFeesForApi,
		logger:                config.GetLoggerFor("Pricing Service"),
	}

	serviceInstance.loadStaticMinAmounts(test_config.TestConfig.Bridge)
}

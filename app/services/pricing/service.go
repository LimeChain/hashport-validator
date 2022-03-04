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
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	decimalHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/decimal"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/asset"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/pricing"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"math/big"
	"sync"
)

var (
	instance          *Service
	initOnce          sync.Once
	hundredPercentage = big.NewInt(100)
)

type Service struct {
	assetsService             service.Assets
	hederaWebApiClient        client.HederaWebAPI
	coinGeckoWebApiClient     client.CoinGeckoWebAPI
	coinMarketCapWebApiClient client.CoinMarketCapWebAPI
	tokenPriceInfoMutex       *sync.RWMutex
	minAmountsForApiMutex     *sync.RWMutex
	coinMarketCapIds          map[uint64]map[string]string
	coinGeckoIds              map[uint64]map[string]string
	tokensPriceInfo           map[uint64]map[string]pricing.TokenPriceInfo
	minAmountsForApi          map[uint64]map[string]*big.Int
	hbarFungibleAssetInfo     asset.FungibleAssetInfo
	hbarNativeAsset           *asset.NativeAsset
	logger                    *log.Entry
}

func NewService(bridgeConfig config.Bridge,
	assetsService service.Assets,
	hederaClient client.HederaWebAPI,
	coinGeckoClient client.CoinGeckoWebAPI,
	coinMarketCapClient client.CoinMarketCapWebAPI) *Service {
	initOnce.Do(func() {
		tokensPriceInfo := make(map[uint64]map[string]pricing.TokenPriceInfo)
		minAmountsForApi := make(map[uint64]map[string]*big.Int)
		for networkId := range constants.NetworksById {
			tokensPriceInfo[networkId] = make(map[string]pricing.TokenPriceInfo)
			minAmountsForApi[networkId] = make(map[string]*big.Int)
		}

		hbarFungibleAssetInfo, _ := assetsService.GetFungibleAssetInfo(constants.HederaNetworkId, constants.Hbar)
		hbarNativeAsset := assetsService.FungibleNativeAsset(constants.HederaNetworkId, constants.Hbar)
		instance = &Service{
			tokensPriceInfo:           tokensPriceInfo,
			minAmountsForApi:          minAmountsForApi,
			hederaWebApiClient:        hederaClient,
			coinGeckoWebApiClient:     coinGeckoClient,
			coinMarketCapWebApiClient: coinMarketCapClient,
			tokenPriceInfoMutex:       new(sync.RWMutex),
			minAmountsForApiMutex:     new(sync.RWMutex),
			assetsService:             assetsService,
			coinGeckoIds:              bridgeConfig.CoinGeckoIds,
			coinMarketCapIds:          bridgeConfig.CoinMarketCapIds,
			hbarFungibleAssetInfo:     hbarFungibleAssetInfo,
			hbarNativeAsset:           hbarNativeAsset,
			logger:                    config.GetLoggerFor("Pricing Service"),
		}

		instance.FetchAndUpdateUsdPrices(true)
	})

	return instance
}

func (s *Service) GetTokenPriceInfo(networkId uint64, tokenAddressOrId string) (priceInfo pricing.TokenPriceInfo, exist bool) {
	s.tokenPriceInfoMutex.RLock()
	defer s.tokenPriceInfoMutex.RUnlock()

	_, exist = s.tokensPriceInfo[networkId]
	if !exist {
		return priceInfo, false
	}

	priceInfo, exist = s.tokensPriceInfo[networkId][tokenAddressOrId]

	return priceInfo, exist
}

func (s *Service) FetchAndUpdateUsdPrices(initialFetch bool) {
	s.minAmountsForApiMutex.Lock()
	defer s.minAmountsForApiMutex.Unlock()
	s.tokenPriceInfoMutex.Lock()
	defer s.tokenPriceInfoMutex.Unlock()

	results := s.fetchUsdPricesFromAPIs(initialFetch)

	if results.AllPricesErr == nil {
		s.updatePricesWithoutHbar(results.AllPrices)
	}

	s.updateHbarPrice(results)
}

func (s *Service) GetMinAmountsForAPI() map[uint64]map[string]*big.Int {
	s.minAmountsForApiMutex.RLock()
	defer s.minAmountsForApiMutex.Unlock()

	return s.minAmountsForApi
}

func (s *Service) updateHbarPrice(results fetchResults) {

	var priceInUsd decimal.Decimal
	if results.HbarErr != nil {
		if results.AllPricesErr != nil {
			return
		}
		priceInUsd = results.AllPrices[constants.HederaNetworkId][constants.Hbar]
	} else {
		priceInUsd = results.HbarPrice
	}

	minAmountWithFee := s.calculateMinAmountWithFee(s.hbarNativeAsset, s.hbarFungibleAssetInfo.Decimals, priceInUsd)

	tokenPriceInfo := pricing.TokenPriceInfo{
		UsdPrice:              priceInUsd,
		MinAmountInUsdWithFee: minAmountWithFee,
	}

	s.updatePriceInfoContainers(s.hbarNativeAsset, tokenPriceInfo)
}

func (s *Service) calculateMinAmountWithFee(nativeAsset *asset.NativeAsset, decimals uint8, priceInUsd decimal.Decimal) *big.Int {

	feePercentageBigInt := big.NewInt(nativeAsset.FeePercentage)
	minAmountInUsd := nativeAsset.MinFeeAmountInUsd.Div(priceInUsd)
	minAmountInUsdToSmallest := decimalHelper.ToLowestDenomination(minAmountInUsd, decimals)
	amountMultiplier := big.NewInt(0).Div(constants.FeeMaxPercentageBigInt, feePercentageBigInt)
	minAmountWithFee := big.NewInt(0).Mul(minAmountInUsdToSmallest, amountMultiplier)
	return minAmountWithFee
}

func (s *Service) updatePriceInfoContainers(nativeAsset *asset.NativeAsset, tokenPriceInfo pricing.TokenPriceInfo) {
	s.tokensPriceInfo[nativeAsset.ChainId][nativeAsset.Asset] = tokenPriceInfo
	s.minAmountsForApi[nativeAsset.ChainId][nativeAsset.Asset] = tokenPriceInfo.MinAmountInUsdWithFee

	for networkId := range constants.NetworksById {
		if networkId == nativeAsset.ChainId {
			continue
		}

		// Calculating Min Amount for wrapped tokens
		wrappedToken := s.assetsService.NativeToWrapped(nativeAsset.Asset, nativeAsset.ChainId, networkId)
		wrappedAssetInfo, _ := s.assetsService.GetFungibleAssetInfo(networkId, wrappedToken)
		wrappedMinAmountWithFee := s.calculateMinAmountWithFee(nativeAsset, wrappedAssetInfo.Decimals, tokenPriceInfo.UsdPrice)

		tokenPriceInfo.MinAmountInUsdWithFee = wrappedMinAmountWithFee
		s.tokensPriceInfo[networkId][wrappedToken] = tokenPriceInfo

		s.minAmountsForApi[networkId][wrappedToken] = wrappedMinAmountWithFee
	}
}

func (s *Service) updatePricesWithoutHbar(pricesByNetworkAndAddress map[uint64]map[string]decimal.Decimal) {

	for networkId, pricesByAddress := range pricesByNetworkAndAddress {
		for address, price := range pricesByAddress {
			if address == constants.Hbar {
				continue
			}

			fungibleAssetInfo, exist := s.assetsService.GetFungibleAssetInfo(networkId, address)
			if !exist {
				continue
			}
			nativeAsset := s.assetsService.FungibleNativeAsset(networkId, address)
			minAmountWithFee := s.calculateMinAmountWithFee(nativeAsset, fungibleAssetInfo.Decimals, price)

			tokenPriceInfo := pricing.TokenPriceInfo{
				UsdPrice:              price,
				MinAmountInUsdWithFee: minAmountWithFee,
			}

			s.updatePriceInfoContainers(nativeAsset, tokenPriceInfo)
		}
	}
}

type fetchResults struct {
	HbarPrice    decimal.Decimal
	HbarErr      error
	AllPrices    map[uint64]map[string]decimal.Decimal
	AllPricesErr error
}

func (s *Service) fetchUsdPricesFromAPIs(initialFetch bool) (fetchResults fetchResults) {
	fetchResults.HbarPrice, fetchResults.HbarErr = s.hederaWebApiClient.GetHBARUsdPrice()

	fetchResults.AllPrices, fetchResults.AllPricesErr = s.coinGeckoWebApiClient.GetUsdPrices(s.coinGeckoIds)

	if fetchResults.AllPricesErr != nil { // Fetch from CoinMarketCap if CoinGecko fetch fails
		fetchResults.AllPrices, fetchResults.AllPricesErr = s.coinMarketCapWebApiClient.GetUsdPrices(s.coinMarketCapIds)
		if fetchResults.AllPricesErr != nil { // If CoinMarketCap fetch fails this means the whole update failed
			msg := "Couldn't fetch prices from any of the Web APIs."
			if initialFetch {
				s.logger.Fatalf("Couldn't fetch prices from any of the Web APIs.")
			}
			s.logger.Error(msg)
		}
	}

	return fetchResults
}

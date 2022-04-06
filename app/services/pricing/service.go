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
	"fmt"
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

type Service struct {
	assetsService         service.Assets
	mirrorNodeClient      client.MirrorNode
	coinGeckoClient       client.Pricing
	coinMarketCapClient   client.Pricing
	tokenPriceInfoMutex   *sync.RWMutex
	minAmountsForApiMutex *sync.RWMutex
	coinMarketCapIds      map[uint64]map[string]string
	coinGeckoIds          map[uint64]map[string]string
	tokensPriceInfo       map[uint64]map[string]pricing.TokenPriceInfo
	minAmountsForApi      map[uint64]map[string]string
	hbarFungibleAssetInfo asset.FungibleAssetInfo
	hbarNativeAsset       *asset.NativeAsset
	logger                *log.Entry
}

func NewService(bridgeConfig config.Bridge,
	assetsService service.Assets,
	mirrorNodeClient client.MirrorNode,
	coinGeckoClient client.Pricing,
	coinMarketCapClient client.Pricing) *Service {
	tokensPriceInfo := make(map[uint64]map[string]pricing.TokenPriceInfo)
	minAmountsForApi := make(map[uint64]map[string]string)
	for networkId := range constants.NetworksById {
		tokensPriceInfo[networkId] = make(map[string]pricing.TokenPriceInfo)
		minAmountsForApi[networkId] = make(map[string]string)
	}

	logger := config.GetLoggerFor("Pricing Service")
	hbarFungibleAssetInfo, _ := assetsService.FungibleAssetInfo(constants.HederaNetworkId, constants.Hbar)
	hbarNativeAsset := assetsService.FungibleNativeAsset(constants.HederaNetworkId, constants.Hbar)
	instance := &Service{
		tokensPriceInfo:       tokensPriceInfo,
		minAmountsForApi:      minAmountsForApi,
		mirrorNodeClient:      mirrorNodeClient,
		coinGeckoClient:       coinGeckoClient,
		coinMarketCapClient:   coinMarketCapClient,
		tokenPriceInfoMutex:   new(sync.RWMutex),
		minAmountsForApiMutex: new(sync.RWMutex),
		assetsService:         assetsService,
		coinGeckoIds:          bridgeConfig.CoinGeckoIds,
		coinMarketCapIds:      bridgeConfig.CoinMarketCapIds,
		hbarFungibleAssetInfo: hbarFungibleAssetInfo,
		hbarNativeAsset:       hbarNativeAsset,
		logger:                logger,
	}

	for networkId, minAmountsByTokenAddress := range bridgeConfig.MinAmounts {
		for tokenAddress, minAmount := range minAmountsByTokenAddress {
			tokensPriceInfo[networkId][tokenAddress] = pricing.TokenPriceInfo{
				MinAmountWithFee: minAmount,
			}
			minAmountsForApi[networkId][tokenAddress] = minAmount.String()
		}
	}

	err := instance.FetchAndUpdateUsdPrices()
	if err != nil {
		panic(fmt.Sprintf("Failed to initially fetch USD prices. Error: [%s]", err.Error()))
	}

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

func (s *Service) FetchAndUpdateUsdPrices() error {
	s.minAmountsForApiMutex.Lock()
	defer s.minAmountsForApiMutex.Unlock()
	s.tokenPriceInfoMutex.Lock()
	defer s.tokenPriceInfoMutex.Unlock()

	results := s.fetchUsdPricesFromAPIs()
	if results.AllPricesErr == nil {
		err := s.updatePricesWithoutHbar(results.AllPrices)
		if err != nil {
			err = errors.New(fmt.Sprintf("Failed to fetch prices for all tokens without HBAR. Error [%s]", err))
			return err
		}
	}

	err := s.updateHbarPrice(results)
	if err != nil {
		err = errors.New(fmt.Sprintf("Failed to fetch price for HBAR. Error [%s]", err))
		return err
	}

	return nil
}

func (s *Service) GetMinAmountsForAPI() map[uint64]map[string]string {
	s.minAmountsForApiMutex.RLock()
	defer s.minAmountsForApiMutex.RUnlock()

	return s.minAmountsForApi
}

func (s *Service) updateHbarPrice(results fetchResults) error {

	var priceInUsd decimal.Decimal
	if results.HbarErr != nil {
		if results.AllPricesErr != nil {
			return results.HbarErr
		}
		priceInUsd = results.AllPrices[constants.HederaNetworkId][constants.Hbar]
	} else {
		priceInUsd = results.HbarPrice
	}

	minAmountWithFee, err := s.calculateMinAmountWithFee(s.hbarNativeAsset, s.hbarFungibleAssetInfo.Decimals, priceInUsd)
	if err != nil {
		return err
	}

	tokenPriceInfo := pricing.TokenPriceInfo{
		UsdPrice:         priceInUsd,
		MinAmountWithFee: minAmountWithFee,
	}

	err = s.updatePriceInfoContainers(s.hbarNativeAsset, tokenPriceInfo)
	if err != nil {
		err = errors.New(fmt.Sprintf("Failed to update price info containers. Error: [%s]", err))
	}

	return err
}

func (s *Service) calculateMinAmountWithFee(nativeAsset *asset.NativeAsset, decimals uint8, priceInUsd decimal.Decimal) (minAmountWithFee *big.Int, err error) {
	if nativeAsset.MinFeeAmountInUsd.Equal(decimal.NewFromFloat(0.0)) {
		return big.NewInt(0), nil
	}

	feePercentageBigInt := big.NewInt(nativeAsset.FeePercentage)
	minFeeAmountMultiplier, err := decimal.NewFromString(big.NewInt(0).Div(constants.FeeMaxPercentageBigInt, feePercentageBigInt).String())
	if err != nil {
		return nil, err
	}

	minAmountInUsdWithFee := nativeAsset.MinFeeAmountInUsd.Mul(minFeeAmountMultiplier)
	minAmountWithFeeAsDecimal := minAmountInUsdWithFee.Div(priceInUsd)

	return decimalHelper.ToLowestDenomination(minAmountWithFeeAsDecimal, decimals), nil
}

func (s *Service) updatePriceInfoContainers(nativeAsset *asset.NativeAsset, tokenPriceInfo pricing.TokenPriceInfo) error {
	s.tokensPriceInfo[nativeAsset.ChainId][nativeAsset.Asset] = tokenPriceInfo
	s.minAmountsForApi[nativeAsset.ChainId][nativeAsset.Asset] = tokenPriceInfo.MinAmountWithFee.String()

	msgTemplate := "Updating UsdPrice [%s] and MinAmountWithFee [%s] for %s asset [%s]"
	s.logger.Infof(msgTemplate, tokenPriceInfo.UsdPrice.String(), tokenPriceInfo.MinAmountWithFee.String(), "native", nativeAsset.Asset)

	for networkId := range constants.NetworksById {
		if networkId == nativeAsset.ChainId {
			continue
		}

		// Calculating Min Amount for wrapped tokens
		wrappedToken := s.assetsService.NativeToWrapped(nativeAsset.Asset, nativeAsset.ChainId, networkId)
		if wrappedToken == "" {
			continue
		}
		wrappedAssetInfo, _ := s.assetsService.FungibleAssetInfo(networkId, wrappedToken)
		wrappedMinAmountWithFee, err := s.calculateMinAmountWithFee(nativeAsset, wrappedAssetInfo.Decimals, tokenPriceInfo.UsdPrice)
		if err != nil {
			return err
		}

		tokenPriceInfo.MinAmountWithFee = wrappedMinAmountWithFee
		s.tokensPriceInfo[networkId][wrappedToken] = tokenPriceInfo
		s.minAmountsForApi[networkId][wrappedToken] = wrappedMinAmountWithFee.String()
		s.logger.Infof(msgTemplate, tokenPriceInfo.UsdPrice.String(), wrappedMinAmountWithFee.String(), "wrapped", wrappedToken)
	}

	return nil
}

func (s *Service) updatePricesWithoutHbar(pricesByNetworkAndAddress map[uint64]map[string]decimal.Decimal) error {

	for networkId, pricesByAddress := range pricesByNetworkAndAddress {
		for assetAddress, usdPrice := range pricesByAddress {
			if assetAddress == constants.Hbar {
				continue
			}

			fungibleAssetInfo, exist := s.assetsService.FungibleAssetInfo(networkId, assetAddress)
			if !exist {
				continue
			}
			nativeAsset := s.assetsService.FungibleNativeAsset(networkId, assetAddress)
			minAmountWithFee, err := s.calculateMinAmountWithFee(nativeAsset, fungibleAssetInfo.Decimals, usdPrice)
			if err != nil {
				err = errors.New(fmt.Sprintf("Failed to calculate 'MinAmountWithFee' for asset: [%s]. Error: [%s]", assetAddress, err))
				return err
			}

			tokenPriceInfo := pricing.TokenPriceInfo{
				UsdPrice:         usdPrice,
				MinAmountWithFee: minAmountWithFee,
			}

			err = s.updatePriceInfoContainers(nativeAsset, tokenPriceInfo)
			if err != nil {
				err = errors.New(fmt.Sprintf("Failed to update price info containers. Error: [%s]", err))
				return err
			}
		}
	}

	return nil
}

type fetchResults struct {
	HbarPrice    decimal.Decimal
	HbarErr      error
	AllPrices    map[uint64]map[string]decimal.Decimal
	AllPricesErr error
}

func (s *Service) fetchUsdPricesFromAPIs() (fetchResults fetchResults) {
	fetchResults.HbarPrice, fetchResults.HbarErr = s.mirrorNodeClient.GetHBARUsdPrice()

	fetchResults.AllPrices, fetchResults.AllPricesErr = s.coinGeckoClient.GetUsdPrices(s.coinGeckoIds)
	if fetchResults.AllPricesErr != nil {
		s.logger.Errorf("Couldn't fetch prices from CoinGecko Web API. Error: [%s]", fetchResults.AllPricesErr)
	}

	if fetchResults.AllPricesErr != nil { // Fetch from CoinMarketCap if CoinGecko fetch fails
		s.logger.Infof("Fallback to fetching prices from Coin Market Cap ...")
		fetchResults.AllPrices, fetchResults.AllPricesErr = s.coinMarketCapClient.GetUsdPrices(s.coinMarketCapIds)
		if fetchResults.AllPricesErr != nil { // If CoinMarketCap fetch fails this means the whole update failed
			msg := fmt.Sprintf("Couldn't fetch prices from Coin Market Cap Web API. Error: [%s]", fetchResults.AllPricesErr)
			s.logger.Error(msg)
		}
	}

	return fetchResults
}

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
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/gookit/event"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	decimalHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/decimal"
	eventHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/events"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/asset"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/pricing"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	assetsService         service.Assets
	mirrorNodeClient      client.MirrorNode
	coinGeckoClient       client.Pricing
	coinMarketCapClient   client.Pricing
	tokenPriceInfoMutex   *sync.RWMutex
	minAmountsForApiMutex *sync.RWMutex
	nftFeesForApiMutex    *sync.RWMutex
	coinMarketCapIds      map[uint64]map[string]string
	coinGeckoIds          map[uint64]map[string]string
	tokensPriceInfo       map[uint64]map[string]pricing.TokenPriceInfo
	minAmountsForApi      map[uint64]map[string]string
	hbarFungibleAssetInfo *asset.FungibleAssetInfo
	hbarNativeAsset       *asset.NativeAsset
	hederaNftDynamicFees  map[string]decimal.Decimal
	hederaNftFees         map[string]int64
	nftFeesForApi         map[uint64]map[string]pricing.NonFungibleFee
	diamondRouters        map[uint64]client.DiamondRouter
	logger                *log.Entry
}

func NewService(bridgeConfig *config.Bridge, assetsService service.Assets, diamondRouters map[uint64]client.DiamondRouter, mirrorNodeClient client.MirrorNode, coinGeckoClient client.Pricing, coinMarketCapClient client.Pricing) *Service {
	instance := initialize(bridgeConfig, assetsService, mirrorNodeClient, coinGeckoClient, coinMarketCapClient, diamondRouters)
	event.On(constants.EventBridgeConfigUpdate, event.ListenerFunc(func(e event.Event) error {
		return bridgeCfgEventHandler(e, assetsService, mirrorNodeClient, coinGeckoClient, coinMarketCapClient, instance)
	}), constants.ServiceEventPriority)

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
			err = fmt.Errorf("Failed to update prices for all tokens without HBAR. Error [%s]", err)
			return err
		}
	}

	err := s.updateHbarPrice(results)
	if err != nil {
		err = fmt.Errorf("Failed to fetch price for HBAR. Error [%s]", err)
		return err
	}

	return nil
}

func (s *Service) fetchAndUpdateNftFeesForApi() error {
	s.nftFeesForApiMutex.Lock()
	defer s.nftFeesForApiMutex.Unlock()
	s.logger.Infof("Populating NFT fees for API")

	res := make(map[uint64]map[string]pricing.NonFungibleFee)
	assets := s.assetsService.NonFungibleNetworkAssets()
	for networkId, nfts := range assets {
		res[networkId] = make(map[string]pricing.NonFungibleFee)
		for _, id := range nfts {
			assetInfo, ok := s.assetsService.NonFungibleAssetInfo(networkId, id)
			if !ok {
				s.logger.Errorf("Failed to get asset info for [%s]", id)
				return fmt.Errorf("Failed to get asset info for [%s]", id)
			}

			if networkId == constants.HederaNetworkId {
				fee, err := s.hederaNativeNftFee(id, networkId)
				if err != nil {
					return err
				}
				res[networkId][id] = *fee
				continue
			}

			diamondRouter, ok := s.diamondRouters[networkId]
			if !ok {
				return fmt.Errorf("could not get diamond router for network %d", networkId)
			}

			if assetInfo.IsNative {
				fee, err := s.evmNativeNftFee(id, diamondRouter)
				if err != nil {
					return err
				}
				res[networkId][id] = *fee
			} else {
				fee, err := s.evmWrappedNftFee(id, diamondRouter)
				if err != nil {
					return err
				}
				res[networkId][id] = *fee
			}
		}
	}

	s.logger.Infof("fetched all NFT fees and payment tokens successfully")
	s.nftFeesForApi = res

	return nil
}

func (s *Service) NftFees() map[uint64]map[string]pricing.NonFungibleFee {
	s.nftFeesForApiMutex.RLock()
	defer s.nftFeesForApiMutex.RUnlock()

	return s.nftFeesForApi
}

func (s *Service) evmNativeNftFee(id string, diamondRouter client.DiamondRouter) (*pricing.NonFungibleFee, error) {
	fee, ok := s.GetHederaNftFee(id)
	if !ok {
		return nil, fmt.Errorf("could not get fee for asset %s", id)
	}

	nftFee := &pricing.NonFungibleFee{
		Fee: decimal.NewFromInt(fee),
	}

	paymentToken, err := diamondRouter.Erc721Payment(&bind.CallOpts{}, common.HexToAddress(id))
	if err != nil {
		s.logger.Errorf("Failed to get payment token for asset %s. Error [%s]", id, err)
		return nil, err
	}

	nftFee.PaymentToken = paymentToken.String()
	nftFee.IsNative = true

	return nftFee, nil
}

func (s *Service) evmWrappedNftFee(id string, diamondRouter client.DiamondRouter) (*pricing.NonFungibleFee, error) {
	paymentToken, err := diamondRouter.Erc721Payment(&bind.CallOpts{}, common.HexToAddress(id))
	if err != nil {
		s.logger.Errorf("Failed to get payment token for asset %s. Error [%s]", id, err)
		return nil, err
	}

	fee, err := diamondRouter.Erc721Fee(&bind.CallOpts{}, common.HexToAddress(id))
	if err != nil {
		s.logger.Errorf("Failed to get fee for asset %s. Error [%s]", id, err)
		return nil, err
	}

	return &pricing.NonFungibleFee{
		IsNative:     false,
		PaymentToken: paymentToken.String(),
		Fee:          decimal.NewFromBigInt(fee, 0),
	}, nil
}

func (s *Service) hederaNativeNftFee(id string, networkId uint64) (*pricing.NonFungibleFee, error) {
	fee, ok := s.hederaNftFees[id]
	if !ok {
		errMsg := fmt.Sprintf("No fee found for NFT [%s] on network [%d]", id, networkId)
		s.logger.Errorf(errMsg)
		return nil, errors.New(errMsg)
	}

	return &pricing.NonFungibleFee{
		IsNative:     true,
		Fee:          decimal.NewFromInt(fee),
		PaymentToken: constants.Hbar,
	}, nil
}

func (s *Service) GetMinAmountsForAPI() map[uint64]map[string]string {
	s.minAmountsForApiMutex.RLock()
	defer s.minAmountsForApiMutex.RUnlock()

	return s.minAmountsForApi
}

func (s *Service) GetHederaNftFee(token string) (int64, bool) {
	s.tokenPriceInfoMutex.RLock()
	defer s.tokenPriceInfoMutex.RUnlock()

	fee, exists := s.hederaNftFees[token]
	return fee, exists
}

func (s *Service) loadStaticMinAmounts(bridgeConfig *config.Bridge) {
	for networkId, minAmountsByTokenAddress := range bridgeConfig.MinAmounts {
		for tokenAddress, minAmount := range minAmountsByTokenAddress {
			s.tokensPriceInfo[networkId][tokenAddress] = pricing.TokenPriceInfo{
				MinAmountWithFee: minAmount,
				DefaultMinAmount: minAmount,
			}
			s.minAmountsForApi[networkId][tokenAddress] = minAmount.String()
		}
	}
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

	defaultMinAmount := s.tokensPriceInfo[constants.HederaNetworkId][constants.Hbar].DefaultMinAmount
	previousUsdPrice := s.tokensPriceInfo[constants.HederaNetworkId][constants.Hbar].UsdPrice

	// Use the cached priceInUsd in case the price fetching failed
	if priceInUsd.Cmp(decimal.NewFromFloat(0.0)) == 0 {
		priceInUsd = previousUsdPrice
		s.logger.Warnf("Using the cached price for [%s]", constants.Hbar)
	}

	minAmountWithFee, err := s.calculateMinAmountWithFee(s.hbarNativeAsset, s.hbarFungibleAssetInfo.Decimals, priceInUsd)
	if err != nil {
		return err
	}

	tokenPriceInfo := pricing.TokenPriceInfo{
		UsdPrice:         priceInUsd,
		MinAmountWithFee: minAmountWithFee,
		DefaultMinAmount: defaultMinAmount,
	}

	err = s.updatePriceInfoContainers(s.hbarNativeAsset, tokenPriceInfo)
	if err != nil {
		return fmt.Errorf("Failed to update price info containers. Error: [%s]", err)
	}

	s.updateHederaNftDynamicFeesBasedOnHbar(priceInUsd, s.hbarFungibleAssetInfo.Decimals)
	err = s.fetchAndUpdateNftFeesForApi()
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) calculateMinAmountWithFee(nativeAsset *asset.NativeAsset, decimals uint8, priceInUsd decimal.Decimal) (minAmountWithFee *big.Int, err error) {
	if priceInUsd.Cmp(decimal.NewFromFloat(0.0)) <= 0 {
		return nil, fmt.Errorf("price in USD [%s] for NativeAsset [%s] must be a positive number", priceInUsd.String(), nativeAsset.Asset)
	}
	if nativeAsset.MinFeeAmountInUsd.Equal(decimal.NewFromFloat(0.0)) {
		return big.NewInt(0), nil
	}

	feePercentageBigInt := big.NewInt(nativeAsset.FeePercentage)
	if feePercentageBigInt.Cmp(big.NewInt(0)) <= 0 {
		return nil, fmt.Errorf("FeePercentage [%d] for NativeAsset [%s] must be a positive number", nativeAsset.FeePercentage, nativeAsset.Asset)
	}
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
		defaultMinAmount := tokenPriceInfo.DefaultMinAmount
		wrappedMinAmountWithFee, err := s.calculateMinAmountWithFee(nativeAsset, wrappedAssetInfo.Decimals, tokenPriceInfo.UsdPrice)
		if err != nil {
			s.logger.Errorf("Failed to calculate 'MinAmountWithFee' for asset: [%s]. Error: [%s]", wrappedToken, err)
			if defaultMinAmount.Cmp(big.NewInt(0)) <= 0 {
				return fmt.Errorf("Default min_amount for asset: [%s] is not set. Error: [%s]", wrappedToken, err)
			}
			s.logger.Infof("Updating MinAmountWithFee for [%s] to equal the defaultMinAmount", wrappedToken)
			wrappedMinAmountWithFee = defaultMinAmount
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
			defaultMinAmount := s.tokensPriceInfo[networkId][assetAddress].DefaultMinAmount
			previousUsdPrice := s.tokensPriceInfo[networkId][assetAddress].UsdPrice

			// Use the cached priceInUsd in case the price fetching failed
			if usdPrice.Cmp(decimal.NewFromFloat(0.0)) == 0 {
				usdPrice = previousUsdPrice
				s.logger.Warnf("Using the cached price for [%s]", assetAddress)
			}

			minAmountWithFee, err := s.calculateMinAmountWithFee(nativeAsset, fungibleAssetInfo.Decimals, usdPrice)
			if err != nil {
				s.logger.Errorf("Failed to calculate 'MinAmountWithFee' for asset: [%s]. Error: [%s]", assetAddress, err)
				if defaultMinAmount.Cmp(big.NewInt(0)) <= 0 {
					return fmt.Errorf("Default min_amount for asset: [%s] is not set. Error: [%s]", assetAddress, err)
				}
				s.logger.Infof("Updating MinAmountWithFee for [%s] to equal the defaultMinAmount", assetAddress)
				minAmountWithFee = defaultMinAmount
			}

			tokenPriceInfo := pricing.TokenPriceInfo{
				UsdPrice:         usdPrice,
				MinAmountWithFee: minAmountWithFee,
				DefaultMinAmount: defaultMinAmount,
			}

			err = s.updatePriceInfoContainers(nativeAsset, tokenPriceInfo)
			if err != nil {
				err = fmt.Errorf("Failed to update price info containers. Error: [%s]", err)
				return err
			}
		}
	}

	return nil
}

func (s *Service) updateHederaNftDynamicFeesBasedOnHbar(priceInUsd decimal.Decimal, decimals uint8) {
	for token, feeAmount := range s.hederaNftDynamicFees {
		nftDynamicFee := decimalHelper.ToLowestDenomination(feeAmount.Div(priceInUsd), decimals).Int64()
		s.hederaNftFees[token] = nftDynamicFee

		s.logger.Infof("Updating NFT Dynamic fee for [%s] to HBAR [%d], based on USD constant fee [%s] and HBAR/USD rate [%s]", token, nftDynamicFee, feeAmount, priceInUsd)
	}
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

func bridgeCfgEventHandler(e event.Event, assetsService service.Assets, mirrorNodeClient client.MirrorNode, coinGeckoClient client.Pricing, coinMarketCapClient client.Pricing, instance *Service) error {
	params, err := eventHelper.GetBridgeCfgUpdateEventParams(e)
	if err != nil {
		return err
	}

	newInstance := initialize(params.Bridge, assetsService, mirrorNodeClient, coinGeckoClient, coinMarketCapClient, params.RouterClients)
	*instance = *newInstance

	return nil
}

func initialize(bridgeConfig *config.Bridge, assetsService service.Assets, mirrorNodeClient client.MirrorNode, coinGeckoClient client.Pricing, coinMarketCapClient client.Pricing, diamondRouters map[uint64]client.DiamondRouter) *Service {
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
		nftFeesForApiMutex:    new(sync.RWMutex),
		assetsService:         assetsService,
		coinGeckoIds:          bridgeConfig.CoinGeckoIds,
		coinMarketCapIds:      bridgeConfig.CoinMarketCapIds,
		hbarFungibleAssetInfo: hbarFungibleAssetInfo,
		hbarNativeAsset:       hbarNativeAsset,
		hederaNftFees:         bridgeConfig.Hedera.NftConstantFees,
		hederaNftDynamicFees:  bridgeConfig.Hedera.NftDynamicFees,
		diamondRouters:        diamondRouters,
		logger:                logger,
	}

	instance.loadStaticMinAmounts(bridgeConfig)

	err := instance.FetchAndUpdateUsdPrices()
	if err != nil {
		panic(fmt.Sprintf("Failed to initially fetch USD prices. Error: [%s]", err.Error()))
	}

	return instance
}

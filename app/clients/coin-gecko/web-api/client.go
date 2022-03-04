package web_api

import (
	"fmt"
	coinGeckoHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/coin-gecko"
	httpHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/http"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

var (
	getSimplePriceHeaders = map[string]string{"Accepts": "application/json"}
)

type Client struct {
	apiCfg                config.CoinGeckoWebApi
	fullGetSimplePriceUrl string
	httpClient            *http.Client
	logger                *log.Entry
}

func NewClient(apiCfg config.CoinGeckoWebApi) *Client {
	return &Client{
		apiCfg:                apiCfg,
		httpClient:            new(http.Client),
		fullGetSimplePriceUrl: strings.Join([]string{apiCfg.BaseUrl, apiCfg.Endpoints.GetSimplePriceInUsd}, "/"),
		logger:                config.GetLoggerFor("CoinGecko Web API Client"),
	}
}

func (c *Client) GetUsdPrices(idsByNetworkAndAddress map[uint64]map[string]string) (pricesByNetworkAndAddress map[uint64]map[string]decimal.Decimal, err error) {
	pricesByNetworkAndAddress = make(map[uint64]map[string]decimal.Decimal)

	var ids []string
	for _, addressesWithIds := range idsByNetworkAndAddress {
		for _, id := range addressesWithIds {
			ids = append(ids, id)
		}
	}

	urlWithIds := fmt.Sprintf(c.fullGetSimplePriceUrl, strings.Join(ids, ","))
	responseBodyBytes, err := httpHelper.Get(c.httpClient, urlWithIds, getSimplePriceHeaders, c.logger)
	if err != nil {
		return pricesByNetworkAndAddress, err
	}

	parsedResponse, err := coinGeckoHelper.ParseGetSimplePriceResponse(responseBodyBytes)
	if err == nil {
		for networkId, addressesWithIds := range idsByNetworkAndAddress {
			pricesForCurrNetwork := make(map[string]decimal.Decimal)

			for address, id := range addressesWithIds {
				currPrice := decimal.NewFromFloat(parsedResponse[id].Usd)
				pricesForCurrNetwork[address] = currPrice
			}

			pricesByNetworkAndAddress[networkId] = pricesForCurrNetwork
		}
	}

	return pricesByNetworkAndAddress, err
}

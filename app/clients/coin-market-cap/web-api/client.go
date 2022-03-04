package web_api

import (
	"fmt"
	coinMarketCapHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/coin-market-cap"
	httpHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/http"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

var (
	apiKeyHeaderName       = "X-CMC_PRO_API_KEY"
	getLatestQuotesHeaders = map[string]string{
		"Accepts":        "application/json",
		apiKeyHeaderName: "SET_API_KEY",
	}
)

type Client struct {
	apiCfg                 config.CoinMarketCapWebApi
	fullGetLatestQuotesUrl string
	httpClient             *http.Client
	logger                 *log.Entry
}

func NewClient(apiCfg config.CoinMarketCapWebApi) *Client {
	return &Client{
		apiCfg:                 apiCfg,
		httpClient:             new(http.Client),
		fullGetLatestQuotesUrl: strings.Join([]string{apiCfg.BaseUrl, apiCfg.Endpoints.GetLatestQuotes}, "/"),
		logger:                 config.GetLoggerFor("CoinMarketCap Web API Client"),
	}
}

func (c *Client) GetUsdPrices(idsByNetworkAndAddress map[uint64]map[string]string) (pricesByNetworkAndAddress map[uint64]map[string]decimal.Decimal, err error) {
	index := 0
	ids := make([]string, len(idsByNetworkAndAddress))
	for _, addressesWithIds := range idsByNetworkAndAddress {
		for _, id := range addressesWithIds {
			ids[index] = id
			index += 1
		}
	}

	urlWithIds := fmt.Sprintf(c.fullGetLatestQuotesUrl, strings.Join(ids, ","))
	getLatestQuotesHeaders[apiKeyHeaderName] = c.apiCfg.ApiKey
	responseBodyBytes, err := httpHelper.Get(c.httpClient, urlWithIds, getLatestQuotesHeaders, c.logger)
	if err != nil {
		return pricesByNetworkAndAddress, err
	}

	parsedResponse, err := coinMarketCapHelper.ParseGetLatestQuotesResponse(responseBodyBytes)
	if err == nil {
		for networkId, addressesWithIds := range idsByNetworkAndAddress {
			pricesForCurrNetwork := make(map[string]decimal.Decimal)

			for address, id := range addressesWithIds {
				currPrice := decimal.NewFromFloat(parsedResponse.Data[id].Quote.USD.Price)
				pricesForCurrNetwork[address] = currPrice
			}
			pricesByNetworkAndAddress[networkId] = pricesForCurrNetwork
		}
	}

	return pricesByNetworkAndAddress, err
}

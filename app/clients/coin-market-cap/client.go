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

package coin_market_cap

import (
	"fmt"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	httpHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/http"
	coinMarketCapModel "github.com/limechain/hedera-eth-bridge-validator/app/model/coin-market-cap"
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
	GetLatestQuotesEndpoint = "quotes/latest?id=%v"
)

type Client struct {
	apiCfg                 config.CoinMarketCap
	fullGetLatestQuotesUrl string
	httpClient             client.HttpClient
	logger                 *log.Entry
}

func NewClient(apiCfg config.CoinMarketCap) *Client {
	return &Client{
		apiCfg:                 apiCfg,
		httpClient:             new(http.Client),
		fullGetLatestQuotesUrl: strings.Join([]string{apiCfg.ApiAddress, GetLatestQuotesEndpoint}, ""),
		logger:                 config.GetLoggerFor("CoinMarketCap Client"),
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

	urlWithIds := fmt.Sprintf(c.fullGetLatestQuotesUrl, strings.Join(ids, ","))
	getLatestQuotesHeaders[apiKeyHeaderName] = c.apiCfg.ApiKey
	var parsedResponse coinMarketCapModel.CoinMarketCapResponse

	var statusCode int
	err = httpHelper.Get(c.httpClient, urlWithIds, getLatestQuotesHeaders, &parsedResponse, c.logger, &statusCode)
	if err != nil {
		return pricesByNetworkAndAddress, err
	}

	if statusCode != http.StatusOK {
		err = fmt.Errorf("Coin Market Cap responded with [%v]", statusCode)
		return pricesByNetworkAndAddress, err
	}

	for networkId, addressesWithIds := range idsByNetworkAndAddress {
		pricesForCurrNetwork := make(map[string]decimal.Decimal)

		for address, id := range addressesWithIds {
			currPrice := decimal.NewFromFloat(parsedResponse.Data[id].Quote.USD.Price)
			pricesForCurrNetwork[address] = currPrice
		}
		pricesByNetworkAndAddress[networkId] = pricesForCurrNetwork
	}

	return pricesByNetworkAndAddress, err
}

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
 * imitations under the License.
 */

package constants

import (
	coinMarketCapModel "github.com/limechain/hedera-eth-bridge-validator/app/model/coin-market-cap"
	"github.com/limechain/hedera-eth-bridge-validator/config"
)

var (
	CoinMarketCapApiConfig = config.CoinMarketCap{
		ApiAddress: "https://pro-api.coinmarketcap.com/v2/cryptocurrency/",
		ApiKey:     "0f56c8c8-4cde-4432-80f2-11f1623e5149",
	}

	//	EthereumCoinGeckoId: coin_gecko.PriceResult{Usd: EthereumNativeTokenPriceInUsdFloat64},
	//HbarCoinGeckoId:     coin_gecko.PriceResult{Usd: HbarNativeTokenPriceInUsdFloat64},

	CoinMarketCapResponse = coinMarketCapModel.CoinMarketCapResponse{
		Status: coinMarketCapModel.Status{
			ErrorCode: 0,
		},
		Data: map[string]coinMarketCapModel.TokenInfo{
			EthereumCoinMarketCapId: {
				Quote: coinMarketCapModel.Quote{
					USD: coinMarketCapModel.Usd{
						Price: EthereumNativeTokenPriceInUsdFloat64,
					},
				},
			},
			HbarCoinMarketCapId: {
				Quote: coinMarketCapModel.Quote{
					USD: coinMarketCapModel.Usd{
						Price: HbarNativeTokenPriceInUsdFloat64,
					},
				},
			},
		},
	}
)

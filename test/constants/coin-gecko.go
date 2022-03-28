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
	coin_gecko "github.com/limechain/hedera-eth-bridge-validator/app/model/coin-gecko"
	"github.com/limechain/hedera-eth-bridge-validator/config"
)

var (
	CoinGeckoApiConfig = config.CoinGecko{
		ApiAddress: "https://api.coingecko.com/api/v3/",
	}

	SimplePriceResponse = coin_gecko.SimplePriceResponse{
		EthereumCoinGeckoId: coin_gecko.PriceResult{Usd: EthereumNativeTokenPriceInUsdFloat64},
		HbarCoinGeckoId:     coin_gecko.PriceResult{Usd: HbarNativeTokenPriceInUsdFloat64},
	}
)

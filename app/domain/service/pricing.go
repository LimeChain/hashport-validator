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

package service

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/model/pricing"
)

type Pricing interface {
	// GetTokenPriceInfo gets price for token with the passed networkId and tokenAddressOrId
	GetTokenPriceInfo(networkId uint64, tokenAddressOrId string) (priceInfo pricing.TokenPriceInfo, exist bool)
	// FetchAndUpdateUsdPrices fetches all prices from the Web APIs and updates them in the mapping
	FetchAndUpdateUsdPrices() error
	// GetMinAmountsForAPI getting all prices by networkId
	GetMinAmountsForAPI() map[uint64]map[string]string
	// GetHederaNftFee returns the nft fee for Hedera NFTs based on token id
	GetHederaNftFee(token string) (int64, bool)

	// FetchAndUpdateNftFeesForApi fetches the fees for porting/burning NFTs
	fetchAndUpdateNftFeesForApi() error

	NftFees() map[uint64]map[string]pricing.NonFungibleFee
}

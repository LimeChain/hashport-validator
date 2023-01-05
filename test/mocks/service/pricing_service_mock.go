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
	"github.com/stretchr/testify/mock"
)

type MockPricingService struct {
	mock.Mock
}

// GetTokenPriceInfo gets price for token with the passed networkId and tokenAddressOrId
func (mas *MockPricingService) GetTokenPriceInfo(networkId uint64, tokenAddressOrId string) (priceInfo pricing.TokenPriceInfo, exist bool) {
	args := mas.Called(networkId, tokenAddressOrId)
	priceInfo = args.Get(0).(pricing.TokenPriceInfo)
	exist = args.Get(1).(bool)

	return priceInfo, exist
}

// FetchAndUpdateUsdPrices fetches all prices from the Web APIs and updates them in the mapping
func (mas *MockPricingService) FetchAndUpdateUsdPrices() error {
	args := mas.Called()

	return args.Error(0)
}

// GetMinAmountsForAPI getting all prices by networkId
func (mas *MockPricingService) GetMinAmountsForAPI() map[uint64]map[string]string {
	args := mas.Called()
	result := args.Get(0).(map[uint64]map[string]string)

	return result
}

// GetHederaNftFee returns the nft fee for Hedera NFTs based on token id
func (mas *MockPricingService) GetHederaNftFee(token string) (int64, bool) {
	args := mas.Called(token)
	fee := args.Get(0).(int64)
	exist := args.Get(1).(bool)

	return fee, exist
}

// GetHederaNftPrevFee returns the previous nft fee for Hedera NFTs based on token id
func (mas *MockPricingService) GetHederaNftPrevFee(token string) (int64, bool) {
	args := mas.Called(token)
	fee := args.Get(0).(int64)
	exist := args.Get(1).(bool)

	return fee, exist
}

func (mas *MockPricingService) FetchAndUpdateNftFeesForApi() error {
	args := mas.Called()
	return args.Error(0)
}

func (mas *MockPricingService) NftFees() map[uint64]map[string]pricing.NonFungibleFee {
	args := mas.Called()
	return args.Get(0).(map[uint64]map[string]pricing.NonFungibleFee)
}

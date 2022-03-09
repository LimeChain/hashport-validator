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

package pricing_client

import (
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
)

type MockPricingClient struct {
	mock.Mock
}

func (m *MockPricingClient) GetUsdPrices(idsByNetworkAndAddress map[uint64]map[string]string) (pricesByNetworkAndAddress map[uint64]map[string]decimal.Decimal, err error) {
	args := m.Called(idsByNetworkAndAddress)
	return args.Get(0).(map[uint64]map[string]decimal.Decimal), args.Error(1)
}

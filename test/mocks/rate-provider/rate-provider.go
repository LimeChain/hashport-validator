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

package rate_provider

import (
	"github.com/stretchr/testify/mock"
)

type MockExchangeRateProvider struct {
	mock.Mock
}

func (m *MockExchangeRateProvider) GetEthVsHbarRate() (float64, error) {
	args := m.Called()
	if args.Get(1) == nil {
		return args.Get(0).(float64), nil
	}
	return args.Get(0).(float64), args.Get(1).(error)
}

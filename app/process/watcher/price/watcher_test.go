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

package price

import (
	"errors"
	qi "github.com/limechain/hedera-eth-bridge-validator/app/domain/queue"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	watcher *Watcher
)

func Test_NewWatcher(t *testing.T) {
	setup()

	actualWatcher := NewWatcher(mocks.MPricingService)

	assert.Equal(t, watcher, actualWatcher)
}

func Test_watchIteration(t *testing.T) {
	setup()
	mocks.MPricingService.On("FetchAndUpdateUsdPrices").Return(nil)

	watcher.watchIteration()

	mocks.MPricingService.AssertCalled(t, "FetchAndUpdateUsdPrices")
}

func Test_watchIteration_Error(t *testing.T) {
	setup()
	mocks.MPricingService.On("FetchAndUpdateUsdPrices").Return(errors.New("some error"))

	watcher.watchIteration()

	mocks.MPricingService.AssertCalled(t, "FetchAndUpdateUsdPrices")
}

func Test_Watch(t *testing.T) {
	setup()
	mocks.MPricingService.On("FetchAndUpdateUsdPrices").Return(nil)

	watcher.Watch(qi.Queue(nil))
}

func setup() {
	mocks.Setup()

	watcher = &Watcher{
		pricingService: mocks.MPricingService,
		logger:         config.GetLoggerFor("Price Watcher"),
	}
}

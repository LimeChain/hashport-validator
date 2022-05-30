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

package coin_gecko

import (
	"errors"
	httpHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/http"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	testConstants "github.com/limechain/hedera-eth-bridge-validator/test/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"testing"
)

var (
	fullGetSimplePriceUrl = testConstants.CoinGeckoApiConfig.ApiAddress + GetSimplePriceInUsdEndpoint
	logger                = config.GetLoggerFor("Coin Gecko Client")
	c                     *Client
	nilErr                error
)

func Test_NewClient(t *testing.T) {
	setup()

	actual := NewClient(testConstants.CoinGeckoApiConfig)

	assert.Equal(t, c.apiCfg, actual.apiCfg)
	assert.Equal(t, c.fullGetSimplePriceUrl, actual.fullGetSimplePriceUrl)
	assert.Equal(t, c.logger, actual.logger)
}

func Test_GetUsdPrices(t *testing.T) {
	setup()
	encodedResponseReaderCloser, encodeErr := httpHelper.EncodeBodyContent(testConstants.SimplePriceResponse)
	if encodeErr != nil {
		t.Fatal(encodeErr)
	}
	mocks.MHTTPClient.On("Do", mock.Anything).Return(&http.Response{StatusCode: 200, Body: encodedResponseReaderCloser}, nilErr)

	result, err := c.GetUsdPrices(testConstants.CoinGeckoIds)

	assert.Nil(t, err)
	assert.Equal(t, testConstants.UsdPrices, result)
}

func Test_GetUsdPrices_Err(t *testing.T) {
	setup()

	mocks.MHTTPClient.On("Do", mock.Anything).Return(&http.Response{StatusCode: 500}, errors.New("internal server error"))
	_, err := c.GetUsdPrices(testConstants.CoinGeckoIds)

	assert.NotNil(t, err)
}

func setup() {
	mocks.Setup()

	c = &Client{
		apiCfg:                testConstants.CoinGeckoApiConfig,
		httpClient:            mocks.MHTTPClient,
		fullGetSimplePriceUrl: fullGetSimplePriceUrl,
		logger:                logger,
	}
}

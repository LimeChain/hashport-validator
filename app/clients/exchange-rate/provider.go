/*
 * Copyright 2021 LimeChain Ltd.
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

package exchangerate

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Provider struct representing the ExchangeRate provider client. Used for currency pairs (e.g HBAR/ETH)
type Provider struct {
	httpClient *http.Client
	rateURL    string
	coin       string
	currency   string
	rate       float64
}

// NewProvider creates new instance of the Exchange rate provider client
func NewProvider(coin string, currency string) *Provider {
	return &Provider{
		httpClient: &http.Client{},
		rateURL:    fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=%s", coin, currency),
		coin:       coin,
		currency:   currency,
	}
}

// GetEthVsHbarRate retrieves the current ETH/HBAR rate from the 3rd party API
func (erp Provider) GetEthVsHbarRate() (float64, error) {
	response, err := erp.httpClient.Get(erp.rateURL)
	if err != nil {
		return 0, err
	}

	bodyBytes, err := readResponseBody(response)
	if err != nil {
		return 0, err
	}

	var rates map[string]map[string]float32
	err = json.Unmarshal(bodyBytes, &rates)
	if err != nil {
		return 0, err
	}

	return float64(rates[erp.coin][erp.currency]), nil
}

// readResponseBody parses the http.Response into byte array
func readResponseBody(response *http.Response) ([]byte, error) {
	defer response.Body.Close()

	return ioutil.ReadAll(response.Body)
}

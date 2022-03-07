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
	"encoding/json"
	log "github.com/sirupsen/logrus"
)

type SimplePriceResponse map[string]PriceResult

type PriceResult struct {
	Usd float64 `json:"usd"`
}

func ParseGetSimplePriceResponse(responseBody []byte) (result SimplePriceResponse, err error) {
	err = json.Unmarshal(responseBody, &result)
	if err != nil {
		log.Errorf("Error while parsing CoinGecko Simple Price response Body. Error: [%v]", err)
	}

	return result, err
}

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

package coin_market_cap

import (
	"github.com/shopspring/decimal"
	"time"
)

type CoinMarketCapResponse struct {
	Status Status               `json:"Status"`
	Data   map[string]TokenInfo `json:"data"`
}

type Status struct {
	Timestamp    time.Time   `json:"timestamp"`
	ErrorCode    int         `json:"error_code"`
	ErrorMessage interface{} `json:"error_message"`
	Elapsed      int         `json:"elapsed"`
	CreditCount  int         `json:"credit_count"`
	Notice       interface{} `json:"notice"`
}

type TokenInfo struct {
	Id                            int             `json:"id"`
	Name                          string          `json:"name"`
	Symbol                        string          `json:"symbol"`
	Slug                          string          `json:"slug"`
	NumMarketPairs                int             `json:"num_market_pairs"`
	DateAdded                     time.Time       `json:"date_added"`
	Tags                          []Tag           `json:"tags"`
	MaxSupply                     int             `json:"max_supply"`
	CirculatingSupply             decimal.Decimal `json:"circulating_supply"`
	TotalSupply                   decimal.Decimal `json:"total_supply"`
	IsActive                      int             `json:"is_active"`
	Platform                      interface{}     `json:"platform"`
	CmcRank                       int             `json:"cmc_rank"`
	IsFiat                        int             `json:"is_fiat"`
	SelfReportedCirculatingSupply interface{}     `json:"self_reported_circulating_supply"`
	SelfReportedMarketCap         interface{}     `json:"self_reported_market_cap"`
	LastUpdated                   time.Time       `json:"last_updated"`
	Quote                         Quote           `json:"Quote"`
}

type Tag struct {
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

type Usd struct {
	Price                 float64   `json:"price"`
	Volume24H             float64   `json:"volume_24h"`
	VolumeChange24H       float64   `json:"volume_change_24h"`
	PercentChange1H       float64   `json:"percent_change_1h"`
	PercentChange24H      float64   `json:"percent_change_24h"`
	PercentChange7D       float64   `json:"percent_change_7d"`
	PercentChange30D      float64   `json:"percent_change_30d"`
	PercentChange60D      float64   `json:"percent_change_60d"`
	PercentChange90D      float64   `json:"percent_change_90d"`
	MarketCap             float64   `json:"market_cap"`
	MarketCapDominance    float64   `json:"market_cap_dominance"`
	FullyDilutedMarketCap float64   `json:"fully_diluted_market_cap"`
	LastUpdated           time.Time `json:"last_updated"`
}

type Quote struct {
	USD Usd `json:"Usd"`
}

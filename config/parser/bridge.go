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

package parser

/*
	Structs used to parse the bridge YAML configuration
*/
type Bridge struct {
	TopicId  string              `yaml:"topic_id" json:"topicId,omitempty"`
	Networks map[uint64]*Network `yaml:"networks" json:"networks,omitempty"`
}

type Network struct {
	Name                  string   `yaml:"name" json:"name,omitempty"`
	BridgeAccount         string   `yaml:"bridge_account" json:"bridgeAccount,omitempty"`
	PayerAccount          string   `yaml:"payer_account" json:"payerAccount,omitempty"`
	RouterContractAddress string   `yaml:"router_contract_address" json:"routerContractAddress,omitempty"`
	Members               []string `yaml:"members" json:"members,omitempty"`
	Tokens                Tokens   `yaml:"tokens" json:"tokens,omitempty"`
}

type Tokens struct {
	Fungible map[string]Token `yaml:"fungible" json:"fungible,omitempty"`
	Nft      map[string]Token `yaml:"nft" json:"nft,omitempty"`
}

type Token struct {
	Fee               int64             `yaml:"fee" json:"fee,omitempty"`                                 // Represent a constant fee for Non-Fungible tokens. Applies only for Hedera Native Tokens
	FeePercentage     int64             `yaml:"fee_percentage" json:"feePercentage,omitempty"`            // Represents a constant fee for Fungible Tokens. Applies only for Hedera Native Tokens
	MinFeeAmountInUsd string            `yaml:"min_fee_amount_in_usd" json:"minFeeAmountInUsd,omitempty"` // Represents a constant minimum fee amount in USD which is needed for the validator not to be on a loss
	Networks          map[uint64]string `yaml:"networks" json:"networks,omitempty"`
	CoinGeckoId       string            `yaml:"coin_gecko_id" json:"coinGeckoId,omitempty"`
	CoinMarketCapId   string            `yaml:"coin_market_cap_id" json:"coinMarketCapId,omitempty"`
}

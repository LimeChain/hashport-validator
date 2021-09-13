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

package parser

/*
	Structs used to parse the bridge YAML configuration
*/
type Bridge struct {
	TopicId  string             `yaml:"topic_id"`
	Networks map[int64]*Network `yaml:"networks"`
}

type Network struct {
	BridgeAccount         string           `yaml:"bridge_account"`
	PayerAccount          string           `yaml:"payer_account"`
	RouterContractAddress string           `yaml:"router_contract_address"`
	Members               []string         `yaml:"members"`
	Tokens                map[string]Token `yaml:"tokens"`
}

type Token struct {
	FeePercentage int64            `yaml:"fee_percentage"`
	Networks      map[int64]string `yaml:"networks"`
}

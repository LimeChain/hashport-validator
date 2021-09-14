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

import "time"

/*
	Structs used to parse the node YAML configuration
*/
type Node struct {
	Database  Database `yaml:"database"`
	Clients   Clients  `yaml:"clients"`
	LogLevel  string   `yaml:"log_level"`
	Port      string   `yaml:"port"`
	Validator bool     `yaml:"validator"`
}

type Database struct {
	Host     string `yaml:"host" env:"VALIDATOR_DATABASE_HOST"`
	Name     string `yaml:"name"`
	Password string `yaml:"password"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
}

type Clients struct {
	Evm        map[int64]Evm `yaml:"evm"`
	Hedera     Hedera        `yaml:"hedera"`
	MirrorNode MirrorNode    `yaml:"mirror_node"`
}

type Evm struct {
	BlockConfirmations uint64 `yaml:"block_confirmations"`
	NodeUrl            string `yaml:"node_url"`
	PrivateKey         string `yaml:"private_key"`
	StartBlock         int64  `yaml:"start_block"`
}

type Hedera struct {
	Operator       Operator `yaml:"operator"`
	Network        string   `yaml:"network"`
	StartTimestamp int64    `yaml:"start_timestamp"`
}

type Operator struct {
	AccountId  string `yaml:"account_id"`
	PrivateKey string `yaml:"private_key"`
}

type MirrorNode struct {
	ClientAddress   string        `yaml:"client_address"`
	ApiAddress      string        `yaml:"api_address"`
	PollingInterval time.Duration `yaml:"polling_interval"`
}

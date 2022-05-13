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

package setup

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	e2eParser "github.com/limechain/hedera-eth-bridge-validator/e2e/setup/parser"
)

const (
	// The configuration file for the e2e tests. Placed at ./e2e/setup/application.yml
	e2eConfigPath       = "setup/application.yml"
	e2eBridgeConfigPath = "setup/bridge.yml"
)

// Config used to load and parse from application.yml
type Config struct {
	Hedera         Hedera
	EVM            map[uint64]config.Evm
	Tokens         e2eParser.Tokens
	ValidatorUrl   string
	Bridge         parser.Bridge
	AssetMappings  service.Assets
	FeePercentages map[string]int64
	NftFees        map[string]int64
}

// Hedera props from the application.yml
type Hedera struct {
	NetworkType       string
	BridgeAccount     string
	Members           []string
	TopicID           string
	Sender            Sender
	DbValidationProps []config.Database
	MirrorNode        config.MirrorNode
}

// Sender props from the application.yml
type Sender struct {
	Account    string
	PrivateKey string
}

type Receiver struct {
	Account string
}

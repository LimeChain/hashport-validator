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

package config

import (
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
)

type Bridge struct {
	TopicId string
	Hedera  *BridgeHedera
	EVMs    map[int64]BridgeEvm
	Assets  Assets
}

type BridgeHedera struct {
	BridgeAccount  string
	PayerAccount   string
	Members        []string
	Tokens         map[string]HederaToken
	FeePercentages map[string]int64
}

type HederaToken struct {
	FeePercentage int64
	MinAmount     string
	Networks      map[int64]string
}

type Token struct {
	MinAmount *big.Int
	Networks  map[int64]string
}

type BridgeEvm struct {
	RouterContractAddress string
	Tokens                map[string]Token
}

func NewBridge(bridge parser.Bridge) Bridge {
	config := Bridge{
		TopicId: bridge.TopicId,
		Hedera:  nil,
		EVMs:    make(map[int64]BridgeEvm),
		Assets:  LoadAssets(bridge.Networks),
	}
	for key, value := range bridge.Networks {
		if key == 0 {
			config.Hedera = &BridgeHedera{
				BridgeAccount: value.BridgeAccount,
				PayerAccount:  value.PayerAccount,
				Members:       value.Members,
				Tokens:        make(map[string]HederaToken),
			}

			for name, value := range value.Tokens {
				config.Hedera.Tokens[name] = HederaToken(value)
			}
			config.Hedera.FeePercentages = LoadHederaFeePercentages(value.Tokens)
			continue
		}
		config.EVMs[key] = BridgeEvm{
			RouterContractAddress: value.RouterContractAddress,
			Tokens:                make(map[string]Token),
		}
		for name, value := range value.Tokens {
			config.EVMs[key].Tokens[name] = Token{Networks: value.Networks}
		}
	}

	return config
}

func LoadHederaFeePercentages(tokens map[string]parser.Token) map[string]int64 {
	feePercentages := map[string]int64{}
	for token, value := range tokens {
		feePercentages[token] = value.FeePercentage
	}

	return feePercentages
}

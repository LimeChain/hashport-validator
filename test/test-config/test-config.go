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

package test_config

import "github.com/limechain/hedera-eth-bridge-validator/config"

var (
	TestConfig = config.Config{
		Validator: config.Validator{
			LogLevel:    "debug",
			RestApiOnly: false,
			Port:        "8080",
			Database: config.Database{
				Host:     "127.0.0.1",
				Name:     "validator",
				Password: "validator",
				Port:     "5432",
				Username: "validator",
			},
			Clients: config.Clients{
				Ethereum: config.Ethereum{
					NodeUrl:               "wss://ropsten.infura.io/ws/v3/8b64d65996d24dc0aae2e0c6029e5a9b",
					RouterContractAddress: "B5762f4159e7bFE24B5E7E9a2e829F535744d30e",
					BlockConfirmations:    5,
					PrivateKey:            "9f6da11eecc0fd7cb081d2aee88092ee3436397916c894ad6cd80a79009c0ded",
				},
				MirrorNode: config.MirrorNode{
					ClientAddress:   "",
					ApiAddress:      "",
					PollingInterval: 0,
				},
				Hedera: config.Hedera{
					NetworkType: "testnet",
					Operator: config.Operator{
						AccountId:  "0.0.478300",
						PrivateKey: "302e020100300506032b657004220420479934e1729d3a2a25f3cdec95862d247944635113b4f4a07ec44c5ff8ec0884",
					},
					BridgeAccount: "0.0.476139",
					PayerAccount:  "0.0.476139",
					TopicId:       "0.0.476138",
					FeePercentage: 0,
					Members:       []string{"0.0.123", "0.0.321", "0.0.231"},
				},
			},
			Recovery: config.Recovery{
				StartTimestamp: 5,
			},
		},
	}
)

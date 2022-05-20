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

package test_config

import (
	"math/big"

	"github.com/limechain/hedera-eth-bridge-validator/config"
	testConstants "github.com/limechain/hedera-eth-bridge-validator/test/constants"
)

var (
	TestConfig = config.Config{
		Node: config.Node{
			LogLevel:  "debug",
			Validator: false,
			Port:      "8080",
			Database: config.Database{
				Host:     "127.0.0.1",
				Name:     "validator",
				Password: "validator",
				Port:     "5432",
				Username: "validator",
			},
			Clients: config.Clients{
				Evm: map[uint64]config.Evm{
					3: {
						NodeUrl:            "wss://ropsten.infura.io/ws/v3/64364afbcf794ff9a00deabde636b7e1",
						BlockConfirmations: 5,
						PrivateKey:         "9f6da11eecc0fd7cb081d2aee88092ee3436397916c894ad6cd80a79009c0ded",
					},
				},
				MirrorNode: config.MirrorNode{
					ClientAddress:   "",
					ApiAddress:      "",
					PollingInterval: 0,
					QueryMaxLimit:   100,
				},
				Hedera: config.Hedera{
					Network: "testnet",
					Operator: config.Operator{
						AccountId:  "0.0.478300",
						PrivateKey: "302e020100300506032b657004220420479934e1729d3a2a25f3cdec95862d247944635113b4f4a07ec44c5ff8ec0885",
					},
					StartTimestamp: 5,
				},
				CoinGecko: config.CoinGecko{ApiAddress: "https://api.coingecko.com/api/v3"},
			},

			Monitoring: config.Monitoring{
				Enable:           true,
				DashboardPolling: 1,
			},
		},

		Bridge: &config.Bridge{
			Hedera: &config.BridgeHedera{
				BridgeAccount:   "0.0.578300",
				NftConstantFees: testConstants.HederaNftFees,
			},
			MinAmounts: map[uint64]map[string]*big.Int{
				testConstants.PolygonNetworkId: {
					testConstants.NetworkPolygonFungibleNativeToken: testConstants.PolygonNativeTokenMinAmountWithFee,
				},
			},
		},
	}
)

/*
 * Copyright 2022 LimeChain Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package constants

import (
	fee_policy "github.com/limechain/hedera-eth-bridge-validator/app/model/fee-policy"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
)

var (
	ParsedFeePolicyConfig = parser.FeePolicy{
		LegalEntities: map[string]*parser.LegalEntity{
			"Some LTD": &parser.LegalEntity{
				Addresses: []string{"0.0.101", "0.0.102", "0.0.103"},
				PolicyInfo: parser.PolicyInfo{
					FeeType:  constants.FeePolicyTypeFlat,
					Networks: []uint64{8001},
					Value:    2000,
				},
			},
		},
	}

	FeePolicyConfig = config.FeePolicyStorage{
		StoreAddresses: map[string]fee_policy.FeePolicy{
			"0.0.101": &fee_policy.FlatFeePolicy{Networks: []uint64{8001}, Value: 2000},
			"0.0.102": &fee_policy.FlatFeePolicy{Networks: []uint64{8001}, Value: 2000},
			"0.0.103": &fee_policy.FlatFeePolicy{Networks: []uint64{8001}, Value: 2000},
		},
	}
)

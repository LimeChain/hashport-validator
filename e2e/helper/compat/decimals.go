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

package compat

import (
	"math"

	"github.com/ethereum/go-ethereum/common"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/wtoken"
	"github.com/limechain/hedera-eth-bridge-validator/e2e/setup"
)

func RemoveDecimals(amount int64, asset common.Address, evm setup.EVMUtils) (int64, error) {
	evmAsset, err := wtoken.NewWtoken(asset, evm.EVMClient)
	if err != nil {
		return 0, err
	}

	decimals, err := evmAsset.Decimals(nil)
	if err != nil {
		return 0, err
	}

	adaptation := decimals - 8
	if adaptation > 0 {
		adapted := amount / int64(math.Pow10(int(adaptation)))
		return adapted, nil
	}
	return amount, nil
}

func AddDecimals(amount int64, asset common.Address, evm setup.EVMUtils) (int64, error) {
	evmAsset, err := wtoken.NewWtoken(asset, evm.EVMClient)
	if err != nil {
		return 0, err
	}

	decimals, err := evmAsset.Decimals(nil)
	if err != nil {
		return 0, err
	}
	adaptation := decimals - 8
	if adaptation > 0 {
		adapted := amount * int64(math.Pow10(int(adaptation)))
		return adapted, nil
	}
	return amount, nil
}

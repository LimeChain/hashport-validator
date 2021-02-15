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

package helper

import (
	"errors"
	"fmt"
	"math/big"
)

func ToBigInt(value string) (*big.Int, error) {
	amount := new(big.Int)
	amount, ok := amount.SetString(value, 10)
	if !ok {
		return nil, errors.New(fmt.Sprintf("Failed to parse amount [%s] to big integer.", amount))
	}

	return amount, nil
}

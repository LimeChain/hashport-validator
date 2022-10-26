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

package create

import (
	"github.com/hashgraph/hedera-sdk-go/v2"
)

func CreateNativeFungibleToken(client *hedera.Client, treasuryAccountId hedera.AccountID, name, symbol string, decimals uint, supply uint64, maxTransactionFee hedera.Hbar) (*hedera.TokenID, error) {
	createTokenTX, err := hedera.NewTokenCreateTransaction().
		SetTreasuryAccountID(treasuryAccountId).
		SetTokenName(name).
		SetTokenSymbol(symbol).
		SetDecimals(decimals).
		SetInitialSupply(supply).
		SetMaxTransactionFee(maxTransactionFee).
		Execute(client)
	if err != nil {
		return nil, err
	}
	receipt, err := createTokenTX.GetReceipt(client)
	if err != nil {
		return nil, err
	}
	return receipt.TokenID, nil
}

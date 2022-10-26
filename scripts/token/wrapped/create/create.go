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

func WrappedFungibleToken(client *hedera.Client, treasuryAccountId hedera.AccountID, adminKey hedera.PublicKey, supplyKey *hedera.KeyList, custodianKey []hedera.PrivateKey, name, symbol string, decimals uint, supply uint64) (*hedera.TokenID, error) {
	freezeTokenTX, err := hedera.NewTokenCreateTransaction().
		SetTreasuryAccountID(treasuryAccountId).
		SetAdminKey(adminKey).
		SetSupplyKey(supplyKey).
		SetTokenName(name).
		SetTokenSymbol(symbol).
		SetInitialSupply(supply).
		SetDecimals(decimals).
		FreezeWith(client)

	if err != nil {
		return nil, err
	}

	// add all keys
	for i := 0; i < len(custodianKey); i++ {
		freezeTokenTX = freezeTokenTX.Sign(custodianKey[i])
	}
	createTx, err := freezeTokenTX.Execute(client)
	if err != nil {
		return nil, err
	}
	receipt, err := createTx.GetReceipt(client)
	if err != nil {
		return nil, err
	}

	return receipt.TokenID, nil
}

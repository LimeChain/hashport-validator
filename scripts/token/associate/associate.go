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

package associate

import (
	"fmt"

	"github.com/hashgraph/hedera-sdk-go/v2"
)

func TokenToAccount(client *hedera.Client, token hedera.TokenID, accountId hedera.AccountID) (*hedera.TransactionReceipt, error) {
	associateTX, err := hedera.
		NewTokenAssociateTransaction().
		SetAccountID(accountId).
		SetTokenIDs(token).
		Execute(client)
	if err != nil {
		return nil, fmt.Errorf(
			"Failed to associate token id [%s] with account id [%s]. Error: [%s]",
			token.String(),
			accountId.String(),
			err,
		)
	}

	receipt, err := associateTX.GetReceipt(client)
	if err != nil {
		return nil, fmt.Errorf(
			"Failed to get receipt for associate token id [%s] with account id [%s]. Error: [%s]",
			token.String(),
			accountId.String(),
			err,
		)
	}

	return &receipt, nil
}

func TokenToAccountWithCustodianKey(client *hedera.Client, token hedera.TokenID, accountID hedera.AccountID, custodianKey []hedera.PrivateKey) (*hedera.TransactionReceipt, error) {
	freezedAssociateTX, err := hedera.
		NewTokenAssociateTransaction().
		SetAccountID(accountID).
		SetTokenIDs(token).
		SetMaxTransactionFee(hedera.NewHbar(10)).
		FreezeWith(client)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(custodianKey); i++ {
		freezedAssociateTX = freezedAssociateTX.Sign(custodianKey[i])
	}
	associateTX, err := freezedAssociateTX.Execute(client)
	if err != nil {
		return nil, err
	}
	receipt, err := associateTX.GetReceipt(client)
	if err != nil {
		return nil, err
	}
	return &receipt, nil
}

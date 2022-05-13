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

package expected

import (
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
)

func MirrorNodeExpectedTransfersForBurnEvent(members []hedera.AccountID, hederaClient *hedera.Client, bridgeAccount hedera.AccountID, asset string, amount, fee int64) []transaction.Transfer {
	total := amount + fee
	feePerMember := fee / int64(len(members))

	var expectedTransfers []transaction.Transfer
	expectedTransfers = append(expectedTransfers, transaction.Transfer{
		Account: bridgeAccount.String(),
		Amount:  -total,
	},
		transaction.Transfer{
			Account: hederaClient.GetOperatorAccountID().String(),
			Amount:  amount,
		})

	for _, member := range members {
		expectedTransfers = append(expectedTransfers, transaction.Transfer{
			Account: member.String(),
			Amount:  feePerMember,
		})
	}

	if asset != constants.Hbar {
		for i := range expectedTransfers {
			expectedTransfers[i].Token = asset
		}
	}

	return expectedTransfers
}

func MirrorNodeExpectedTransfersForLockEvent(hederaClient *hedera.Client, bridgeAccount hedera.AccountID, asset string, amount int64) []transaction.Transfer {
	expectedTransfers := []transaction.Transfer{
		{
			Account: bridgeAccount.String(),
			Amount:  -amount,
			Token:   asset,
		},
		{
			Account: hederaClient.GetOperatorAccountID().String(),
			Amount:  amount,
			Token:   asset,
		},
	}

	return expectedTransfers
}

func MirrorNodeExpectedTransfersForHederaTransfer(members []hedera.AccountID, bridgeAccount hedera.AccountID, asset string, fee int64) []transaction.Transfer {
	feePerMember := fee / int64(len(members))

	var expectedTransfers []transaction.Transfer
	expectedTransfers = append(expectedTransfers, transaction.Transfer{
		Account: bridgeAccount.String(),
		Amount:  -fee,
	})

	for _, member := range members {
		expectedTransfers = append(expectedTransfers, transaction.Transfer{
			Account: member.String(),
			Amount:  feePerMember,
		})
	}

	if asset != constants.Hbar {
		for i := range expectedTransfers {
			expectedTransfers[i].Token = asset
		}
	}

	return expectedTransfers
}

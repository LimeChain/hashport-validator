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
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/limechain/hedera-eth-bridge-validator/e2e/setup"
)

func MirrorNodeExpectedTransfersForBurnEvent(setupEnv *setup.Setup, asset string, amount, fee int64) []transaction.Transfer {
	total := amount + fee
	feePerMember := fee / int64(len(setupEnv.Members))

	var expectedTransfers []transaction.Transfer
	expectedTransfers = append(expectedTransfers, transaction.Transfer{
		Account: setupEnv.BridgeAccount.String(),
		Amount:  -total,
	},
		transaction.Transfer{
			Account: setupEnv.Clients.Hedera.GetOperatorAccountID().String(),
			Amount:  amount,
		})

	for _, member := range setupEnv.Members {
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

func MirrorNodeExpectedTransfersForLockEvent(setupEnv *setup.Setup, asset string, amount int64) []transaction.Transfer {
	expectedTransfers := []transaction.Transfer{
		{
			Account: setupEnv.BridgeAccount.String(),
			Amount:  -amount,
			Token:   asset,
		},
		{
			Account: setupEnv.Clients.Hedera.GetOperatorAccountID().String(),
			Amount:  amount,
			Token:   asset,
		},
	}

	return expectedTransfers
}

func MirrorNodeExpectedTransfersForHederaTransfer(setupEnv *setup.Setup, asset string, fee int64) []transaction.Transfer {
	feePerMember := fee / int64(len(setupEnv.Members))

	var expectedTransfers []transaction.Transfer
	expectedTransfers = append(expectedTransfers, transaction.Transfer{
		Account: setupEnv.BridgeAccount.String(),
		Amount:  -fee,
	})

	for _, member := range setupEnv.Members {
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

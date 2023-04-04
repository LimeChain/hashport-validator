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

package blacklist

import (
	"fmt"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/transaction"
)

func IsBlacklistedAccount(blackListedAccounts []string, account string) bool {
	for _, blacklisted := range blackListedAccounts {
		if blacklisted == account {
			return true
		}
	}
	return false
}

// Checks if the transaction contains any blacklisted accounts in any transfer
func CheckTxForBlacklistedAccounts(blackListedAccounts []string, tx transaction.Transaction) error {
	for i := range tx.Transfers {
		fmt.Printf("\nAcc: %v", tx.Transfers[i].Account)
		if IsBlacklistedAccount(blackListedAccounts, tx.Transfers[i].Account) {
			return fmt.Errorf("[%s], Acc:[%v] - Found blacklisted transfer", tx.TransactionID, tx.Transfers[i].Account)
		}
	}

	for i := range tx.TokenTransfers {
		fmt.Printf("\nTokenTransfers Acc: %v", tx.TokenTransfers[i].Account)
		if IsBlacklistedAccount(blackListedAccounts, tx.TokenTransfers[i].Account) {
			return fmt.Errorf("[%s], Acc: [%v] - Found blacklisted transfer", tx.TransactionID, tx.TokenTransfers[i].Account)
		}
	}

	for i := range tx.NftTransfers {
		fmt.Printf("\nNftTransfers Acc: %v", tx.NftTransfers[i].SenderAccountID)
		if IsBlacklistedAccount(blackListedAccounts, tx.NftTransfers[i].SenderAccountID) {
			return fmt.Errorf("[%s], Acc: [%v] - Found blacklisted transfer", tx.TransactionID, tx.NftTransfers[i].SenderAccountID)
		}
	}

	return nil
}

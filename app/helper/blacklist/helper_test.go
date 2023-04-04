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
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/transaction"
	"github.com/stretchr/testify/assert"
	"testing"
)

var blacklist = []string{"0.0.999", "0.0.333"}

func Test_IsBlacklistedAccount(t *testing.T) {
	blacklist := []string{"0x000000", "0x000001"}
	assert.True(t, IsBlacklistedAccount(blacklist, "0x000000"))
	assert.False(t, IsBlacklistedAccount(blacklist, "0x000002"))
}

func Test_CheckNFTTxForBlacklistedAccounts(t *testing.T) {
	tx := setupTX()
	tx.NftTransfers = []transaction.NftTransfer{
		{ReceiverAccountID: "0.0.111",
			SenderAccountID: "0.0.233",
			SerialNumber:    1,
			Token:           "0.0.21241241"},
	}

	assert.NoError(t, CheckTxForBlacklistedAccounts(blacklist, tx))
}

func Test_CheckNFTTxForBlacklistedAccounts_Fails(t *testing.T) {
	tx := setupTX()
	tx.NftTransfers = []transaction.NftTransfer{
		{ReceiverAccountID: "0.0.111",
			SenderAccountID: "0.0.333",
			SerialNumber:    1,
			Token:           "0.0.21241241"},
	}

	assert.Error(t, CheckTxForBlacklistedAccounts(blacklist, tx))
}

func Test_CheckTokenTxForBlacklistedAccounts(t *testing.T) {
	tx := setupTX()
	assert.NoError(t, CheckTxForBlacklistedAccounts(blacklist, tx))
}

func Test_CheckTokenTxForBlacklistedAccounts_Fails(t *testing.T) {
	tx := setupTX()
	tx.Transfers = []transaction.Transfer{
		{
			Account: "0.0.333",
			Amount:  303030303030303030,
			Token:   "0.0.21312",
		},
	}

	assert.Error(t, CheckTxForBlacklistedAccounts(blacklist, tx))
}

func Test_CheckHBARTxForBlacklistedAccounts(t *testing.T) {
	tx := setupTX()
	tx.TokenTransfers = []transaction.Transfer{
		{
			Account: "0.0.233",
			Amount:  303030303030303030,
			Token:   "0.0.21312",
		},
	}
	assert.NoError(t, CheckTxForBlacklistedAccounts(blacklist, tx))
}

func Test_CheckHBARTxForBlacklistedAccounts_Fails(t *testing.T) {
	tx := setupTX()
	tx.TokenTransfers = []transaction.Transfer{
		{
			Account: "0.0.333",
			Amount:  303030303030303030,
			Token:   "0.0.21312",
		},
	}
	assert.Error(t, CheckTxForBlacklistedAccounts(blacklist, tx))
}

func setupTX() transaction.Transaction {
	return transaction.Transaction{
		ConsensusTimestamp: "1631092491.483966000",
		TransactionID:      "0.0.111-1631092491-483966000",
		Transfers: []transaction.Transfer{
			{
				Account: "0.0.3231",
				Amount:  10,
				Token:   "HBAR",
			},
		},
	}
}

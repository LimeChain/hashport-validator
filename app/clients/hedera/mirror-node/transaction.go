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

package mirror_node

import (
	"errors"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	"strconv"
)

type (
	// Transaction struct used by the Hedera Mirror node REST API
	Transaction struct {
		ConsensusTimestamp   string `json:"consensus_timestamp"`
		TransactionHash      string `json:"transaction_hash"`
		ValidStartTimestamp  string `json:"valid_start_timestamp"`
		ChargedTxFee         int    `json:"charged_tx_fee"`
		MemoBase64           string `json:"memo_base64"`
		Result               string `json:"result"`
		Name                 string `json:"name"`
		MaxFee               string `json:"max_fee"`
		ValidDurationSeconds string `json:"valid_duration_seconds"`
		Node                 string `json:"node"`
		TransactionID        string `json:"transaction_id"`
		Transfers            []Transfer
	}
	// Transfer struct used by the Hedera Mirror node REST API
	Transfer struct {
		Account string `json:"account"`
		Amount  int64  `json:"amount"`
	}
	// Response struct used by the Hedera Mirror node REST API and returned once
	// account transactions are queried
	Response struct {
		Transactions []Transaction
		Status       `json:"_status"`
	}
)

// GetIncomingAmountFor returns the amount that is credited to the specified
// account for the given transaction
func (t Transaction) GetIncomingAmountFor(account string) (string, error) {
	for _, tr := range t.Transfers {
		if tr.Account == account {
			return strconv.Itoa(int(tr.Amount)), nil
		}
	}
	return "", errors.New("no incoming transfer found")
}

// GetLatestTxnConsensusTime iterates all transactions and returns the consensus timestamp of the latest one
func (r Response) GetLatestTxnConsensusTime() (int64, error) {
	var max int64 = 0
	for _, t := range r.Transactions {
		ts, err := timestamp.FromString(t.ConsensusTimestamp)
		if err != nil {
			return 0, err
		}
		if ts > max {
			max = ts
		}
	}
	return max, nil
}

// isNotFound traverses all Error messages and searches for Not Found message
func (r Response) isNotFound() bool {
	for _, m := range r.Messages {
		if m.IsNotFound() {
			return true
		}
	}
	return false
}

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
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"strconv"
)

type (
	// Transaction struct used by the Hedera Mirror node REST API
	Transaction struct {
		ConsensusTimestamp   string     `json:"consensus_timestamp"`
		EntityId             string     `json:"entity_id"`
		TransactionHash      string     `json:"transaction_hash"`
		ValidStartTimestamp  string     `json:"valid_start_timestamp"`
		ChargedTxFee         int        `json:"charged_tx_fee"`
		MemoBase64           string     `json:"memo_base64"`
		Result               string     `json:"result"`
		Name                 string     `json:"name"`
		MaxFee               string     `json:"max_fee"`
		ValidDurationSeconds string     `json:"valid_duration_seconds"`
		Node                 string     `json:"node"`
		Scheduled            bool       `json:"scheduled"`
		TransactionID        string     `json:"transaction_id"`
		Transfers            []Transfer `json:"transfers"`
		TokenTransfers       []Transfer `json:"token_transfers"`
	}
	// Transfer struct used by the Hedera Mirror node REST API
	Transfer struct {
		Account string `json:"account"`
		Amount  int64  `json:"amount"`
		// When retrieving ordinary hbar transfers, this field does not get populated
		Token string `json:"token_id"`
	}
	// Response struct used by the Hedera Mirror node REST API and returned once
	// account transactions are queried
	Response struct {
		Transactions []Transaction
		Status       `json:"_status"`
	}
	// Schedule struct used by the Hedera Mirror node REST API to return information
	// regarding a given Schedule entity
	Schedule struct {
		ConsensusTimestamp string `json:"consensus_timestamp"`
		CreatorAccountId   string `json:"creator_account_id"`
		ExecutedTimestamp  string `json:"executed_timestamp"`
		Memo               string `json:"memo"`
		PayerAccountId     string `json:"payer_account_id"`
		ScheduleId         string `json:"schedule_id"`
	}
)

// getIncomingAmountFor returns the amount that is credited to the specified
// account for the given transaction
func (t Transaction) getIncomingAmountFor(account string) (string, string, error) {
	for _, tr := range t.Transfers {
		if tr.Account == account {
			return strconv.Itoa(int(tr.Amount)), constants.Hbar, nil
		}
	}
	return "", "", errors.New("no incoming transfer found")
}

// getIncomingTokenAmountFor returns the token amount that is credited to the specified
// account for the given transaction
func (t Transaction) getIncomingTokenAmountFor(account string) (string, string, error) {
	for _, tr := range t.TokenTransfers {
		if tr.Account == account {
			return strconv.Itoa(int(tr.Amount)), tr.Token, nil
		}
	}
	return "", "", errors.New("no incoming token transfer found")
}

// GetIncomingTransfer returns the token amount OR the hbar amount that is credited to the specified
// account for the given transaction. It depends on getIncomingAmountFor() and getIncomingTokenAmountFor()
func (t Transaction) GetIncomingTransfer(account string) (string, string, error) {
	amount, asset, err := t.getIncomingTokenAmountFor(account)
	if err == nil {
		return amount, asset, err
	}

	amount, asset, err = t.getIncomingAmountFor(account)
	if err == nil {
		return amount, asset, err
	}

	return amount, asset, err
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

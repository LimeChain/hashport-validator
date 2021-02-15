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

package transaction

type (
	HederaTransaction struct {
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
	Transfer struct {
		Account string `json:"account"`
		Amount  int64  `json:"amount"`
	}
	HederaTransactions struct {
		Transactions []HederaTransaction
	}
)

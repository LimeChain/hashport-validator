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

package mirror_node

import "github.com/shopspring/decimal"

type UpdatedFileRateData struct {
	CurrentRate decimal.Decimal
	NextRate    decimal.Decimal
}

type TransactionsResponse struct {
	Transactions []Transaction     `json:"transactions"`
	Links        map[string]string `json:"links"`
}

type Transaction struct {
	Bytes                    interface{}   `json:"bytes"`
	ChargedTxFee             int           `json:"charged_tx_fee"`
	ConsensusTimestamp       string        `json:"consensus_timestamp"`
	EntityId                 string        `json:"entity_id"`
	MaxFee                   string        `json:"max_fee"`
	MemoBase64               string        `json:"memo_base64"`
	Name                     string        `json:"name"`
	Node                     string        `json:"node"`
	Nonce                    int           `json:"nonce"`
	ParentConsensusTimestamp string        `json:"parent_consensus_timestamp"`
	Result                   string        `json:"result"`
	Scheduled                bool          `json:"scheduled"`
	TransactionHash          string        `json:"transaction_hash"`
	TransactionId            string        `json:"transaction_id"`
	Transfers                []interface{} `json:"transfers"`
	ValidDurationSeconds     string        `json:"valid_duration_seconds"`
	ValidStartTimestamp      string        `json:"valid_start_timestamp"`
}

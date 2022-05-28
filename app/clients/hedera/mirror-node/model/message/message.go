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

package message

type (
	InitialTransactionId struct {
		AccountId             string `json:"account_id"`
		Nonce                 uint64 `json:"nonce"`
		Scheduled             bool   `json:"scheduled"`
		TransactionValidStart string `json:"transaction_valid_start"`
	}

	ChunkInfo struct {
		InitialTransactionId InitialTransactionId `json:"initial_transaction_id"`
		Number               int64                `json:"number"`
		Total                int64                `json:"total"`
	}

	// Message struct used by the Hedera Mirror node REST API to represent Topic Message
	Message struct {
		ConsensusTimestamp string     `json:"consensus_timestamp"`
		TopicId            string     `json:"topic_id"`
		Contents           string     `json:"message"`
		RunningHash        string     `json:"running_hash"`
		SequenceNumber     int64      `json:"sequence_number"`
		ChunkInfo          *ChunkInfo `json:"chunk_info"`
	}
	// Messages struct used by the Hedera Mirror node REST API and returned once
	// Topic Messages are queried
	Messages struct {
		Messages []Message
	}
)

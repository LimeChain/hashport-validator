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

package transfer

import (
	"time"
)

// Transfer serves as a model between Transfer Watcher and Handler
type Transfer struct {
	TransactionId string    `json:"transactionId"`
	SourceChainId uint64    `json:"sourceChainId"`
	TargetChainId uint64    `json:"targetChainId"`
	NativeChainId uint64    `json:"nativeChainId"`
	SourceAsset   string    `json:"sourceAsset"`
	TargetAsset   string    `json:"targetAsset"`
	NativeAsset   string    `json:"nativeAsset"`
	Receiver      string    `json:"receiver"`
	Amount        string    `json:"amount,omitempty"`
	SerialNum     int64     `json:"serialNum,omitempty"`
	Metadata      string    `json:"metadata,omitempty"`
	IsNft         bool      `json:"isNft"`
	Originator    string    `json:"originator"`
	Timestamp     time.Time `json:"timestamp"`
	Fee           string    `json:"fee,omitempty"`
	Status        string    `json:"status"`
}

type Paged struct {
	Items      []*Transfer `json:"items"`
	TotalCount int64       `json:"totalCount"`
}

type PagedRequest struct {
	Page     uint64 `json:"page"`
	PageSize uint64 `json:"pageSize"`
	Filter   Filter `json:"filter"`
}

type Filter struct {
	Originator    string    `json:"originator"`
	Timestamp     time.Time `json:"timestamp"`
	TokenId       string    `json:"tokenId"`
	TransactionId string    `json:"transactionId"`
}

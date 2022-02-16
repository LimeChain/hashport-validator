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

// Transfer serves as a model between Transfer Watcher and Handler
type Transfer struct {
	TransactionId string
	SourceChainId uint64
	TargetChainId uint64
	NativeChainId uint64
	SourceAsset   string
	TargetAsset   string
	NativeAsset   string
	Receiver      string
	Amount        string
	SerialNum     int64
	Metadata      string
	IsNft         bool
	Timestamp     string
}

// New instantiates Transfer struct ready for submission to the handler
func New(txId string,
	sourceChainId, targetChainId, nativeChainId uint64,
	receiver, sourceAsset, targetAsset, nativeAsset, amount string) *Transfer {
	return &Transfer{
		TransactionId: txId,
		SourceChainId: sourceChainId,
		TargetChainId: targetChainId,
		NativeChainId: nativeChainId,
		SourceAsset:   sourceAsset,
		TargetAsset:   targetAsset,
		NativeAsset:   nativeAsset,
		Receiver:      receiver,
		Amount:        amount,
		IsNft:         false,
	}
}

// NewNft instantiates a Transfer, consisting of serial num and metadata for a given NFT
func NewNft(
	txId string,
	sourceChainId, targetChainId, nativeChainId uint64, receiver, sourceAsset, targetAsset, nativeAsset string, serialNum int64, metadata string) *Transfer {
	return &Transfer{
		TransactionId: txId,
		SourceChainId: sourceChainId,
		TargetChainId: targetChainId,
		NativeChainId: nativeChainId,
		SourceAsset:   sourceAsset,
		TargetAsset:   targetAsset,
		NativeAsset:   nativeAsset,
		Receiver:      receiver,
		SerialNum:     serialNum,
		Metadata:      metadata,
		IsNft:         true,
	}
}

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

package transfer

// Transfer serves as a model between Transfer Watcher and Handler
type Transfer struct {
	TransactionId string
	SourceChainId int64
	TargetChainId int64
	NativeChainId int64
	SourceAsset   string
	TargetAsset   string
	NativeAsset   string
	Receiver      string
	Amount        string
	RouterAddress string
	TargetChainID int64
}

// New instantiates Transfer struct ready for submission to the handler
func New(txId string,
	sourceChainId, targetChainId, nativeChainId int64,
	receiver, sourceAsset, targetAsset, nativeAsset, amount, routerAddress string) *Transfer {
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
		RouterAddress: routerAddress,
	}
}

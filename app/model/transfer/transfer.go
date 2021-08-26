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

type NativeTransfer struct {
	Transfer
}

type WrappedTransfer struct {
	Transfer
}

// Transfer serves as a model between Transfer Watcher and Handler
type Transfer struct {
	TransactionId string
	Receiver      string
	Amount        string
	NativeAsset   string
	WrappedAsset  string
	RouterAddress string
	TargetChainID int64
}

// NewNative instantiates Transfer struct ready for submission to the handler
func NewNative(txId, receiver, nativeAsset, wrappedAsset, amount, routerAddress string, targetChainID int64) *NativeTransfer {
	return &NativeTransfer{
		Transfer: Transfer{
			TransactionId: txId,
			Receiver:      receiver,
			Amount:        amount,
			NativeAsset:   nativeAsset,
			WrappedAsset:  wrappedAsset,
			RouterAddress: routerAddress,
			TargetChainID: targetChainID,
		},
	}
}

// NewNative instantiates Transfer struct ready for submission to the handler
func NewWrapped(txId, receiver, nativeAsset, wrappedAsset, amount, routerAddress string, targetChainID int64) *WrappedTransfer {
	return &WrappedTransfer{
		Transfer: Transfer{
			TransactionId: txId,
			Receiver:      receiver,
			Amount:        amount,
			NativeAsset:   nativeAsset,
			WrappedAsset:  wrappedAsset,
			RouterAddress: routerAddress,
			TargetChainID: targetChainID,
		},
	}
}

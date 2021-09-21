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

package service

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
)

// Transfers is the major service used for processing Transfers operations
type Transfers interface {
	// SanityCheckTransfer performs any validation required prior to handling the transaction
	// (memo, state proof verification)
	SanityCheckTransfer(tx model.Transaction) (int64, string, error)
	// InitiateNewTransfer Stores the incoming transfer message into the Database
	// aware of already processed transfers
	InitiateNewTransfer(tm transfer.Transfer) (*entity.Transfer, error)
	// ProcessNativeTransfer processes the native transfer message by signing the required
	// authorisation signature submitting it into the required HCS Topic
	ProcessNativeTransfer(tm transfer.Transfer) error
	// ProcessWrappedTransfer processes the wrapped transfer message by signing the required
	// authorisation signature submitting it into the required HCS Topic
	ProcessWrappedTransfer(tm transfer.Transfer) error
	// TransferData returns from the database the given transfer, its signatures and
	// calculates if its messages have reached super majority
	TransferData(txId string) (TransferData, error)
}

type TransferData struct {
	Recipient     string   `json:"recipient"`
	RouterAddress string   `json:"routerAddress"`
	Amount        string   `json:"amount"`
	SourceChainId int64    `json:"sourceChainId"`
	TargetChainId int64    `json:"targetChainId"`
	SourceAsset   string   `json:"sourceAsset"`
	NativeAsset   string   `json:"nativeAsset"`
	TargetAsset   string   `json:"wrappedAsset"`
	Signatures    []string `json:"signatures"`
	Majority      bool     `json:"majority"`
}

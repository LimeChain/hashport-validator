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

package encoding

import (
	model "github.com/limechain/hedera-eth-bridge-validator/proto"
	"google.golang.org/protobuf/proto"
)

type TransferMessage struct {
	*model.TransferMessage
}

// NewTransferMessage instantiates Transfer Message struct ready for submission to the handler
func NewTransferMessage(txId, ethAddress, asset, erc20ContractAddress, amount, fee, gasPriceGwei string) *TransferMessage {
	return &TransferMessage{
		&model.TransferMessage{
			TransactionId: txId,
			EthAddress:    ethAddress,
			Amount:        amount,
			Fee:           fee,
			GasPriceGwei:  gasPriceGwei,
			Asset:         asset,
			Erc20Address:  erc20ContractAddress,
		},
	}
}

// NewTransferMessageFromBytes instantiates new TransferMessage protobuf used internally by the Watchers/Handlers
func NewTransferMessageFromBytes(data []byte) (*TransferMessage, error) {
	transferMsg := &model.TransferMessage{}
	err := proto.Unmarshal(data, transferMsg)
	if err != nil {
		return nil, err
	}
	return &TransferMessage{transferMsg}, nil
}

// ToBytes marshals the underlying protobuf TransferMessage into bytes
func (tm *TransferMessage) ToBytes() ([]byte, error) {
	return proto.Marshal(tm.TransferMessage)
}

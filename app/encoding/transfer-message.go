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
	"errors"
	"fmt"
	model "github.com/limechain/hedera-eth-bridge-validator/proto"
	"google.golang.org/protobuf/proto"
)

type TransferMessage struct {
	*model.TransferMessage
}

// NewTransferMessage instantiates Transfer Message struct ready for submission to the handler
func NewTransferMessage(txId, receiver, nativeToken, wrappedToken, amount, txReimbursement, gasPrice string, executeEthTransaction bool) *TransferMessage {
	return &TransferMessage{
		&model.TransferMessage{
			TransactionId:         txId,
			Receiver:              receiver,
			Amount:                amount,
			TxReimbursement:       txReimbursement,
			GasPrice:              gasPrice,
			NativeToken:           nativeToken,
			WrappedToken:          wrappedToken,
			ExecuteEthTransaction: executeEthTransaction,
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

func NewTransferMessageFromInterface(data interface{}) (*TransferMessage, error) {
	transferMsg, ok := data.(*TransferMessage)
	if !ok {
		return nil, errors.New(fmt.Sprintf("unable to cast to TransferMessage data [%s]", data))
	}
	return transferMsg, nil
}

// ToBytes marshals the underlying protobuf TransferMessage into bytes
func (tm *TransferMessage) ToBytes() ([]byte, error) {
	return proto.Marshal(tm.TransferMessage)
}

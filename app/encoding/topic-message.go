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
	"encoding/base64"
	"github.com/golang/protobuf/proto"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	model "github.com/limechain/hedera-eth-bridge-validator/proto"
)

type TopicMessage struct {
	*model.TopicMessage
}

// NewTopicMessageFromBytes instantiates new TopicMessage protobuf used internally by the Watchers/Handlers
func NewTopicMessageFromBytes(data []byte) (*TopicMessage, error) {
	msg := &model.TopicMessage{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		return nil, err
	}
	return &TopicMessage{msg}, nil
}

// NewTopicMessageFromBytesWithTS instantiates new TopicMessage protobuf used internally by the Watchers/Handlers
func NewTopicMessageFromBytesWithTS(data []byte, ts int64) (*TopicMessage, error) {
	msg, err := NewTopicMessageFromBytes(data)
	if err != nil {
		return nil, err
	}
	msg.TransactionTimestamp = ts
	return msg, nil
}

// NewTopicMessageFromString instantiates new Topic Message protobuf from string `content` and `timestamp`
func NewTopicMessageFromString(data, ts string) (*TopicMessage, error) {
	t, err := timestamp.FromString(ts)
	if err != nil {
		return nil, err
	}

	bytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	return NewTopicMessageFromBytesWithTS(bytes, t)
}

// NewSignatureMessage instantiates Signature Message struct ready for submission to the Bridge Topic
func NewSignatureMessage(txId, ethAddress, amount, txReimbursement, gasPrice, signature string) *TopicMessage {
	topicMsg := &model.TopicMessage{
		Type: model.TopicMessageType_EthSignature,
		Message: &model.TopicMessage_TopicSignatureMessage{
			TopicSignatureMessage: &model.TopicEthSignatureMessage{
				TransactionId: txId,
				EthAddress:    ethAddress,
				Amount:        amount,
				Fee:           txReimbursement,
				Signature:     signature,
				GasPrice:      gasPrice,
			},
		},
	}
	return &TopicMessage{topicMsg}
}

// NewEthereumHashMessage instantiates Ethereum Transaction Hash Message struct ready for submission to the Bridge Topic
func NewEthereumHashMessage(txId, messageHash, ethereumTxHash string) *TopicMessage {
	topicMsg := &model.TopicMessage{
		Type: model.TopicMessageType_EthTransaction,
		Message: &model.TopicMessage_TopicEthTransactionMessage{
			TopicEthTransactionMessage: &model.TopicEthTransactionMessage{
				TransactionId: txId,
				Hash:          messageHash,
				EthTxHash:     ethereumTxHash,
			},
		},
	}
	return &TopicMessage{topicMsg}
}

// ToBytes marshals the underlying protobuf TopicMessage into bytes
func (tm *TopicMessage) ToBytes() ([]byte, error) {
	return proto.Marshal(tm.TopicMessage)
}

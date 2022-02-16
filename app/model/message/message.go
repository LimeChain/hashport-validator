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

import (
	"encoding/base64"
	"github.com/golang/protobuf/proto"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	model "github.com/limechain/hedera-eth-bridge-validator/proto"
)

// Message serves as a model between Topic Message Watcher and Handler
type Message struct {
	*model.TopicMessage
	TransactionTimestamp int64
}

// FromBytes instantiates new TopicMessage protobuf used internally by the Watchers/Handlers
func FromBytes(data []byte) (*Message, error) {
	msg := &model.TopicMessage{}
	err := proto.Unmarshal(data, msg)
	if err != nil {
		return nil, err
	}
	switch msg.Message.(type) {
	case *model.TopicMessage_NftSignatureMessage:
		return &Message{TopicMessage: msg}, nil
	case *model.TopicMessage_FungibleSignatureMessage:
		return &Message{TopicMessage: msg}, nil
	default: // try to parse it to backward compatible type
		oldFungibleMessage := &model.TopicEthSignatureMessage{}
		err = proto.Unmarshal(data, oldFungibleMessage)
		if err != nil {
			return nil, err
		}

		result := &model.TopicMessage{
			Message: &model.TopicMessage_FungibleSignatureMessage{
				FungibleSignatureMessage: oldFungibleMessage,
			},
		}
		return &Message{TopicMessage: result}, nil
	}
}

// FromBytesWithTS instantiates new TopicMessage protobuf used internally by the Watchers/Handlers
func FromBytesWithTS(data []byte, ts int64) (*Message, error) {
	msg, err := FromBytes(data)
	if err != nil {
		return nil, err
	}
	msg.TransactionTimestamp = ts
	return msg, nil
}

// FromString instantiates new Topic Message protobuf from string `content` and `timestamp`
func FromString(data, ts string) (*Message, error) {
	t, err := timestamp.FromString(ts)
	if err != nil {
		return nil, err
	}

	bytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	return FromBytesWithTS(bytes, t)
}

// NewFungibleSignature instantiates Signature Message struct ready for submission to the Bridge Topic
func NewFungibleSignature(topicMsg *model.TopicEthSignatureMessage) *Message {
	return &Message{TopicMessage: &model.TopicMessage{Message: &model.TopicMessage_FungibleSignatureMessage{FungibleSignatureMessage: topicMsg}}}
}

// NewNftSignature instantiates Signature Message struct ready for submission to the Bridge Topic
func NewNftSignature(topicMsg *model.TopicEthNftSignatureMessage) *Message {
	return &Message{TopicMessage: &model.TopicMessage{Message: &model.TopicMessage_NftSignatureMessage{NftSignatureMessage: topicMsg}}}
}

// ToBytes marshals the underlying protobuf Message into bytes
func (tm *Message) ToBytes() ([]byte, error) {
	return proto.Marshal(tm.TopicMessage)
}

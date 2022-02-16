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

package service

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
	"github.com/stretchr/testify/mock"
)

type MockMessageService struct {
	mock.Mock
}

func (m *MockMessageService) SignFungibleMessage(transfer transfer.Transfer) ([]byte, error) {
	args := m.Called(transfer)
	if args[1] == nil {
		return args[0].([]byte), nil
	}
	return args[0].([]byte), args[1].(error)
}

func (m *MockMessageService) SignNftMessage(transfer transfer.Transfer) ([]byte, error) {
	args := m.Called(transfer)
	if args[1] == nil {
		return args[0].([]byte), nil
	}
	return args[0].([]byte), args[1].(error)
}

// SanityCheckFungibleSignature performs any validation required prior handling the topic message
// (verifies metadata against the corresponding Transaction record)
func (m *MockMessageService) SanityCheckFungibleSignature(tm *proto.TopicEthSignatureMessage) (bool, error) {
	args := m.Called(tm)
	if args[1] == nil {
		return args[0].(bool), nil
	}
	return args[0].(bool), args[1].(error)
}

// SanityCheckNftSignature performs any validation required prior handling the topic message
// (verifies metadata against the corresponding Transaction record)
func (m *MockMessageService) SanityCheckNftSignature(tm *proto.TopicEthNftSignatureMessage) (bool, error) {
	args := m.Called(tm)
	if args[1] == nil {
		return args[0].(bool), nil
	}
	return args[0].(bool), args[1].(error)
}

// ProcessSignature processes the signature message, verifying and updating all necessary fields in the DB
func (m *MockMessageService) ProcessSignature(transferID, signature string, targetChainId, timestamp int64, authMsg []byte) error {
	args := m.Called(transferID, signature, targetChainId, timestamp, authMsg)
	if args[0] == nil {
		return nil
	}
	return args[0].(error)
}

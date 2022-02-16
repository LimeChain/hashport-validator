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
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
)

type Messages interface {
	// SanityCheckFungibleSignature performs any validation required prior handling the topic message
	// (verifies input data against the corresponding Transaction record)
	SanityCheckFungibleSignature(tm *proto.TopicEthSignatureMessage) (bool, error)
	// SanityCheckNftSignature performs any validation required prior handling the topic message
	// (verifies input data against the corresponding Transaction record)
	SanityCheckNftSignature(tm *proto.TopicEthNftSignatureMessage) (bool, error)
	// ProcessSignature processes the signature message, verifying and updating all necessary fields in the DB
	ProcessSignature(transferID, signature string, targetChainId uint64, timestamp int64, authMsg []byte) error
	// SignFungibleMessage signs a Fungible message based on Transfer
	SignFungibleMessage(transfer model.Transfer) ([]byte, error)
	// SignNftMessage signs an NFT messaged based on Transfer
	SignNftMessage(transfer model.Transfer) ([]byte, error)
}

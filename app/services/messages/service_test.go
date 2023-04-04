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

package messages

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/payload"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	serviceInstance  *Service
	topicId          hedera.TopicID
	ethSigners       map[uint64]service.Signer
	contractServices map[uint64]service.Contracts
	ethClients       map[uint64]client.EVM

	sourceChainId = constants.HederaNetworkId
	targetChainId = uint64(80001)
	asset         = "0.0.1"

	topicEthFungibleMessage = &proto.TopicEthSignatureMessage{
		SourceChainId: sourceChainId,
		TargetChainId: targetChainId,
		TransferID:    "some-transfer-id",
		Asset:         asset,
		Recipient:     "0xb083879B1e10C8476802016CB12cd2F25a896691",
		Amount:        "95",
		Signature:     "custom-signature",
	}

	topicFungibleMessage = message.Message{
		TopicMessage: &proto.TopicMessage{
			Message: &proto.TopicMessage_FungibleSignatureMessage{
				FungibleSignatureMessage: topicEthFungibleMessage,
			},
		},
	}

	topicEthNftMessage = &proto.TopicEthNftSignatureMessage{
		Recipient:     "0xb083879B1e10C8476802016CB12cd2F25a896691",
		TokenId:       42,
		Metadata:      "nft-metadata",
		Asset:         asset,
		TargetChainId: targetChainId,
		SourceChainId: sourceChainId,
		TransferID:    "some-transfer-id",
	}

	topicNftMessage = message.Message{
		TopicMessage: &proto.TopicMessage{
			Message: &proto.TopicMessage_NftSignatureMessage{
				NftSignatureMessage: topicEthNftMessage,
			},
		},
	}
)

func Test_NewService(t *testing.T) {
	setup()

	actualService := NewService(
		ethSigners,
		contractServices,
		mocks.MTransferRepository,
		mocks.MMessageRepository,
		mocks.MHederaMirrorClient,
		ethClients,
		"0.0.1",
		mocks.MAssetsService,
	)
	actualService.retryAttempts = 1

	assert.Equal(t, serviceInstance, actualService)
}

func Test_SanityCheckFungibleSignature_ShouldReturnError(t *testing.T) {
	setup()

	mocks.MTransferRepository.On("GetByTransactionId", topicEthFungibleMessage.TransferID).Return(nil, errors.New("some-error"))

	ok, err := serviceInstance.SanityCheckFungibleSignature(topicFungibleMessage.GetFungibleSignatureMessage())
	assert.False(t, ok)
	assert.NotNil(t, err)
}

func Test_awaitTransfer_ShouldDropTransfer(t *testing.T) {
	setup()

	mocks.MTransferRepository.On("GetByTransactionId", topicEthFungibleMessage.TransferID).Return((*entity.Transfer)(nil), nil)
	_, err := serviceInstance.awaitTransfer(topicFungibleMessage.GetFungibleSignatureMessage().TransferID)
	assert.NotNil(t, err)
}

func Test_SanityCheckFungibleSignature_ShouldReturnTrue(t *testing.T) {
	setup()

	transfer := &entity.Transfer{
		TransactionID: topicEthFungibleMessage.TransferID,
		SourceChainID: topicEthFungibleMessage.SourceChainId,
		TargetChainID: topicEthFungibleMessage.TargetChainId,
		NativeChainID: constants.HederaNetworkId,
		TargetAsset:   topicEthFungibleMessage.Asset,
		Amount:        "100",
		Fee:           "5",
		Receiver:      topicEthFungibleMessage.Recipient,
	}

	mocks.MTransferRepository.On("GetByTransactionId", topicEthFungibleMessage.TransferID).Return(transfer, nil)

	ok, err := serviceInstance.SanityCheckFungibleSignature(topicFungibleMessage.GetFungibleSignatureMessage())
	assert.True(t, ok)
	assert.Nil(t, err)
}

func Test_SanityCheckNftSignature_ShouldReturnError(t *testing.T) {
	setup()

	mocks.MTransferRepository.On("GetByTransactionId", topicEthNftMessage.TransferID).Return(nil, errors.New("some-error"))

	ok, err := serviceInstance.SanityCheckNftSignature(topicNftMessage.GetNftSignatureMessage())
	assert.False(t, ok)
	assert.NotNil(t, err)
}

func Test_SanityCheckNftSignature(t *testing.T) {
	setup()

	transfer := &entity.Transfer{
		Receiver:      topicEthNftMessage.Recipient,
		SerialNumber:  int64(topicEthNftMessage.TokenId),
		Metadata:      topicEthNftMessage.Metadata,
		TargetAsset:   topicEthNftMessage.Asset,
		TargetChainID: topicEthNftMessage.TargetChainId,
		SourceChainID: topicEthNftMessage.SourceChainId,
		TransactionID: topicEthNftMessage.TransferID,
	}

	mocks.MTransferRepository.On("GetByTransactionId", topicEthNftMessage.TransferID).Return(transfer, nil)

	ok, err := serviceInstance.SanityCheckNftSignature(topicNftMessage.GetNftSignatureMessage())
	assert.True(t, ok)
	assert.Nil(t, err)
}

func Test_SignFungibleMessage_ShouldReturnError(t *testing.T) {
	setup()

	tm := payload.Transfer{}

	bytes, err := serviceInstance.SignFungibleMessage(tm)
	assert.Nil(t, bytes)
	assert.NotNil(t, err)

	tm = payload.Transfer{
		SourceChainId: topicEthFungibleMessage.SourceChainId,
		TargetChainId: topicEthFungibleMessage.TargetChainId,
		TransactionId: topicEthFungibleMessage.TransferID,
		TargetAsset:   topicEthFungibleMessage.Asset,
		Receiver:      topicEthFungibleMessage.Recipient,
		Amount:        topicEthFungibleMessage.Amount,
	}

	mocks.MSignerService.On("Sign", mock.Anything).Return(nil, errors.New("some-error"))

	bytes, err = serviceInstance.SignFungibleMessage(tm)
	assert.Nil(t, bytes)
	assert.NotNil(t, err)
}

func Test_SignFungibleMessage(t *testing.T) {
	setup()

	tm := payload.Transfer{
		SourceChainId: topicEthFungibleMessage.SourceChainId,
		TargetChainId: topicEthFungibleMessage.TargetChainId,
		TransactionId: topicEthFungibleMessage.TransferID,
		TargetAsset:   topicEthFungibleMessage.Asset,
		Receiver:      topicEthFungibleMessage.Recipient,
		Amount:        topicEthFungibleMessage.Amount,
	}

	mocks.MSignerService.On("Sign", mock.Anything).Return([]byte{}, nil)

	bytes, err := serviceInstance.SignFungibleMessage(tm)
	assert.NotNil(t, bytes)
	assert.Nil(t, err)
}

func Test_SignNftMessage_ShouldReturnError(t *testing.T) {
	setup()

	tm := payload.Transfer{
		SourceChainId: topicEthNftMessage.SourceChainId,
		TargetChainId: topicEthNftMessage.TargetChainId,
		TransactionId: topicEthNftMessage.TransferID,
		TargetAsset:   topicEthNftMessage.Asset,
		Receiver:      topicEthNftMessage.Recipient,
		SerialNum:     int64(topicEthNftMessage.TokenId),
		IsNft:         true,
	}

	mocks.MSignerService.On("Sign", mock.Anything).Return(nil, errors.New("some-error"))

	bytes, err := serviceInstance.SignNftMessage(tm)
	assert.Nil(t, bytes)
	assert.NotNil(t, err)
}

func Test_SignNftMessage(t *testing.T) {
	setup()

	tm := payload.Transfer{
		SourceChainId: topicEthNftMessage.SourceChainId,
		TargetChainId: topicEthNftMessage.TargetChainId,
		TransactionId: topicEthNftMessage.TransferID,
		TargetAsset:   topicEthNftMessage.Asset,
		Receiver:      topicEthNftMessage.Recipient,
		SerialNum:     int64(topicEthNftMessage.TokenId),
		IsNft:         true,
	}

	mocks.MSignerService.On("Sign", mock.Anything).Return([]byte{}, nil)

	bytes, err := serviceInstance.SignNftMessage(tm)
	assert.NotNil(t, bytes)
	assert.Nil(t, err)
}

func Test_ProcessSignature(t *testing.T) {
	setup()

	err := serviceInstance.ProcessSignature(
		topicEthNftMessage.TransferID,
		"signature",
		topicEthNftMessage.TargetChainId,
		time.Now().UnixNano(),
		[]byte{},
	)

	assert.NotNil(t, err)
}

func setup() {
	mocks.Setup()

	ethSigners = map[uint64]service.Signer{
		80001: mocks.MSignerService,
	}

	contractServices = map[uint64]service.Contracts{
		80001: mocks.MBridgeContractService,
	}

	topicId = hedera.TopicID{
		Shard: 0,
		Realm: 0,
		Topic: 1,
	}

	ethClients = map[uint64]client.EVM{
		80001: mocks.MEVMClient,
	}

	serviceInstance = &Service{
		ethSigners:         ethSigners,
		contractServices:   contractServices,
		transferRepository: mocks.MTransferRepository,
		messageRepository:  mocks.MMessageRepository,
		topicID:            topicId,
		mirrorClient:       mocks.MHederaMirrorClient,
		ethClients:         ethClients,
		logger:             config.GetLoggerFor(fmt.Sprintf("Messages Service")),
		assetsService:      mocks.MAssetsService,
		retryAttempts:      1,
	}
}

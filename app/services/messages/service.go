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

package messages

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	ethhelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/ethereum"
	auth_message "github.com/limechain/hedera-eth-bridge-validator/app/model/auth-message"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	ethSigner          service.Signer
	contractsService   service.Contracts
	transferRepository repository.Transfer
	messageRepository  repository.Message
	topicID            hedera.TopicID
	hederaClient       client.HederaNode
	mirrorClient       client.MirrorNode
	ethClient          client.Ethereum
	logger             *log.Entry
}

func NewService(
	ethSigner service.Signer,
	contractsService service.Contracts,
	transferRepository repository.Transfer,
	messageRepository repository.Message,
	hederaClient client.HederaNode,
	mirrorClient client.MirrorNode,
	ethClient client.Ethereum,
	topicID string,
) *Service {
	tID, e := hedera.TopicIDFromString(topicID)
	if e != nil {
		panic(fmt.Sprintf("Invalid monitoring Topic ID [%s] - Error: [%s]", topicID, e))
	}

	return &Service{
		ethSigner:          ethSigner,
		contractsService:   contractsService,
		messageRepository:  messageRepository,
		transferRepository: transferRepository,
		logger:             config.GetLoggerFor(fmt.Sprintf("Messages Service")),
		topicID:            tID,
		hederaClient:       hederaClient,
		mirrorClient:       mirrorClient,
		ethClient:          ethClient,
	}
}

// SanityCheckSignature performs validation on the topic message metadata.
// Validates it against the Transaction Record metadata from DB
func (ss *Service) SanityCheckSignature(topicMessage message.Message) (bool, error) {
	// In case a topic message for given transfer is being processed before the actual transfer
	t, err := ss.awaitTransfer(topicMessage.TransferID)
	if err != nil {
		ss.logger.Errorf("[%s] - Failed to await incoming transfer. Error: [%s]", topicMessage.TransferID, err)
		return false, err
	}

	amount, err := strconv.ParseInt(t.Amount, 10, 64)
	if err != nil {
		ss.logger.Errorf("[%s] - Failed to parse transfer amount. Error [%s]", topicMessage.TransferID, err)
		return false, err
	}

	feeAmount, err := strconv.ParseInt(t.Fee.Amount, 10, 64)
	if err != nil {
		ss.logger.Errorf("[%s] - Failed to parse fee amount. Error [%s]", topicMessage.TransferID, err)
		return false, err
	}
	signedAmount := strconv.FormatInt(amount-feeAmount, 10)

	wrappedAsset, err := ss.contractsService.ToWrapped(t.NativeAsset)
	if err != nil {
		ss.logger.Errorf("[%s] - Could not parse native asset [%s] - Error: [%s]", t.TransactionID, t.NativeAsset, err)
		return false, err
	}

	match := t.Receiver == topicMessage.Receiver &&
		topicMessage.Amount == signedAmount &&
		topicMessage.WrappedAsset == wrappedAsset
	return match, nil
}

// ProcessSignature processes the signature message, verifying and updating all necessary fields in the DB
func (ss *Service) ProcessSignature(tsm message.Message) error {
	// Parse incoming message
	authMsgBytes, err := auth_message.EncodeBytesFrom(tsm.TransferID, tsm.WrappedAsset, tsm.Receiver, tsm.Amount)
	if err != nil {
		ss.logger.Errorf("[%s] - Failed to encode the authorisation signature. Error: [%s]", tsm.TransferID, err)
		return err
	}

	// Prepare Signature
	signatureBytes, signatureHex, err := ethhelper.DecodeSignature(tsm.GetSignature())
	if err != nil {
		ss.logger.Errorf("[%s] - Decoding Signature [%s] for TX failed. Error: [%s]", tsm.TransferID, tsm.GetSignature(), err)
		return err
	}
	authMessageStr := hex.EncodeToString(authMsgBytes)

	// Check for duplicated signature
	exists, err := ss.messageRepository.Exist(tsm.TransferID, signatureHex, authMessageStr)
	if err != nil {
		ss.logger.Errorf("[%s] - An error occurred while checking existence from DB. Error: [%s]", tsm.TransferID, err)
		return err
	}
	if exists {
		ss.logger.Errorf("[%s] - Signature already received", tsm.TransferID)
		return err
	}

	// Verify Signature
	address, err := ss.verifySignature(err, authMsgBytes, signatureBytes, tsm.TransferID, authMessageStr)
	if err != nil {
		return err
	}

	ss.logger.Debugf("[%s] - Successfully verified new Signature from [%s]", tsm.TransferID, address.String())

	// Persist in DB
	err = ss.messageRepository.Create(&entity.Message{
		TransferID:           tsm.TransferID,
		Signature:            signatureHex,
		Hash:                 authMessageStr,
		Signer:               address.String(),
		TransactionTimestamp: tsm.TransactionTimestamp,
	})
	if err != nil {
		ss.logger.Errorf("[%s] - Failed to save Transaction Message in DB with Signature [%s]. Error: [%s]", tsm.TransferID, signatureHex, err)
		return err
	}

	ss.logger.Infof("[%s] - Successfully processed Signature Message from [%s]", tsm.TransferID, address.String())
	return nil
}

func (ss *Service) verifySignature(err error, authMsgBytes []byte, signatureBytes []byte, transferID, authMessageStr string) (common.Address, error) {
	publicKey, err := crypto.Ecrecover(authMsgBytes, signatureBytes)
	if err != nil {
		ss.logger.Errorf("[%s] - Failed to recover public key. Hash [%s]. Error: [%s]", transferID, authMessageStr, err)
		return common.Address{}, err
	}
	unmarshalledPublicKey, err := crypto.UnmarshalPubkey(publicKey)
	if err != nil {
		ss.logger.Errorf("[%s] - Failed to unmarshall public key. Error: [%s]", transferID, err)
		return common.Address{}, err
	}
	address := crypto.PubkeyToAddress(*unmarshalledPublicKey)
	if !ss.contractsService.IsMember(address.String()) {
		ss.logger.Errorf("[%s] - Received Signature [%s] is not signed by Bridge member", transferID, authMessageStr)
		return common.Address{}, errors.New(fmt.Sprintf("signer is not signatures member"))
	}
	return address, nil
}

// awaitTransfer checks until given transfer is found
func (ss *Service) awaitTransfer(transferID string) (*entity.Transfer, error) {
	for {
		t, err := ss.transferRepository.GetWithFee(transferID)
		if err != nil {
			ss.logger.Errorf("[%s] - Failed to retrieve Transaction Record. Error: [%s]", transferID, err)
			return nil, err
		}

		if t != nil && t.Fee.TransactionID != "" {
			return t, nil
		}
		ss.logger.Debugf("[%s] - Transfer not yet added. Querying after 5 seconds", transferID)
		time.Sleep(5 * time.Second)
	}
}

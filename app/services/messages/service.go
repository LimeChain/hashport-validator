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
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/encoding"
	"github.com/limechain/hedera-eth-bridge-validator/app/encoding/auth-message"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper"
	ethhelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/ethereum"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
	"strings"
)

type Service struct {
	ethSigner          service.Signer
	contractsService   service.Contracts
	scheduler          service.Scheduler
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
	scheduler service.Scheduler,
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
		scheduler:          scheduler,
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
func (ss *Service) SanityCheckSignature(tm encoding.TopicMessage) (bool, error) {
	topicMessage := tm.GetTopicSignatureMessage()
	t, err := ss.transferRepository.GetByTransactionId(topicMessage.TransferID)
	if err != nil {
		ss.logger.Errorf("[%s] - Failed to retrieve Transaction Record. Error: [%s]", topicMessage.TransferID, err)
		return false, err
	}

	valid, erc20address := ss.contractsService.IsValidBridgeAsset(t.SourceAsset)
	if !valid {
		ss.logger.Errorf("[%s] - Provided Asset is not supported - [%s]", topicMessage.TransferID, t.SourceAsset)
		return false, err
	}

	match := t.Receiver == topicMessage.Receiver &&
		t.Amount == topicMessage.Amount &&
		t.TxReimbursement == topicMessage.TxReimbursement &&
		t.GasPrice == topicMessage.GasPrice &&
		topicMessage.TargetAsset == erc20address
	return match, nil
}

// ProcessSignature processes the signature message, verifying and updating all necessary fields in the DB
func (ss *Service) ProcessSignature(tm encoding.TopicMessage) error {
	// Parse incoming message
	tsm := tm.GetTopicSignatureMessage()
	authMsgBytes, err := auth_message.EncodeBytesFrom(tsm.TransferID, tsm.Receiver, tsm.TargetAsset, tsm.Amount, tsm.TxReimbursement, tsm.GasPrice)
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
		TransactionTimestamp: tm.TransactionTimestamp,
	})
	if err != nil {
		ss.logger.Errorf("[%s] - Failed to save Transaction Message in DB with Signature [%s]. Error: [%s]", tsm.TransferID, signatureHex, err)
		return err
	}

	ss.logger.Infof("[%s] - Successfully processed Signature Message from [%s]", tsm.TransferID, address.String())
	return nil
}

// ScheduleForSubmission computes the execution slot and schedules the Eth TX for submission
func (ss *Service) ScheduleEthereumTxForSubmission(transferID string) error {
	transfer, err := ss.transferRepository.GetByTransactionId(transferID)
	if err != nil {
		ss.logger.Errorf("[%s] - Failed to query transfer. Error: [%s].", transferID, err)
	}
	signatureMessages, err := ss.messageRepository.Get(transferID)
	if err != nil {
		ss.logger.Errorf("[%s] - Failed to query all Signature Messages. Error: [%s]", transferID, err)
		return err
	}

	slot, isFound := ss.computeExecutionSlot(signatureMessages)
	if !isFound {
		ss.logger.Debugf("[%s] - Operator [%s] has not been found as signer amongst the signatures collected.", transferID, ss.ethSigner.Address())
		return nil
	}

	amount := transfer.Amount
	txReimbursement := transfer.TxReimbursement
	ethAddress := transfer.Receiver
	messageHash := transfer.EthTxHash
	gasPriceWei := transfer.GasPrice
	signatures, err := getSignatures(signatureMessages)
	if err != nil {
		return err
	}

	ethereumMintTask := ss.prepareEthereumMintTask(transferID, ethAddress, amount, txReimbursement, gasPriceWei, signatures, messageHash)
	err = ss.scheduler.Schedule(transferID, signatureMessages[0].TransactionTimestamp, slot, ethereumMintTask)
	if err != nil {
		return err
	}
	return nil
}

// prepareEthereumMintTask returns the function to be executed for processing the
// Ethereum Mint transaction and HCS topic message with the ethereum TX hash after that
func (ss *Service) prepareEthereumMintTask(transferID string, ethAddress string, amount string, txReimbursement string, gasPriceWei string, signatures [][]byte, messageHash string) func() {
	ethereumMintTask := func() {
		// Submit and monitor Ethereum TX
		ethTransactor, err := ss.ethSigner.NewKeyTransactor(ss.ethClient.ChainID())
		if err != nil {
			ss.logger.Errorf("[%s] - Failed to establish key transactor. Error: [%s].", transferID, err)
			return
		}

		ethTransactor.GasPrice, err = helper.ToBigInt(gasPriceWei)
		if err != nil {
			ss.logger.Errorf("[%s] - Failed to parse provided gas price. Error: [%s].", transferID, err)
			return
		}

		ethTx, err := ss.contractsService.SubmitSignatures(ethTransactor, transferID, ethAddress, amount, txReimbursement, signatures)
		if err != nil {
			ss.logger.Errorf("[%s] - Failed to Submit Signatures. Error: [%s]", transferID, err)
			return
		}
		err = ss.transferRepository.UpdateEthTxSubmitted(transferID, ethTx.Hash().String())
		if err != nil {
			ss.logger.Errorf("[%s] - Failed to update status. Error: [%s].", transferID, err)
			return
		}
		ss.logger.Infof("[%s] - Submitted Ethereum Mint TX [%s]", transferID, ethTx.Hash().String())

		onEthTxSuccess, onEthTxRevert := ss.ethTxCallbacks(transferID, ethTx.Hash().String())
		ss.ethClient.WaitForTransaction(ethTx.Hash().String(), onEthTxSuccess, onEthTxRevert, func(err error) {})

		// Submit and monitor HCS Message for Ethereum TX Hash
		hcsTx, err := ss.submitEthTxTopicMessage(transferID, messageHash, ethTx.Hash().String())
		if err != nil {
			ss.logger.Errorf("[%s] - Failed to submit Ethereum TX Hash to Bridge Topic. Error: [%s].", transferID, err)
			return
		}
		err = ss.transferRepository.UpdateStatusEthTxMsgSubmitted(transferID)
		if err != nil {
			ss.logger.Errorf("[%s] - Failed to update status for. Error: [%s].", transferID, err)
			return
		}
		ss.logger.Infof("[%s] - Submitted Ethereum TX Hash [%s] to HCS. Transaction ID [%s].", transferID, ethTx.Hash().String(), hcsTx.String())

		onHcsMessageSuccess, onHcsMessageFail := ss.hcsTxCallbacks(transferID)
		ss.mirrorClient.WaitForTransaction(hcsTx.String(), onHcsMessageSuccess, onHcsMessageFail)

		ss.logger.Infof("[%s] - Successfully processed Ethereum Minting", transferID)
	}
	return ethereumMintTask
}

func getSignatures(messages []entity.Message) ([][]byte, error) {
	var signatures [][]byte

	for _, msg := range messages {
		signature, err := hex.DecodeString(msg.Signature)
		if err != nil {
			return nil, err
		}
		signatures = append(signatures, signature)
	}

	return signatures, nil
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

// computeExecutionSlot - computes the slot order in which the TX will execute
// Important! Transaction messages ARE expected to be sorted by ascending Timestamp
func (ss *Service) computeExecutionSlot(messages []entity.Message) (slot int64, isFound bool) {
	for i := 0; i < len(messages); i++ {
		if strings.ToLower(messages[i].Signer) == strings.ToLower(ss.ethSigner.Address()) {
			return int64(i), true
		}
	}

	return -1, false
}

func (ss *Service) submitEthTxTopicMessage(transferID, messageHash, ethereumTxHash string) (*hedera.TransactionID, error) {
	ethTxHashMessage := encoding.NewEthereumHashMessage(transferID, messageHash, ethereumTxHash)
	ethTxHashBytes, err := ethTxHashMessage.ToBytes()
	if err != nil {
		ss.logger.Errorf("[%s] - Failed to encode Eth TX Hash Message to bytes. Error: [%s]", transferID, err)
		return nil, err
	}

	return ss.hederaClient.SubmitTopicConsensusMessage(ss.topicID, ethTxHashBytes)
}

func (ss *Service) ethTxCallbacks(transferID, hash string) (onSuccess, onRevert func()) {
	onSuccess = func() {
		ss.logger.Infof("[%s] - Ethereum TX [%s] was successfully mined", transferID, hash)
		err := ss.transferRepository.UpdateEthTxMined(transferID)
		if err != nil {
			ss.logger.Errorf("[%s] - Failed to update status. Error [%s].", transferID, err)
			return
		}
	}

	onRevert = func() {
		ss.logger.Infof("[%s] - Ethereum TX [%s] reverted", transferID, hash)
		err := ss.transferRepository.UpdateEthTxReverted(transferID)
		if err != nil {
			ss.logger.Errorf("[%s] - Failed to update status. Error [%s].", transferID, err)
			return
		}
	}
	return onSuccess, onRevert
}

func (ss *Service) hcsTxCallbacks(txId string) (onSuccess, onFailure func()) {
	onSuccess = func() {
		ss.logger.Infof("[%s] - Ethereum TX Hash message was successfully mined", txId)
		err := ss.transferRepository.UpdateStatusEthTxMsgMined(txId)
		if err != nil {
			ss.logger.Errorf("Failed to update status for TX [%s]. Error [%s].", txId, err)
			return
		}
	}

	onFailure = func() {
		ss.logger.Infof("[%s] - Ethereum TX Hash message failed", txId)
		err := ss.transferRepository.UpdateStatusEthTxMsgFailed(txId)
		if err != nil {
			ss.logger.Errorf("Failed to update status for TX [%s]. Error [%s].", txId, err)
			return
		}
	}
	return onSuccess, onFailure
}

// VerifyEthereumTxAuthenticity performs the validation required prior handling the topic message
// (verifies the submitted TX against the required target contract and arguments passed)
func (ss *Service) VerifyEthereumTxAuthenticity(tm encoding.TopicMessage) (bool, error) {
	ethTxMessage := tm.GetTopicEthTransactionMessage()
	tx, _, err := ss.ethClient.GetClient().TransactionByHash(context.Background(), common.HexToHash(ethTxMessage.EthTxHash))
	if err != nil {
		ss.logger.Warnf("[%s] - Failed to get eth transaction by hash [%s]. Error [%s].", ethTxMessage.TransferID, ethTxMessage.EthTxHash, err)
		return false, err
	}

	// Verify Ethereum TX `to` property
	if strings.ToLower(tx.To().String()) != strings.ToLower(ss.contractsService.GetBridgeContractAddress().String()) {
		ss.logger.Debugf("[%s] - ETH TX [%s] - Failed authenticity - Different To Address [%s].", ethTxMessage.TransferID, ethTxMessage.EthTxHash, tx.To().String())
		return false, nil
	}
	// Verify Ethereum TX `call data`
	txId, ethAddress, amount, txReimbursement, erc20address, signatures, err := ethhelper.DecodeBridgeMintFunction(tx.Data())
	if err != nil {
		if errors.Is(err, ethhelper.ErrorInvalidMintFunctionParameters) {
			ss.logger.Debugf("[%s] - ETH TX [%s] - Invalid Mint parameters provided", ethTxMessage.TransferID, ethTxMessage.EthTxHash)
			return false, nil
		}
		return false, err
	}

	if txId != ethTxMessage.TransferID {
		ss.logger.Debugf("[%s] - ETH TX [%s] - Different txn id [%s].", ethTxMessage.TransferID, ethTxMessage.EthTxHash, txId)
		return false, nil
	}

	dbTx, err := ss.transferRepository.GetByTransactionId(ethTxMessage.TransferID)
	if err != nil {
		return false, err
	}
	if dbTx == nil {
		ss.logger.Debugf("[%s] - ETH TX [%s] - Transaction not found in database.", ethTxMessage.TransferID, ethTxMessage.EthTxHash)
		return false, nil
	}

	if dbTx.Amount != amount ||
		dbTx.Receiver != ethAddress ||
		dbTx.TxReimbursement != txReimbursement ||
		// TODO: Add validation for erc20address, once the contracts support it
		tx.GasPrice().String() != dbTx.GasPrice {
		ss.logger.Debugf("[%s] - ETH TX [%s] - Invalid arguments.", ethTxMessage.TransferID, ethTxMessage.EthTxHash)
		return false, nil
	}

	// Verify Ethereum TX provided `signatures` authenticity
	messageHash, err := auth_message.EncodeBytesFrom(txId, ethAddress, erc20address, amount, txReimbursement, dbTx.GasPrice)
	if err != nil {
		ss.logger.Errorf("[%s] - Failed to encode the authorisation signature to reconstruct required Signature. Error: [%s]", txId, err)
		return false, err
	}

	checkedAddresses := make(map[string]bool)
	for _, signature := range signatures {
		address, err := ethhelper.GetAddressBySignature(messageHash, signature)
		if err != nil {
			return false, err
		}
		if checkedAddresses[address] {
			return false, err
		}

		if !ss.contractsService.IsMember(address) {
			ss.logger.Debugf("[%s] - ETH TX [%s] - Invalid operator process - [%s].", txId, ethTxMessage.EthTxHash, address)
			return false, nil
		}
		checkedAddresses[address] = true
	}

	return true, nil
}

func (ss *Service) ProcessEthereumTxMessage(tm encoding.TopicMessage) error {
	etm := tm.GetTopicEthTransactionMessage()
	err := ss.transferRepository.UpdateEthTxSubmitted(etm.TransferID, etm.EthTxHash)
	if err != nil {
		ss.logger.Errorf("[%s] - Failed to update status to [%s]. Error [%s].", etm.TransferID, transfer.StatusEthTxSubmitted, err)
		return err
	}

	onEthTxSuccess, onEthTxRevert := ss.ethTxCallbacks(etm.TransferID, etm.EthTxHash)
	ss.ethClient.WaitForTransaction(etm.EthTxHash, onEthTxSuccess, onEthTxRevert, func(err error) {})

	ss.scheduler.Cancel(etm.TransferID)
	return nil
}

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
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transfer"
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
		ss.logger.Errorf("Failed to retrieve Transaction Record for TX ID [%s]. Error: %s", topicMessage.TransferID, err)
		return false, err
	}

	valid, erc20address := ss.contractsService.IsValidBridgeAsset(t.SourceAsset)
	if !valid {
		ss.logger.Errorf("Provided Asset is not supported - [%s]", t.SourceAsset)
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
		ss.logger.Errorf("Failed to encode the authorisation signature for TX ID [%s]. Error: %s", tsm.TransferID, err)
		return err
	}

	// Prepare Signature
	signatureBytes, signatureHex, err := ethhelper.DecodeSignature(tsm.GetSignature())
	if err != nil {
		ss.logger.Errorf("[%s] - Decoding Signature [%s] for TX failed. Error: %s", tsm.TransferID, tsm.GetSignature(), err)
		return err
	}
	authMessageStr := hex.EncodeToString(authMsgBytes)

	// Check for duplicated signature
	exists, err := ss.messageRepository.Exist(tsm.TransferID, signatureHex, authMessageStr)
	if err != nil {
		ss.logger.Errorf("An error occurred while getting TX [%s] from DB. Error: %s", tsm.TransferID, err)
		return err
	}
	if exists {
		ss.logger.Errorf("Signature already received for TX [%s]", tsm.TransferID)
		return err
	}

	// Verify Signature
	address, err := ss.verifySignature(err, authMsgBytes, signatureBytes, tsm.TransferID, authMessageStr)
	if err != nil {
		return err
	}

	ss.logger.Debugf("Successfully verified new Signature from [%s] for TX [%s]", address.String(), tsm.TransferID)

	// Persist in DB
	err = ss.messageRepository.Create(&message.Message{
		TransferID:           tsm.TransferID,
		Signature:            signatureHex,
		Hash:                 authMessageStr,
		Signer:               address.String(),
		TransactionTimestamp: tm.TransactionTimestamp,
	})
	if err != nil {
		ss.logger.Errorf("[%s] - Failed to save Transaction Message in DB with Signature [%s]. Error: %s", tsm.TransferID, signatureHex, err)
		return err
	}

	ss.logger.Infof("Successfully processed Signature Message from [%s] for TX [%s]", address.String(), tsm.TransferID)
	return nil
}

// ScheduleForSubmission computes the execution slot and schedules the Eth TX for submission
func (ss *Service) ScheduleEthereumTxForSubmission(txId string) error {
	transfer, err := ss.transferRepository.GetByTransactionId(txId)
	if err != nil {
		ss.logger.Errorf("Failed eto query transfer for TX [%s]. Error: [%s].", txId, err)
	}
	signatureMessages, err := ss.messageRepository.Get(txId)
	if err != nil {
		ss.logger.Errorf("Failed to query all Signature Messages for TX [%s]. Error: %s", txId, err)
		return err
	}

	slot, isFound := ss.computeExecutionSlot(signatureMessages)
	if !isFound {
		ss.logger.Debugf("TX [%s] - Operator [%s] has not been found as signer amongst the signatures collected.", txId, ss.ethSigner.Address())
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

	ethereumMintTask := ss.prepareEthereumMintTask(txId, ethAddress, amount, txReimbursement, gasPriceWei, signatures, messageHash)
	err = ss.scheduler.Schedule(txId, signatureMessages[0].TransactionTimestamp, slot, ethereumMintTask)
	if err != nil {
		return err
	}
	return nil
}

// prepareEthereumMintTask returns the function to be executed for processing the
// Ethereum Mint transaction and HCS topic message with the ethereum TX hash after that
func (ss *Service) prepareEthereumMintTask(txId string, ethAddress string, amount string, txReimbursement string, gasPriceWei string, signatures [][]byte, messageHash string) func() {
	ethereumMintTask := func() {
		// Submit and monitor Ethereum TX
		ethTransactor, err := ss.ethSigner.NewKeyTransactor(ss.ethClient.ChainID())
		if err != nil {
			ss.logger.Errorf("Failed to establish key transactor. Error %s", err)
			return
		}

		ethTransactor.GasPrice, err = helper.ToBigInt(gasPriceWei)
		if err != nil {
			ss.logger.Errorf("Failed to parse provided gas price for TX [%s]. Error: %s", txId, err)
			return
		}

		ethTx, err := ss.contractsService.SubmitSignatures(ethTransactor, txId, ethAddress, amount, txReimbursement, signatures)
		if err != nil {
			ss.logger.Errorf("Failed to Submit Signatures for TX [%s]. Error: %s", txId, err)
			return
		}
		err = ss.transferRepository.UpdateEthTxSubmitted(txId, ethTx.Hash().String())
		if err != nil {
			ss.logger.Errorf("Failed to update status for TX [%s]. Error [%s].", txId, err)
			return
		}
		ss.logger.Infof("Submitted Ethereum Mint TX [%s] for TX [%s]", ethTx.Hash().String(), txId)

		onEthTxSuccess, onEthTxRevert := ss.ethTxCallbacks(txId, ethTx.Hash().String())
		ss.ethClient.WaitForTransaction(ethTx.Hash().String(), onEthTxSuccess, onEthTxRevert, func(err error) {})

		// Submit and monitor HCS Message for Ethereum TX Hash
		hcsTx, err := ss.submitEthTxTopicMessage(txId, messageHash, ethTx.Hash().String())
		if err != nil {
			ss.logger.Errorf("Failed to submit Ethereum TX Hash to Bridge Topic for TX [%s]. Error %s", txId, err)
			return
		}
		err = ss.transferRepository.UpdateStatusEthTxMsgSubmitted(txId)
		if err != nil {
			ss.logger.Errorf("Failed to update status for TX [%s]. Error [%s].", txId, err)
			return
		}
		ss.logger.Infof("Submitted Ethereum TX Hash [%s] for TX [%s] to HCS. Transaction ID [%s]", ethTx.Hash().String(), txId, hcsTx.String())

		onHcsMessageSuccess, onHcsMessageFail := ss.hcsTxCallbacks(txId)
		ss.mirrorClient.WaitForTransaction(hcsTx.String(), onHcsMessageSuccess, onHcsMessageFail)

		ss.logger.Infof("Successfully processed Ethereum Minting for TX [%s]", txId)
	}
	return ethereumMintTask
}

func getSignatures(messages []message.Message) ([][]byte, error) {
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

func (ss *Service) verifySignature(err error, authMsgBytes []byte, signatureBytes []byte, txId, authMessageStr string) (common.Address, error) {
	publicKey, err := crypto.Ecrecover(authMsgBytes, signatureBytes)
	if err != nil {
		ss.logger.Errorf("[%s] - Failed to recover public key. Hash [%s]. Error: %s", txId, authMessageStr, err)
		return common.Address{}, err
	}
	unmarshalledPublicKey, err := crypto.UnmarshalPubkey(publicKey)
	if err != nil {
		ss.logger.Errorf("[%s] - Failed to unmarshall public key. Error: %s", txId, err)
		return common.Address{}, err
	}
	address := crypto.PubkeyToAddress(*unmarshalledPublicKey)
	if !ss.contractsService.IsMember(address.String()) {
		ss.logger.Errorf("[%s] - Received Signature [%s] is not signed by Bridge member", txId, authMessageStr)
		return common.Address{}, errors.New(fmt.Sprintf("signer is not signatures member"))
	}
	return address, nil
}

// computeExecutionSlot - computes the slot order in which the TX will execute
// Important! Transaction messages ARE expected to be sorted by ascending Timestamp
func (ss *Service) computeExecutionSlot(messages []message.Message) (slot int64, isFound bool) {
	for i := 0; i < len(messages); i++ {
		if strings.ToLower(messages[i].Signer) == strings.ToLower(ss.ethSigner.Address()) {
			return int64(i), true
		}
	}

	return -1, false
}

func (ss *Service) submitEthTxTopicMessage(txId, messageHash, ethereumTxHash string) (*hedera.TransactionID, error) {
	ethTxHashMessage := encoding.NewEthereumHashMessage(txId, messageHash, ethereumTxHash)
	ethTxHashBytes, err := ethTxHashMessage.ToBytes()
	if err != nil {
		ss.logger.Errorf("Failed to encode Eth TX Hash Message to bytes for TX [%s]. Error: %s", txId, err)
		return nil, err
	}

	return ss.hederaClient.SubmitTopicConsensusMessage(ss.topicID, ethTxHashBytes)
}

func (ss *Service) ethTxCallbacks(txId, hash string) (onSuccess, onRevert func()) {
	onSuccess = func() {
		ss.logger.Infof("Ethereum TX [%s] for TX [%s] was successfully mined", hash, txId)
		err := ss.transferRepository.UpdateEthTxMined(txId)
		if err != nil {
			ss.logger.Errorf("Failed to update status for TX [%s]. Error [%s].", txId, err)
			return
		}
	}

	onRevert = func() {
		ss.logger.Infof("Ethereum TX [%s] for TX [%s] reverted", hash, txId)
		err := ss.transferRepository.UpdateEthTxReverted(txId)
		if err != nil {
			ss.logger.Errorf("Failed to update status for TX [%s]. Error [%s].", txId, err)
			return
		}
	}
	return onSuccess, onRevert
}

func (ss *Service) hcsTxCallbacks(txId string) (onSuccess, onFailure func()) {
	onSuccess = func() {
		ss.logger.Infof("Ethereum TX Hash message was successfully mined for TX [%s]", txId)
		err := ss.transferRepository.UpdateStatusEthTxMsgMined(txId)
		if err != nil {
			ss.logger.Errorf("Failed to update status for TX [%s]. Error [%s].", txId, err)
			return
		}
	}

	onFailure = func() {
		ss.logger.Infof("Ethereum TX Hash message failed for TX ID [%s]", txId)
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
		ss.logger.Errorf("Failed to encode the authorisation signature to reconstruct required Signature for TX ID [%s]. Error: %s", txId, err)
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
		ss.logger.Errorf("Failed to update status to [%s] of transaction with TransactionID [%s]. Error [%s].", transfer.StatusEthTxSubmitted, etm.TransferID, err)
		return err
	}

	onEthTxSuccess, onEthTxRevert := ss.ethTxCallbacks(etm.TransferID, etm.EthTxHash)
	ss.ethClient.WaitForTransaction(etm.EthTxHash, onEthTxSuccess, onEthTxRevert, func(err error) {})

	ss.scheduler.Cancel(etm.TransferID)
	return nil
}

// ShouldTransactionBeScheduled checks the database for ExecuteEthTransaction flag
func (ss *Service) ShouldTransactionBeScheduled(transactionId string) (bool, error) {
	t, err := ss.transferRepository.GetByTransactionId(transactionId)
	if err != nil {
		ss.logger.Errorf("Could not load transaction info for TX [%s] Error: %s", transactionId, err)
		return false, err
	}
	return t.ExecuteEthTransaction, nil
}

// TransactionData returns from the database all messages for specific transactionId and
// calculates if messages have reached super majority
func (ss *Service) TransactionData(transactionId string) (service.TransactionData, error) {
	transfer, err := ss.transferRepository.GetByTransactionId(transactionId)
	if err != nil {
		ss.logger.Errorf("Failed to query Transfer for TX [%s]. Error: [%s]", transactionId, err)
	}
	messages, err := ss.messageRepository.Get(transactionId)
	if err != nil {
		ss.logger.Errorf("Failed to query Signature Messages for TX [%s]. Error: [%s].", transactionId, err)
		return service.TransactionData{}, err
	}

	if len(messages) == 0 {
		return service.TransactionData{}, nil
	}

	var signatures []string
	for _, m := range messages {
		signatures = append(signatures, m.Signature)
	}

	requiredSigCount := len(ss.contractsService.GetMembers())/2 + 1
	reachedMajority := len(messages) >= requiredSigCount

	return service.TransactionData{
		Recipient:   transfer.Receiver,
		Amount:      transfer.Amount,
		Fee:         transfer.TxReimbursement,
		SourceAsset: transfer.SourceAsset,
		TargetAsset: transfer.TargetAsset,
		Signatures:  signatures,
		Majority:    reachedMajority,
		GasPrice:    transfer.GasPrice,
	}, nil
}

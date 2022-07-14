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

package e2e

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	read_only "github.com/limechain/hedera-eth-bridge-validator/app/services/read-only"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/werc721"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/wtoken"
	mirror_node "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/transaction"
	hederahelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	auth_message "github.com/limechain/hedera-eth-bridge-validator/app/model/auth-message"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/schedule"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/limechain/hedera-eth-bridge-validator/e2e/util"

	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/e2e/service/database"
	"github.com/limechain/hedera-eth-bridge-validator/e2e/setup"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashgraph/hedera-sdk-go/v2"
	validatorproto "github.com/limechain/hedera-eth-bridge-validator/proto"
	"google.golang.org/protobuf/proto"
)

var (
	now time.Time
)

const (
	expectedValidatorsCount = 3
	firstEvmChainId         = uint64(80001) // represents Polygon Mumbai Testnet (e2e config must have configuration for that particular network)
	secondEvmChainId        = uint64(43113) // represents Avalanche Fuji Testnet (e2e config must have configuration for that particular network)
)

// Test_HBAR recreates a real life situation of a user who wants to bridge a Hedera HBARs to the EVM Network infrastructure. The wrapped token on the EVM network(corresponding to the native Hedera Hashgraph's HBARs) gets minted, then transferred to the recipient account on the EVM network.
func Test_HBAR(t *testing.T) {
	amount := int64(1000000000) // 10 HBAR
	setupEnv := setup.Load()
	now = time.Now()

	chainId := firstEvmChainId
	evm := setupEnv.Clients.EVM[chainId]
	receiver := evm.Receiver
	memo := fmt.Sprintf("%d-%s", chainId, evm.Receiver.String())

	targetAsset, err := setup.NativeToWrappedAsset(setupEnv.AssetMappings, constants.HederaNetworkId, chainId, constants.Hbar)
	if err != nil {
		t.Fatalf("Expecting Token [%s] is not supported. - Error: [%s]", constants.Hbar, err)
	}

	mintAmount, fee := calculateReceiverAndFeeAmounts(setupEnv, constants.Hbar, amount)

	// Step 1 - Verify the transfer of Hbars to the Bridge Account
	transactionResponse, wrappedBalanceBefore := verifyTransferToBridgeAccount(setupEnv, targetAsset, evm, memo, receiver, amount, t)

	// Step 2 - Verify the submitted topic messages
	receivedSignatures := verifyTopicMessages(setupEnv, hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String(), t)

	// Step 3 - Validate fee scheduled transaction
	scheduledTxID, scheduleID := validateMembersScheduledTxs(setupEnv, constants.Hbar, generateMirrorNodeExpectedTransfersForHederaTransfer(setupEnv, constants.Hbar, fee), t)

	// Step 4 - Verify Transfer retrieved from Validator API
	transactionData := verifyFungibleTransferFromValidatorAPI(setupEnv, evm, hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String(), constants.Hbar, fmt.Sprint(mintAmount), targetAsset, t)

	// Step 4.1 - Get the consensus timestamp of transfer tx
	tx, err := setupEnv.Clients.MirrorNode.GetSuccessfulTransaction(hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String())
	if err != nil {
		t.Fatal("failed to get successful transaction", err)
	}
	nanos, err := timestamp.FromString(tx.ConsensusTimestamp)
	if err != nil {
		t.Fatal("failed to parse consensus timestamp", err)
	}
	ts := timestamp.FromNanos(nanos)

	// Step 5 - Submit Mint transaction
	txHash := submitMintTransaction(evm, hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String(), transactionData, common.HexToAddress(targetAsset), t)

	// Step 6 - Wait for transaction to be mined
	waitForTransaction(evm, txHash, t)

	// Step 7 - Validate Token balances
	verifyWrappedAssetBalance(evm, targetAsset, big.NewInt(mintAmount), wrappedBalanceBefore, receiver, t)

	// Step 8 - Prepare Comparable Expected Transfer Record
	expectedTxRecord := util.PrepareExpectedTransfer(
		constants.HederaNetworkId,
		chainId,
		constants.HederaNetworkId,
		hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String(),
		constants.Hbar,
		targetAsset,
		constants.Hbar,
		strconv.FormatInt(amount, 10),
		receiver.String(),
		status.Completed,
		setupEnv.Clients.Hedera.GetOperatorAccountID().String(),
		entity.NanoTime{Time: ts})
	// and:
	expectedFeeRecord := util.PrepareExpectedFeeRecord(
		scheduledTxID,
		scheduleID,
		fee,
		hederahelper.ToMirrorNodeTransactionID(transactionResponse.TransactionID.String()))

	authMsgBytes, err := auth_message.
		EncodeFungibleBytesFrom(
			expectedTxRecord.SourceChainID,
			expectedTxRecord.TargetChainID,
			expectedTxRecord.TransactionID,
			expectedTxRecord.TargetAsset,
			expectedTxRecord.Receiver,
			strconv.FormatInt(mintAmount, 10))
	if err != nil {
		t.Fatalf("[%s] - Failed to encode the authorisation signature. Error: [%s]", expectedTxRecord.TransactionID, err)
	}

	// Step 9 - Verify Database Records
	verifyTransferRecordAndSignatures(setupEnv.DbValidator, expectedTxRecord, authMsgBytes, receivedSignatures, t)
	// and:
	verifyFeeRecord(setupEnv.DbValidator, expectedFeeRecord, t)
}

// Test_E2E_Token_Transfer recreates a real life situation of a user who wants to bridge a Hedera native token to the EVM Network infrastructure. The wrapped token on the EVM network(corresponding to the native Hedera Hashgraph's one) gets minted, then transferred to the recipient account on the EVM network.
func Test_E2E_Token_Transfer(t *testing.T) {
	amount := int64(1000000000) // 10 HBAR
	setupEnv := setup.Load()
	now = time.Now()

	chainId := firstEvmChainId
	evm := setupEnv.Clients.EVM[chainId]
	memo := fmt.Sprintf("%d-%s", chainId, evm.Receiver.String())
	mintAmount, fee := calculateReceiverAndFeeAmounts(setupEnv, setupEnv.TokenID.String(), amount)

	targetAsset, err := setup.NativeToWrappedAsset(setupEnv.AssetMappings, constants.HederaNetworkId, chainId, setupEnv.TokenID.String())
	if err != nil {
		t.Fatalf("Expecting Token [%s] is not supported. - Error: [%s]", constants.Hbar, err)
	}

	// Step 1 - Verify the transfer of HTS to the Bridge Account
	transactionResponse, wrappedBalanceBefore := verifyTokenTransferToBridgeAccount(setupEnv, targetAsset, setupEnv.TokenID, evm, memo, evm.Receiver, amount, t)

	// Step 2 - Verify the submitted topic messages
	receivedSignatures := verifyTopicMessages(setupEnv, hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String(), t)

	// Step 3 - Validate fee scheduled transaction
	scheduledTxID, scheduleID := validateMembersScheduledTxs(setupEnv, setupEnv.TokenID.String(), generateMirrorNodeExpectedTransfersForHederaTransfer(setupEnv, setupEnv.TokenID.String(), fee), t)

	// Step 4 - Verify Transfer retrieved from Validator API
	transactionData := verifyFungibleTransferFromValidatorAPI(setupEnv, evm, hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String(), setupEnv.TokenID.String(), fmt.Sprint(mintAmount), targetAsset, t)

	// Step 4.1 - Get the consensus timestamp of the transfer
	tx, err := setupEnv.Clients.MirrorNode.GetSuccessfulTransaction(hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String())
	if err != nil {
		t.Fatal("failed to get successful transaction", err)
	}
	nanos, err := timestamp.FromString(tx.ConsensusTimestamp)
	if err != nil {
		t.Fatal("failed to parse consensus timestamp", err)
	}
	ts := timestamp.FromNanos(nanos)

	// Step 5 - Submit Mint transaction
	txHash := submitMintTransaction(evm, hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String(), transactionData, common.HexToAddress(targetAsset), t)

	// Step 6 - Wait for transaction to be mined
	waitForTransaction(evm, txHash, t)

	// Step 7 - Validate Token balances
	verifyWrappedAssetBalance(evm, targetAsset, big.NewInt(mintAmount), wrappedBalanceBefore, evm.Receiver, t)

	// Step 8 - Verify Database records
	expectedTxRecord := util.PrepareExpectedTransfer(
		constants.HederaNetworkId,
		chainId,
		constants.HederaNetworkId,
		hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String(),
		setupEnv.TokenID.String(),
		targetAsset,
		setupEnv.TokenID.String(),
		strconv.FormatInt(amount, 10),
		evm.Receiver.String(),
		status.Completed,
		setupEnv.Clients.Hedera.GetOperatorAccountID().String(),
		entity.NanoTime{Time: ts})
	// and:
	expectedFeeRecord := util.PrepareExpectedFeeRecord(
		scheduledTxID,
		scheduleID, fee,
		hederahelper.ToMirrorNodeTransactionID(transactionResponse.TransactionID.String()))

	authMsgBytes, err := auth_message.
		EncodeFungibleBytesFrom(
			expectedTxRecord.SourceChainID,
			expectedTxRecord.TargetChainID,
			expectedTxRecord.TransactionID,
			expectedTxRecord.TargetAsset,
			expectedTxRecord.Receiver,
			strconv.FormatInt(mintAmount, 10))
	if err != nil {
		t.Fatalf("[%s] - Failed to encode the authorisation signature. Error: [%s]", expectedTxRecord.TransactionID, err)
	}

	// Step 9 - Verify Database Records
	verifyTransferRecordAndSignatures(setupEnv.DbValidator, expectedTxRecord, authMsgBytes, receivedSignatures, t)
	// and:
	verifyFeeRecord(setupEnv.DbValidator, expectedFeeRecord, t)
}

// Test_EVM_Hedera_HBAR recreates a real life situation of a user who wants to return a Hedera native HBARs from the EVM Network infrastructure. The wrapped HBARs on the EVM network(corresponding to the native Hedera Hashgraph's one) gets burned, then the locked HBARs on the Hedera bridge account get unlocked, forwarding them to the recipient account.
func Test_EVM_Hedera_HBAR(t *testing.T) {
	amount := int64(100000000) // 1 HBAR
	setupEnv := setup.Load()

	chainId := firstEvmChainId
	evm := setupEnv.Clients.EVM[chainId]
	now = time.Now()
	accountBalanceBefore := util.GetHederaAccountBalance(setupEnv.Clients.Hedera, setupEnv.Clients.Hedera.GetOperatorAccountID(), t)

	targetAsset, err := setup.NativeToWrappedAsset(setupEnv.AssetMappings, constants.HederaNetworkId, chainId, constants.Hbar)
	if err != nil {
		t.Fatalf("Expecting Token [%s] is not supported. - Error: [%s]", constants.Hbar, err)
	}

	// 1. Calculate Expected Receive And Fee Amounts
	expectedReceiveAmount, fee := calculateReceiverAndFeeAmounts(setupEnv, constants.Hbar, amount)

	// 2. Submit burn transaction to the bridge contract
	burnTxReceipt, expectedRouterBurn := sendBurnEthTransaction(setupEnv.AssetMappings, evm, constants.Hbar, constants.HederaNetworkId, chainId, setupEnv.Clients.Hedera.GetOperatorAccountID().ToBytes(), amount, t)

	// 2.1 Get the block timestamp of burn event
	block, err := evm.EVMClient.BlockByNumber(context.Background(), burnTxReceipt.BlockNumber)
	if err != nil {
		t.Fatal("failed to get block by number", err)
	}
	blockTimestamp := time.Unix(int64(block.Time()), 0).UTC()

	// 3. Validate that the burn transaction went through and emitted the correct events
	expectedId := validateBurnEvent(burnTxReceipt, expectedRouterBurn, t)

	// 4. Validate that a scheduled transaction was submitted
	transactionID, scheduleID := validateSubmittedScheduledTx(setupEnv, constants.Hbar, generateMirrorNodeExpectedTransfersForBurnEvent(setupEnv, constants.Hbar, expectedReceiveAmount, fee), t)

	// 5. Validate Event Transaction ID retrieved from Validator API
	validateEventTransactionIDFromValidatorAPI(setupEnv, expectedId, transactionID, t)

	// 6. Validate that the balance of the receiver account (hedera) was changed with the correct amount
	validateReceiverAccountBalance(setupEnv, uint64(expectedReceiveAmount), accountBalanceBefore, constants.Hbar, t)

	// 7. Prepare Expected Database Records
	expectedBurnEventRecord := util.PrepareExpectedTransfer(
		chainId,
		constants.HederaNetworkId,
		constants.HederaNetworkId,
		expectedId,
		targetAsset,
		constants.Hbar,
		constants.Hbar,
		strconv.FormatInt(amount, 10),
		setupEnv.Clients.Hedera.GetOperatorAccountID().String(),
		status.Completed,
		evm.Signer.Address(),
		entity.NanoTime{Time: blockTimestamp})
	// and:
	expectedFeeRecord := util.PrepareExpectedFeeRecord(transactionID, scheduleID, fee, expectedId)

	// 8. Wait for validators to update DB state after Scheduled TX is mined
	time.Sleep(20 * time.Second)

	// 9. Validate Database Records
	verifyTransferRecord(setupEnv.DbValidator, expectedBurnEventRecord, t)
	// and:
	verifyFeeRecord(setupEnv.DbValidator, expectedFeeRecord, t)
}

// Test_EVM_Hedera_Token recreates a real life situation of a user who wants to return a Hedera native token from the EVM Network infrastructure. The wrapped token on the EVM network(corresponding to the native Hedera one) gets burned, then the amount gets unlocked on the Hedera bridge account, forwarding it to the recipient account.
func Test_EVM_Hedera_Token(t *testing.T) {
	amount := int64(100000000) // 1 HBAR
	setupEnv := setup.Load()

	chainId := firstEvmChainId
	evm := setupEnv.Clients.EVM[chainId]
	now = time.Now()
	accountBalanceBefore := util.GetHederaAccountBalance(setupEnv.Clients.Hedera, setupEnv.Clients.Hedera.GetOperatorAccountID(), t)

	targetAsset, err := setup.NativeToWrappedAsset(setupEnv.AssetMappings, constants.HederaNetworkId, chainId, setupEnv.TokenID.String())
	if err != nil {
		t.Fatalf("Expecting Token [%s] is not supported. - Error: [%s]", setupEnv.TokenID.String(), err)
	}

	// 1. Calculate Expected Receive Amount
	expectedReceiveAmount, fee := calculateReceiverAndFeeAmounts(setupEnv, setupEnv.TokenID.String(), amount)

	// 2. Submit burn transaction to the bridge contract
	burnTxReceipt, expectedRouterBurn := sendBurnEthTransaction(
		setupEnv.AssetMappings,
		evm,
		setupEnv.TokenID.String(),
		constants.HederaNetworkId,
		chainId,
		setupEnv.Clients.Hedera.GetOperatorAccountID().ToBytes(),
		amount,
		t)

	// 2.1 Get the block timestamp of burn event
	block, err := evm.EVMClient.BlockByNumber(context.Background(), burnTxReceipt.BlockNumber)
	if err != nil {
		t.Fatal("failed to get block by number", err)
	}
	blockTimestamp := time.Unix(int64(block.Time()), 0).UTC()

	// 3. Validate that the burn transaction went through and emitted the correct events
	expectedId := validateBurnEvent(burnTxReceipt, expectedRouterBurn, t)

	// 4. Validate that a scheduled transaction was submitted
	transactionID, scheduleID := validateSubmittedScheduledTx(setupEnv, setupEnv.TokenID.String(), generateMirrorNodeExpectedTransfersForBurnEvent(setupEnv, setupEnv.TokenID.String(), expectedReceiveAmount, fee), t)

	// 5. Validate Event Transaction ID retrieved from Validator API
	validateEventTransactionIDFromValidatorAPI(setupEnv, expectedId, transactionID, t)

	// 6. Validate that the balance of the receiver account (hedera) was changed with the correct amount
	validateReceiverAccountBalance(setupEnv, uint64(expectedReceiveAmount), accountBalanceBefore, setupEnv.TokenID.String(), t)

	// 7. Prepare Expected Database Records
	expectedBurnEventRecord := util.PrepareExpectedTransfer(
		chainId,
		constants.HederaNetworkId,
		constants.HederaNetworkId,
		expectedId,
		targetAsset,
		setupEnv.TokenID.String(),
		setupEnv.TokenID.String(),
		strconv.FormatInt(amount, 10),
		setupEnv.Clients.Hedera.GetOperatorAccountID().String(),
		status.Completed,
		evm.Signer.Address(),
		entity.NanoTime{Time: blockTimestamp})
	// and:
	expectedFeeRecord := util.PrepareExpectedFeeRecord(transactionID, scheduleID, fee, expectedId)

	// 8. Wait for validators to update DB state after Scheduled TX is mined
	time.Sleep(20 * time.Second)

	// 9. Validate Database Records
	verifyTransferRecord(setupEnv.DbValidator, expectedBurnEventRecord, t)
	// and:
	verifyFeeRecord(setupEnv.DbValidator, expectedFeeRecord, t)
}

// Test_EVM_Hedera_Native_Token recreates a real life situation of a user who wants to bridge an EVM native token to the Hedera infrastructure. A new wrapped token (corresponding to the native EVM one) gets minted to the bridge account, then gets transferred to the recipient account.
func Test_EVM_Hedera_Native_Token(t *testing.T) {
	amount := int64(1000000000000) // 1 000 gwei
	// Step 1: Initialize setup, smart contracts, etc.
	setupEnv := setup.Load()

	chainId := firstEvmChainId
	evm := setupEnv.Clients.EVM[chainId]
	now = time.Now()
	bridgeAccountBalanceBefore := util.GetHederaAccountBalance(setupEnv.Clients.Hedera, setupEnv.BridgeAccount, t)
	receiverAccountBalanceBefore := util.GetHederaAccountBalance(setupEnv.Clients.Hedera, setupEnv.Clients.Hedera.GetOperatorAccountID(), t)
	targetAsset, err := setup.NativeToWrappedAsset(setupEnv.AssetMappings, chainId, constants.HederaNetworkId, setupEnv.NativeEvmToken)
	if err != nil {
		t.Fatal(err)
	}

	// Step 2: Submit Lock Txn from a deployed smart contract
	receipt, expectedLockEventLog := sendLockEthTransaction(evm, setupEnv.NativeEvmToken, constants.HederaNetworkId, setupEnv.Clients.Hedera.GetOperatorAccountID().ToBytes(), amount, t)

	// Step 2.1 - Get the block timestamp of lock event
	block, err := evm.EVMClient.BlockByNumber(context.Background(), receipt.BlockNumber)
	if err != nil {
		t.Fatal("failed to get block by number", err)
	}
	blockTimestamp := time.Unix(int64(block.Time()), 0).UTC()

	// Step 3: Validate Lock Event was emitted with correct data
	lockEventId := validateLockEvent(receipt, expectedLockEventLog, t)

	bridgedAmount := new(big.Int).Sub(expectedLockEventLog.Amount, expectedLockEventLog.ServiceFee)
	expectedAmount, err := removeDecimals(bridgedAmount.Int64(), common.HexToAddress(setupEnv.NativeEvmToken), evm)
	if err != nil {
		t.Fatal(err)
	}

	mintTransfer := []transaction.Transfer{
		{
			Account: setupEnv.BridgeAccount.String(),
			Amount:  expectedAmount,
			Token:   targetAsset,
		},
	}

	// Step 4: Validate that a scheduled token mint txn was submitted successfully
	bridgeMintTransactionID, bridgeMintScheduleID := validateScheduledMintTx(setupEnv, setupEnv.BridgeAccount, setupEnv.TokenID.String(), mintTransfer, t)

	// Wait for validators to update DB state after Scheduled TX is mined
	time.Sleep(20 * time.Second)

	// Step 5: Validate that Database statuses were changed correctly
	expectedLockEventRecord := util.PrepareExpectedTransfer(
		chainId,
		constants.HederaNetworkId,
		chainId,
		lockEventId,
		setupEnv.NativeEvmToken,
		targetAsset,
		setupEnv.NativeEvmToken,
		strconv.FormatInt(expectedAmount, 10),
		setupEnv.Clients.Hedera.GetOperatorAccountID().String(),
		status.Completed,
		evm.Signer.Address(),
		entity.NanoTime{Time: blockTimestamp})

	expectedScheduleMintRecord := &entity.Schedule{
		TransactionID: bridgeMintTransactionID,
		ScheduleID:    bridgeMintScheduleID,
		Operation:     schedule.MINT,
		Status:        status.Completed,
		TransferID: sql.NullString{
			String: lockEventId,
			Valid:  true,
		},
	}
	// Step 6: Verify that records have been created successfully
	verifyTransferRecord(setupEnv.DbValidator, expectedLockEventRecord, t)
	verifyScheduleRecord(setupEnv.DbValidator, expectedScheduleMintRecord, t)

	// Step 7: Validate that a scheduled transfer txn was submitted successfully
	bridgeTransferTransactionID, bridgeTransferScheduleID := validateScheduledTx(
		setupEnv,
		setupEnv.Clients.Hedera.GetOperatorAccountID(),
		setupEnv.TokenID.String(),
		generateMirrorNodeExpectedTransfersForLockEvent(setupEnv, targetAsset, expectedAmount),
		t)

	// Wait for validators to update DB state after Scheduled TX is mined
	time.Sleep(20 * time.Second)

	// Step 8: Validate that database statuses were updated correctly for the Schedule Transfer
	expectedScheduleTransferRecord := &entity.Schedule{
		TransactionID: bridgeTransferTransactionID,
		ScheduleID:    bridgeTransferScheduleID,
		Operation:     schedule.TRANSFER,
		HasReceiver:   true,
		Status:        status.Completed,
		TransferID: sql.NullString{
			String: lockEventId,
			Valid:  true,
		},
	}

	verifyScheduleRecord(setupEnv.DbValidator, expectedScheduleTransferRecord, t)
	// Step 9: Validate Treasury(BridgeAccount) Balance and Receiver Balance
	validateAccountBalance(setupEnv, setupEnv.BridgeAccount, 0, bridgeAccountBalanceBefore, targetAsset, t)
	validateAccountBalance(setupEnv, setupEnv.Clients.Hedera.GetOperatorAccountID(), uint64(expectedAmount), receiverAccountBalanceBefore, targetAsset, t)
}

func removeDecimals(amount int64, asset common.Address, evm setup.EVMUtils) (int64, error) {
	evmAsset, err := wtoken.NewWtoken(asset, evm.EVMClient)
	if err != nil {
		return 0, err
	}

	decimals, err := evmAsset.Decimals(nil)
	if err != nil {
		return 0, err
	}

	adaptation := decimals - 8
	if adaptation > 0 {
		adapted := amount / int64(math.Pow10(int(adaptation)))
		return adapted, nil
	}
	return amount, nil
}

func addDecimals(amount int64, asset common.Address, evm setup.EVMUtils) (int64, error) {
	evmAsset, err := wtoken.NewWtoken(asset, evm.EVMClient)
	if err != nil {
		return 0, err
	}

	decimals, err := evmAsset.Decimals(nil)
	if err != nil {
		return 0, err
	}
	adaptation := decimals - 8
	if adaptation > 0 {
		adapted := amount * int64(math.Pow10(int(adaptation)))
		return adapted, nil
	}
	return amount, nil
}

// Test_E2E_Hedera_EVM_Native_Token recreates a real life situation of a user who wants to bridge a Hedera wrapped token to the EVM Native Network infrastructure. The wrapped token on the EVM network(corresponding to the native Hedera Hashgraph's one) gets minted, then transferred to the recipient account on the EVM network.
func Test_E2E_Hedera_EVM_Native_Token(t *testing.T) {
	setupEnv := setup.Load()
	now = time.Now()

	chainId := firstEvmChainId
	evm := setupEnv.Clients.EVM[chainId]
	memo := fmt.Sprintf("%d-%s", chainId, evm.Receiver.String())
	unlockAmount := int64(10) // Amount, which converted to 18 decimals is 100000000000 (100 gwei)

	// Step 1 - Verify the transfer of HTS to the Bridge Account
	wrappedAsset, err := setup.NativeToWrappedAsset(setupEnv.AssetMappings, chainId, constants.HederaNetworkId, setupEnv.NativeEvmToken)
	if err != nil {
		t.Fatal(err)
	}

	tokenID, err := hedera.TokenIDFromString(wrappedAsset)
	if err != nil {
		t.Fatal(err)
	}

	expectedSubmitUnlockAmount, err := addDecimals(unlockAmount, common.HexToAddress(setupEnv.NativeEvmToken), evm)
	if err != nil {
		t.Fatal(err)
	}

	transactionResponse, nativeBalanceBefore := verifyTokenTransferToBridgeAccount(setupEnv, setupEnv.NativeEvmToken, tokenID, evm, memo, evm.Receiver, unlockAmount, t)
	burnTransfer := []transaction.Transfer{
		{
			Account: setupEnv.BridgeAccount.String(),
			Amount:  -unlockAmount,
			Token:   wrappedAsset,
		},
	}

	// Step 2 - Verify the submitted topic messages
	receivedSignatures := verifyTopicMessages(setupEnv, hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String(), t)

	// Step 3 - Validate burn scheduled transaction
	burnTransactionID, burnScheduleID := validateScheduledBurnTx(setupEnv, setupEnv.BridgeAccount, setupEnv.TokenID.String(), burnTransfer, t)

	// Step 4 - Verify Transfer retrieved from Validator API
	transactionData := verifyFungibleTransferFromValidatorAPI(setupEnv, evm, hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String(), setupEnv.NativeEvmToken, fmt.Sprint(expectedSubmitUnlockAmount), setupEnv.NativeEvmToken, t)

	// Step 4.1 - Get the consensus timestamp of the transfer
	tx, err := setupEnv.Clients.MirrorNode.GetSuccessfulTransaction(hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String())
	if err != nil {
		t.Fatal("failed to get successful transaction", err)
	}
	nanos, err := timestamp.FromString(tx.ConsensusTimestamp)
	if err != nil {
		t.Fatal("failed to parse consensus timestamp", err)
	}
	ts := timestamp.FromNanos(nanos)

	// Step 5 - Submit Unlock transaction
	txHash := submitUnlockTransaction(evm, hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String(), transactionData, common.HexToAddress(setupEnv.NativeEvmToken), t)

	// Step 6 - Wait for transaction to be mined
	waitForTransaction(evm, txHash, t)

	expectedUnlockedAmount := calculateExpectedUnlockAmount(evm, setupEnv.NativeEvmToken, expectedSubmitUnlockAmount, t)

	//Step 7 - Validate Token balances
	verifyWrappedAssetBalance(evm, setupEnv.NativeEvmToken, expectedUnlockedAmount, nativeBalanceBefore, evm.Receiver, t)

	// Step 8 - Verify Database records
	expectedTxRecord := util.PrepareExpectedTransfer(
		constants.HederaNetworkId,
		chainId,
		chainId,
		hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String(),
		wrappedAsset,
		setupEnv.NativeEvmToken,
		setupEnv.NativeEvmToken,
		strconv.FormatInt(expectedSubmitUnlockAmount, 10),
		evm.Receiver.String(),
		status.Completed,
		setupEnv.Clients.Hedera.GetOperatorAccountID().String(),
		entity.NanoTime{Time: ts})

	// Step 8: Validate that database statuses were updated correctly for the Schedule Burn
	expectedScheduleBurnRecord := &entity.Schedule{
		TransactionID: burnTransactionID,
		ScheduleID:    burnScheduleID,
		Operation:     schedule.BURN,
		Status:        status.Completed,
		TransferID: sql.NullString{
			String: hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String(),
			Valid:  true,
		},
	}

	authMsgBytes, err := auth_message.
		EncodeFungibleBytesFrom(
			expectedTxRecord.SourceChainID,
			expectedTxRecord.TargetChainID,
			expectedTxRecord.TransactionID,
			expectedTxRecord.TargetAsset,
			expectedTxRecord.Receiver,
			strconv.FormatInt(expectedSubmitUnlockAmount, 10))
	if err != nil {
		t.Fatalf("[%s] - Failed to encode the authorisation signature. Error: [%s]", expectedTxRecord.TransactionID, err)
	}

	// Step 9 - Verify Database Records
	verifyTransferRecordAndSignatures(setupEnv.DbValidator, expectedTxRecord, authMsgBytes, receivedSignatures, t)
	// and
	verifyScheduleRecord(setupEnv.DbValidator, expectedScheduleBurnRecord, t)
}

// Test_EVM_Native_to_EVM_Token recreates a real life situation of a user who wants to bridge an EVM native token to another EVM chain.
func Test_EVM_Native_to_EVM_Token(t *testing.T) {
	amount := int64(1000000000000) // 1000 gwei
	// Step 1 - Initialize setup, smart contracts, etc.
	setupEnv := setup.Load()

	chainId := firstEvmChainId
	evm := setupEnv.Clients.EVM[chainId]
	now = time.Now()
	targetChainID := secondEvmChainId
	wrappedAsset, err := setup.NativeToWrappedAsset(setupEnv.AssetMappings, chainId, targetChainID, setupEnv.NativeEvmToken)
	if err != nil {
		t.Fatal(err)
	}

	wrappedEvm := setupEnv.Clients.EVM[targetChainID]
	wrappedInstance, err := setup.InitAssetContract(wrappedAsset, wrappedEvm.EVMClient)
	if err != nil {
		t.Fatal(err)
	}

	wrappedBalanceBefore, err := wrappedInstance.BalanceOf(&bind.CallOpts{}, evm.Receiver)
	if err != nil {
		t.Fatal(err)
	}

	// Step 2 - Submit Lock Txn from a deployed smart contract
	receipt, expectedLockEventLog := sendLockEthTransaction(evm, setupEnv.NativeEvmToken, targetChainID, evm.Receiver.Bytes(), amount, t)

	expectedAmount := new(big.Int).Sub(expectedLockEventLog.Amount, expectedLockEventLog.ServiceFee)

	// Step 2.1 - Get the block timestamp of the lock event
	block, err := evm.EVMClient.BlockByNumber(context.Background(), receipt.BlockNumber)
	if err != nil {
		t.Fatal("failed to get block by number", err)
	}
	blockTimestamp := time.Unix(int64(block.Time()), 0).UTC()

	// Step 3 - Validate Lock Event was emitted with correct data
	lockEventId := validateLockEvent(receipt, expectedLockEventLog, t)

	// Step 4 - Verify the submitted topic messages
	receivedSignatures := verifyTopicMessages(setupEnv, lockEventId, t)

	// Step 5 - Verify Transfer retrieved from Validator API
	transactionData := verifyFungibleTransferFromValidatorAPI(setupEnv, evm, lockEventId, setupEnv.NativeEvmToken, expectedAmount.String(), wrappedAsset, t)

	// Step 6 - Submit Mint transaction
	txHash := submitMintTransaction(wrappedEvm, lockEventId, transactionData, common.HexToAddress(wrappedAsset), t)

	// Step 7 - Wait for transaction to be mined
	waitForTransaction(wrappedEvm, txHash, t)

	// Step 8 - Validate Token balances
	verifyWrappedAssetBalance(wrappedEvm, wrappedAsset, expectedAmount, wrappedBalanceBefore, evm.Receiver, t)

	// Step 9 - Prepare expected Transfer record
	expectedLockEventRecord := util.PrepareExpectedTransfer(
		chainId,
		targetChainID,
		chainId,
		lockEventId,
		setupEnv.NativeEvmToken,
		wrappedAsset,
		setupEnv.NativeEvmToken,
		expectedAmount.String(),
		evm.Receiver.String(),
		status.Completed,
		evm.Signer.Address(),
		entity.NanoTime{Time: blockTimestamp})

	authMsgBytes, err := auth_message.
		EncodeFungibleBytesFrom(
			expectedLockEventRecord.SourceChainID,
			expectedLockEventRecord.TargetChainID,
			expectedLockEventRecord.TransactionID,
			expectedLockEventRecord.TargetAsset,
			expectedLockEventRecord.Receiver,
			expectedAmount.String())
	if err != nil {
		t.Fatalf("[%s] - Failed to encode the authorisation signature. Error: [%s]", expectedLockEventRecord.TransactionID, err)
	}

	// Step 10 - Verify Database Records
	verifyTransferRecordAndSignatures(setupEnv.DbValidator, expectedLockEventRecord, authMsgBytes, receivedSignatures, t)
}

// Test_EVM_Wrapped_to_EVM_Token recreates a real life situation of a user who wants to bridge an EVM native token to another EVM chain.
func Test_EVM_Wrapped_to_EVM_Token(t *testing.T) {
	amount := int64(100000000000) // 100 gwei
	// Step 1 - Initialize setup, smart contracts, etc.
	setupEnv := setup.Load()

	chainId := firstEvmChainId
	sourceChain := secondEvmChainId
	wrappedEvm := setupEnv.Clients.EVM[sourceChain]
	now = time.Now()
	sourceAsset, err := setup.NativeToWrappedAsset(setupEnv.AssetMappings, chainId, sourceChain, setupEnv.NativeEvmToken)
	if err != nil {
		t.Fatal(err)
	}

	nativeEvm := setupEnv.Clients.EVM[chainId]
	nativeInstance, err := setup.InitAssetContract(setupEnv.NativeEvmToken, nativeEvm.EVMClient)
	if err != nil {
		t.Fatal(err)
	}

	nativeBalanceBefore, err := nativeInstance.BalanceOf(&bind.CallOpts{}, nativeEvm.Receiver)
	if err != nil {
		t.Fatal(err)
	}

	// Step 2 - Submit Lock Txn from a deployed smart contract
	receipt, expectedLockEventLog := sendBurnEthTransaction(setupEnv.AssetMappings, wrappedEvm, setupEnv.NativeEvmToken, chainId, sourceChain, nativeEvm.Receiver.Bytes(), amount, t)

	// Step 2.1 - Get the block timestamp of the burn event
	block, err := wrappedEvm.EVMClient.BlockByNumber(context.Background(), receipt.BlockNumber)
	if err != nil {
		t.Fatal("failed to get block by number", err)
	}
	blockTimestamp := time.Unix(int64(block.Time()), 0).UTC()

	// Step 3 - Validate Burn Event was emitted with correct data
	burnEventId := validateBurnEvent(receipt, expectedLockEventLog, t)

	// Step 4 - Verify the submitted topic messages
	receivedSignatures := verifyTopicMessages(setupEnv, burnEventId, t)

	// Step 5 - Verify Transfer retrieved from Validator API
	transactionData := verifyFungibleTransferFromValidatorAPI(setupEnv, nativeEvm, burnEventId, setupEnv.NativeEvmToken, fmt.Sprint(amount), setupEnv.NativeEvmToken, t)

	// Step 6 - Submit Mint transaction
	txHash := submitUnlockTransaction(nativeEvm, burnEventId, transactionData, common.HexToAddress(setupEnv.NativeEvmToken), t)

	// Step 7 - Wait for transaction to be mined
	waitForTransaction(nativeEvm, txHash, t)

	expectedUnlockedAmount := calculateExpectedUnlockAmount(nativeEvm, setupEnv.NativeEvmToken, amount, t)

	// Step 8 - Validate Token balances
	verifyWrappedAssetBalance(nativeEvm, setupEnv.NativeEvmToken, expectedUnlockedAmount, nativeBalanceBefore, nativeEvm.Receiver, t)

	// Step 9 - Prepare expected Transfer record
	expectedLockEventRecord := util.PrepareExpectedTransfer(
		sourceChain,
		chainId,
		chainId,
		burnEventId,
		sourceAsset,
		setupEnv.NativeEvmToken,
		setupEnv.NativeEvmToken,
		strconv.FormatInt(amount, 10),
		nativeEvm.Receiver.String(),
		status.Completed,
		wrappedEvm.Signer.Address(),
		entity.NanoTime{Time: blockTimestamp})

	authMsgBytes, err := auth_message.
		EncodeFungibleBytesFrom(
			expectedLockEventRecord.SourceChainID,
			expectedLockEventRecord.TargetChainID,
			expectedLockEventRecord.TransactionID,
			expectedLockEventRecord.TargetAsset,
			expectedLockEventRecord.Receiver,
			strconv.FormatInt(amount, 10))
	if err != nil {
		t.Fatalf("[%s] - Failed to encode the authorisation signature. Error: [%s]", expectedLockEventRecord.TransactionID, err)
	}

	// Step 9 - Verify Database Records
	verifyTransferRecordAndSignatures(setupEnv.DbValidator, expectedLockEventRecord, authMsgBytes, receivedSignatures, t)
}

// Test_Hedera_Native_EVM_NFT_Transfer recreates User who wants to portal a Hedera Native NFT to an EVM chain.
func Test_Hedera_Native_EVM_NFT_Transfer(t *testing.T) {
	now = time.Now()
	setupEnv := setup.Load()
	nftToken := setupEnv.NftTokenID.String()
	serialNumber := setupEnv.NftSerialNumber

	chainId := firstEvmChainId
	evm := setupEnv.Clients.EVM[chainId]
	receiver := evm.Receiver

	targetAsset, err := setup.NativeToWrappedAsset(setupEnv.AssetMappings, constants.HederaNetworkId, chainId, nftToken)
	if err != nil {
		t.Fatalf("Expecting Token [%s] is not supported. - Error: [%s]", setupEnv.NftTokenID.String(), err)
	}

	transferFee := setupEnv.NftConstantFees[nftToken]
	validatorsFee := setupEnv.Clients.Distributor.ValidAmount(transferFee)

	// Step 1 - Get Token Metadata
	nftData, err := setupEnv.Clients.MirrorNode.GetNft(nftToken, serialNumber)

	if err != nil {
		t.Fatalf("Failed to get mirror node nft. Error [%s]", err)
	}
	originator := nftData.AccountID
	nftInfo, ok := setupEnv.AssetMappings.NonFungibleAssetInfo(constants.HederaNetworkId, nftToken)
	if !ok {
		t.Fatalf("Failed to asset info for NFT [%s]. Error [%s]", nftToken, err)
	}
	if originator != nftInfo.TreasuryAccountId {
		transferFee += nftInfo.CustomFeeTotalAmounts.FallbackFeeAmountInHbar
	}

	decodedMetadata, e := base64.StdEncoding.DecodeString(nftData.Metadata)
	if e != nil {
		t.Fatalf("Failed to decode metadata [%s]. Error [%s]", nftData.Metadata, e)
	}

	nftIDString := fmt.Sprintf("%d@%s", serialNumber, nftToken)
	nftID, err := hedera.NftIDFromString(nftIDString)
	if err != nil {
		t.Fatalf("Failed to parse NFT ID [%s]. Error: [%s]", nftIDString, err)
	}

	// Step 2 - Send the NFT Allowance for the Payer Account
	_, err = sendNFTAllowance(setupEnv, nftID, setupEnv.Clients.Hedera.GetOperatorAccountID(), setupEnv.PayerAccount)
	if err != nil {
		t.Fatalf("Failed to send Allowance for NFT [%s]. Error: [%s]", nftIDString, err)
	}
	signaturesStartTime := time.Now().UnixNano()
	// Step 3 - Send the NFT transfer, including the fee to the Bridge Account
	feeResponse, err := sendFeeForNFTToBridgeAccount(setupEnv, evm.Receiver, chainId, nftID, transferFee)
	if err != nil {
		t.Fatalf("Failed to send Fee and allowance for NFT transfer. Error: [%s]", err)
	}
	transactionID := hederahelper.FromHederaTransactionID(feeResponse.TransactionID).String()

	// Step 4 - Validate that a scheduled NFT transaction to the bridge account was submitted by a validator
	scheduledTxID, scheduleID := validateScheduledNftTransfer(setupEnv, transactionID, nftToken, serialNumber, originator, setupEnv.BridgeAccount.String(), t)

	// Step 5 - Verify the submitted topic messages
	receivedSignatures := verifyTopicMessagesWithStartTime(setupEnv, hederahelper.FromHederaTransactionID(feeResponse.TransactionID).String(), signaturesStartTime, t)

	// Step 6 - Validate members fee scheduled transaction
	scheduledTxID, scheduleID = validateMembersScheduledTxs(setupEnv, constants.Hbar, generateMirrorNodeExpectedTransfersForHederaTransfer(setupEnv, constants.Hbar, validatorsFee), t)

	// Step 7 - Verify Non-Fungible Transfer retrieved from Validator API
	transactionData := verifyNonFungibleTransferFromValidatorAPI(
		setupEnv,
		evm,
		transactionID,
		nftToken, string(decodedMetadata), serialNumber, targetAsset, t)

	// Step 7.1 - Get the consensus timestamp of the transfer
	tx, err := setupEnv.Clients.MirrorNode.GetSuccessfulTransaction(transactionID)
	if err != nil {
		t.Fatal("failed to get successful transaction", err)
	}
	nanos, err := timestamp.FromString(tx.ConsensusTimestamp)
	if err != nil {
		t.Fatal("failed to parse consensus timestamp", err)
	}
	ts := timestamp.FromNanos(nanos)

	// Step 8 - Submit Mint ERC-721 transaction
	txHash := submitMintERC721Transaction(evm, transactionID, transactionData, t)

	// Step 9 - Wait for transaction to be mined
	waitForTransaction(evm, txHash, t)

	// Step 10 - Validate EVM TokenId
	verifyERC721TokenId(evm.EVMClient, targetAsset, serialNumber, receiver.String(), string(decodedMetadata), t)

	// Step 11 - Verify Database records
	expectedTxRecord := &entity.Transfer{
		TransactionID: transactionID,
		SourceChainID: constants.HederaNetworkId,
		TargetChainID: chainId,
		NativeChainID: constants.HederaNetworkId,
		SourceAsset:   nftToken,
		TargetAsset:   targetAsset,
		NativeAsset:   nftToken,
		Receiver:      receiver.String(),
		Amount:        "",
		Fee:           strconv.FormatInt(validatorsFee, 10),
		Status:        status.Completed,
		SerialNumber:  serialNumber,
		Metadata:      string(decodedMetadata),
		IsNft:         true,
		Originator:    setupEnv.Clients.Hedera.GetOperatorAccountID().String(),
		Timestamp:     entity.NanoTime{Time: ts},
	}
	// and:
	expectedFeeRecord := util.PrepareExpectedFeeRecord(
		scheduledTxID,
		scheduleID, validatorsFee,
		transactionID)
	// and:
	expectedScheduleTransferRecord := &entity.Schedule{
		TransactionID: scheduledTxID,
		ScheduleID:    scheduleID,
		Operation:     schedule.TRANSFER,
		HasReceiver:   false,
		Status:        status.Completed,
		TransferID: sql.NullString{
			String: transactionID,
			Valid:  true,
		},
	}
	// and recreate signed auth message
	authMsgBytes, err := auth_message.
		EncodeNftBytesFrom(
			expectedTxRecord.SourceChainID,
			expectedTxRecord.TargetChainID,
			expectedTxRecord.TransactionID,
			expectedTxRecord.TargetAsset,
			expectedTxRecord.SerialNumber,
			expectedTxRecord.Metadata,
			expectedTxRecord.Receiver)
	if err != nil {
		t.Fatalf("[%s] - Failed to encode the authorisation signature. Error: [%s]", expectedTxRecord.TransactionID, err)
	}

	// Step 12 - Verify Database Records
	verifyTransferRecordAndSignatures(setupEnv.DbValidator, expectedTxRecord, authMsgBytes, receivedSignatures, t)
	// and:
	verifyFeeRecord(setupEnv.DbValidator, expectedFeeRecord, t)
	// and:
	verifyScheduleRecord(setupEnv.DbValidator, expectedScheduleTransferRecord, t)
}

// Test_Hedera_EVM_BurnERC721_Transfer recreates User who wants to portal back a Hedera Native NFT from an EVM chain.
func Test_Hedera_EVM_BurnERC721_Transfer(t *testing.T) {
	now = time.Now()
	setupEnv := setup.Load()
	nftToken := setupEnv.NftTokenID.String()
	serialNumber := setupEnv.NftSerialNumber

	chainId := firstEvmChainId
	evm := setupEnv.Clients.EVM[chainId]

	wrappedAsset, err := setup.NativeToWrappedAsset(setupEnv.AssetMappings, constants.HederaNetworkId, chainId, nftToken)
	if err != nil {
		t.Fatalf("Expecting Token [%s] is not supported. - Error: [%s]", nftToken, err)
	}

	// 1. Validate that NFT spender is the bridge account
	validateNftOwner(setupEnv, nftToken, serialNumber, setupEnv.BridgeAccount, t)

	// 2. Submit burnERC721 transaction to the bridge contract
	burnTxReceipt, expectedRouterBurnERC721 := sendBurnERC721Transaction(evm, wrappedAsset, constants.HederaNetworkId, setupEnv.Clients.Hedera.GetOperatorAccountID().ToBytes(), serialNumber, t)

	// 2.1 - Get the block timestamp of the burn event
	block, err := evm.EVMClient.BlockByNumber(context.Background(), burnTxReceipt.BlockNumber)
	if err != nil {
		t.Fatal("failed to get block by number", err)
	}
	blockTimestamp := time.Unix(int64(block.Time()), 0).UTC()

	// 3. Validate that the burn ERC-721 transaction went through and emitted the correct event
	expectedTxId := validateBurnERC721Event(burnTxReceipt, expectedRouterBurnERC721, t)

	// 4. Validate that a scheduled NFT transaction was submitted
	scheduledTxID, scheduleID := validateScheduledNftAllowanceApprove(t, setupEnv, expectedTxId, blockTimestamp.UnixNano())

	// 5. Validate Event Transaction ID retrieved from Validator API
	validateEventTransactionIDFromValidatorAPI(setupEnv, expectedTxId, scheduledTxID, t)

	// 6. Validate that the NFT was sent to the receiver account
	validateNftSpender(t, setupEnv, nftToken, serialNumber, setupEnv.Clients.Hedera.GetOperatorAccountID())

	// 7. Prepare Expected Database Records
	expectedTxRecord := &entity.Transfer{
		TransactionID: expectedTxId,
		SourceChainID: chainId,
		TargetChainID: constants.HederaNetworkId,
		NativeChainID: constants.HederaNetworkId,
		SourceAsset:   wrappedAsset,
		TargetAsset:   nftToken,
		NativeAsset:   nftToken,
		Receiver:      setupEnv.Clients.Hedera.GetOperatorAccountID().String(),
		Status:        status.Completed,
		SerialNumber:  serialNumber,
		IsNft:         true,
		Originator:    evm.Signer.Address(),
		Timestamp:     entity.NanoTime{Time: blockTimestamp},
	}
	expectedScheduleTransferRecord := &entity.Schedule{
		TransactionID: scheduledTxID,
		ScheduleID:    scheduleID,
		Operation:     schedule.TRANSFER,
		HasReceiver:   true,
		Status:        status.Completed,
		TransferID: sql.NullString{
			String: expectedTxId,
			Valid:  true,
		},
	}

	// 8. Wait for validators to update DB state after Scheduled TX is mined
	time.Sleep(20 * time.Second)

	// 9. Validate Database Records
	verifyTransferRecord(setupEnv.DbValidator, expectedTxRecord, t)
	// and:
	verifyScheduleRecord(setupEnv.DbValidator, expectedScheduleTransferRecord, t)
}

func validateNftOwner(setup *setup.Setup, tokenID string, serialNumber int64, expectedOwner hedera.AccountID, t *testing.T) {
	nftID, err := hedera.NftIDFromString(fmt.Sprintf("%d@%s", serialNumber, tokenID))
	if err != nil {
		t.Fatal(err)
	}

	nftInfo, err := hedera.NewTokenNftInfoQuery().
		SetNftID(nftID).
		Execute(setup.Clients.Hedera)
	if err != nil {
		t.Fatal(err)
	}

	if len(nftInfo) != 1 {
		t.Fatalf("Invalid NFT Info [%s] length result. Result: [%v]", nftID.String(), nftInfo)
	}

	owner := nftInfo[0].AccountID
	if owner != expectedOwner {
		t.Fatalf("Invalid NftID [%s] owner. Expected [%s], actual [%s].", nftID.String(), expectedOwner, owner)
	}
}

func validateNftSpender(t *testing.T, setup *setup.Setup, tokenID string, serialNumber int64, expectedSpender hedera.AccountID) {
	tokenIdFromString, err := hedera.TokenIDFromString(tokenID)
	if err != nil {
		t.Fatal(err)
	}

	nftId := hedera.NftID{
		TokenID:      tokenIdFromString,
		SerialNumber: serialNumber,
	}

	nftInfo, err := hedera.NewTokenNftInfoQuery().
		SetNftID(nftId).
		Execute(setup.Clients.Hedera)
	if err != nil {
		t.Fatal(err)
	}

	if len(nftInfo) != 1 {
		t.Fatalf("Invalid NFT Info [%s] length result. Result: [%v]", nftId.String(), nftInfo)
	}

	spender := nftInfo[0].SpenderID
	if spender != expectedSpender {
		t.Fatalf("Invalid NftID [%s] spender. Expected [%s], actual [%s].", nftId.String(), expectedSpender, spender)
	}
}

func validateReceiverAccountBalance(setup *setup.Setup, expectedReceiveAmount uint64, beforeHbarBalance hedera.AccountBalance, asset string, t *testing.T) {
	afterHbarBalance := util.GetHederaAccountBalance(setup.Clients.Hedera, setup.Clients.Hedera.GetOperatorAccountID(), t)

	var beforeTransfer uint64
	var afterTransfer uint64

	if asset == constants.Hbar {
		beforeTransfer = uint64(beforeHbarBalance.Hbars.AsTinybar())
		afterTransfer = uint64(afterHbarBalance.Hbars.AsTinybar())
	} else {
		beforeTransfer = beforeHbarBalance.Token[setup.TokenID]
		afterTransfer = afterHbarBalance.Token[setup.TokenID]
	}

	if afterTransfer-beforeTransfer != expectedReceiveAmount {
		t.Fatalf("[%s] Expected %s balance after - [%d], but was [%d]. Expected to receive [%d], but was [%d]", setup.Clients.Hedera.GetOperatorAccountID(), asset, beforeTransfer+expectedReceiveAmount, afterTransfer, expectedReceiveAmount, afterTransfer-beforeTransfer)
	}
}

func validateAccountBalance(setup *setup.Setup, hederaID hedera.AccountID, expectedReceiveAmount uint64, beforeHbarBalance hedera.AccountBalance, asset string, t *testing.T) {
	afterHbarBalance := util.GetHederaAccountBalance(setup.Clients.Hedera, hederaID, t)

	tokenAsset, err := hedera.TokenIDFromString(asset)
	if err != nil {
		t.Fatal(err)
	}

	beforeTransfer := beforeHbarBalance.Token[tokenAsset]
	afterTransfer := afterHbarBalance.Token[tokenAsset]

	if afterTransfer-beforeTransfer != expectedReceiveAmount {
		t.Fatalf("[%s] Expected %s balance after - [%d], but was [%d]. Expected to receive [%d], but was [%d]", setup.Clients.Hedera.GetOperatorAccountID(), asset, beforeTransfer+expectedReceiveAmount, afterTransfer, expectedReceiveAmount, afterTransfer-beforeTransfer)
	}
}

func validateBurnEvent(txReceipt *types.Receipt, expectedRouterBurn *router.RouterBurn, t *testing.T) string {
	parsedAbi, err := abi.JSON(strings.NewReader(router.RouterABI))
	if err != nil {
		t.Fatal(err)
	}

	routerBurn := router.RouterBurn{}
	eventSignature := []byte("Burn(uint256,address,uint256,bytes)")
	eventSignatureHash := crypto.Keccak256Hash(eventSignature)
	for _, log := range txReceipt.Logs {
		if log.Topics[0] != eventSignatureHash {
			continue
		}

		//account := log.Topics[1]
		//wrappedAsset := log.Topics[2]
		err := parsedAbi.UnpackIntoInterface(&routerBurn, "Burn", log.Data)
		if err != nil {
			t.Fatal(err)
		}

		if routerBurn.Amount.String() != expectedRouterBurn.Amount.String() {
			t.Fatalf("Expected Burn Event Amount [%v], but actually was [%v]", expectedRouterBurn.Amount, routerBurn.Amount)
		}

		//if wrappedAsset != expectedRouterBurn.WrappedAsset.Hash() {
		//	t.Fatalf("Expected Burn Event Wrapped Token [%v], but actually was [%v]", expectedRouterBurn.WrappedAsset, routerBurn.WrappedAsset)
		//}

		if !reflect.DeepEqual(routerBurn.Receiver, expectedRouterBurn.Receiver) {
			t.Fatalf("Expected Burn Event Receiver [%v], but actually was [%v]", expectedRouterBurn.Receiver, routerBurn.Receiver)
		}

		//if account != expectedRouterBurn.Account.Hash() {
		//	t.Fatalf("Expected Burn Event Account [%v], but actually was [%v]", expectedRouterBurn.Account, routerBurn.Account)
		//}

		expectedId := fmt.Sprintf("%s-%d", log.TxHash, log.Index)
		return expectedId
	}

	t.Fatal("Could not retrieve valid Burn Event Log information.")
	return ""
}

func validateLockEvent(txReceipt *types.Receipt, expectedRouterLock *router.RouterLock, t *testing.T) string {
	parsedAbi, err := abi.JSON(strings.NewReader(router.RouterABI))
	if err != nil {
		t.Fatal(err)
	}

	routerLock := router.RouterLock{}
	eventSignature := []byte("Lock(uint256,address,bytes,uint256,uint256)")
	eventSignatureHash := crypto.Keccak256Hash(eventSignature)
	for _, log := range txReceipt.Logs {
		if log.Topics[0] != eventSignatureHash {
			continue
		}

		err := parsedAbi.UnpackIntoInterface(&routerLock, "Lock", log.Data)
		if err != nil {
			t.Fatal(err)
		}

		if routerLock.Amount.String() != expectedRouterLock.Amount.String() {
			t.Fatalf("Expected Lock Event Amount [%v], actual [%v]", expectedRouterLock.Amount, routerLock.Amount)
		}

		if routerLock.TargetChain.String() != expectedRouterLock.TargetChain.String() {
			t.Fatalf("Expected Lock Event Target Chain [%v], actual [%v]", expectedRouterLock.TargetChain, routerLock.TargetChain)
		}

		if routerLock.Token.String() != expectedRouterLock.Token.String() {
			t.Fatalf("Expected Lock Event Token [%v], actual [%v]", expectedRouterLock.Token, routerLock.Token)
		}

		if routerLock.ServiceFee.String() != expectedRouterLock.ServiceFee.String() {
			t.Fatalf("Expected Lock Event Service Fee [%v], actual [%s]", expectedRouterLock.ServiceFee, routerLock.ServiceFee)
		}

		if !reflect.DeepEqual(routerLock.Receiver, expectedRouterLock.Receiver) {
			t.Fatalf("Expected Lock Event Receiver [%v], actual [%v]", expectedRouterLock.Receiver, routerLock.Receiver)
		}

		expectedId := fmt.Sprintf("%s-%d", log.TxHash, log.Index)
		return expectedId
	}

	t.Fatal("Could not retrieve valid Burn Event Log information.")
	return ""
}

func validateBurnERC721Event(txReceipt *types.Receipt, expectedRouterLock *router.RouterBurnERC721, t *testing.T) string {
	parsedAbi, err := abi.JSON(strings.NewReader(router.RouterABI))
	if err != nil {
		t.Fatal(err)
	}

	event := router.RouterBurnERC721{}
	eventSignature := parsedAbi.Events["BurnERC721"].ID
	for _, log := range txReceipt.Logs {
		if log.Topics[0] != eventSignature {
			continue
		}

		err := parsedAbi.UnpackIntoInterface(&event, "BurnERC721", log.Data)
		if err != nil {
			t.Fatal(err)
		}

		if event.TokenId.String() != expectedRouterLock.TokenId.String() {
			t.Fatalf("Expected Burn-ERC721 Event TokenId [%v], actual [%v]", expectedRouterLock.TokenId, event.TokenId)
		}

		if event.TargetChain.String() != expectedRouterLock.TargetChain.String() {
			t.Fatalf("Expected Burn-ERC721 Event Target Chain [%v], actual [%v]", expectedRouterLock.TargetChain, event.TargetChain)
		}

		if !reflect.DeepEqual(event.WrappedToken, expectedRouterLock.WrappedToken) {
			t.Fatalf("Expected Burn-ERC721 Event Token [%v], actual [%v]", expectedRouterLock.WrappedToken, event.WrappedToken)
		}

		if !reflect.DeepEqual(event.Receiver, expectedRouterLock.Receiver) {
			t.Fatalf("Expected Burn-ERC721 Event Receiver [%v], actual [%v]", expectedRouterLock.Receiver, event.Receiver)
		}

		expectedId := fmt.Sprintf("%s-%d", log.TxHash, log.Index)
		return expectedId
	}

	t.Fatal("Could not retrieve valid Burn Event Log information.")
	return ""
}

func validateSubmittedScheduledTx(setupEnv *setup.Setup, asset string, expectedTransfers []transaction.Transfer, t *testing.T) (transactionID, scheduleID string) {
	receiverTransactionID, receiverScheduleID := validateScheduledTx(setupEnv, setupEnv.Clients.Hedera.GetOperatorAccountID(), asset, expectedTransfers, t)

	membersTransactionID, membersScheduleID := validateMembersScheduledTxs(setupEnv, asset, expectedTransfers, t)

	if receiverTransactionID != membersTransactionID {
		t.Fatalf("Scheduled Transactions between members are different. Receiver [%s], Member [%s]", receiverTransactionID, membersTransactionID)
	}

	if receiverScheduleID != membersScheduleID {
		t.Fatalf("Scheduled IDs between members are different. Receiver [%s], Member [%s]", receiverScheduleID, membersScheduleID)
	}

	return receiverTransactionID, receiverScheduleID
}

func validateScheduledMintTx(setupEnv *setup.Setup, account hedera.AccountID, asset string, expectedTransfers []transaction.Transfer, t *testing.T) (transactionID, scheduleID string) {
	timeLeft := 180
	for {
		response, err := setupEnv.Clients.MirrorNode.GetAccountTokenMintTransactionsAfterTimestamp(account, now.UnixNano())
		if err != nil {
			t.Fatal(err)
		}

		if len(response.Transactions) > 1 {
			t.Fatalf("[%s] - Found [%d] new transactions, must be 1.", account, len(response.Transactions))
		}

		txId, entityId := listenForTx(response, setupEnv.Clients.MirrorNode, expectedTransfers, asset, t)
		if txId != "" && entityId != "" {
			return txId, entityId
		}

		if timeLeft > 0 {
			fmt.Println(fmt.Sprintf("Could not find any scheduled transactions for account [%s]. Trying again. Time left: ~[%d] seconds", account, timeLeft))
			timeLeft -= 10
			time.Sleep(10 * time.Second)
			continue
		}
		break
	}

	t.Fatalf("Could not find any scheduled transactions for account [%s]", setupEnv.Clients.Hedera.GetOperatorAccountID())
	return "", ""
}

func validateScheduledBurnTx(setupEnv *setup.Setup, account hedera.AccountID, asset string, expectedTransfers []transaction.Transfer, t *testing.T) (transactionID, scheduleID string) {
	timeLeft := 180
	for {
		response, err := setupEnv.Clients.MirrorNode.GetAccountTokenBurnTransactionsAfterTimestamp(account, now.UnixNano())
		if err != nil {
			t.Fatal(err)
		}

		if len(response.Transactions) > 1 {
			t.Fatalf("[%s] - Found [%d] new transactions, must be 1.", account, len(response.Transactions))
		}

		txId, entityId := listenForTx(response, setupEnv.Clients.MirrorNode, expectedTransfers, asset, t)
		if txId != "" && entityId != "" {
			return txId, entityId
		}

		if timeLeft > 0 {
			fmt.Println(fmt.Sprintf("Could not find any scheduled transactions for account [%s]. Trying again. Time left: ~[%d] seconds", account, timeLeft))
			timeLeft -= 10
			time.Sleep(10 * time.Second)
			continue
		}
		break
	}

	t.Fatalf("Could not find any scheduled transactions for account [%s]", setupEnv.Clients.Hedera.GetOperatorAccountID())
	return "", ""
}

func validateScheduledNftAllowanceApprove(t *testing.T, setupEnv *setup.Setup, expectedTransactionID string, startTimestamp int64) (transactionID, scheduleID string) {
	receiver := setupEnv.Clients.Hedera.GetOperatorAccountID()
	timeLeft := 180

	for {
		scheduleCreates, err := setupEnv.Clients.MirrorNode.GetTransactionsAfterTimestamp(setupEnv.BridgeAccount, startTimestamp, read_only.CryptoApproveAllowance)
		if err != nil {
			t.Fatal(err)
		}

		for _, scheduleCreate := range scheduleCreates {
			scheduledTransaction, err := setupEnv.Clients.MirrorNode.GetScheduledTransaction(scheduleCreate.TransactionID)
			if err != nil {
				t.Fatalf("Could not get scheduled transaction for [%s]", scheduleCreate.TransactionID)
			}

			for _, tx := range scheduledTransaction.Transactions {
				schedule, err := setupEnv.Clients.MirrorNode.GetSchedule(tx.EntityId)
				if err != nil {
					t.Fatalf("Could not get schedule entity for [%s]", tx.EntityId)
				}

				if schedule.Memo == expectedTransactionID {
					return tx.TransactionID, tx.EntityId
				}
			}
		}

		if timeLeft > 0 {
			fmt.Println(fmt.Sprintf("Could not find any scheduled transactions for NFT Transfer for account [%s]. Trying again. Time left: ~[%d] seconds", receiverAccountId, timeLeft))
			timeLeft -= 10
			time.Sleep(10 * time.Second)
			continue
		}
		break
	}

	t.Fatalf("Could not find any scheduled transactions for account [%s]", setupEnv.Clients.Hedera.GetOperatorAccountID())
	return "", ""
}

func listenForTx(response *transaction.Response, mirrorNode *mirror_node.Client, expectedTransfers []transaction.Transfer, asset string, t *testing.T) (string, string) {
	for _, transaction := range response.Transactions {
		if transaction.Scheduled == true {
			scheduleCreateTx, err := mirrorNode.GetTransaction(transaction.TransactionID)
			if err != nil {
				t.Fatal(err)
			}

			for _, expectedTransfer := range expectedTransfers {
				found := false
				if asset == constants.Hbar {
					for _, transfer := range transaction.Transfers {
						if expectedTransfer == transfer {
							found = true
							break
						}
					}
				} else {
					for _, transfer := range transaction.TokenTransfers {
						if expectedTransfer == transfer {
							found = true
							break
						}
					}
				}

				if !found {
					t.Fatalf("[%s] - Expected transfer [%v] not found.", transaction.TransactionID, expectedTransfer)
				}
			}

			for _, tx := range scheduleCreateTx.Transactions {
				if tx.EntityId != "" {
					return tx.TransactionID, tx.EntityId
				}
			}
		}
	}
	return "", ""
}

func validateScheduledTx(setupEnv *setup.Setup, account hedera.AccountID, asset string, expectedTransfers []transaction.Transfer, t *testing.T) (transactionID, scheduleID string) {
	timeLeft := 180
	for {
		response, err := setupEnv.Clients.MirrorNode.GetAccountCreditTransactionsAfterTimestamp(account, now.UnixNano())
		if err != nil {
			t.Fatal(err)
		}

		if len(response.Transactions) > 1 {
			t.Fatalf("[%s] - Found [%d] new transactions, must be 1.", account, len(response.Transactions))
		}

		txId, entityId := listenForTx(response, setupEnv.Clients.MirrorNode, expectedTransfers, asset, t)
		if txId != "" && entityId != "" {
			return txId, entityId
		}

		if timeLeft > 0 {
			fmt.Println(fmt.Sprintf("Could not find any scheduled transactions for account [%s]. Trying again. Time left: ~[%d] seconds", setupEnv.Clients.Hedera.GetOperatorAccountID(), timeLeft))
			timeLeft -= 10
			time.Sleep(10 * time.Second)
			continue
		}
		break
	}

	t.Fatalf("Could not find any scheduled transactions for account [%s]", setupEnv.Clients.Hedera.GetOperatorAccountID())
	return "", ""
}

func validateMembersScheduledTxs(setupEnv *setup.Setup, asset string, expectedTransfers []transaction.Transfer, t *testing.T) (transactionID, scheduleID string) {
	if len(setupEnv.Members) == 0 {
		return "", ""
	}

	var transactions []string
	var scheduleIDs []string
	for _, member := range setupEnv.Members {
		txID, scheduleID := validateScheduledTx(setupEnv, member, asset, expectedTransfers, t)
		transactions = append(transactions, txID)

		if !util.AllSame(transactions) {
			t.Fatalf("Transaction [%s] does not match with previously added transactions.", txID)
		}
		scheduleIDs = append(scheduleIDs, scheduleID)

		if !util.AllSame(scheduleIDs) {
			t.Fatalf("ScheduleID [%s] does not match with previously added ids", scheduleID)
		}
	}

	return transactions[0], scheduleIDs[0]
}

func calculateReceiverAndFeeAmounts(setup *setup.Setup, token string, amount int64) (receiverAmount, fee int64) {
	fee, remainder := setup.Clients.FeeCalculator.CalculateFee(token, amount)
	validFee := setup.Clients.Distributor.ValidAmount(fee)
	if validFee != fee {
		remainder += fee - validFee
	}

	return remainder, validFee
}

func calculateExpectedUnlockAmount(evm setup.EVMUtils, token string, amount int64, t *testing.T) *big.Int {
	amountBn := big.NewInt(amount)

	feeData, err := evm.RouterContract.TokenFeeData(nil, common.HexToAddress(token))
	if err != nil {
		t.Fatal(err)
	}

	precision, err := evm.RouterContract.ServiceFeePrecision(nil)
	if err != nil {
		t.Fatal(err)
	}

	multiplied := new(big.Int).Mul(amountBn, feeData.ServiceFeePercentage)
	serviceFee := new(big.Int).Div(multiplied, precision)

	return new(big.Int).Sub(amountBn, serviceFee)
}

func submitMintTransaction(evm setup.EVMUtils, txId string, transactionData *service.FungibleTransferData, tokenAddress common.Address, t *testing.T) common.Hash {
	var signatures [][]byte
	for i := 0; i < len(transactionData.Signatures); i++ {
		signature, err := hex.DecodeString(transactionData.Signatures[i])
		if err != nil {
			t.Fatalf("Failed to decode signature with error: [%s]", err)
		}
		signatures = append(signatures, signature)
	}
	mintAmount, ok := new(big.Int).SetString(transactionData.Amount, 10)
	if !ok {
		t.Fatalf("Could not convert mint amount [%s] to big int", transactionData.Amount)
	}

	res, err := evm.RouterContract.Mint(
		evm.KeyTransactor,
		new(big.Int).SetUint64(transactionData.SourceChainId),
		[]byte(txId),
		tokenAddress,
		evm.Receiver,
		mintAmount,
		signatures,
	)

	if err != nil {
		t.Fatalf("Cannot execute transaction - Error: [%s].", err)
	}
	return res.Hash()
}

func submitMintERC721Transaction(evm setup.EVMUtils, txId string, transactionData *service.NonFungibleTransferData, t *testing.T) common.Hash {
	var signatures [][]byte
	for i := 0; i < len(transactionData.Signatures); i++ {
		signature, err := hex.DecodeString(transactionData.Signatures[i])
		if err != nil {
			t.Fatalf("Failed to decode signature with error: [%s]", err)
		}
		signatures = append(signatures, signature)
	}

	res, err := evm.RouterContract.MintERC721(
		evm.KeyTransactor,
		new(big.Int).SetUint64(transactionData.SourceChainId),
		[]byte(txId),
		common.HexToAddress(transactionData.TargetAsset),
		big.NewInt(transactionData.TokenId),
		transactionData.Metadata,
		evm.Receiver,
		signatures,
	)

	if err != nil {
		t.Fatalf("Cannot execute transaction - Error: [%s].", err)
	}
	return res.Hash()
}

func submitUnlockTransaction(evm setup.EVMUtils, txId string, transactionData *service.FungibleTransferData, tokenAddress common.Address, t *testing.T) common.Hash {
	var signatures [][]byte
	for i := 0; i < len(transactionData.Signatures); i++ {
		signature, err := hex.DecodeString(transactionData.Signatures[i])
		if err != nil {
			t.Fatalf("Failed to decode signature with error: [%s]", err)
		}
		signatures = append(signatures, signature)
	}
	mintAmount, ok := new(big.Int).SetString(transactionData.Amount, 10)
	if !ok {
		t.Fatalf("Could not convert mint amount [%s] to big int", transactionData.Amount)
	}

	res, err := evm.RouterContract.Unlock(
		evm.KeyTransactor,
		new(big.Int).SetUint64(transactionData.SourceChainId),
		[]byte(txId),
		tokenAddress,
		mintAmount,
		evm.Receiver,
		signatures,
	)

	if err != nil {
		t.Fatalf("Cannot execute transaction - Error: [%s].", err)
	}
	return res.Hash()
}

func generateMirrorNodeExpectedTransfersForBurnEvent(setupEnv *setup.Setup, asset string, amount, fee int64) []transaction.Transfer {
	total := amount + fee
	feePerMember := fee / int64(len(setupEnv.Members))

	var expectedTransfers []transaction.Transfer
	expectedTransfers = append(expectedTransfers, transaction.Transfer{
		Account: setupEnv.BridgeAccount.String(),
		Amount:  -total,
	},
		transaction.Transfer{
			Account: setupEnv.Clients.Hedera.GetOperatorAccountID().String(),
			Amount:  amount,
		})

	for _, member := range setupEnv.Members {
		expectedTransfers = append(expectedTransfers, transaction.Transfer{
			Account: member.String(),
			Amount:  feePerMember,
		})
	}

	if asset != constants.Hbar {
		for i := range expectedTransfers {
			expectedTransfers[i].Token = asset
		}
	}

	return expectedTransfers
}

func generateMirrorNodeExpectedTransfersForLockEvent(setupEnv *setup.Setup, asset string, amount int64) []transaction.Transfer {
	expectedTransfers := []transaction.Transfer{
		{
			Account: setupEnv.BridgeAccount.String(),
			Amount:  -amount,
			Token:   asset,
		},
		{
			Account: setupEnv.Clients.Hedera.GetOperatorAccountID().String(),
			Amount:  amount,
			Token:   asset,
		},
	}

	return expectedTransfers
}

func generateMirrorNodeExpectedTransfersForHederaTransfer(setupEnv *setup.Setup, asset string, fee int64) []transaction.Transfer {
	feePerMember := fee / int64(len(setupEnv.Members))

	var expectedTransfers []transaction.Transfer
	expectedTransfers = append(expectedTransfers, transaction.Transfer{
		Account: setupEnv.BridgeAccount.String(),
		Amount:  -fee,
	})

	for _, member := range setupEnv.Members {
		expectedTransfers = append(expectedTransfers, transaction.Transfer{
			Account: member.String(),
			Amount:  feePerMember,
		})
	}

	if asset != constants.Hbar {
		for i := range expectedTransfers {
			expectedTransfers[i].Token = asset
		}
	}

	return expectedTransfers
}

func sendBurnEthTransaction(assetsService service.Assets, evm setup.EVMUtils, asset string, sourceChainId, targetChainId uint64, receiver []byte, amount int64, t *testing.T) (*types.Receipt, *router.RouterBurn) {
	wrappedAsset, err := setup.NativeToWrappedAsset(assetsService, sourceChainId, targetChainId, asset)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("Parsed [%s] to ETH Token [%s]", asset, wrappedAsset))

	approvedValue := big.NewInt(amount)

	instance, err := setup.InitAssetContract(wrappedAsset, evm.EVMClient)
	if err != nil {
		t.Fatal(err)
	}

	approveTx, err := instance.Approve(evm.KeyTransactor, evm.RouterAddress, approvedValue)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(fmt.Sprintf("[%s] Waiting for Approval Transaction", approveTx.Hash()))
	waitForTransaction(evm, approveTx.Hash(), t)

	burnTx, err := evm.RouterContract.Burn(evm.KeyTransactor, new(big.Int).SetUint64(sourceChainId), common.HexToAddress(wrappedAsset), approvedValue, receiver)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("[%s] Submitted Burn Transaction", burnTx.Hash()))

	expectedRouterBurn := &router.RouterBurn{
		//Account:      common.HexToAddress(evm.Signer.Address()),
		//WrappedAsset: *wrappedAsset,
		Amount:   approvedValue,
		Receiver: receiver,
	}

	burnTxHash := burnTx.Hash()

	fmt.Println(fmt.Sprintf("[%s] Waiting for Burn Transaction Receipt.", burnTxHash))
	burnTxReceipt, err := evm.EVMClient.WaitForTransactionReceipt(burnTxHash)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("[%s] Burn Transaction mined and retrieved receipt.", burnTxHash))

	return burnTxReceipt, expectedRouterBurn
}

func sendLockEthTransaction(evm setup.EVMUtils, asset string, targetChainId uint64, receiver []byte, amount int64, t *testing.T) (*types.Receipt, *router.RouterLock) {
	approvedValue := big.NewInt(amount)

	instance, err := setup.InitAssetContract(asset, evm.EVMClient)
	if err != nil {
		t.Fatal(err)
	}
	approveTx, err := instance.Approve(evm.KeyTransactor, evm.RouterAddress, approvedValue)
	if err != nil {
		t.Fatal(err)
	}

	feeData, err := evm.RouterContract.TokenFeeData(nil, common.HexToAddress(asset))
	if err != nil {
		t.Fatal(err)
	}

	precision, err := evm.RouterContract.ServiceFeePrecision(nil)
	if err != nil {
		t.Fatal(err)
	}

	multiplied := new(big.Int).Mul(approvedValue, feeData.ServiceFeePercentage)
	serviceFee := new(big.Int).Div(multiplied, precision)

	fmt.Println(fmt.Sprintf("[%s] Waiting for Approval Transaction", approveTx.Hash()))
	waitForTransaction(evm, approveTx.Hash(), t)

	lockTx, err := evm.RouterContract.Lock(evm.KeyTransactor, new(big.Int).SetUint64(targetChainId), common.HexToAddress(asset), approvedValue, receiver)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("[%s] Submitted Lock Transaction", lockTx.Hash()))

	expectedRouterLock := &router.RouterLock{
		TargetChain: new(big.Int).SetUint64(targetChainId),
		Token:       common.HexToAddress(asset),
		Receiver:    receiver,
		Amount:      approvedValue,
		ServiceFee:  serviceFee,
		Raw:         types.Log{},
	}

	lockTxHash := lockTx.Hash()

	fmt.Println(fmt.Sprintf("[%s] Waiting for Lock Transaction Receipt.", lockTxHash))
	lockTxReceipt, err := evm.EVMClient.WaitForTransactionReceipt(lockTxHash)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("[%s] Lock Transaction mined and retrieved receipt.", lockTxHash))

	return lockTxReceipt, expectedRouterLock
}

func sendBurnERC721Transaction(evm setup.EVMUtils, wrappedToken string, targetChainId uint64, receiver []byte, serialNumber int64, t *testing.T) (*types.Receipt, *router.RouterBurnERC721) {
	wrappedAddress := common.HexToAddress(wrappedToken)

	paymentToken, err := evm.RouterContract.Erc721Payment(nil, wrappedAddress)
	if err != nil {
		t.Fatal(err)
	}

	fee, err := evm.RouterContract.Erc721Fee(nil, wrappedAddress)
	if err != nil {
		t.Fatal(err)
	}

	erc20Contract, err := wtoken.NewWtoken(paymentToken, evm.EVMClient.GetClient())
	if err != nil {
		t.Fatal(err)
	}

	approveERC20Tx, err := erc20Contract.Approve(evm.KeyTransactor, evm.RouterAddress, fee)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(fmt.Sprintf("[%s] Waiting for ERC-20 Approval Transaction", approveERC20Tx.Hash()))
	waitForTransaction(evm, approveERC20Tx.Hash(), t)

	erc721Contract, err := werc721.NewWerc721(wrappedAddress, evm.EVMClient)
	if err != nil {
		t.Fatal(err)
	}
	tokenId := big.NewInt(serialNumber)

	approveERC721Tx, err := erc721Contract.Approve(evm.KeyTransactor, evm.RouterAddress, tokenId)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(fmt.Sprintf("[%s] Waiting for ERC-721 Approval Transaction", approveERC721Tx.Hash()))
	waitForTransaction(evm, approveERC721Tx.Hash(), t)
	targetChainIdBigInt := new(big.Int).SetUint64(targetChainId)
	burnTx, err := evm.RouterContract.BurnERC721(evm.KeyTransactor, targetChainIdBigInt, wrappedAddress, tokenId, paymentToken, fee, receiver)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("[%s] Submitted Burn Transaction", burnTx.Hash()))

	expectedRouterBurn := &router.RouterBurnERC721{
		TargetChain:  targetChainIdBigInt,
		WrappedToken: common.HexToAddress(wrappedToken),
		TokenId:      tokenId,
		Receiver:     receiver,
	}

	burnTxHash := burnTx.Hash()

	fmt.Println(fmt.Sprintf("[%s] Waiting for Burn ERC-721 Transaction Receipt.", burnTxHash))
	burnTxReceipt, err := evm.EVMClient.WaitForTransactionReceipt(burnTxHash)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("[%s] Burn ERC-721 Transaction mined and retrieved receipt.", burnTxHash))

	return burnTxReceipt, expectedRouterBurn
}

func waitForTransaction(evm setup.EVMUtils, txHash common.Hash, t *testing.T) {
	receipt, err := evm.EVMClient.WaitForTransactionReceipt(txHash)
	if err != nil {
		t.Fatalf("[%s] - Error occurred while monitoring. Error: [%s]", txHash, err)
		return
	}

	if receipt.Status == 1 {
		fmt.Println(fmt.Sprintf("TX [%s] was successfully mined", txHash))
	} else {
		t.Fatalf("TX [%s] reverted", txHash)
	}
}

func verifyWrappedAssetBalance(evm setup.EVMUtils, nativeAsset string, mintAmount *big.Int, wrappedBalanceBefore *big.Int, wTokenReceiverAddress common.Address, t *testing.T) {
	instance, err := setup.InitAssetContract(nativeAsset, evm.EVMClient)
	if err != nil {
		t.Fatal(err)
	}

	wrappedBalanceAfter, err := instance.BalanceOf(&bind.CallOpts{}, wTokenReceiverAddress)
	if err != nil {
		t.Fatal(err)
	}

	if err != nil {
		t.Fatal(err)
	}

	expectedBalance := new(big.Int).Add(wrappedBalanceBefore, mintAmount)

	if wrappedBalanceAfter.Cmp(expectedBalance) != 0 {
		t.Fatalf("Incorrect token balance. Expected to be [%s], but was [%s].", expectedBalance, wrappedBalanceAfter)
	}
}

func verifyERC721TokenId(evm *evm.Client, wrappedToken string, serialNumber int64, receiver string, expectedMetadata string, t *testing.T) {
	contract, err := werc721.NewWerc721(common.HexToAddress(wrappedToken), evm)
	if err != nil {
		t.Fatal(err)
	}

	owner, err := contract.OwnerOf(nil, big.NewInt(serialNumber))
	if err != nil {
		t.Fatal(err)
	}

	if owner.String() != receiver {
		t.Fatalf("Invalid owner. Expected owner for serial number [%d] to be [%s], but was [%s]", serialNumber, receiver, owner.String())
	}

	tokenURI, err := contract.TokenURI(nil, big.NewInt(serialNumber))
	if err != nil {
		t.Fatal(err)
	}

	if expectedMetadata != tokenURI {
		t.Fatalf("Invalid token URI. Expected token URI for serial number [%d] to be [%s], but was [%s]", serialNumber, expectedMetadata, tokenURI)
	}
}

func validateEventTransactionIDFromValidatorAPI(setupEnv *setup.Setup, eventID, expectedTxID string, t *testing.T) {
	actualTxID, err := setupEnv.Clients.ValidatorClient.GetEventTransactionID(eventID)
	if err != nil {
		t.Fatalf("[%s] - Failed to get event transaction ID. Error: [%s]", eventID, err)
	}

	if actualTxID != expectedTxID {
		t.Fatalf("Expected Event TX ID [%s] did not match actual TX ID [%s]", expectedTxID, actualTxID)
	}
}

func verifyFungibleTransferFromValidatorAPI(setupEnv *setup.Setup, evm setup.EVMUtils, txId, tokenID, expectedSendAmount, targetAsset string, t *testing.T) *service.FungibleTransferData {
	bytes, err := setupEnv.Clients.ValidatorClient.GetTransferData(txId)
	if err != nil {
		t.Fatalf("Cannot fetch transaction data - Error: [%s].", err)
	}
	var transferDataResponse *service.FungibleTransferData
	err = json.Unmarshal(bytes, &transferDataResponse)
	if err != nil {
		t.Fatalf("Failed to parse JSON transaction data [%s]. Error: [%s]", bytes, err)
	}

	if transferDataResponse.IsNft {
		t.Fatalf("Transaction data mismatch: Expected response data to not be NFT related.")
	}
	if transferDataResponse.Amount != expectedSendAmount {
		t.Fatalf("Transaction data mismatch: Expected [%s], but was [%s]", expectedSendAmount, transferDataResponse.Amount)
	}
	if transferDataResponse.NativeAsset != tokenID {
		t.Fatalf("Native Token mismatch: Expected [%s], but was [%s]", setupEnv.TokenID.String(), transferDataResponse.NativeAsset)
	}
	if transferDataResponse.Recipient != evm.Receiver.String() {
		t.Fatalf("Receiver address mismatch: Expected [%s], but was [%s]", evm.Receiver.String(), transferDataResponse.Recipient)
	}
	if transferDataResponse.TargetAsset != targetAsset {
		t.Fatalf("Token address mismatch: Expected [%s], but was [%s]", targetAsset, transferDataResponse.TargetAsset)
	}

	return transferDataResponse
}

func verifyNonFungibleTransferFromValidatorAPI(setupEnv *setup.Setup, evm setup.EVMUtils, txId, tokenID, metadata string, tokenIdOrSerialNum int64, targetAsset string, t *testing.T) *service.NonFungibleTransferData {
	bytes, err := setupEnv.Clients.ValidatorClient.GetTransferData(txId)
	if err != nil {
		t.Fatalf("Cannot fetch transaction data - Error: [%s].", err)
	}
	var transferDataResponse *service.NonFungibleTransferData
	err = json.Unmarshal(bytes, &transferDataResponse)
	if err != nil {
		t.Fatalf("Failed to parse JSON transaction data [%s]. Error: [%s]", bytes, err)
	}

	if !transferDataResponse.IsNft {
		t.Fatalf("Transaction data mismatch: Expected response data to be NFT related.")
	}
	if transferDataResponse.Metadata != metadata {
		t.Fatalf("Transaction data mismatch: Expected [%s], but was [%s]", metadata, transferDataResponse.Metadata)
	}
	if transferDataResponse.TokenId != tokenIdOrSerialNum {
		t.Fatalf("Transaction tokenId/serialNum mismatch: Expected [%d], but was [%d]", tokenIdOrSerialNum, transferDataResponse.TokenId)
	}
	if transferDataResponse.NativeAsset != tokenID {
		t.Fatalf("Native Token mismatch: Expected [%s], but was [%s]", setupEnv.TokenID.String(), transferDataResponse.NativeAsset)
	}
	if transferDataResponse.Recipient != evm.Receiver.String() {
		t.Fatalf("Receiver address mismatch: Expected [%s], but was [%s]", evm.Receiver.String(), transferDataResponse.Recipient)
	}
	if transferDataResponse.TargetAsset != targetAsset {
		t.Fatalf("Token address mismatch: Expected [%s], but was [%s]", targetAsset, transferDataResponse.TargetAsset)
	}

	return transferDataResponse
}

func verifyFeeRecord(dbValidation *database.Service, expectedRecord *entity.Fee, t *testing.T) {
	ok, err := dbValidation.VerifyFeeRecord(expectedRecord)
	if err != nil {
		t.Fatalf("[%s] - Verification of database records failed - Error: [%s].", expectedRecord.TransactionID, err)
	}
	if !ok {
		t.Fatalf("[%s] - Database does not contain expected fee records", expectedRecord.TransactionID)
	}
}

func verifyTransferRecord(dbValidation *database.Service, expectedRecord *entity.Transfer, t *testing.T) {
	exist, err := dbValidation.VerifyTransferRecord(expectedRecord)
	if err != nil {
		t.Fatalf("[%s] - Verification of database records failed - Error: [%s].", expectedRecord.TransactionID, err)
	}
	if !exist {
		t.Fatalf("[%s] - Database does not contain expected transfer records", expectedRecord.TransactionID)
	}
}

func verifyScheduleRecord(dbValidation *database.Service, expectedRecord *entity.Schedule, t *testing.T) {
	exist, err := dbValidation.VerifyScheduleRecord(expectedRecord)
	if err != nil {
		t.Fatalf("[%s] - Verification of database records failed - Error: [%s].", expectedRecord.TransactionID, err)
	}
	if !exist {
		t.Fatalf("[%s] - Database does not contain expected schedule records", expectedRecord.TransactionID)
	}
}

func verifyTransferRecordAndSignatures(dbValidation *database.Service, expectedRecord *entity.Transfer, authMsgBytes []byte, signatures []string, t *testing.T) {
	exist, err := dbValidation.VerifyTransferAndSignatureRecords(expectedRecord, authMsgBytes, signatures)
	if err != nil {
		t.Fatalf("[%s] - Verification of database records failed - Error: [%s].", expectedRecord.TransactionID, err)
	}
	if !exist {
		t.Fatalf("[%s] - Database does not contain expected records", expectedRecord.TransactionID)
	}
}

func verifyTransferToBridgeAccount(s *setup.Setup, wrappedAsset string, evm setup.EVMUtils, memo string, whbarReceiverAddress common.Address, expectedAmount int64, t *testing.T) (hedera.TransactionResponse, *big.Int) {
	instance, err := setup.InitAssetContract(wrappedAsset, evm.EVMClient)
	if err != nil {
		t.Fatal(err)
	}
	// Get the wrapped hbar balance of the receiver before the transfer
	whbarBalanceBefore, err := instance.BalanceOf(&bind.CallOpts{}, whbarReceiverAddress)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(fmt.Sprintf("WHBAR balance before transaction: [%s]", whbarBalanceBefore))
	// Get bridge account hbar balance before transfer
	receiverBalance := util.GetHederaAccountBalance(s.Clients.Hedera, s.BridgeAccount, t).Hbars.AsTinybar()

	fmt.Println(fmt.Sprintf("Bridge account balance HBAR balance before transaction: [%d]", receiverBalance))

	// Get the transaction receipt to verify the transaction was executed
	transactionResponse, err := sendHbarsToBridgeAccount(s, memo, expectedAmount)
	if err != nil {
		t.Fatalf("Unable to send HBARs to Bridge Account, Error: [%s]", err)
	}

	transactionReceipt, err := transactionResponse.GetReceipt(s.Clients.Hedera)
	if err != nil {
		t.Fatalf("Transaction unsuccessful, Error: [%s]", err)
	}

	fmt.Println(fmt.Sprintf("Successfully sent HBAR to bridge account, Status: [%s]", transactionReceipt.Status))

	// Get bridge account hbar balance after transfer
	receiverBalanceNew := util.GetHederaAccountBalance(s.Clients.Hedera, s.BridgeAccount, t).Hbars.AsTinybar()

	fmt.Println(fmt.Sprintf("Bridge Account HBAR balance after transaction: [%d]", receiverBalanceNew))

	// Verify that the custodial address has received exactly the amount sent
	amount := receiverBalanceNew - receiverBalance

	// Verify that the bridge account has received exactly the amount sent
	if amount != expectedAmount {
		t.Fatalf("Expected to receive the exact transfer amount of hbar: [%v], but was [%v]", expectedAmount, amount)
	}

	return *transactionResponse, whbarBalanceBefore
}

func verifyTokenTransferToBridgeAccount(s *setup.Setup, evmAsset string, tokenID hedera.TokenID, evm setup.EVMUtils, memo string, wTokenReceiverAddress common.Address, amount int64, t *testing.T) (hedera.TransactionResponse, *big.Int) {
	instance, err := setup.InitAssetContract(evmAsset, evm.EVMClient)
	if err != nil {
		t.Fatal(err)
	}
	// Get the wrapped hts token balance of the receiver before the transfer
	wrappedBalanceBefore, err := instance.BalanceOf(&bind.CallOpts{}, wTokenReceiverAddress)
	if err != nil {
		t.Fatalf("Unable to query the token balance of the receiver account. Error: [%s]", err)
	}

	fmt.Println(fmt.Sprintf("Token balance before transaction: [%s]", wrappedBalanceBefore))
	// Get bridge account token balance before transfer
	receiverBalance := util.GetHederaAccountBalance(s.Clients.Hedera, s.BridgeAccount, t)

	fmt.Println(fmt.Sprintf("Bridge account Token balance before transaction: [%d]", receiverBalance.Token[s.TokenID]))
	// Get the transaction receipt to verify the transaction was executed
	transactionResponse, err := sendTokensToBridgeAccount(s, tokenID, memo, amount)
	if err != nil {
		t.Fatalf(fmt.Sprintf("Unable to send Tokens to Bridge Account, Error: [%s]", err))
	}
	transactionReceipt, err := transactionResponse.GetReceipt(s.Clients.Hedera)
	if err != nil {
		t.Fatalf(fmt.Sprintf("Transaction unsuccessful, Error: [%s]", err))
	}
	fmt.Println(fmt.Sprintf("Successfully sent Tokens to bridge account, Status: [%s]", transactionReceipt.Status))

	// Get bridge account HTS token balance after transfer
	receiverBalanceNew := util.GetHederaAccountBalance(s.Clients.Hedera, s.BridgeAccount, t)

	fmt.Println(fmt.Sprintf("Bridge Account Token balance after transaction: [%d]", receiverBalanceNew.Token[s.TokenID]))

	// Verify that the custodial address has received exactly the amount sent
	resultAmount := receiverBalanceNew.Token[tokenID] - receiverBalance.Token[tokenID]
	// Verify that the bridge account has received exactly the amount sent
	if resultAmount != uint64(amount) {
		t.Fatalf("Expected to receive the exact transfer amount of hbar: [%v], but received: [%v]", amount, resultAmount)
	}

	return *transactionResponse, wrappedBalanceBefore
}

func verifyTopicMessages(setup *setup.Setup, txId string, t *testing.T) []string {
	fmt.Println(fmt.Sprintf("Waiting for Signatures & TX Hash to be published to Topic [%v]", setup.TopicID.String()))

	return verifyTopicMessagesWithStartTime(setup, txId, time.Now().UnixNano(), t)
}

func verifyTopicMessagesWithStartTime(setup *setup.Setup, txId string, startTime int64, t *testing.T) []string {
	ethSignaturesCollected := 0
	var receivedSignatures []string

	fmt.Println(fmt.Sprintf("Waiting for Signatures & TX Hash to be published to Topic [%v]", setup.TopicID.String()))

	// Subscribe to Topic
	subscription, err := hedera.NewTopicMessageQuery().
		SetStartTime(time.Unix(0, startTime)).
		SetTopicID(setup.TopicID).
		Subscribe(
			setup.Clients.Hedera,
			func(response hedera.TopicMessage) {
				msg := &validatorproto.TopicMessage{}
				err := proto.Unmarshal(response.Contents, msg)
				if err != nil {
					t.Fatal(err)
				}

				var transferID string
				var signature string
				switch msg.Message.(type) {
				case *validatorproto.TopicMessage_FungibleSignatureMessage:
					message := msg.GetFungibleSignatureMessage()
					transferID = message.TransferID
					signature = message.Signature
					break
				case *validatorproto.TopicMessage_NftSignatureMessage:
					message := msg.GetNftSignatureMessage()
					transferID = message.TransferID
					signature = message.Signature
				}

				//Verify that all the submitted messages have signed the same transaction
				if transferID != txId {
					fmt.Println(fmt.Sprintf(`Expected signature message to contain the transaction id: [%s]`, txId))
				} else {
					receivedSignatures = append(receivedSignatures, signature)
					ethSignaturesCollected++
					fmt.Println(fmt.Sprintf("Received Auth Signature [%s]", signature))
				}
			},
		)
	if err != nil {
		t.Fatalf("Unable to subscribe to Topic [%s]", setup.TopicID)
	}

	select {
	case <-time.After(120 * time.Second):
		if ethSignaturesCollected != expectedValidatorsCount {
			t.Fatalf("Expected the count of collected signatures to equal the number of validators: [%v], but was: [%v]", expectedValidatorsCount, ethSignaturesCollected)
		}
		subscription.Unsubscribe()
		return receivedSignatures
	}
	// Not possible end-case
	return nil
}

func sendHbarsToBridgeAccount(setup *setup.Setup, memo string, amount int64) (*hedera.TransactionResponse, error) {
	hbarSendAmount := hedera.HbarFromTinybar(amount)
	hbarRemovalAmount := hedera.HbarFromTinybar(-amount)
	fmt.Println(fmt.Sprintf("Sending [%v] Hbars through the Bridge. Transaction Memo: [%s]", hbarSendAmount, memo))

	res, err := hedera.NewTransferTransaction().
		AddHbarTransfer(setup.Clients.Hedera.GetOperatorAccountID(), hbarRemovalAmount).
		AddHbarTransfer(setup.BridgeAccount, hbarSendAmount).
		SetTransactionMemo(memo).
		Execute(setup.Clients.Hedera)
	if err != nil {
		return nil, err
	}
	rec, err := res.GetReceipt(setup.Clients.Hedera)
	if err != nil {
		return nil, err
	}

	fmt.Println(fmt.Sprintf("TX broadcasted. ID [%s], Status: [%s]", res.TransactionID, rec.Status))
	time.Sleep(1 * time.Second)

	return &res, err
}

func sendNFTAllowance(setup *setup.Setup, nftId hedera.NftID, ownerAccountId, spenderAccountId hedera.AccountID) (*hedera.TransactionResponse, error) {
	fmt.Println(fmt.Sprintf("Sending Allowance for NFT [%s] to account [%s]", nftId.String(), spenderAccountId.String()))
	res, err := hedera.NewAccountAllowanceApproveTransaction().
		ApproveTokenNftAllowance(
			nftId,
			ownerAccountId,
			spenderAccountId,
		).Execute(setup.Clients.Hedera)

	if err != nil {
		return nil, err
	}
	rec, err := res.GetReceipt(setup.Clients.Hedera)
	if err != nil {
		return nil, err
	}
	fmt.Println(fmt.Sprintf("TX broadcasted. ID [%s], Status: [%s]", res.TransactionID, rec.Status))
	time.Sleep(4 * time.Second)

	return &res, err
}

func sendTokensToBridgeAccount(setup *setup.Setup, tokenID hedera.TokenID, memo string, amount int64) (*hedera.TransactionResponse, error) {
	fmt.Println(fmt.Sprintf("Sending [%v] Tokens to the Bridge. Transaction Memo: [%s]", amount, memo))

	res, err := hedera.NewTransferTransaction().
		SetTransactionMemo(memo).
		AddTokenTransfer(tokenID, setup.Clients.Hedera.GetOperatorAccountID(), -amount).
		AddTokenTransfer(tokenID, setup.BridgeAccount, amount).
		Execute(setup.Clients.Hedera)
	if err != nil {
		return nil, err
	}
	rec, err := res.GetReceipt(setup.Clients.Hedera)
	if err != nil {
		return nil, err
	}

	fmt.Println(fmt.Sprintf("TX broadcasted. ID [%s], Status: [%s]", res.TransactionID, rec.Status))
	time.Sleep(1 * time.Second)

	return &res, err
}

func sendFeeForNFTToBridgeAccount(setup *setup.Setup, receiver common.Address, chainId uint64, nftID hedera.NftID, fee int64) (*hedera.TransactionResponse, error) {
	hbarSendAmount := hedera.HbarFromTinybar(fee)
	hbarRemovalAmount := hedera.HbarFromTinybar(-fee)

	memo := fmt.Sprintf("%d-%s-%s", chainId, receiver, nftID.String())
	fmt.Println(fmt.Sprintf("Sending Fungible Fee for NFT [%s] through the Portal. Transaction Memo: [%s]", nftID.String(), memo))

	res, err := hedera.NewTransferTransaction().
		AddHbarTransfer(setup.Clients.Hedera.GetOperatorAccountID(), hbarRemovalAmount).
		AddHbarTransfer(setup.BridgeAccount, hbarSendAmount).
		SetTransactionMemo(memo).
		Execute(setup.Clients.Hedera)
	if err != nil {
		return nil, err
	}
	rec, err := res.GetReceipt(setup.Clients.Hedera)
	if err != nil {
		return nil, err
	}

	fmt.Println(fmt.Sprintf("TX broadcasted. ID [%s], Status: [%s]", res.TransactionID, rec.Status))
	time.Sleep(1 * time.Second)

	return &res, err
}

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

package e2e

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/wtoken"
	mirror_node "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	hederahelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/schedule"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/limechain/hedera-eth-bridge-validator/e2e/util"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	entity_transfer "github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/e2e/service/database"
	"github.com/limechain/hedera-eth-bridge-validator/e2e/setup"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashgraph/hedera-sdk-go/v2"
	validatorproto "github.com/limechain/hedera-eth-bridge-validator/proto"
	"google.golang.org/protobuf/proto"
)

var (
	receiveAmount     int64 = 100
	tinyBarAmount     int64 = 1000000000
	hBarSendAmount          = hedera.HbarFromTinybar(tinyBarAmount)
	hbarRemovalAmount       = hedera.HbarFromTinybar(-tinyBarAmount)
	now               time.Time
)

const (
	expectedValidatorsCount = 3
)

// Test_HBAR recreates a real life situation of a user who wants to bridge a Hedera HBARs to the EVM Network infrastructure. The wrapped token on the EVM network(corresponding to the native Hedera Hashgraph's HBARs) gets minted, then transferred to the recipient account on the EVM network.
func Test_HBAR(t *testing.T) {
	setupEnv := setup.Load()
	now = time.Now()

	chainId := int64(80001) // represents Polygon Mumbai Testnet (e2e config must have configuration for that particular network)
	evm := setupEnv.Clients.EVM[chainId]
	receiver := evm.Receiver
	memo := fmt.Sprintf("%d-%s", chainId, evm.Receiver.String())

	targetAsset, err := setup.NativeToWrappedAsset(setupEnv.AssetMappings, 0, chainId, constants.Hbar)
	if err != nil {
		t.Fatalf("Expecting Token [%s] is not supported. - Error: [%s]", constants.Hbar, err)
	}

	mintAmount, fee := calculateReceiverAndFeeAmounts(setupEnv, constants.Hbar, hBarSendAmount.AsTinybar())

	// Step 1 - Verify the transfer of Hbars to the Bridge Account
	transactionResponse, wrappedBalanceBefore := verifyTransferToBridgeAccount(setupEnv, targetAsset, evm, memo, receiver, t)

	// Step 2 - Verify the submitted topic messages
	receivedSignatures := verifyTopicMessages(setupEnv, hederahelper.FromHederaTransactionID(&transactionResponse.TransactionID).String(), t)

	// Step 3 - Validate fee scheduled transaction
	scheduledTxID, scheduleID := validateMembersScheduledTxs(setupEnv, constants.Hbar, generateMirrorNodeExpectedTransfersForHederaTransfer(setupEnv, constants.Hbar, fee), t)

	// Step 4 - Verify Transfer retrieved from Validator API
	transactionData := verifyTransferFromValidatorAPI(setupEnv, evm, hederahelper.FromHederaTransactionID(&transactionResponse.TransactionID).String(), constants.Hbar, mintAmount, targetAsset, t)

	// Step 5 - Submit Mint transaction
	txHash := submitMintTransaction(evm, hederahelper.FromHederaTransactionID(&transactionResponse.TransactionID).String(), transactionData, common.HexToAddress(targetAsset), t)

	// Step 6 - Wait for transaction to be mined
	waitForTransaction(evm, txHash, t)

	// Step 7 - Validate Token balances
	verifyWrappedAssetBalance(evm, targetAsset, big.NewInt(mintAmount), wrappedBalanceBefore, receiver, t)

	// Step 8 - Prepare Comparable Expected Transfer Record
	expectedTxRecord := util.PrepareExpectedTransfer(
		0,
		chainId,
		0,
		hederahelper.FromHederaTransactionID(&transactionResponse.TransactionID).String(),
		constants.Hbar,
		targetAsset,
		constants.Hbar,
		strconv.FormatInt(hBarSendAmount.AsTinybar(), 10),
		receiver.String(),
		database.ExpectedStatuses{
			Status: entity_transfer.StatusCompleted,
		})
	// and:
	expectedFeeRecord := util.PrepareExpectedFeeRecord(
		scheduledTxID,
		scheduleID,
		fee,
		hederahelper.ToMirrorNodeTransactionID(transactionResponse.TransactionID.String()))

	// Step 9 - Verify Database Records
	verifyTransferRecordAndSignatures(setupEnv.DbValidator, expectedTxRecord, strconv.FormatInt(mintAmount, 10), receivedSignatures, t)
	// and:
	verifyFeeRecord(setupEnv.DbValidator, expectedFeeRecord, t)
}

// Test_E2E_Token_Transfer recreates a real life situation of a user who wants to bridge a Hedera native token to the EVM Network infrastructure. The wrapped token on the EVM network(corresponding to the native Hedera Hashgraph's one) gets minted, then transferred to the recipient account on the EVM network.
func Test_E2E_Token_Transfer(t *testing.T) {
	setupEnv := setup.Load()
	now = time.Now()

	chainId := int64(80001) // represents Polygon Mumbai Testnet (e2e config must have configuration for that particular network)
	evm := setupEnv.Clients.EVM[chainId]
	memo := fmt.Sprintf("%d-%s", chainId, evm.Receiver.String())
	mintAmount, fee := calculateReceiverAndFeeAmounts(setupEnv, setupEnv.TokenID.String(), tinyBarAmount)

	targetAsset, err := setup.NativeToWrappedAsset(setupEnv.AssetMappings, 0, chainId, setupEnv.TokenID.String())
	if err != nil {
		t.Fatalf("Expecting Token [%s] is not supported. - Error: [%s]", constants.Hbar, err)
	}

	// Step 1 - Verify the transfer of HTS to the Bridge Account
	transactionResponse, wrappedBalanceBefore := verifyTokenTransferToBridgeAccount(setupEnv, targetAsset, setupEnv.TokenID, evm, memo, evm.Receiver, tinyBarAmount, t)

	// Step 2 - Verify the submitted topic messages
	receivedSignatures := verifyTopicMessages(setupEnv, hederahelper.FromHederaTransactionID(&transactionResponse.TransactionID).String(), t)

	// Step 3 - Validate fee scheduled transaction
	scheduledTxID, scheduleID := validateMembersScheduledTxs(setupEnv, setupEnv.TokenID.String(), generateMirrorNodeExpectedTransfersForHederaTransfer(setupEnv, setupEnv.TokenID.String(), fee), t)

	// Step 4 - Verify Transfer retrieved from Validator API
	transactionData := verifyTransferFromValidatorAPI(setupEnv, evm, hederahelper.FromHederaTransactionID(&transactionResponse.TransactionID).String(), setupEnv.TokenID.String(), mintAmount, targetAsset, t)

	// Step 5 - Submit Mint transaction
	txHash := submitMintTransaction(evm, hederahelper.FromHederaTransactionID(&transactionResponse.TransactionID).String(), transactionData, common.HexToAddress(targetAsset), t)

	// Step 6 - Wait for transaction to be mined
	waitForTransaction(evm, txHash, t)

	// Step 7 - Validate Token balances
	verifyWrappedAssetBalance(evm, targetAsset, big.NewInt(mintAmount), wrappedBalanceBefore, evm.Receiver, t)

	// Step 8 - Verify Database records
	expectedTxRecord := util.PrepareExpectedTransfer(
		0,
		chainId,
		0,
		hederahelper.FromHederaTransactionID(&transactionResponse.TransactionID).String(),
		setupEnv.TokenID.String(),
		targetAsset,
		setupEnv.TokenID.String(),
		strconv.FormatInt(tinyBarAmount, 10),
		evm.Receiver.String(),
		database.ExpectedStatuses{
			Status: entity_transfer.StatusCompleted,
		})
	// and:
	expectedFeeRecord := util.PrepareExpectedFeeRecord(
		scheduledTxID,
		scheduleID, fee,
		hederahelper.ToMirrorNodeTransactionID(transactionResponse.TransactionID.String()))

	// Step 9 - Verify Database Records
	verifyTransferRecordAndSignatures(setupEnv.DbValidator, expectedTxRecord, strconv.FormatInt(mintAmount, 10), receivedSignatures, t)
	// and:
	verifyFeeRecord(setupEnv.DbValidator, expectedFeeRecord, t)
}

// Test_EVM_Hedera_HBAR recreates a real life situation of a user who wants to return a Hedera native HBARs from the EVM Network infrastructure. The wrapped HBARs on the EVM network(corresponding to the native Hedera Hashgraph's one) gets burned, then the locked HBARs on the Hedera bridge account get unlocked, forwarding them to the recipient account.
func Test_EVM_Hedera_HBAR(t *testing.T) {
	setupEnv := setup.Load()

	chainId := int64(80001) // represents Polygon Mumbai Testnet (e2e config must have configuration for that particular network)
	evm := setupEnv.Clients.EVM[chainId]
	now = time.Now()
	accountBalanceBefore := util.GetHederaAccountBalance(setupEnv.Clients.Hedera, setupEnv.Clients.Hedera.GetOperatorAccountID(), t)

	targetAsset, err := setup.NativeToWrappedAsset(setupEnv.AssetMappings, 0, chainId, constants.Hbar)
	if err != nil {
		t.Fatalf("Expecting Token [%s] is not supported. - Error: [%s]", constants.Hbar, err)
	}

	// 1. Calculate Expected Receive And Fee Amounts
	expectedReceiveAmount, fee := calculateReceiverAndFeeAmounts(setupEnv, constants.Hbar, receiveAmount)

	// 2. Submit burn transaction to the bridge contract
	burnTxReceipt, expectedRouterBurn := sendBurnEthTransaction(setupEnv.AssetMappings, evm, constants.Hbar, 0, chainId, setupEnv.Clients.Hedera.GetOperatorAccountID().ToBytes(), t)

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
		0,
		0,
		expectedId,
		targetAsset,
		constants.Hbar,
		constants.Hbar,
		strconv.FormatInt(receiveAmount, 10),
		setupEnv.Clients.Hedera.GetOperatorAccountID().String(),
		database.ExpectedStatuses{
			Status: entity_transfer.StatusCompleted,
		})
	// and:
	expectedFeeRecord := util.PrepareExpectedFeeRecord(transactionID, scheduleID, fee, expectedId)

	// 8. Wait for validators to update DB state after Scheduled TX is mined
	time.Sleep(10 * time.Second)

	// 9. Validate Database Records
	verifyTransferRecord(setupEnv.DbValidator, expectedBurnEventRecord, t)
	// and:
	verifyFeeRecord(setupEnv.DbValidator, expectedFeeRecord, t)
}

// Test_EVM_Hedera_Token recreates a real life situation of a user who wants to return a Hedera native token from the EVM Network infrastructure. The wrapped token on the EVM network(corresponding to the native Hedera one) gets burned, then the amount gets unlocked on the Hedera bridge account, forwarding it to the recipient account.
func Test_EVM_Hedera_Token(t *testing.T) {
	setupEnv := setup.Load()

	chainId := int64(80001) // represents Polygon Mumbai Testnet (e2e config must have configuration for that particular network)
	evm := setupEnv.Clients.EVM[chainId]
	now = time.Now()
	accountBalanceBefore := util.GetHederaAccountBalance(setupEnv.Clients.Hedera, setupEnv.Clients.Hedera.GetOperatorAccountID(), t)

	targetAsset, err := setup.NativeToWrappedAsset(setupEnv.AssetMappings, 0, chainId, setupEnv.TokenID.String())
	if err != nil {
		t.Fatalf("Expecting Token [%s] is not supported. - Error: [%s]", setupEnv.TokenID.String(), err)
	}

	// 1. Calculate Expected Receive Amount
	expectedReceiveAmount, fee := calculateReceiverAndFeeAmounts(setupEnv, setupEnv.TokenID.String(), receiveAmount)

	// 2. Submit burn transaction to the bridge contract
	burnTxReceipt, expectedRouterBurn := sendBurnEthTransaction(
		setupEnv.AssetMappings,
		evm,
		setupEnv.TokenID.String(),
		0,
		chainId,
		setupEnv.Clients.Hedera.GetOperatorAccountID().ToBytes(),
		t)

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
		0,
		0,
		expectedId,
		targetAsset,
		setupEnv.TokenID.String(),
		setupEnv.TokenID.String(),
		strconv.FormatInt(receiveAmount, 10),
		setupEnv.Clients.Hedera.GetOperatorAccountID().String(),
		database.ExpectedStatuses{
			Status: entity_transfer.StatusCompleted,
		})
	// and:
	expectedFeeRecord := util.PrepareExpectedFeeRecord(transactionID, scheduleID, fee, expectedId)

	// 8. Wait for validators to update DB state after Scheduled TX is mined
	time.Sleep(10 * time.Second)

	// 9. Validate Database Records
	verifyTransferRecord(setupEnv.DbValidator, expectedBurnEventRecord, t)
	// and:
	verifyFeeRecord(setupEnv.DbValidator, expectedFeeRecord, t)
}

// Test_EVM_Hedera_Native_Token recreates a real life situation of a user who wants to bridge an EVM native token to the Hedera infrastructure. A new wrapped token (corresponding to the native EVM one) gets minted to the bridge account, then gets transferred to the recipient account.
func Test_EVM_Hedera_Native_Token(t *testing.T) {
	// Step 1: Initialize setup, smart contracts, etc.
	setupEnv := setup.Load()

	chainId := int64(80001) // represents Polygon Mumbai Testnet (e2e config must have configuration for that particular network)
	evm := setupEnv.Clients.EVM[chainId]
	now = time.Now()
	bridgeAccountBalanceBefore := util.GetHederaAccountBalance(setupEnv.Clients.Hedera, setupEnv.BridgeAccount, t)
	receiverAccountBalanceBefore := util.GetHederaAccountBalance(setupEnv.Clients.Hedera, setupEnv.Clients.Hedera.GetOperatorAccountID(), t)
	targetAsset, err := setup.NativeToWrappedAsset(setupEnv.AssetMappings, chainId, 0, setupEnv.NativeEvmToken)
	if err != nil {
		t.Fatal(err)
	}

	// Step 2: Submit Lock Txn from a deployed smart contract
	receipt, expectedLockEventLog := sendLockEthTransaction(evm, setupEnv.NativeEvmToken, 0, setupEnv.Clients.Hedera.GetOperatorAccountID().ToBytes(), t)

	// Step 3: Validate Lock Event was emitted with correct data
	lockEventId := validateLockEvent(receipt, expectedLockEventLog, t)

	expectedAmount, err := removeDecimals(receiveAmount, common.HexToAddress(setupEnv.NativeEvmToken), evm)
	if err != nil {
		t.Fatal(err)
	}

	mintTransfer := []mirror_node.Transfer{
		{
			Account: setupEnv.BridgeAccount.String(),
			Amount:  expectedAmount,
			Token:   targetAsset,
		},
	}

	// Step 4: Validate that a scheduled token mint txn was submitted successfully
	bridgeMintTransactionID, bridgeMintScheduleID := validateScheduledMintTx(setupEnv, setupEnv.BridgeAccount, setupEnv.TokenID.String(), mintTransfer, t)

	// Wait for validators to update DB state after Scheduled TX is mined
	time.Sleep(10 * time.Second)

	// Step 5: Validate that Database statuses were changed correctly
	expectedLockEventRecord := util.PrepareExpectedTransfer(
		chainId,
		0,
		chainId,
		lockEventId,
		setupEnv.NativeEvmToken,
		targetAsset,
		setupEnv.NativeEvmToken,
		strconv.FormatInt(expectedAmount, 10),
		setupEnv.Clients.Hedera.GetOperatorAccountID().String(),
		database.ExpectedStatuses{
			Status: entity_transfer.StatusCompleted,
		})
	expectedScheduleMintRecord := &entity.Schedule{
		TransactionID: bridgeMintTransactionID,
		ScheduleID:    bridgeMintScheduleID,
		Operation:     schedule.MINT,
		Status:        schedule.StatusCompleted,
		TransferID: sql.NullString{
			String: lockEventId,
			Valid:  true,
		},
	}
	// Step 6: Verify that records have been created successfully
	verifyTransferRecord(setupEnv.DbValidator, expectedLockEventRecord, t)
	verifyScheduleRecord(setupEnv.DbValidator, expectedScheduleMintRecord, t)

	// Step 7: Validate that a scheduled transfer txn was submitted successfully
	bridgeTransferTransactionID, bridgeTransferScheduleID := validateScheduledTx(setupEnv, setupEnv.Clients.Hedera.GetOperatorAccountID(), setupEnv.TokenID.String(), generateMirrorNodeExpectedTransfersForLockEvent(setupEnv, targetAsset, receiveAmount), t)

	// Wait for validators to update DB state after Scheduled TX is mined
	time.Sleep(10 * time.Second)

	// Step 8: Validate that database statuses were updated correctly for the Schedule Transfer
	expectedScheduleTransferRecord := &entity.Schedule{
		TransactionID: bridgeTransferTransactionID,
		ScheduleID:    bridgeTransferScheduleID,
		Operation:     schedule.TRANSFER,
		Status:        schedule.StatusCompleted,
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

// Test_E2E_Hedera_EVM_Native_Token recreates a real life situation of a user who wants to bridge a Hedera native token to the EVM Network infrastructure. The wrapped token on the EVM network(corresponding to the native Hedera Hashgraph's one) gets minted, then transferred to the recipient account on the EVM network.
func Test_E2E_Hedera_EVM_Native_Token(t *testing.T) {
	setupEnv := setup.Load()
	now = time.Now()

	chainId := int64(80001) // represents Polygon Mumbai Testnet (e2e config must have configuration for that particular network)
	evm := setupEnv.Clients.EVM[chainId]
	memo := fmt.Sprintf("%d-%s", chainId, evm.Receiver.String())
	unlockAmount := int64(10)

	// Step 1 - Verify the transfer of HTS to the Bridge Account
	wrappedAsset, err := setup.NativeToWrappedAsset(setupEnv.AssetMappings, chainId, 0, setupEnv.NativeEvmToken)
	if err != nil {
		t.Fatal(err)
	}

	tokenID, err := hedera.TokenIDFromString(wrappedAsset)
	if err != nil {
		t.Fatal(err)
	}

	expectedUnlockAmount, err := addDecimals(unlockAmount, common.HexToAddress(setupEnv.NativeEvmToken), evm)
	if err != nil {
		t.Fatal(err)
	}

	transactionResponse, nativeBalanceBefore := verifyTokenTransferToBridgeAccount(setupEnv, setupEnv.NativeEvmToken, tokenID, evm, memo, evm.Receiver, expectedUnlockAmount, t)
	burnTransfer := []mirror_node.Transfer{
		{
			Account: setupEnv.BridgeAccount.String(),
			Amount:  -expectedUnlockAmount, // TODO: examine what amount exactly will be sent
			Token:   wrappedAsset,
		},
	}
	// Step 2 - Verify the submitted topic messages
	receivedSignatures := verifyTopicMessages(setupEnv, hederahelper.FromHederaTransactionID(&transactionResponse.TransactionID).String(), t)

	// Step 3 - Validate burn scheduled transaction
	burnTransactionID, burnScheduleID := validateScheduledBurnTx(setupEnv, setupEnv.BridgeAccount, setupEnv.TokenID.String(), burnTransfer, t)

	// Step 4 - Verify Transfer retrieved from Validator API
	transactionData := verifyTransferFromValidatorAPI(setupEnv, evm, hederahelper.FromHederaTransactionID(&transactionResponse.TransactionID).String(), setupEnv.NativeEvmToken, expectedUnlockAmount, setupEnv.NativeEvmToken, t)

	// Step 5 - Submit Unlock transaction
	txHash := submitUnlockTransaction(evm, hederahelper.FromHederaTransactionID(&transactionResponse.TransactionID).String(), transactionData, common.HexToAddress(setupEnv.NativeEvmToken), t)

	// Step 6 - Wait for transaction to be mined
	waitForTransaction(evm, txHash, t)

	//Step 7 - Validate Token balances
	verifyWrappedAssetBalance(evm, setupEnv.NativeEvmToken, big.NewInt(expectedUnlockAmount), nativeBalanceBefore, evm.Receiver, t)

	// Step 8 - Verify Database records
	expectedTxRecord := util.PrepareExpectedTransfer(
		0,
		chainId,
		chainId,
		hederahelper.FromHederaTransactionID(&transactionResponse.TransactionID).String(),
		wrappedAsset,
		setupEnv.NativeEvmToken,
		setupEnv.NativeEvmToken,
		strconv.FormatInt(expectedUnlockAmount, 10),
		evm.Receiver.String(),
		database.ExpectedStatuses{
			Status: entity_transfer.StatusCompleted,
		})

	// Step 8: Validate that database statuses were updated correctly for the Schedule Burn
	expectedScheduleBurnRecord := &entity.Schedule{
		TransactionID: burnTransactionID,
		ScheduleID:    burnScheduleID,
		Operation:     schedule.BURN,
		Status:        schedule.StatusCompleted,
		TransferID: sql.NullString{
			String: hederahelper.FromHederaTransactionID(&transactionResponse.TransactionID).String(),
			Valid:  true,
		},
	}

	// Step 9 - Verify Database Records
	verifyTransferRecordAndSignatures(setupEnv.DbValidator, expectedTxRecord, strconv.FormatInt(expectedUnlockAmount, 10), receivedSignatures, t)
	// and
	verifyScheduleRecord(setupEnv.DbValidator, expectedScheduleBurnRecord, t)
}

// Test_EVM_Native_to_EVM_Token recreates a real life situation of a user who wants to bridge an EVM native token to another EVM chain.
func Test_EVM_Native_to_EVM_Token(t *testing.T) {
	// Step 1 - Initialize setup, smart contracts, etc.
	setupEnv := setup.Load()

	chainId := int64(80001) // represents Polygon Mumbai Testnet (e2e config must have configuration for that particular network)
	evm := setupEnv.Clients.EVM[chainId]
	now = time.Now()
	targetChainID := int64(5) // represents Ethereum Goerli Testnet (e2e config must have configuration for that particular network)
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
	receipt, expectedLockEventLog := sendLockEthTransaction(evm, setupEnv.NativeEvmToken, targetChainID, evm.Receiver.Bytes(), t)

	// Step 3 - Validate Lock Event was emitted with correct data
	lockEventId := validateLockEvent(receipt, expectedLockEventLog, t)

	// Step 4 - Verify the submitted topic messages
	receivedSignatures := verifyTopicMessages(setupEnv, lockEventId, t)

	// Step 5 - Verify Transfer retrieved from Validator API
	transactionData := verifyTransferFromValidatorAPI(setupEnv, evm, lockEventId, setupEnv.NativeEvmToken, receiveAmount, wrappedAsset, t)

	// Step 6 - Submit Mint transaction
	txHash := submitMintTransaction(wrappedEvm, lockEventId, transactionData, common.HexToAddress(wrappedAsset), t)

	// Step 7 - Wait for transaction to be mined
	waitForTransaction(wrappedEvm, txHash, t)

	// Step 8 - Validate Token balances
	verifyWrappedAssetBalance(wrappedEvm, wrappedAsset, big.NewInt(receiveAmount), wrappedBalanceBefore, evm.Receiver, t)

	// Step 9 - Prepare expected Transfer record
	expectedLockEventRecord := util.PrepareExpectedTransfer(
		chainId,
		targetChainID,
		chainId,
		lockEventId,
		setupEnv.NativeEvmToken,
		wrappedAsset,
		setupEnv.NativeEvmToken,
		strconv.FormatInt(receiveAmount, 10),
		evm.Receiver.String(),
		database.ExpectedStatuses{
			Status: entity_transfer.StatusCompleted,
		})

	// Step 10 - Verify Database Records
	verifyTransferRecordAndSignatures(setupEnv.DbValidator, expectedLockEventRecord, strconv.FormatInt(receiveAmount, 10), receivedSignatures, t)
}

// Test_EVM_Wrapped_to_EVM_Token recreates a real life situation of a user who wants to bridge an EVM native token to another EVM chain.
func Test_EVM_Wrapped_to_EVM_Token(t *testing.T) {
	// Step 1 - Initialize setup, smart contracts, etc.
	setupEnv := setup.Load()

	chainId := int64(80001) // represents Polygon Mumbai Testnet (e2e config must have configuration for that particular network)
	sourceChain := int64(5) // represents Ethereum Goerli Testnet (e2e config must have configuration for that particular network)
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
	receipt, expectedLockEventLog := sendBurnEthTransaction(setupEnv.AssetMappings, wrappedEvm, setupEnv.NativeEvmToken, chainId, sourceChain, nativeEvm.Receiver.Bytes(), t)

	// Step 3 - Validate Burn Event was emitted with correct data
	burnEventId := validateBurnEvent(receipt, expectedLockEventLog, t)

	// Step 4 - Verify the submitted topic messages
	receivedSignatures := verifyTopicMessages(setupEnv, burnEventId, t)

	// Step 5 - Verify Transfer retrieved from Validator API
	transactionData := verifyTransferFromValidatorAPI(setupEnv, nativeEvm, burnEventId, setupEnv.NativeEvmToken, receiveAmount, setupEnv.NativeEvmToken, t)

	// Step 6 - Submit Mint transaction
	txHash := submitUnlockTransaction(nativeEvm, burnEventId, transactionData, common.HexToAddress(setupEnv.NativeEvmToken), t)

	// Step 7 - Wait for transaction to be mined
	waitForTransaction(nativeEvm, txHash, t)

	// Step 8 - Validate Token balances
	verifyWrappedAssetBalance(nativeEvm, setupEnv.NativeEvmToken, big.NewInt(receiveAmount), nativeBalanceBefore, nativeEvm.Receiver, t)

	// Step 9 - Prepare expected Transfer record
	expectedLockEventRecord := util.PrepareExpectedTransfer(
		sourceChain,
		chainId,
		chainId,
		burnEventId,
		sourceAsset,
		setupEnv.NativeEvmToken,
		setupEnv.NativeEvmToken,
		strconv.FormatInt(receiveAmount, 10),
		nativeEvm.Receiver.String(),
		database.ExpectedStatuses{
			Status: entity_transfer.StatusCompleted,
		})

	// Step 9 - Verify Database Records
	verifyTransferRecordAndSignatures(setupEnv.DbValidator, expectedLockEventRecord, strconv.FormatInt(receiveAmount, 10), receivedSignatures, t)
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
			t.Fatalf("Expected Lock Event Amount [%v], but actually was [%v]", expectedRouterLock.Amount, routerLock.Amount)
		}

		if routerLock.TargetChain.String() != expectedRouterLock.TargetChain.String() {
			t.Fatalf("Expected Lock Event Target Chain [%v], but actually was [%v]", expectedRouterLock.TargetChain, routerLock.TargetChain)
		}

		if routerLock.Token.String() != expectedRouterLock.Token.String() {
			t.Fatalf("Expected Lock Event Token [%v], but actually was [%v]", expectedRouterLock.Token, routerLock.Token)
		}

		if !reflect.DeepEqual(routerLock.Receiver, expectedRouterLock.Receiver) {
			t.Fatalf("Expected Lock Event Receiver [%v], but actually was [%v]", expectedRouterLock.Receiver, routerLock.Receiver)
		}

		expectedId := fmt.Sprintf("%s-%d", log.TxHash, log.Index)
		return expectedId
	}

	t.Fatal("Could not retrieve valid Burn Event Log information.")
	return ""
}

func validateSubmittedScheduledTx(setupEnv *setup.Setup, asset string, expectedTransfers []mirror_node.Transfer, t *testing.T) (transactionID, scheduleID string) {
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

func validateScheduledMintTx(setupEnv *setup.Setup, account hedera.AccountID, asset string, expectedTransfers []mirror_node.Transfer, t *testing.T) (transactionID, scheduleID string) {
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

func validateScheduledBurnTx(setupEnv *setup.Setup, account hedera.AccountID, asset string, expectedTransfers []mirror_node.Transfer, t *testing.T) (transactionID, scheduleID string) {
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

func listenForTx(response *mirror_node.Response, mirrorNode *mirror_node.Client, expectedTransfers []mirror_node.Transfer, asset string, t *testing.T) (string, string) {
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

func validateScheduledTx(setupEnv *setup.Setup, account hedera.AccountID, asset string, expectedTransfers []mirror_node.Transfer, t *testing.T) (transactionID, scheduleID string) {
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

func validateMembersScheduledTxs(setupEnv *setup.Setup, asset string, expectedTransfers []mirror_node.Transfer, t *testing.T) (transactionID, scheduleID string) {
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

func submitMintTransaction(evm setup.EVMUtils, txId string, transactionData *service.TransferData, tokenAddress common.Address, t *testing.T) common.Hash {
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
		big.NewInt(transactionData.SourceChainId),
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

func submitUnlockTransaction(evm setup.EVMUtils, txId string, transactionData *service.TransferData, tokenAddress common.Address, t *testing.T) common.Hash {
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
		big.NewInt(transactionData.SourceChainId),
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

func generateMirrorNodeExpectedTransfersForBurnEvent(setupEnv *setup.Setup, asset string, amount, fee int64) []mirror_node.Transfer {
	total := amount + fee
	feePerMember := fee / int64(len(setupEnv.Members))

	var expectedTransfers []mirror_node.Transfer
	expectedTransfers = append(expectedTransfers, mirror_node.Transfer{
		Account: setupEnv.BridgeAccount.String(),
		Amount:  -total,
	},
		mirror_node.Transfer{
			Account: setupEnv.Clients.Hedera.GetOperatorAccountID().String(),
			Amount:  amount,
		})

	for _, member := range setupEnv.Members {
		expectedTransfers = append(expectedTransfers, mirror_node.Transfer{
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

func generateMirrorNodeExpectedTransfersForLockEvent(setupEnv *setup.Setup, asset string, amount int64) []mirror_node.Transfer {
	expectedTransfers := []mirror_node.Transfer{
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

func generateMirrorNodeExpectedTransfersForHederaTransfer(setupEnv *setup.Setup, asset string, fee int64) []mirror_node.Transfer {
	feePerMember := fee / int64(len(setupEnv.Members))

	var expectedTransfers []mirror_node.Transfer
	expectedTransfers = append(expectedTransfers, mirror_node.Transfer{
		Account: setupEnv.BridgeAccount.String(),
		Amount:  -fee,
	})

	for _, member := range setupEnv.Members {
		expectedTransfers = append(expectedTransfers, mirror_node.Transfer{
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

func sendBurnEthTransaction(assetMappings config.Assets, evm setup.EVMUtils, asset string, sourceChainId, targetChainId int64, receiver []byte, t *testing.T) (*types.Receipt, *router.RouterBurn) {
	wrappedAsset, err := setup.NativeToWrappedAsset(assetMappings, sourceChainId, targetChainId, asset)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("Parsed [%s] to ETH Token [%s]", asset, wrappedAsset))

	approvedValue := big.NewInt(receiveAmount)

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

	burnTx, err := evm.RouterContract.Burn(evm.KeyTransactor, big.NewInt(sourceChainId), common.HexToAddress(wrappedAsset), approvedValue, receiver)
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

func sendLockEthTransaction(evm setup.EVMUtils, asset string, targetChainId int64, receiver []byte, t *testing.T) (*types.Receipt, *router.RouterLock) {
	approvedValue := big.NewInt(receiveAmount)

	instance, err := setup.InitAssetContract(asset, evm.EVMClient)
	if err != nil {
		t.Fatal(err)
	}
	approveTx, err := instance.Approve(evm.KeyTransactor, evm.RouterAddress, approvedValue)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(fmt.Sprintf("[%s] Waiting for Approval Transaction", approveTx.Hash()))
	waitForTransaction(evm, approveTx.Hash(), t)

	lockTx, err := evm.RouterContract.Lock(evm.KeyTransactor, big.NewInt(targetChainId), common.HexToAddress(asset), approvedValue, receiver)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("[%s] Submitted Lock Transaction", lockTx.Hash()))

	expectedRouterLock := &router.RouterLock{
		TargetChain: big.NewInt(targetChainId),
		Token:       common.HexToAddress(asset),
		Receiver:    receiver,
		Amount:      approvedValue,
		ServiceFee:  big.NewInt(0),
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

func validateEventTransactionIDFromValidatorAPI(setupEnv *setup.Setup, eventID, expectedTxID string, t *testing.T) {
	actualTxID, err := setupEnv.Clients.ValidatorClient.GetEventTransactionID(eventID)
	if err != nil {
		t.Fatalf("[%s] - Failed to get event transaction ID. Error: [%s]", eventID, err)
	}

	if actualTxID != expectedTxID {
		t.Fatalf("Expected Event TX ID [%s] did not match actual TX ID [%s]", expectedTxID, actualTxID)
	}
}

func verifyTransferFromValidatorAPI(setupEnv *setup.Setup, evm setup.EVMUtils, txId string, tokenID string, expectedSendAmount int64, targetAsset string, t *testing.T) *service.TransferData {
	transactionData, err := setupEnv.Clients.ValidatorClient.GetTransferData(txId)
	if err != nil {
		t.Fatalf("Cannot fetch transaction data - Error: [%s].", err)
	}
	if transactionData.Amount != fmt.Sprint(expectedSendAmount) {
		t.Fatalf("Transaction data mismatch: Expected [%d], but was [%s]", expectedSendAmount, transactionData.Amount)
	}
	if transactionData.NativeAsset != tokenID {
		t.Fatalf("Native Token mismatch: Expected [%s], but was [%s]", setupEnv.TokenID.String(), transactionData.NativeAsset)
	}
	if transactionData.Recipient != evm.Receiver.String() {
		t.Fatalf("Receiver address mismatch: Expected [%s], but was [%s]", evm.Receiver.String(), transactionData.Recipient)
	}
	if transactionData.TargetAsset != targetAsset {
		t.Fatalf("Token address mismatch: Expected [%s], but was [%s]", targetAsset, transactionData.TargetAsset)
	}

	return transactionData
}

func verifyFeeRecord(dbValidation *database.Service, expectedRecord *entity.Fee, t *testing.T) {
	ok, err := dbValidation.VerifyFeeRecord(expectedRecord)
	if err != nil {
		t.Fatalf("[%s] - Verification of database records failed - Error: [%s].", expectedRecord.TransactionID, err)
	}
	if !ok {
		t.Fatalf("[%s] - Database does not contain expected records", expectedRecord.TransactionID)
	}
}

func verifyTransferRecord(dbValidation *database.Service, expectedRecord *entity.Transfer, t *testing.T) {
	exist, err := dbValidation.VerifyTransferRecord(expectedRecord)
	if err != nil {
		t.Fatalf("[%s] - Verification of database records failed - Error: [%s].", expectedRecord.TransactionID, err)
	}
	if !exist {
		t.Fatalf("[%s] - Database does not contain expected records", expectedRecord.TransactionID)
	}
}

func verifyScheduleRecord(dbValidation *database.Service, expectedRecord *entity.Schedule, t *testing.T) {
	exist, err := dbValidation.VerifyScheduleRecord(expectedRecord)
	if err != nil {
		t.Fatalf("[%s] - Verification of database records failed - Error: [%s].", expectedRecord.TransactionID, err)
	}
	if !exist {
		t.Fatalf("[%s] - Database does not contain expected records", expectedRecord.TransactionID)
	}
}

func verifyTransferRecordAndSignatures(dbValidation *database.Service, expectedRecord *entity.Transfer, amount string, signatures []string, t *testing.T) {
	exist, err := dbValidation.VerifyTransferAndSignatureRecords(expectedRecord, amount, signatures)
	if err != nil {
		t.Fatalf("[%s] - Verification of database records failed - Error: [%s].", expectedRecord.TransactionID, err)
	}
	if !exist {
		t.Fatalf("[%s] - Database does not contain expected records", expectedRecord.TransactionID)
	}
}

func verifyTransferToBridgeAccount(s *setup.Setup, wrappedAsset string, evm setup.EVMUtils, memo string, whbarReceiverAddress common.Address, t *testing.T) (hedera.TransactionResponse, *big.Int) {
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
	transactionResponse, err := sendHbarsToBridgeAccount(s, memo)
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
	if amount != tinyBarAmount {
		t.Fatalf("Expected to receive the exact transfer amount of hbar: [%v], but was [%v]", tinyBarAmount, amount)
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
		t.Fatalf("Expected to receive the exact transfer amount of hbar: [%v], but received: [%v]", hBarSendAmount.AsTinybar(), tinyBarAmount)
	}

	return *transactionResponse, wrappedBalanceBefore
}

func verifyTopicMessages(setup *setup.Setup, txId string, t *testing.T) []string {
	ethSignaturesCollected := 0
	var receivedSignatures []string

	fmt.Println(fmt.Sprintf("Waiting for Signatures & TX Hash to be published to Topic [%v]", setup.TopicID.String()))

	// Subscribe to Topic
	_, err := hedera.NewTopicMessageQuery().
		SetStartTime(time.Unix(0, time.Now().UnixNano())).
		SetTopicID(setup.TopicID).
		Subscribe(
			setup.Clients.Hedera,
			func(response hedera.TopicMessage) {
				msg := &validatorproto.TopicEthSignatureMessage{}
				err := proto.Unmarshal(response.Contents, msg)
				if err != nil {
					t.Fatal(err)
				}

				//Verify that all the submitted messages have signed the same transaction
				if msg.TransferID != txId {
					fmt.Println(fmt.Sprintf(`Expected signature message to contain the transaction id: [%s]`, txId))
				} else {
					receivedSignatures = append(receivedSignatures, msg.Signature)
					ethSignaturesCollected++
					fmt.Println(fmt.Sprintf("Received Auth Signature [%s]", msg.Signature))
				}
			},
		)
	if err != nil {
		t.Fatalf("Unable to subscribe to Topic [%s]", setup.TopicID)
	}

	select {
	case <-time.After(90 * time.Second):
		if ethSignaturesCollected != expectedValidatorsCount {
			t.Fatalf("Expected the count of collected signatures to equal the number of validators: [%v], but was: [%v]", expectedValidatorsCount, ethSignaturesCollected)
		}
		return receivedSignatures
	}
	// Not possible end-case
	return nil
}

func sendHbarsToBridgeAccount(setup *setup.Setup, memo string) (*hedera.TransactionResponse, error) {
	fmt.Println(fmt.Sprintf("Sending [%v] Hbars through the Bridge. Transaction Memo: [%s]", hBarSendAmount, memo))

	res, err := hedera.NewTransferTransaction().
		AddHbarTransfer(setup.Clients.Hedera.GetOperatorAccountID(), hbarRemovalAmount).
		AddHbarTransfer(setup.BridgeAccount, hBarSendAmount).
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

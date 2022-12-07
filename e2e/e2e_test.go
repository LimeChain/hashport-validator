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
	"fmt"
	"math/big"
	"strconv"
	"testing"
	"time"

	evmSetup "github.com/limechain/hedera-eth-bridge-validator/e2e/setup/evm"

	"github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"

	"github.com/limechain/hedera-eth-bridge-validator/e2e/helper/fetch"

	"github.com/limechain/hedera-eth-bridge-validator/e2e/helper/compat"

	"github.com/limechain/hedera-eth-bridge-validator/e2e/helper/submit"

	"github.com/limechain/hedera-eth-bridge-validator/e2e/helper/expected"
	"github.com/limechain/hedera-eth-bridge-validator/e2e/helper/verify"

	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/transaction"
	hederahelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/hedera"
	auth_message "github.com/limechain/hedera-eth-bridge-validator/app/model/auth-message"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/schedule"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/limechain/hedera-eth-bridge-validator/e2e/setup"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashgraph/hedera-sdk-go/v2"
)

const (
	firstEvmChainId  = uint64(80001) // represents Polygon Mumbai Testnet (e2e config must have configuration for that particular network)
	secondEvmChainId = uint64(43113) // represents Avalanche Fuji Testnet (e2e config must have configuration for that particular network)
)

// Test_HBAR recreates a real life situation of a user who wants to bridge a Hedera HBARs to the EVM Network infrastructure. The wrapped token on the EVM network(corresponding to the native Hedera Hashgraph's HBARs) gets minted, then transferred to the recipient account on the EVM network.
func Test_HBAR(t *testing.T) {
	amount := int64(1000000000) // 10 HBAR
	setupEnv := setup.Load()
	now := time.Now()

	chainId := firstEvmChainId
	evm := setupEnv.Clients.EVM[chainId]
	receiver := evm.Receiver
	memo := fmt.Sprintf("%d-%s", chainId, evm.Receiver.String())

	targetAsset, err := evmSetup.NativeToWrappedAsset(setupEnv.AssetMappings, constants.HederaNetworkId, chainId, constants.Hbar)
	if err != nil {
		t.Fatalf("Expecting Token [%s] is not supported. - Error: [%s]", constants.Hbar, err)
	}

	mintAmount, fee := expected.ReceiverAndFeeAmounts(setupEnv.Clients.FeeCalculator, setupEnv.Clients.Distributor, constants.Hbar, amount)

	// Step 1 - Verify the transfer of Hbars to the Bridge Account
	transactionResponse, wrappedBalanceBefore := verify.TransferToBridgeAccount(t, setupEnv.Clients.Hedera, setupEnv.BridgeAccount, targetAsset, evm, memo, receiver, amount)

	// Step 2 - Verify the submitted topic messages
	receivedSignatures := verify.TopicMessages(t, setupEnv.Clients.Hedera, hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String(), setupEnv.TopicID)

	// Step 3 - Validate fee scheduled transaction
	expectedTransfers := expected.MirrorNodeExpectedTransfersForHederaTransfer(setupEnv.Members, setupEnv.BridgeAccount, constants.Hbar, fee)
	scheduledTxID, scheduleID := verify.MembersScheduledTxs(t, setupEnv.Clients.Hedera, setupEnv.Clients.MirrorNode, setupEnv.Members, constants.Hbar, expectedTransfers, now)

	// Step 4 - Verify Transfer retrieved from Validator API
	transactionData := verify.FungibleTransferFromValidatorAPI(t, setupEnv.Clients.ValidatorClient, setupEnv.TokenID, evm, hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String(), constants.Hbar, fmt.Sprint(mintAmount), targetAsset)

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
	txHash := submit.MintTransaction(t, evm, hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String(), transactionData, common.HexToAddress(targetAsset))

	// Step 6 - Wait for transaction to be mined
	submit.WaitForTransaction(t, evm, txHash)

	// Step 7 - Validate Token balances
	verify.WrappedAssetBalance(t, evm, targetAsset, big.NewInt(mintAmount), wrappedBalanceBefore, receiver)

	// Step 8 - Prepare Comparable Expected Transfer Record
	expectedTxRecord := expected.FungibleTransferRecord(
		constants.HederaNetworkId,
		chainId,
		constants.HederaNetworkId,
		hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String(),
		constants.Hbar,
		targetAsset,
		constants.Hbar,
		strconv.FormatInt(amount, 10),
		strconv.FormatInt(fee, 10),
		receiver.String(),
		status.Completed,
		setupEnv.Clients.Hedera.GetOperatorAccountID().String(),
		entity.NanoTime{Time: ts})
	// and:
	expectedFeeRecord := expected.FeeRecord(
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
	verify.TransferRecordAndSignatures(t, setupEnv.DbValidator, expectedTxRecord, authMsgBytes, receivedSignatures)
	// and:
	verify.FeeRecord(t, setupEnv.DbValidator, expectedFeeRecord)
}

// Test_E2E_Token_Transfer recreates a real life situation of a user who wants to bridge a Hedera native token to the EVM Network infrastructure. The wrapped token on the EVM network(corresponding to the native Hedera Hashgraph's one) gets minted, then transferred to the recipient account on the EVM network.
func Test_E2E_Token_Transfer(t *testing.T) {
	amount := int64(1000000000) // 10 HBAR
	setupEnv := setup.Load()
	now := time.Now()

	chainId := firstEvmChainId
	evm := setupEnv.Clients.EVM[chainId]
	memo := fmt.Sprintf("%d-%s", chainId, evm.Receiver.String())
	mintAmount, fee := expected.ReceiverAndFeeAmounts(setupEnv.Clients.FeeCalculator, setupEnv.Clients.Distributor, setupEnv.TokenID.String(), amount)

	targetAsset, err := evmSetup.NativeToWrappedAsset(setupEnv.AssetMappings, constants.HederaNetworkId, chainId, setupEnv.TokenID.String())
	if err != nil {
		t.Fatalf("Expecting Token [%s] is not supported. - Error: [%s]", constants.Hbar, err)
	}

	// Step 1 - Verify the transfer of HTS to the Bridge Account
	transactionResponse, wrappedBalanceBefore := verify.TokenTransferToBridgeAccount(t, setupEnv.Clients.Hedera, setupEnv.BridgeAccount, targetAsset, setupEnv.TokenID, evm, memo, evm.Receiver, amount)

	// Step 2 - Verify the submitted topic messages
	receivedSignatures := verify.TopicMessages(
		t, setupEnv.Clients.Hedera, hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String(), setupEnv.TopicID)

	// Step 3 - Validate fee scheduled transaction
	expectedTransfers := expected.MirrorNodeExpectedTransfersForHederaTransfer(setupEnv.Members, setupEnv.BridgeAccount, setupEnv.TokenID.String(), fee)
	scheduledTxID, scheduleID := verify.MembersScheduledTxs(t, setupEnv.Clients.Hedera, setupEnv.Clients.MirrorNode, setupEnv.Members, setupEnv.TokenID.String(), expectedTransfers, now)

	// Step 4 - Verify Transfer retrieved from Validator API
	transactionData := verify.FungibleTransferFromValidatorAPI(
		t, setupEnv.Clients.ValidatorClient, setupEnv.TokenID, evm, hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String(), setupEnv.TokenID.String(), fmt.Sprint(mintAmount), targetAsset)

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
	txHash := submit.MintTransaction(t, evm,
		hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String(), transactionData, common.HexToAddress(targetAsset))

	// Step 6 - Wait for transaction to be mined
	submit.WaitForTransaction(t, evm, txHash)

	// Step 7 - Validate Token balances
	verify.WrappedAssetBalance(t, evm, targetAsset, big.NewInt(mintAmount), wrappedBalanceBefore, evm.Receiver)

	// Step 8 - Verify Database records
	expectedTxRecord := expected.FungibleTransferRecord(
		constants.HederaNetworkId,
		chainId,
		constants.HederaNetworkId,
		hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String(),
		setupEnv.TokenID.String(),
		targetAsset,
		setupEnv.TokenID.String(),
		strconv.FormatInt(amount, 10),
		strconv.FormatInt(fee, 10),
		evm.Receiver.String(),
		status.Completed,
		setupEnv.Clients.Hedera.GetOperatorAccountID().String(),
		entity.NanoTime{Time: ts})
	// and:
	expectedFeeRecord := expected.FeeRecord(
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
	verify.TransferRecordAndSignatures(t, setupEnv.DbValidator, expectedTxRecord, authMsgBytes, receivedSignatures)
	// and:
	verify.FeeRecord(t, setupEnv.DbValidator, expectedFeeRecord)
}

// Test_EVM_Hedera_HBAR recreates a real life situation of a user who wants to return a Hedera native HBARs from the EVM Network infrastructure. The wrapped HBARs on the EVM network(corresponding to the native Hedera Hashgraph's one) gets burned, then the locked HBARs on the Hedera bridge account get unlocked, forwarding them to the recipient account.
func Test_EVM_Hedera_HBAR(t *testing.T) {
	amount := int64(100000000) // 1 HBAR
	setupEnv := setup.Load()

	chainId := firstEvmChainId
	evm := setupEnv.Clients.EVM[chainId]
	now := time.Now()
	accountBalanceBefore := fetch.HederaAccountBalance(t, setupEnv.Clients.Hedera, setupEnv.Clients.Hedera.GetOperatorAccountID())

	targetAsset, err := evmSetup.NativeToWrappedAsset(setupEnv.AssetMappings, constants.HederaNetworkId, chainId, constants.Hbar)
	if err != nil {
		t.Fatalf("Expecting Token [%s] is not supported. - Error: [%s]", constants.Hbar, err)
	}

	// 1. Calculate Expected Receive And Fee Amounts
	expectedReceiveAmount, fee := expected.ReceiverAndFeeAmounts(setupEnv.Clients.FeeCalculator, setupEnv.Clients.Distributor, constants.Hbar, amount)

	// 2. Submit burn transaction to the bridge contract
	burnTxReceipt, expectedRouterBurn := submit.BurnEthTransaction(
		t, setupEnv.AssetMappings, evm, constants.Hbar, constants.HederaNetworkId, chainId, setupEnv.Clients.Hedera.GetOperatorAccountID().ToBytes(), amount)

	// 2.1 Get the block timestamp of burn event
	block, err := evm.EVMClient.BlockByNumber(context.Background(), burnTxReceipt.BlockNumber)
	if err != nil {
		t.Fatal("failed to get block by number", err)
	}
	blockTimestamp := time.Unix(int64(block.Time()), 0).UTC()

	// 3. Validate that the burn transaction went through and emitted the correct events
	expectedId := verify.BurnEvent(t, burnTxReceipt, expectedRouterBurn)

	// 4. Validate that a scheduled transaction was submitted
	expectedTransfers := expected.MirrorNodeExpectedTransfersForBurnEvent(setupEnv.Members, setupEnv.Clients.Hedera, setupEnv.BridgeAccount, constants.Hbar, expectedReceiveAmount, fee)
	transactionID, scheduleID := verify.SubmittedScheduledTx(t, setupEnv.Clients.Hedera, setupEnv.Clients.MirrorNode, setupEnv.Members, constants.Hbar, expectedTransfers, now)

	// 5. Validate Event Transaction ID retrieved from Validator API
	verify.EventTransactionIDFromValidatorAPI(t, setupEnv.Clients.ValidatorClient, expectedId, transactionID)

	// 6. Validate that the balance of the receiver account (hedera) was changed with the correct amount
	verify.ReceiverAccountBalance(t, setupEnv.Clients.Hedera, uint64(expectedReceiveAmount), accountBalanceBefore, constants.Hbar, setupEnv.TokenID)

	// 7. Prepare Expected Database Records
	expectedBurnEventRecord := expected.FungibleTransferRecord(
		chainId,
		constants.HederaNetworkId,
		constants.HederaNetworkId,
		expectedId,
		targetAsset,
		constants.Hbar,
		constants.Hbar,
		strconv.FormatInt(amount, 10),
		strconv.FormatInt(fee, 10),
		setupEnv.Clients.Hedera.GetOperatorAccountID().String(),
		status.Completed,
		evm.Signer.Address(),
		entity.NanoTime{Time: blockTimestamp})
	// and:
	expectedFeeRecord := expected.FeeRecord(transactionID, scheduleID, fee, expectedId)

	// 8. Wait for validators to update DB state after Scheduled TX is mined
	time.Sleep(20 * time.Second)

	// 9. Validate Database Records
	verify.TransferRecord(t, setupEnv.DbValidator, expectedBurnEventRecord)
	// and:
	verify.FeeRecord(t, setupEnv.DbValidator, expectedFeeRecord)
}

// Test_EVM_Hedera_Token recreates a real life situation of a user who wants to return a Hedera native token from the EVM Network infrastructure. The wrapped token on the EVM network(corresponding to the native Hedera one) gets burned, then the amount gets unlocked on the Hedera bridge account, forwarding it to the recipient account.
func Test_EVM_Hedera_Token(t *testing.T) {
	amount := int64(100000000) // 1 HBAR
	setupEnv := setup.Load()

	chainId := firstEvmChainId
	evm := setupEnv.Clients.EVM[chainId]
	now := time.Now()
	accountBalanceBefore := fetch.HederaAccountBalance(t, setupEnv.Clients.Hedera, setupEnv.Clients.Hedera.GetOperatorAccountID())

	targetAsset, err := evmSetup.NativeToWrappedAsset(setupEnv.AssetMappings, constants.HederaNetworkId, chainId, setupEnv.TokenID.String())
	if err != nil {
		t.Fatalf("Expecting Token [%s] is not supported. - Error: [%s]", setupEnv.TokenID.String(), err)
	}

	// 1. Calculate Expected Receive Amount
	expectedReceiveAmount, fee := expected.ReceiverAndFeeAmounts(setupEnv.Clients.FeeCalculator, setupEnv.Clients.Distributor, setupEnv.TokenID.String(), amount)

	// 2. Submit burn transaction to the bridge contract
	burnTxReceipt, expectedRouterBurn := submit.BurnEthTransaction(t, setupEnv.AssetMappings, evm, setupEnv.TokenID.String(), constants.HederaNetworkId, chainId, setupEnv.Clients.Hedera.GetOperatorAccountID().ToBytes(), amount)

	// 2.1 Get the block timestamp of burn event
	block, err := evm.EVMClient.BlockByNumber(context.Background(), burnTxReceipt.BlockNumber)
	if err != nil {
		t.Fatal("failed to get block by number", err)
	}
	blockTimestamp := time.Unix(int64(block.Time()), 0).UTC()

	// 3. Validate that the burn transaction went through and emitted the correct events
	expectedId := verify.BurnEvent(t, burnTxReceipt, expectedRouterBurn)

	// 4. Validate that a scheduled transaction was submitted
	expectedTransfers := expected.MirrorNodeExpectedTransfersForBurnEvent(
		setupEnv.Members, setupEnv.Clients.Hedera, setupEnv.BridgeAccount, setupEnv.TokenID.String(), expectedReceiveAmount, fee)
	transactionID, scheduleID := verify.SubmittedScheduledTx(
		t, setupEnv.Clients.Hedera, setupEnv.Clients.MirrorNode, setupEnv.Members, setupEnv.TokenID.String(), expectedTransfers, now)

	// 5. Validate Event Transaction ID retrieved from Validator API
	verify.EventTransactionIDFromValidatorAPI(t, setupEnv.Clients.ValidatorClient, expectedId, transactionID)

	// 6. Validate that the balance of the receiver account (hedera) was changed with the correct amount
	verify.ReceiverAccountBalance(t, setupEnv.Clients.Hedera, uint64(expectedReceiveAmount), accountBalanceBefore, setupEnv.TokenID.String(), setupEnv.TokenID)

	// 7. Prepare Expected Database Records
	expectedBurnEventRecord := expected.FungibleTransferRecord(
		chainId,
		constants.HederaNetworkId,
		constants.HederaNetworkId,
		expectedId,
		targetAsset,
		setupEnv.TokenID.String(),
		setupEnv.TokenID.String(),
		strconv.FormatInt(amount, 10),
		strconv.FormatInt(fee, 10),
		setupEnv.Clients.Hedera.GetOperatorAccountID().String(),
		status.Completed,
		evm.Signer.Address(),
		entity.NanoTime{Time: blockTimestamp})
	// and:
	expectedFeeRecord := expected.FeeRecord(transactionID, scheduleID, fee, expectedId)

	// 8. Wait for validators to update DB state after Scheduled TX is mined
	time.Sleep(20 * time.Second)

	// 9. Validate Database Records
	verify.TransferRecord(t, setupEnv.DbValidator, expectedBurnEventRecord)
	// and:
	verify.FeeRecord(t, setupEnv.DbValidator, expectedFeeRecord)
}

// Test_EVM_Hedera_Native_Token recreates a real life situation of a user who wants to bridge an EVM native token to the Hedera infrastructure. A new wrapped token (corresponding to the native EVM one) gets minted to the bridge account, then gets transferred to the recipient account.
func Test_EVM_Hedera_Native_Token(t *testing.T) {
	amount := int64(1000000000000) // 1 000 gwei
	// Step 1: Initialize setup, smart contracts, etc.
	setupEnv := setup.Load()

	chainId := firstEvmChainId
	evm := setupEnv.Clients.EVM[chainId]
	now := time.Now()
	bridgeAccountBalanceBefore := fetch.HederaAccountBalance(t, setupEnv.Clients.Hedera, setupEnv.BridgeAccount)
	receiverAccountBalanceBefore := fetch.HederaAccountBalance(t, setupEnv.Clients.Hedera, setupEnv.Clients.Hedera.GetOperatorAccountID())
	targetAsset, err := evmSetup.NativeToWrappedAsset(setupEnv.AssetMappings, chainId, constants.HederaNetworkId, setupEnv.NativeEvmToken)
	if err != nil {
		t.Fatal(err)
	}
	feeData, err := evm.RouterContract.TokenFeeData(nil, common.HexToAddress(setupEnv.NativeEvmToken))
	if err != nil {
		t.Fatalf("failed to get token fee data: %v", err)
	}
	fee := amount * (feeData.ServiceFeePercentage.Int64() / constants.FeeMaxPercentage)

	// Step 2: Submit Lock Txn from a deployed smart contract
	receipt, expectedLockEventLog := submit.LockEthTransaction(t, evm, setupEnv.NativeEvmToken, constants.HederaNetworkId, setupEnv.Clients.Hedera.GetOperatorAccountID().ToBytes(), amount)

	// Step 2.1 - Get the block timestamp of lock event
	block, err := evm.EVMClient.BlockByNumber(context.Background(), receipt.BlockNumber)
	if err != nil {
		t.Fatal("failed to get block by number", err)
	}
	blockTimestamp := time.Unix(int64(block.Time()), 0).UTC()

	// Step 3: Validate Lock Event was emitted with correct data
	lockEventId := verify.LockEvent(t, receipt, expectedLockEventLog)

	bridgedAmount := new(big.Int).Sub(expectedLockEventLog.Amount, expectedLockEventLog.ServiceFee)
	expectedAmount, err := compat.RemoveDecimals(bridgedAmount.Int64(), common.HexToAddress(setupEnv.NativeEvmToken), evm)
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
	bridgeMintTransactionID, bridgeMintScheduleID := verify.ScheduledMintTx(
		t, setupEnv.Clients.Hedera, setupEnv.Clients.MirrorNode, setupEnv.BridgeAccount, setupEnv.TokenID.String(), mintTransfer, now)

	// Wait for validators to update DB state after Scheduled TX is mined
	time.Sleep(20 * time.Second)

	// Step 5: Validate that Database statuses were changed correctly
	expectedLockEventRecord := expected.FungibleTransferRecord(
		chainId,
		constants.HederaNetworkId,
		chainId,
		lockEventId,
		setupEnv.NativeEvmToken,
		targetAsset,
		setupEnv.NativeEvmToken,
		strconv.FormatInt(expectedAmount, 10),
		strconv.FormatInt(fee, 10),
		setupEnv.Clients.Hedera.GetOperatorAccountID().String(),
		status.Completed,
		evm.Signer.Address(),
		entity.NanoTime{Time: blockTimestamp})

	expectedScheduleMintRecord := expected.ScheduleRecord(
		bridgeMintTransactionID,
		bridgeMintScheduleID,
		schedule.MINT,
		false,
		status.Completed,
		sql.NullString{
			String: bridgeMintTransactionID,
			Valid:  true,
		})
	// Step 6: Verify that records have been created successfully
	verify.TransferRecord(t, setupEnv.DbValidator, expectedLockEventRecord)
	verify.ScheduleRecord(t, setupEnv.DbValidator, expectedScheduleMintRecord)

	// Step 7: Validate that a scheduled transfer txn was submitted successfully
	bridgeTransferTransactionID, bridgeTransferScheduleID := verify.ScheduledTx(
		t,
		setupEnv.Clients.Hedera,
		setupEnv.Clients.MirrorNode,
		setupEnv.Clients.Hedera.GetOperatorAccountID(),
		setupEnv.TokenID.String(),
		expected.MirrorNodeExpectedTransfersForLockEvent(setupEnv.Clients.Hedera, setupEnv.BridgeAccount, targetAsset, expectedAmount),
		now)

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

	verify.ScheduleRecord(t, setupEnv.DbValidator, expectedScheduleTransferRecord)
	// Step 9: Validate Treasury(BridgeAccount) Balance and Receiver Balance
	verify.AccountBalance(t, setupEnv.Clients.Hedera, setupEnv.BridgeAccount, 0, bridgeAccountBalanceBefore, targetAsset)
	verify.AccountBalance(t, setupEnv.Clients.Hedera, setupEnv.Clients.Hedera.GetOperatorAccountID(), uint64(expectedAmount), receiverAccountBalanceBefore, targetAsset)
}

// Test_E2E_Hedera_EVM_Native_Token recreates a real life situation of a user who wants to bridge a Hedera wrapped token to the EVM Native Network infrastructure. The wrapped token on the EVM network(corresponding to the native Hedera Hashgraph's one) gets minted, then transferred to the recipient account on the EVM network.
func Test_E2E_Hedera_EVM_Native_Token(t *testing.T) {
	setupEnv := setup.Load()
	now := time.Now()

	chainId := firstEvmChainId
	evm := setupEnv.Clients.EVM[chainId]
	memo := fmt.Sprintf("%d-%s", chainId, evm.Receiver.String())
	unlockAmount := int64(10) // Amount, which converted to 18 decimals is 100000000000 (100 gwei)

	// Step 1 - Verify the transfer of HTS to the Bridge Account
	wrappedAsset, err := evmSetup.NativeToWrappedAsset(setupEnv.AssetMappings, chainId, constants.HederaNetworkId, setupEnv.NativeEvmToken)
	if err != nil {
		t.Fatal(err)
	}

	tokenID, err := hedera.TokenIDFromString(wrappedAsset)
	if err != nil {
		t.Fatal(err)
	}

	expectedSubmitUnlockAmount, err := compat.AddDecimals(unlockAmount, common.HexToAddress(setupEnv.NativeEvmToken), evm)
	if err != nil {
		t.Fatal(err)
	}

	transactionResponse, nativeBalanceBefore := verify.TokenTransferToBridgeAccount(
		t, setupEnv.Clients.Hedera, setupEnv.BridgeAccount, setupEnv.NativeEvmToken, tokenID, evm, memo, evm.Receiver, unlockAmount)
	burnTransfer := []transaction.Transfer{
		{
			Account: setupEnv.BridgeAccount.String(),
			Amount:  -unlockAmount,
			Token:   wrappedAsset,
		},
	}

	// Step 2 - Verify the submitted topic messages
	receivedSignatures := verify.TopicMessages(
		t, setupEnv.Clients.Hedera, hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String(), setupEnv.TopicID)

	// Step 3 - Validate burn scheduled transaction
	burnTransactionID, burnScheduleID := verify.ScheduledBurnTx(t, setupEnv.Clients.Hedera, setupEnv.Clients.MirrorNode, setupEnv.BridgeAccount, setupEnv.TokenID.String(), burnTransfer, now)

	// Step 4 - Verify Transfer retrieved from Validator API
	transactionData := verify.FungibleTransferFromValidatorAPI(
		t, setupEnv.Clients.ValidatorClient, setupEnv.TokenID, evm,
		hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String(), setupEnv.NativeEvmToken, fmt.Sprint(expectedSubmitUnlockAmount), setupEnv.NativeEvmToken)

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
	txHash := submit.UnlockTransaction(
		t, evm, hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String(), transactionData, common.HexToAddress(setupEnv.NativeEvmToken))

	// Step 6 - Wait for transaction to be mined
	submit.WaitForTransaction(t, evm, txHash)

	expectedUnlockedAmount := expected.UnlockAmount(t, evm, setupEnv.NativeEvmToken, expectedSubmitUnlockAmount)

	//Step 7 - Validate Token balances
	verify.WrappedAssetBalance(t, evm, setupEnv.NativeEvmToken, expectedUnlockedAmount, nativeBalanceBefore, evm.Receiver)

	// Step 8 - Verify Database records
	expectedTxRecord := expected.FungibleTransferRecord(
		constants.HederaNetworkId,
		chainId,
		chainId,
		hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String(),
		wrappedAsset,
		setupEnv.NativeEvmToken,
		setupEnv.NativeEvmToken,
		strconv.FormatInt(expectedSubmitUnlockAmount, 10),
		"",
		evm.Receiver.String(),
		status.Completed,
		setupEnv.Clients.Hedera.GetOperatorAccountID().String(),
		entity.NanoTime{Time: ts})

	// Step 8: Validate that database statuses were updated correctly for the Schedule Burn
	expectedScheduleBurnRecord := expected.ScheduleRecord(
		burnTransactionID,
		burnScheduleID,
		schedule.BURN,
		false,
		status.Completed,
		sql.NullString{
			String: burnTransactionID,
			Valid:  true,
		})
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
	verify.TransferRecordAndSignatures(t, setupEnv.DbValidator, expectedTxRecord, authMsgBytes, receivedSignatures)
	// and
	verify.ScheduleRecord(t, setupEnv.DbValidator, expectedScheduleBurnRecord)
}

// Test_EVM_Native_to_EVM_Token recreates a real life situation of a user who wants to bridge an EVM native token to another EVM chain.
func Test_EVM_Native_to_EVM_Token(t *testing.T) {
	amount := int64(1000000000000) // 1000 gwei
	// Step 1 - Initialize setup, smart contracts, etc.
	setupEnv := setup.Load()

	chainId := firstEvmChainId
	evm := setupEnv.Clients.EVM[chainId]
	targetChainID := secondEvmChainId
	wrappedAsset, err := evmSetup.NativeToWrappedAsset(setupEnv.AssetMappings, chainId, targetChainID, setupEnv.NativeEvmToken)
	if err != nil {
		t.Fatal(err)
	}
	feeData, err := evm.RouterContract.TokenFeeData(nil, common.HexToAddress(setupEnv.NativeEvmToken))
	if err != nil {
		t.Fatalf("failed to get token fee data: %v", err)
	}
	fee := amount * (feeData.ServiceFeePercentage.Int64() / constants.FeeMaxPercentage)

	wrappedEvm := setupEnv.Clients.EVM[targetChainID]
	wrappedInstance, err := evmSetup.InitAssetContract(wrappedAsset, wrappedEvm.EVMClient)
	if err != nil {
		t.Fatal(err)
	}

	wrappedBalanceBefore, err := wrappedInstance.BalanceOf(&bind.CallOpts{}, evm.Receiver)
	if err != nil {
		t.Fatal(err)
	}

	// Step 2 - Submit Lock Txn from a deployed smart contract
	receipt, expectedLockEventLog := submit.LockEthTransaction(t, evm, setupEnv.NativeEvmToken, targetChainID, evm.Receiver.Bytes(), amount)

	expectedAmount := new(big.Int).Sub(expectedLockEventLog.Amount, expectedLockEventLog.ServiceFee)

	// Step 2.1 - Get the block timestamp of the lock event
	block, err := evm.EVMClient.BlockByNumber(context.Background(), receipt.BlockNumber)
	if err != nil {
		t.Fatal("failed to get block by number", err)
	}
	blockTimestamp := time.Unix(int64(block.Time()), 0).UTC()

	// Step 3 - Validate Lock Event was emitted with correct data
	lockEventId := verify.LockEvent(t, receipt, expectedLockEventLog)

	// Step 4 - Verify the submitted topic messages
	receivedSignatures := verify.TopicMessages(t, setupEnv.Clients.Hedera, lockEventId, setupEnv.TopicID)

	// Step 5 - Verify Transfer retrieved from Validator API
	transactionData := verify.FungibleTransferFromValidatorAPI(
		t, setupEnv.Clients.ValidatorClient, setupEnv.TokenID, evm, lockEventId, setupEnv.NativeEvmToken, expectedAmount.String(), wrappedAsset)

	// Step 6 - Submit Mint transaction
	txHash := submit.MintTransaction(t, wrappedEvm, lockEventId, transactionData, common.HexToAddress(wrappedAsset))

	// Step 7 - Wait for transaction to be mined
	submit.WaitForTransaction(t, wrappedEvm, txHash)

	// Step 8 - Validate Token balances
	verify.WrappedAssetBalance(t, wrappedEvm, wrappedAsset, expectedAmount, wrappedBalanceBefore, evm.Receiver)

	// Step 9 - Prepare expected Transfer record
	expectedLockEventRecord := expected.FungibleTransferRecord(
		chainId,
		targetChainID,
		chainId,
		lockEventId,
		setupEnv.NativeEvmToken,
		wrappedAsset,
		setupEnv.NativeEvmToken,
		expectedAmount.String(),
		strconv.FormatInt(fee, 10),
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
	verify.TransferRecordAndSignatures(t, setupEnv.DbValidator, expectedLockEventRecord, authMsgBytes, receivedSignatures)
}

// Test_EVM_Wrapped_to_EVM_Token recreates a real life situation of a user who wants to bridge an EVM native token to another EVM chain.
func Test_EVM_Wrapped_to_EVM_Token(t *testing.T) {
	amount := int64(100000000000) // 100 gwei
	// Step 1 - Initialize setup, smart contracts, etc.
	setupEnv := setup.Load()

	chainId := firstEvmChainId
	sourceChain := secondEvmChainId
	wrappedEvm := setupEnv.Clients.EVM[sourceChain]
	sourceAsset, err := evmSetup.NativeToWrappedAsset(setupEnv.AssetMappings, chainId, sourceChain, setupEnv.NativeEvmToken)
	if err != nil {
		t.Fatal(err)
	}

	nativeEvm := setupEnv.Clients.EVM[chainId]
	nativeInstance, err := evmSetup.InitAssetContract(setupEnv.NativeEvmToken, nativeEvm.EVMClient)
	if err != nil {
		t.Fatal(err)
	}

	nativeBalanceBefore, err := nativeInstance.BalanceOf(&bind.CallOpts{}, nativeEvm.Receiver)
	if err != nil {
		t.Fatal(err)
	}

	// Step 2 - Submit Lock Txn from a deployed smart contract
	receipt, expectedLockEventLog := submit.BurnEthTransaction(
		t, setupEnv.AssetMappings, wrappedEvm, setupEnv.NativeEvmToken, chainId, sourceChain, nativeEvm.Receiver.Bytes(), amount)

	// Step 2.1 - Get the block timestamp of the burn event
	block, err := wrappedEvm.EVMClient.BlockByNumber(context.Background(), receipt.BlockNumber)
	if err != nil {
		t.Fatal("failed to get block by number", err)
	}
	blockTimestamp := time.Unix(int64(block.Time()), 0).UTC()

	// Step 3 - Validate Burn Event was emitted with correct data
	burnEventId := verify.BurnEvent(t, receipt, expectedLockEventLog)

	// Step 4 - Verify the submitted topic messages
	receivedSignatures := verify.TopicMessages(t, setupEnv.Clients.Hedera, burnEventId, setupEnv.TopicID)

	// Step 5 - Verify Transfer retrieved from Validator API
	transactionData := verify.FungibleTransferFromValidatorAPI(
		t, setupEnv.Clients.ValidatorClient, setupEnv.TokenID, nativeEvm, burnEventId, setupEnv.NativeEvmToken, fmt.Sprint(amount), setupEnv.NativeEvmToken)

	// Step 6 - Submit Mint transaction
	txHash := submit.UnlockTransaction(t, nativeEvm, burnEventId, transactionData, common.HexToAddress(setupEnv.NativeEvmToken))

	// Step 7 - Wait for transaction to be mined
	submit.WaitForTransaction(t, nativeEvm, txHash)

	expectedUnlockedAmount := expected.UnlockAmount(t, nativeEvm, setupEnv.NativeEvmToken, amount)

	// Step 8 - Validate Token balances
	verify.WrappedAssetBalance(t, nativeEvm, setupEnv.NativeEvmToken, expectedUnlockedAmount, nativeBalanceBefore, nativeEvm.Receiver)

	// Step 9 - Prepare expected Transfer record
	expectedLockEventRecord := expected.FungibleTransferRecord(
		sourceChain,
		chainId,
		chainId,
		burnEventId,
		sourceAsset,
		setupEnv.NativeEvmToken,
		setupEnv.NativeEvmToken,
		strconv.FormatInt(amount, 10),
		"",
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
	verify.TransferRecordAndSignatures(t, setupEnv.DbValidator, expectedLockEventRecord, authMsgBytes, receivedSignatures)
}

// Test_Hedera_Native_EVM_NFT_Transfer recreates User who wants to portal a Hedera Native NFT to an EVM chain.
func Test_Hedera_Native_EVM_NFT_Transfer(t *testing.T) {
	now := time.Now()
	setupEnv := setup.Load()
	nftToken := setupEnv.NftTokenID.String()
	serialNumber := setupEnv.NftSerialNumber

	chainId := firstEvmChainId
	evm := setupEnv.Clients.EVM[chainId]
	receiver := evm.Receiver
	memo := fmt.Sprintf("%d-%s", chainId, evm.Receiver.String())

	targetAsset, err := evmSetup.NativeToWrappedAsset(setupEnv.AssetMappings, constants.HederaNetworkId, chainId, nftToken)
	if err != nil {
		t.Fatalf("Expecting Token [%s] is not supported. - Error: [%s]", setupEnv.NftTokenID.String(), err)
	}

	transferFee, ok := setupEnv.NftConstantFees[nftToken]
	if !ok {
		fee, ok1 := setupEnv.NftDynamicFees[nftToken]
		if !ok1 {
			t.Fatalf("could not find fee for NFT: %v", nftToken)
		}
		transferFee = fee.IntPart()
	}
	validatorsFee := setupEnv.Clients.Distributor.ValidAmount(transferFee)

	// Step 1 - Get Token Metadata
	nftData, err := setupEnv.Clients.MirrorNode.GetNft(nftToken, serialNumber)
	if err != nil {
		t.Fatalf("Failed to get mirror node nft. Error [%s]", err)
	}
	decodedMetadata, e := base64.StdEncoding.DecodeString(nftData.Metadata)
	if e != nil {
		t.Fatalf("Failed to decode metadata [%s]. Error [%s]", nftData.Metadata, e)
	}

	// Step 2 - Send the NFT transfer, including the fee to the Bridge Account
	transactionResponse, err := submit.NFTWithFeeToBridgeAccount(setupEnv.Clients.Hedera, setupEnv.BridgeAccount, memo, nftToken, serialNumber, transferFee)
	if err != nil {
		t.Fatalf("Failed to send NFT transfer. Error: [%s]", err)
	}

	transactionID := hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String()

	// Step 3 - Verify the submitted topic messages
	receivedSignatures := verify.TopicMessages(
		t, setupEnv.Clients.Hedera, hederahelper.FromHederaTransactionID(transactionResponse.TransactionID).String(), setupEnv.TopicID)

	// Step 4 - Validate members fee scheduled transaction
	scheduledTxID, scheduleID := verify.MembersScheduledTxs(t, setupEnv.Clients.Hedera, setupEnv.Clients.MirrorNode, setupEnv.Members, constants.Hbar,
		expected.MirrorNodeExpectedTransfersForHederaTransfer(setupEnv.Members, setupEnv.BridgeAccount, constants.Hbar, validatorsFee), now)

	// Step 5 - Verify Non-Fungible Transfer retrieved from Validator API
	transactionData := verify.NonFungibleTransferFromValidatorAPI(t, setupEnv.Clients.ValidatorClient, setupEnv.TokenID, evm, transactionID, nftToken, string(decodedMetadata), serialNumber, targetAsset)

	// Step 5.1 - Get the consensus timestamp of the transfer
	tx, err := setupEnv.Clients.MirrorNode.GetSuccessfulTransaction(transactionID)
	if err != nil {
		t.Fatal("failed to get successful transaction", err)
	}
	nanos, err := timestamp.FromString(tx.ConsensusTimestamp)
	if err != nil {
		t.Fatal("failed to parse consensus timestamp", err)
	}
	ts := timestamp.FromNanos(nanos)

	// Step 5 - Submit Mint ERC-721 transaction
	txHash := submit.MintERC721Transaction(t, evm, transactionID, transactionData)

	// Step 6 - Wait for transaction to be mined
	submit.WaitForTransaction(t, evm, txHash)

	// Step 7 - Validate EVM TokenId
	verify.ERC721TokenId(t, evm.EVMClient, targetAsset, serialNumber, receiver.String(), string(decodedMetadata))

	// Step 8 - Verify Database records
	expectedTxRecord := expected.NonFungibleTransferRecord(
		constants.HederaNetworkId,
		chainId,
		constants.HederaNetworkId,
		transactionID,
		nftToken,
		targetAsset,
		nftToken,
		receiver.String(),
		status.Completed,
		strconv.FormatInt(validatorsFee, 10),
		serialNumber,
		string(decodedMetadata),
		setupEnv.Clients.Hedera.GetOperatorAccountID().String(),
		entity.NanoTime{Time: ts},
	)
	// and:
	expectedFeeRecord := expected.FeeRecord(
		scheduledTxID,
		scheduleID, validatorsFee,
		transactionID)
	// and:
	expectedScheduleTransferRecord := expected.ScheduleRecord(
		scheduledTxID,
		scheduleID,
		schedule.TRANSFER,
		false,
		status.Completed,
		sql.NullString{
			String: transactionID,
			Valid:  true,
		})
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

	// Step 9 - Verify Database Records
	verify.TransferRecordAndSignatures(t, setupEnv.DbValidator, expectedTxRecord, authMsgBytes, receivedSignatures)
	// and:
	verify.FeeRecord(t, setupEnv.DbValidator, expectedFeeRecord)
	// and:
	verify.ScheduleRecord(t, setupEnv.DbValidator, expectedScheduleTransferRecord)
}

// Test_Hedera_EVM_BurnERC721_Transfer recreates User who wants to portal back a Hedera Native NFT from an EVM chain.
func Test_Hedera_EVM_BurnERC721_Transfer(t *testing.T) {
	setupEnv := setup.Load()
	nftToken := setupEnv.NftTokenID.String()
	serialNumber := setupEnv.NftSerialNumber

	chainId := firstEvmChainId
	evm := setupEnv.Clients.EVM[chainId]

	wrappedAsset, err := evmSetup.NativeToWrappedAsset(setupEnv.AssetMappings, constants.HederaNetworkId, chainId, nftToken)
	if err != nil {
		t.Fatalf("Expecting Token [%s] is not supported. - Error: [%s]", nftToken, err)
	}

	// 1. Validate that NFT owner is the bridge account
	verify.NftOwner(t, setupEnv.Clients.Hedera, nftToken, serialNumber, setupEnv.BridgeAccount)

	// 2. Submit burnERC721 transaction to the bridge contract
	burnTxReceipt, expectedRouterBurnERC721 := submit.BurnERC721Transaction(t, evm, wrappedAsset, constants.HederaNetworkId, setupEnv.Clients.Hedera.GetOperatorAccountID().ToBytes(), serialNumber)

	// 2.1 - Get the block timestamp of the burn event
	block, err := evm.EVMClient.BlockByNumber(context.Background(), burnTxReceipt.BlockNumber)
	if err != nil {
		t.Fatal("failed to get block by number", err)
	}
	blockTimestamp := time.Unix(int64(block.Time()), 0).UTC()

	// 3. Validate that the burn ERC-721 transaction went through and emitted the correct event
	expectedTxId := verify.BurnERC721Event(t, burnTxReceipt, expectedRouterBurnERC721)

	// 4. Validate that a scheduled NFT transaction was submitted
	scheduledTxID, scheduleID := verify.ScheduledNftTransfer(
		t, setupEnv.Clients.Hedera, setupEnv.Clients.MirrorNode, setupEnv.BridgeAccount, expectedTxId, nftToken, serialNumber)

	// 5. Validate Event Transaction ID retrieved from Validator API
	verify.EventTransactionIDFromValidatorAPI(t, setupEnv.Clients.ValidatorClient, expectedTxId, scheduledTxID)

	// 6. Validate that the NFT was sent to the receiver account
	verify.NftOwner(t, setupEnv.Clients.Hedera, nftToken, serialNumber, setupEnv.Clients.Hedera.GetOperatorAccountID())

	// 7. Prepare Expected Database Records
	expectedTxRecord := expected.NonFungibleTransferRecord(
		chainId,
		constants.HederaNetworkId,
		constants.HederaNetworkId,
		expectedTxId,
		wrappedAsset,
		nftToken,
		nftToken,
		setupEnv.Clients.Hedera.GetOperatorAccountID().String(),
		status.Completed,
		"",
		serialNumber,
		"	",
		evm.Signer.Address(),
		entity.NanoTime{Time: blockTimestamp})

	expectedScheduleTransferRecord := expected.ScheduleRecord(
		scheduledTxID,
		scheduleID,
		schedule.TRANSFER,
		true,
		status.Completed,
		sql.NullString{
			String: expectedTxId,
			Valid:  true,
		})

	// 8. Wait for validators to update DB state after Scheduled TX is mined
	time.Sleep(20 * time.Second)

	// 9. Validate Database Records
	verify.TransferRecord(t, setupEnv.DbValidator, expectedTxRecord)
	// and:
	verify.ScheduleRecord(t, setupEnv.DbValidator, expectedScheduleTransferRecord)
}

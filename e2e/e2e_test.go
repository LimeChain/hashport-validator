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
	mirror_node "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	hederahelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/hedera"
	lock_event "github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/lock-event"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/limechain/hedera-eth-bridge-validator/e2e/util"
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

	// TODO: id
	chainId := int64(80001)
	evm := setupEnv.Clients.EVM[chainId]
	receiver := evm.Receiver
	memo := evm.Receiver.String()
	mintAmount, fee := calculateReceiverAndFeeAmounts(setupEnv, hBarSendAmount.AsTinybar())

	// Step 1 - Verify the transfer of Hbars to the Bridge Account
	transactionResponse, wrappedBalanceBefore := verifyTransferToBridgeAccount(setupEnv, evm, memo, receiver, t)

	// Step 2 - Verify the submitted topic messages
	receivedSignatures := verifyTopicMessages(setupEnv, transactionResponse, t)

	// Step 3 - Validate fee scheduled transaction
	scheduledTxID, scheduleID := validateMembersScheduledTxs(setupEnv, constants.Hbar, generateMirrorNodeExpectedTransfersForHederaTransfer(setupEnv, constants.Hbar, fee), t)

	// Step 4 - Verify Transfer retrieved from Validator API
	transactionData, tokenAddress := verifyTransferFromValidatorAPI(setupEnv, evm, transactionResponse, constants.Hbar, mintAmount, 0, chainId, t)

	// Step 5 - Submit Mint transaction
	txHash := submitMintTransaction(evm, transactionResponse, transactionData, tokenAddress, t)

	// Step 6 - Wait for transaction to be mined
	waitForTransaction(evm, txHash, t)

	// Step 7 - Validate Token balances
	verifyWrappedAssetBalance(evm, constants.Hbar, big.NewInt(mintAmount), wrappedBalanceBefore, receiver, t)

	// Step 8 - Prepare Comparable Expected Transfer Record
	expectedTxRecord := util.PrepareExpectedTransfer(
		setupEnv.AssetMappings,
		0,
		chainId,
		transactionResponse.TransactionID,
		evm.RouterAddress.String(),
		constants.Hbar,
		strconv.FormatInt(hBarSendAmount.AsTinybar(), 10),
		receiver.String(),
		database.ExpectedStatuses{
			Status:          entity_transfer.StatusCompleted,
			StatusSignature: entity_transfer.StatusSignatureMined,
		}, t)
	// and:
	expectedFeeRecord := util.PrepareExpectedFeeRecord(
		scheduledTxID,
		scheduleID, fee,
		hederahelper.ToMirrorNodeTransactionID(transactionResponse.TransactionID.String()),
		"")

	// Step 9 - Verify Database Records
	verifyTransferRecordAndSignatures(setupEnv.DbValidator, expectedTxRecord, strconv.FormatInt(mintAmount, 10), receivedSignatures, t)
	// and:
	verifyFeeRecord(setupEnv.DbValidator, expectedFeeRecord, t)
}

// Test_E2E_Token_Transfer recreates a real life situation of a user who wants to bridge a Hedera native token to the EVM Network infrastructure. The wrapped token on the EVM network(corresponding to the native Hedera Hashgraph's one) gets minted, then transferred to the recipient account on the EVM network.
func Test_E2E_Token_Transfer(t *testing.T) {
	setupEnv := setup.Load()
	now = time.Now()

	// TODO: id
	chainId := int64(80001)
	evm := setupEnv.Clients.EVM[chainId]
	memo := evm.Receiver.String()
	mintAmount, fee := calculateReceiverAndFeeAmounts(setupEnv, tinyBarAmount)

	// Step 1 - Verify the transfer of HTS to the Bridge Account
	transactionResponse, wrappedBalanceBefore := verifyTokenTransferToBridgeAccount(setupEnv, evm, memo, evm.Receiver, t)

	// Step 2 - Verify the submitted topic messages
	receivedSignatures := verifyTopicMessages(setupEnv, transactionResponse, t)

	// Step 3 - Validate fee scheduled transaction
	scheduledTxID, scheduleID := validateMembersScheduledTxs(setupEnv, setupEnv.TokenID.String(), generateMirrorNodeExpectedTransfersForHederaTransfer(setupEnv, setupEnv.TokenID.String(), fee), t)

	// Step 4 - Verify Transfer retrieved from Validator API
	transactionData, tokenAddress := verifyTransferFromValidatorAPI(setupEnv, evm, transactionResponse, setupEnv.TokenID.String(), mintAmount, 0, chainId, t)

	// Step 5 - Submit Mint transaction
	txHash := submitMintTransaction(evm, transactionResponse, transactionData, tokenAddress, t)

	// Step 6 - Wait for transaction to be mined
	waitForTransaction(evm, txHash, t)

	// Step 7 - Validate Token balances
	verifyWrappedAssetBalance(evm, setupEnv.TokenID.String(), big.NewInt(mintAmount), wrappedBalanceBefore, evm.Receiver, t)

	// Step 8 - Verify Database records
	expectedTxRecord := util.PrepareExpectedTransfer(
		setupEnv.AssetMappings,
		0,
		chainId,
		transactionResponse.TransactionID,
		evm.RouterAddress.String(),
		setupEnv.TokenID.String(),
		strconv.FormatInt(tinyBarAmount, 10),
		evm.Receiver.String(),
		database.ExpectedStatuses{
			Status:          entity_transfer.StatusCompleted,
			StatusSignature: entity_transfer.StatusSignatureMined,
		}, t)
	// and:
	expectedFeeRecord := util.PrepareExpectedFeeRecord(
		scheduledTxID,
		scheduleID, fee,
		hederahelper.ToMirrorNodeTransactionID(transactionResponse.TransactionID.String()),
		"")

	// Step 9 - Verify Database Records
	verifyTransferRecordAndSignatures(setupEnv.DbValidator, expectedTxRecord, strconv.FormatInt(mintAmount, 10), receivedSignatures, t)
	// and:
	verifyFeeRecord(setupEnv.DbValidator, expectedFeeRecord, t)
}

// Test_EVM_Hedera_HBAR recreates a real life situation of a user who wants to return a Hedera native HBARs from the EVM Network infrastructure. The wrapped HBARs on the EVM network(corresponding to the native Hedera Hashgraph's one) gets burned, then the locked HBARs on the Hedera bridge account get unlocked, forwarding them to the recipient account.
func Test_EVM_Hedera_HBAR(t *testing.T) {
	setupEnv := setup.Load()
	// TODO: id
	chainId := int64(80001)
	evm := setupEnv.Clients.EVM[chainId]
	now = time.Now()
	accountBalanceBefore := util.GetHederaAccountBalance(setupEnv.Clients.Hedera, setupEnv.Clients.Hedera.GetOperatorAccountID(), t)

	// 1. Calculate Expected Receive And Fee Amounts
	expectedReceiveAmount, fee := calculateReceiverAndFeeAmounts(setupEnv, receiveAmount)

	// 2. Submit burn transaction to the bridge contract
	burnTxReceipt, expectedRouterBurn := sendBurnEthTransaction(setupEnv.Clients.Hedera, setupEnv.AssetMappings, evm, constants.Hbar, 0, chainId, t)

	// 3. Validate that the burn transaction went through and emitted the correct events
	expectedId := validateBurnEvent(burnTxReceipt, expectedRouterBurn, t)

	// 4. Validate that a scheduled transaction was submitted
	transactionID, scheduleID := validateSubmittedScheduledTx(setupEnv, constants.Hbar, generateMirrorNodeExpectedTransfersForBurnEvent(setupEnv, constants.Hbar, expectedReceiveAmount, fee), t)

	// 5. Validate Event Transaction ID retrieved from Validator API
	validateEventTransactionIDFromValidatorAPI(setupEnv, expectedId, transactionID, t)

	// 6. Validate that the balance of the receiver account (hedera) was changed with the correct amount
	validateReceiverAccountBalance(setupEnv, uint64(expectedReceiveAmount), accountBalanceBefore, constants.Hbar, t)

	// 7. Prepare Expected Database Records
	expectedBurnEventRecord := util.PrepareExpectedBurnEventRecord(
		scheduleID,
		receiveAmount,
		setupEnv.Clients.Hedera.GetOperatorAccountID(),
		expectedId,
		transactionID)
	// and:
	expectedFeeRecord := util.PrepareExpectedFeeRecord(transactionID, scheduleID, fee, "", expectedId)

	// 8. Wait for validators to update DB state after Scheduled TX is mined
	time.Sleep(10 * time.Second)

	// 9. Validate Database Records
	verifyBurnEventRecord(setupEnv.DbValidator, expectedBurnEventRecord, t)
	// and:
	verifyFeeRecord(setupEnv.DbValidator, expectedFeeRecord, t)
}

// Test_EVM_Hedera_Token recreates a real life situation of a user who wants to return a Hedera native token from the EVM Network infrastructure. The wrapped token on the EVM network(corresponding to the native Hedera one) gets burned, then the amount gets unlocked on the Hedera bridge account, forwarding it to the recipient account.
func Test_EVM_Hedera_Token(t *testing.T) {
	setupEnv := setup.Load()
	// TODO: id
	chainId := int64(80001)
	evm := setupEnv.Clients.EVM[chainId]
	now = time.Now()
	accountBalanceBefore := util.GetHederaAccountBalance(setupEnv.Clients.Hedera, setupEnv.Clients.Hedera.GetOperatorAccountID(), t)

	// 1. Calculate Expected Receive Amount
	expectedReceiveAmount, fee := calculateReceiverAndFeeAmounts(setupEnv, receiveAmount)

	// 2. Submit burn transaction to the bridge contract
	burnTxReceipt, expectedRouterBurn := sendBurnEthTransaction(setupEnv.Clients.Hedera, setupEnv.AssetMappings, evm, setupEnv.TokenID.String(), 0, chainId, t)

	// 3. Validate that the burn transaction went through and emitted the correct events
	expectedId := validateBurnEvent(burnTxReceipt, expectedRouterBurn, t)

	// 4. Validate that a scheduled transaction was submitted
	transactionID, scheduleID := validateSubmittedScheduledTx(setupEnv, setupEnv.TokenID.String(), generateMirrorNodeExpectedTransfersForBurnEvent(setupEnv, setupEnv.TokenID.String(), expectedReceiveAmount, fee), t)

	// 5. Validate Event Transaction ID retrieved from Validator API
	validateEventTransactionIDFromValidatorAPI(setupEnv, expectedId, transactionID, t)

	// 6. Validate that the balance of the receiver account (hedera) was changed with the correct amount
	validateReceiverAccountBalance(setupEnv, uint64(expectedReceiveAmount), accountBalanceBefore, setupEnv.TokenID.String(), t)

	// 7. Prepare Expected Database Records
	expectedBurnEventRecord := util.PrepareExpectedBurnEventRecord(
		scheduleID,
		receiveAmount,
		setupEnv.Clients.Hedera.GetOperatorAccountID(),
		expectedId,
		transactionID)
	// and:
	expectedFeeRecord := util.PrepareExpectedFeeRecord(transactionID, scheduleID, fee, "", expectedId)

	// 8. Wait for validators to update DB state after Scheduled TX is mined
	time.Sleep(10 * time.Second)

	// 9. Validate Database Records
	verifyBurnEventRecord(setupEnv.DbValidator, expectedBurnEventRecord, t)
	// and:
	verifyFeeRecord(setupEnv.DbValidator, expectedFeeRecord, t)
}

// Test_EVM_Hedera_Native_Token recreates a real life situation of a user who wants to bridge an EVM native token to the Hedera infrastructure. A new wrapped token (corresponding to the native EVM one) gets minted to the bridge account, then gets transferred to the recipient account.
func Test_EVM_Hedera_Native_Token(t *testing.T) {
	// Step 1: Initialize setup, smart contracts, etc.
	setupEnv := setup.Load()
	chainId := int64(3)
	evm := setupEnv.Clients.EVM[chainId]
	now = time.Now()
	bridgeAccountBalanceBefore := util.GetHederaAccountBalance(setupEnv.Clients.Hedera, setupEnv.BridgeAccount, t)
	receiverAccountBalanceBefore := util.GetHederaAccountBalance(setupEnv.Clients.Hedera, setupEnv.Clients.Hedera.GetOperatorAccountID(), t)
	wrappedAsset, err := setup.NativeToWrappedAsset(setupEnv.AssetMappings, 0, chainId, setupEnv.TokenID.String())
	if err != nil {
		panic(fmt.Sprintf("Token [%s] not supported", setupEnv.TokenID.String()))
	}

	// Step 2: Submit Lock Txn from a deployed smart contract
	receipt, expectedLockEventLog := sendLockEthTransaction(setupEnv.Clients.Hedera, evm, setupEnv.AssetMappings, setupEnv.TokenID.String(), chainId, 0, t)

	// Step 3: Validate Lock Event was emitted with correct data
	lockEventId := validateLockEvent(receipt, expectedLockEventLog, t)

	mintTransfer := []mirror_node.Transfer{
		{
			Account: setupEnv.BridgeAccount.String(),
			Amount:  receiveAmount,
			Token:   setupEnv.TokenID.String(),
		},
	}

	// Step 4: Validate that a scheduled token mint txn was submitted successfully
	bridgeMintTransactionID, bridgeScheduleID := validateScheduledMintTx(setupEnv, setupEnv.BridgeAccount, setupEnv.TokenID.String(), mintTransfer, t)

	// Step 5: Validate that Database statuses were changed correctly
	expectedLockEventRecord := util.PrepareExpectedLockEventRecord(receiveAmount,
		setupEnv.Clients.Hedera.GetOperatorAccountID(),
		lockEventId,
		"",
		bridgeMintTransactionID,
		"",
		bridgeScheduleID,
		setupEnv.TokenID.String(),
		wrappedAsset.String(),
		chainId,
		0,
		lock_event.StatusMintCompleted)
	verifyLockEventRecord(setupEnv.DbValidator, expectedLockEventRecord, t)

	// Step 7: Validate that a scheduled transfer txn was submitted successfully
	receiverTokenTransferTransactionID, receiverScheduleID := validateScheduledTx(setupEnv, setupEnv.Clients.Hedera.GetOperatorAccountID(), setupEnv.TokenID.String(), generateMirrorNodeExpectedTransfersForLockEvent(setupEnv, setupEnv.TokenID.String(), receiveAmount), t)

	// Step 8: Validate that database statuses were updated correctly
	expectedLockEventRecord.ScheduleTransferID = receiverTokenTransferTransactionID
	expectedLockEventRecord.ScheduleTransferTxId = sql.NullString{
		String: receiverScheduleID,
		Valid:  true,
	}
	expectedLockEventRecord.Status = lock_event.StatusCompleted
	verifyLockEventRecord(setupEnv.DbValidator, expectedLockEventRecord, t)

	// Step 9: Validate Treasury(BridgeAccount) Balance and Receiver Balance
	validateAccountBalance(setupEnv, setupEnv.BridgeAccount, 0, bridgeAccountBalanceBefore, setupEnv.TokenID.String(), t)
	validateAccountBalance(setupEnv, setupEnv.Clients.Hedera.GetOperatorAccountID(), uint64(receiveAmount), receiverAccountBalanceBefore, setupEnv.TokenID.String(), t)
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

func calculateReceiverAndFeeAmounts(setup *setup.Setup, amount int64) (receiverAmount, fee int64) {
	fee, remainder := setup.Clients.FeeCalculator.CalculateFee(amount)
	validFee := setup.Clients.Distributor.ValidAmount(fee)
	if validFee != fee {
		remainder += fee - validFee
	}

	return remainder, validFee
}

func submitMintTransaction(evm setup.EVMUtils, transactionResponse hedera.TransactionResponse, transactionData *service.TransferData, tokenAddress *common.Address, t *testing.T) common.Hash {
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
		big.NewInt(0),
		[]byte(hederahelper.FromHederaTransactionID(&transactionResponse.TransactionID).String()),
		*tokenAddress,
		evm.Receiver,
		mintAmount,
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

func sendBurnEthTransaction(hedera *hedera.Client, assetMappings config.AssetMappings, evm setup.EVMUtils, asset string, sourceChainId, targetChainId int64, t *testing.T) (*types.Receipt, *router.RouterBurn) {
	wrappedAsset, err := setup.NativeToWrappedAsset(assetMappings, sourceChainId, targetChainId, asset)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("Parsed [%s] to ETH Token [%s]", asset, wrappedAsset))

	approvedValue := big.NewInt(receiveAmount)

	var approveTx *types.Transaction
	if asset == constants.Hbar {
		approveTx, err = evm.WHbarContract.Approve(evm.KeyTransactor, evm.RouterAddress, approvedValue)
		if err != nil {
			t.Fatal(err)
		}
	} else {
		approveTx, err = evm.WTokenContract.Approve(evm.KeyTransactor, evm.RouterAddress, approvedValue)
		if err != nil {
			t.Fatal(err)
		}
	}

	fmt.Println(fmt.Sprintf("[%s] Waiting for Approval Transaction", approveTx.Hash()))
	waitForTransaction(evm, approveTx.Hash(), t)

	// TODO: ID
	burnTx, err := evm.RouterContract.Burn(evm.KeyTransactor, big.NewInt(0), *wrappedAsset, approvedValue, hedera.GetOperatorAccountID().ToBytes())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("[%s] Submitted Burn Transaction", burnTx.Hash()))

	expectedRouterBurn := &router.RouterBurn{
		//Account:      common.HexToAddress(evm.Signer.Address()),
		//WrappedAsset: *wrappedAsset,
		Amount:   approvedValue,
		Receiver: hedera.GetOperatorAccountID().ToBytes(),
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

func sendLockEthTransaction(hedera *hedera.Client, evm setup.EVMUtils, assetMappings config.AssetMappings, asset string, sourceChainId, targetChainId int64, t *testing.T) (*types.Receipt, *router.RouterLock) {
	wrappedHederaAsset, err := setup.NativeToWrappedAsset(assetMappings, targetChainId, sourceChainId, asset)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("Parsed [%s] to ETH Token [%s]", asset, wrappedHederaAsset))

	approvedValue := big.NewInt(receiveAmount)

	var approveTx *types.Transaction
	if asset == constants.Hbar {
		approveTx, err = evm.WHbarContract.Approve(evm.KeyTransactor, evm.RouterAddress, approvedValue)
		if err != nil {
			t.Fatal(err)
		}
	} else {
		approveTx, err = evm.WTokenContract.Approve(evm.KeyTransactor, evm.RouterAddress, approvedValue)
		if err != nil {
			t.Fatal(err)
		}
	}

	fmt.Println(fmt.Sprintf("[%s] Waiting for Approval Transaction", approveTx.Hash()))
	waitForTransaction(evm, approveTx.Hash(), t)

	lockTx, err := evm.RouterContract.Lock(evm.KeyTransactor, big.NewInt(0), *wrappedHederaAsset, approvedValue, hedera.GetOperatorAccountID().ToBytes())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("[%s] Submitted Lock Transaction", lockTx.Hash()))

	expectedRouterLock := &router.RouterLock{
		TargetChain: big.NewInt(targetChainId),
		Token:       *wrappedHederaAsset,
		Receiver:    hedera.GetOperatorAccountID().ToBytes(),
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
	var wrappedBalanceAfter *big.Int
	var err error
	if nativeAsset == constants.Hbar {
		wrappedBalanceAfter, err = evm.WHbarContract.BalanceOf(&bind.CallOpts{}, wTokenReceiverAddress)
	} else {
		wrappedBalanceAfter, err = evm.WTokenContract.BalanceOf(&bind.CallOpts{}, wTokenReceiverAddress)
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

func verifyTransferFromValidatorAPI(setupEnv *setup.Setup, evm setup.EVMUtils, txResponse hedera.TransactionResponse, tokenID string, expectedSendAmount int64, sourceChainId, targetChainId int64, t *testing.T) (*service.TransferData, *common.Address) {
	tokenAddress, err := setup.NativeToWrappedAsset(setupEnv.AssetMappings, sourceChainId, targetChainId, tokenID)
	if err != nil {
		t.Fatalf("Expecting Token [%s] is not supported. - Error: [%s]", tokenID, err)
	}

	transactionData, err := setupEnv.Clients.ValidatorClient.GetTransferData(hederahelper.FromHederaTransactionID(&txResponse.TransactionID).String())
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
	if transactionData.WrappedAsset != tokenAddress.String() {
		t.Fatalf("Token address mismatch: Expected [%s], but was [%s]", tokenAddress.String(), transactionData.WrappedAsset)
	}

	return transactionData, tokenAddress
}

func verifyBurnEventRecord(dbValidation *database.Service, expectedRecord *entity.BurnEvent, t *testing.T) {
	exist, err := dbValidation.VerifyBurnRecord(expectedRecord)
	if err != nil {
		t.Fatalf("[%s] - Verification of database records failed - Error: [%s].", expectedRecord.Id, err)
	}
	if !exist {
		t.Fatalf("[%s] - Database does not contain expected records", expectedRecord.Id)
	}
}

func verifyLockEventRecord(dbValidation *database.Service, expectedRecord *entity.LockEvent, t *testing.T) {
	exist, err := dbValidation.VerifyLockRecord(expectedRecord)
	if err != nil {
		t.Fatalf("[%s] - Verification of database records failed - Error: [%s].", expectedRecord.Id, err)
	}
	if !exist {
		t.Fatalf("[%s] - Database does not contain expected records", expectedRecord.Id)
	}
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

func verifyTransferRecordAndSignatures(dbValidation *database.Service, expectedRecord *entity.Transfer, mintAmount string, signatures []string, t *testing.T) {
	exist, err := dbValidation.VerifyTransferAndSignatureRecords(expectedRecord, mintAmount, signatures)
	if err != nil {
		t.Fatalf("[%s] - Verification of database records failed - Error: [%s].", expectedRecord.TransactionID, err)
	}
	if !exist {
		t.Fatalf("[%s] - Database does not contain expected records", expectedRecord.TransactionID)
	}
}

func verifyTransferToBridgeAccount(setup *setup.Setup, evm setup.EVMUtils, memo string, whbarReceiverAddress common.Address, t *testing.T) (hedera.TransactionResponse, *big.Int) {
	// Get the wrapped hbar balance of the receiver before the transfer
	whbarBalanceBefore, err := evm.WHbarContract.BalanceOf(&bind.CallOpts{}, whbarReceiverAddress)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(fmt.Sprintf("WHBAR balance before transaction: [%s]", whbarBalanceBefore))
	// Get bridge account hbar balance before transfer
	receiverBalance := util.GetHederaAccountBalance(setup.Clients.Hedera, setup.BridgeAccount, t).Hbars.AsTinybar()

	fmt.Println(fmt.Sprintf("Bridge account balance HBAR balance before transaction: [%d]", receiverBalance))

	// Get the transaction receipt to verify the transaction was executed
	transactionResponse, err := sendHbarsToBridgeAccount(setup, memo)
	if err != nil {
		t.Fatalf("Unable to send HBARs to Bridge Account, Error: [%s]", err)
	}

	transactionReceipt, err := transactionResponse.GetReceipt(setup.Clients.Hedera)
	if err != nil {
		t.Fatalf("Transaction unsuccessful, Error: [%s]", err)
	}

	fmt.Println(fmt.Sprintf("Successfully sent HBAR to bridge account, Status: [%s]", transactionReceipt.Status))

	// Get bridge account hbar balance after transfer
	receiverBalanceNew := util.GetHederaAccountBalance(setup.Clients.Hedera, setup.BridgeAccount, t).Hbars.AsTinybar()

	fmt.Println(fmt.Sprintf("Bridge Account HBAR balance after transaction: [%d]", receiverBalanceNew))

	// Verify that the custodial address has received exactly the amount sent
	amount := receiverBalanceNew - receiverBalance

	// Verify that the bridge account has received exactly the amount sent
	if amount != tinyBarAmount {
		t.Fatalf("Expected to receive the exact transfer amount of hbar: [%v], but was [%v]", tinyBarAmount, amount)
	}

	return *transactionResponse, whbarBalanceBefore
}

func verifyTokenTransferToBridgeAccount(setup *setup.Setup, evm setup.EVMUtils, memo string, wTokenReceiverAddress common.Address, t *testing.T) (hedera.TransactionResponse, *big.Int) {
	// Get the wrapped hts token balance of the receiver before the transfer
	wrappedBalanceBefore, err := evm.WTokenContract.BalanceOf(&bind.CallOpts{}, wTokenReceiverAddress)
	if err != nil {
		t.Fatalf("Unable to query the token balance of the receiver account. Error: [%s]", err)
	}

	fmt.Println(fmt.Sprintf("Token balance before transaction: [%s]", wrappedBalanceBefore))
	// Get bridge account token balance before transfer
	receiverBalance := util.GetHederaAccountBalance(setup.Clients.Hedera, setup.BridgeAccount, t)

	fmt.Println(fmt.Sprintf("Bridge account Token balance before transaction: [%d]", receiverBalance.Token[setup.TokenID]))

	// Get the transaction receipt to verify the transaction was executed
	transactionResponse, err := sendTokensToBridgeAccount(setup, memo)
	if err != nil {
		t.Fatalf(fmt.Sprintf("Unable to send Tokens to Bridge Account, Error: [%s]", err))
	}
	transactionReceipt, err := transactionResponse.GetReceipt(setup.Clients.Hedera)
	if err != nil {
		t.Fatalf(fmt.Sprintf("Transaction unsuccessful, Error: [%s]", err))
	}
	fmt.Println(fmt.Sprintf("Successfully sent Tokens to bridge account, Status: [%s]", transactionReceipt.Status))

	// Get bridge account HTS token balance after transfer
	receiverBalanceNew := util.GetHederaAccountBalance(setup.Clients.Hedera, setup.BridgeAccount, t)

	fmt.Println(fmt.Sprintf("Bridge Account Token balance after transaction: [%d]", receiverBalanceNew.Token[setup.TokenID]))

	// Verify that the custodial address has received exactly the amount sent
	resultAmount := receiverBalanceNew.Token[setup.TokenID] - receiverBalance.Token[setup.TokenID]
	// Verify that the bridge account has received exactly the amount sent
	if resultAmount != uint64(tinyBarAmount) {
		t.Fatalf("Expected to receive the exact transfer amount of hbar: [%v], but received: [%v]", hBarSendAmount.AsTinybar(), tinyBarAmount)
	}

	return *transactionResponse, wrappedBalanceBefore
}

func verifyTopicMessages(setup *setup.Setup, transactionResponse hedera.TransactionResponse, t *testing.T) []string {
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
				topicSubmissionMessageSign := hederahelper.FromHederaTransactionID(&transactionResponse.TransactionID)
				if msg.TransferID != topicSubmissionMessageSign.String() {
					fmt.Println(fmt.Sprintf(`Expected signature message to contain the transaction id: [%s]`, topicSubmissionMessageSign.String()))
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
	case <-time.After(60 * time.Second):
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

func sendTokensToBridgeAccount(setup *setup.Setup, memo string) (*hedera.TransactionResponse, error) {
	fmt.Println(fmt.Sprintf("Sending [%v] Tokens to the Bridge. Transaction Memo: [%s]", tinyBarAmount, memo))

	res, err := hedera.NewTransferTransaction().
		SetTransactionMemo(memo).
		AddTokenTransfer(setup.TokenID, setup.Clients.Hedera.GetOperatorAccountID(), -tinyBarAmount).
		AddTokenTransfer(setup.TokenID, setup.BridgeAccount, tinyBarAmount).
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

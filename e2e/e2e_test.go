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
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	hederahelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/limechain/hedera-eth-bridge-validator/e2e/util"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	routerContract "github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum/contracts/router"
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
	receiveAmount     uint64 = 100
	tinyBarAmount     int64  = 1000000000
	hBarSendAmount           = hedera.HbarFromTinybar(tinyBarAmount)
	hbarRemovalAmount        = hedera.HbarFromTinybar(-tinyBarAmount)
	now                      = time.Now()
)

const (
	expectedValidatorsCount = 3
)

func Test_Ethereum_Hedera_HBAR(t *testing.T) {
	setupEnv := setup.Load()
	now = time.Now()
	accountBalanceBefore := util.GetHederaAccountBalance(setupEnv.Clients.Hedera, setupEnv.Clients.Hedera.GetOperatorAccountID(), t)

	// 1. Calculate Expected Receive Amount
	expectedReceiveAmount := calculateValidAmount(setupEnv, int64(receiveAmount))

	// 2. Submit burn transaction to the bridge contract
	burnTxReceipt, expectedRouterBurn := sendEthTransaction(setupEnv, constants.Hbar, t)

	// 3. Validate that the burn transaction went through and emitted the correct events
	expectedId := validateBurnEvent(burnTxReceipt, expectedRouterBurn, t)

	// 4. Validate that a scheduled transaction was submitted
	entityId := validateScheduledTx(setupEnv, t)

	// 5. Validate that the balance of the receiver account (hedera) was changed with the correct amount
	validateReceiverAccountBalance(setupEnv, uint64(expectedReceiveAmount), accountBalanceBefore, constants.Hbar, t)

	// 6. Prepare Expected Database Record
	expectedBurnEventRecord := util.PrepareExpectedBurnEventRecord(
		entityId,
		int64(receiveAmount),
		setupEnv.Clients.Hedera.GetOperatorAccountID(),
		expectedId)

	// 7. Wait for potential front-run database
	time.Sleep(10 * time.Second)

	// 8. Validate Database Record
	verifyBurnEventRecord(setupEnv.DbValidator, expectedBurnEventRecord, t)
}

func Test_Ethereum_Hedera_Token(t *testing.T) {
	setupEnv := setup.Load()
	now = time.Now()
	accountBalanceBefore := util.GetHederaAccountBalance(setupEnv.Clients.Hedera, setupEnv.Clients.Hedera.GetOperatorAccountID(), t)

	// 1. Calculate Expected Receive Amount
	expectedReceiveAmount := calculateValidAmount(setupEnv, int64(receiveAmount))

	// 2. Submit burn transaction to the bridge contract
	burnTxReceipt, expectedRouterBurn := sendEthTransaction(setupEnv, setupEnv.TokenID.String(), t)

	// 3. Validate that the burn transaction went through and emitted the correct events
	expectedId := validateBurnEvent(burnTxReceipt, expectedRouterBurn, t)

	// 4. Validate that a scheduled transaction was submitted
	entityId := validateScheduledTx(setupEnv, t)

	// 5. Validate that the balance of the receiver account (hedera) was changed with the correct amount
	validateReceiverAccountBalance(setupEnv, uint64(expectedReceiveAmount), accountBalanceBefore, setupEnv.TokenID.String(), t)

	// 6. Prepare Expected Database Record
	expectedBurnEventRecord := util.PrepareExpectedBurnEventRecord(
		entityId,
		int64(receiveAmount),
		setupEnv.Clients.Hedera.GetOperatorAccountID(),
		expectedId)

	// 7. Wait for potential front-run database
	time.Sleep(10 * time.Second)

	// 8. Validate Database Record
	verifyBurnEventRecord(setupEnv.DbValidator, expectedBurnEventRecord, t)
}

func Test_HBAR(t *testing.T) {
	setupEnv := setup.Load()

	memo := setupEnv.EthReceiver.String()

	// Step 1 - Verify the transfer of Hbars to the Bridge Account
	transactionResponse, wrappedBalanceBefore := verifyTransferToBridgeAccount(setupEnv, memo, setupEnv.EthReceiver, t)

	// Step 2 - Verify the submitted topic messages
	receivedSignatures := verifyTopicMessages(setupEnv, transactionResponse, t)

	mintAmount := calculateValidAmount(setupEnv, hBarSendAmount.AsTinybar())

	// Step 3 - Verify Transfer retrieved from Validator API
	transactionData, tokenAddress := verifyTransferFromValidatorAPI(setupEnv, transactionResponse, constants.Hbar, mintAmount, t)

	// Step 4 - Submit Mint transaction
	txHash := submitMintTransaction(setupEnv, transactionResponse, transactionData, tokenAddress, t)

	// Step 5 - Wait for transaction to be mined
	waitForTransaction(setupEnv, txHash, t)

	// Step 6 - Validate Token balances
	verifyWrappedAssetBalance(setupEnv, constants.Hbar, big.NewInt(mintAmount), wrappedBalanceBefore, setupEnv.EthReceiver, t)

	// Step 7 - Prepare Comparable Expected Transfer Record
	expectedTxRecord := util.PrepareExpectedTransfer(
		setupEnv.Clients.RouterContract,
		transactionResponse.TransactionID,
		constants.Hbar,
		strconv.FormatInt(hBarSendAmount.AsTinybar(), 10),
		setupEnv.EthReceiver.String(),
		database.ExpectedStatuses{
			Status:          entity_transfer.StatusCompleted,
			StatusSignature: entity_transfer.StatusSignatureMined,
		}, t)

	// Step 8 - Verify Database Records
	verifyTransferRecordAndSignatures(setupEnv.DbValidator, expectedTxRecord, strconv.FormatInt(mintAmount, 10), receivedSignatures, t)
}

func Test_E2E_Token_Transfer(t *testing.T) {
	setupEnv := setup.Load()

	memo := setupEnv.EthReceiver.String()

	// Step 1 - Verify the transfer of HTS to the Bridge Account
	transactionResponse, wrappedBalanceBefore := verifyTokenTransferToBridgeAccount(setupEnv, memo, setupEnv.EthReceiver, t)

	// Step 2 - Verify the submitted topic messages
	receivedSignatures := verifyTopicMessages(setupEnv, transactionResponse, t)

	mintAmount := calculateValidAmount(setupEnv, tinyBarAmount)

	// Step 3 - Verify Transfer retrieved from Validator API
	transactionData, tokenAddress := verifyTransferFromValidatorAPI(setupEnv, transactionResponse, setupEnv.TokenID.String(), mintAmount, t)

	// Step 4 - Submit Mint transaction
	txHash := submitMintTransaction(setupEnv, transactionResponse, transactionData, tokenAddress, t)

	// Step 5 - Wait for transaction to be mined
	waitForTransaction(setupEnv, txHash, t)

	// Step 6 - Validate Token balances
	verifyWrappedAssetBalance(setupEnv, setupEnv.TokenID.String(), big.NewInt(mintAmount), wrappedBalanceBefore, setupEnv.EthReceiver, t)

	// Step 7 - Verify Database records
	expectedTxRecord := util.PrepareExpectedTransfer(
		setupEnv.Clients.RouterContract,
		transactionResponse.TransactionID,
		setupEnv.TokenID.String(),
		strconv.FormatInt(tinyBarAmount, 10),
		setupEnv.EthReceiver.String(),
		database.ExpectedStatuses{
			Status:          entity_transfer.StatusCompleted,
			StatusSignature: entity_transfer.StatusSignatureMined,
		}, t)

	// Step 8 - Verify Database Records
	verifyTransferRecordAndSignatures(setupEnv.DbValidator, expectedTxRecord, strconv.FormatInt(mintAmount, 10), receivedSignatures, t)
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

func validateBurnEvent(txReceipt *types.Receipt, expectedRouterBurn *routerContract.RouterBurn, t *testing.T) string {
	parsedAbi, err := abi.JSON(strings.NewReader(routerContract.RouterABI))
	if err != nil {
		t.Fatal(err)
	}

	routerBurn := routerContract.RouterBurn{}
	eventSignature := []byte("Burn(address,address,uint256,bytes)")
	eventSignatureHash := crypto.Keccak256Hash(eventSignature)
	for _, log := range txReceipt.Logs {
		if log.Topics[0] != eventSignatureHash {
			continue
		}

		account := log.Topics[1]
		wrappedAsset := log.Topics[2]
		err := parsedAbi.UnpackIntoInterface(&routerBurn, "Burn", log.Data)
		if err != nil {
			t.Fatal(err)
		}

		if routerBurn.Amount.String() != expectedRouterBurn.Amount.String() {
			t.Fatalf("Expected Burn Event Amount [%v], but actually was [%v]", expectedRouterBurn.Amount, routerBurn.Amount)
		}

		if wrappedAsset != expectedRouterBurn.WrappedAsset.Hash() {
			t.Fatalf("Expected Burn Event Wrapped Token [%v], but actually was [%v]", expectedRouterBurn.WrappedAsset, routerBurn.WrappedAsset)
		}

		if !reflect.DeepEqual(routerBurn.Receiver, expectedRouterBurn.Receiver) {
			t.Fatalf("Expected Burn Event Receiver [%v], but actually was [%v]", expectedRouterBurn.Receiver, routerBurn.Receiver)
		}

		if account != expectedRouterBurn.Account.Hash() {
			t.Fatalf("Expected Burn Event Account [%v], but actually was [%v]", expectedRouterBurn.Account, routerBurn.Account)
		}

		expectedId := fmt.Sprintf("%s-%d", log.TxHash, log.Index)
		return expectedId
	}

	t.Fatal("Could not retrieve valid Burn Event Log information.")
	return ""
}

func validateScheduledTx(setupEnv *setup.Setup, t *testing.T) string {
	timeLeft := 180
	for {
		transactions, err := setupEnv.Clients.MirrorNode.GetAccountCreditTransactionsAfterTimestamp(setupEnv.Clients.Hedera.GetOperatorAccountID(), now.UnixNano())
		if err != nil {
			t.Fatal(err)
		}

		for _, transaction := range transactions.Transactions {
			if transaction.Scheduled == true {
				scheduleCreateTx, err := setupEnv.Clients.MirrorNode.GetTransaction(transaction.TransactionID)
				if err != nil {
					t.Fatal(err)
				}

				for _, tx := range scheduleCreateTx.Transactions {
					if tx.EntityId != "" {
						return tx.EntityId
					}
				}
			}
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
	return ""
}

func calculateValidAmount(setup *setup.Setup, amount int64) int64 {
	fee, remainder := setup.Clients.FeeCalculator.CalculateFee(amount)
	validFee := setup.Clients.Distributor.ValidAmount(fee)
	if validFee != fee {
		remainder += fee - validFee
	}

	return remainder
}

func submitMintTransaction(setupEnv *setup.Setup, transactionResponse hedera.TransactionResponse, transactionData *service.TransferData, tokenAddress *common.Address, t *testing.T) common.Hash {
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

	res, err := setupEnv.Clients.RouterContract.Mint(
		setupEnv.Clients.KeyTransactor,
		[]byte(hederahelper.FromHederaTransactionID(&transactionResponse.TransactionID).String()),
		*tokenAddress,
		setupEnv.EthReceiver,
		mintAmount,
		signatures,
	)

	if err != nil {
		t.Fatalf("Cannot execute transaction - Error: [%s].", err)
	}
	return res.Hash()
}

func sendEthTransaction(setupEnv *setup.Setup, asset string, t *testing.T) (*types.Receipt, *routerContract.RouterBurn) {
	wrappedAsset, err := setup.WrappedAsset(setupEnv.Clients.RouterContract, asset)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("Parsed [%s] to ETH Token [%s]", asset, wrappedAsset))

	approvedValue := new(big.Int).SetUint64(receiveAmount)

	controller, err := setupEnv.Clients.RouterContract.Controller(nil)
	if err != nil {
		t.Fatal(err)
	}

	var approveTx *types.Transaction
	if asset == constants.Hbar {
		approveTx, err = setupEnv.Clients.WHbarContract.Approve(setupEnv.Clients.KeyTransactor, controller, approvedValue)
	} else {
		approveTx, err = setupEnv.Clients.WTokenContract.Approve(setupEnv.Clients.KeyTransactor, controller, approvedValue)
	}
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(fmt.Sprintf("[%s] Waiting for Approval Transaction", approveTx.Hash()))
	waitForTransaction(setupEnv, approveTx.Hash(), t)

	burnTx, err := setupEnv.Clients.RouterContract.Burn(setupEnv.Clients.KeyTransactor, approvedValue, setupEnv.Clients.Hedera.GetOperatorAccountID().ToBytes(), *wrappedAsset)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("[%s] Submitted Burn Transaction", burnTx.Hash()))

	expectedRouterBurn := &routerContract.RouterBurn{
		Account:      common.HexToAddress(setupEnv.Clients.Signer.Address()),
		WrappedAsset: *wrappedAsset,
		Amount:       approvedValue,
		Receiver:     setupEnv.Clients.Hedera.GetOperatorAccountID().ToBytes(),
	}

	burnTxHash := burnTx.Hash()

	fmt.Println(fmt.Sprintf("[%s] Waiting for Burn Transaction Receipt.", burnTxHash))
	burnTxReceipt, err := setupEnv.Clients.EthClient.WaitForTransactionReceipt(burnTxHash)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("[%s] Burn Transaction mined and retrieved receipt.", burnTxHash))

	return burnTxReceipt, expectedRouterBurn
}

func waitForTransaction(setupEnv *setup.Setup, txHash common.Hash, t *testing.T) {
	receipt, err := setupEnv.Clients.EthClient.WaitForTransactionReceipt(txHash)
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

func verifyWrappedAssetBalance(setupEnv *setup.Setup, nativeAsset string, mintAmount *big.Int, wrappedBalanceBefore *big.Int, wTokenReceiverAddress common.Address, t *testing.T) {
	var wrappedBalanceAfter *big.Int
	var err error
	if nativeAsset == constants.Hbar {
		wrappedBalanceAfter, err = setupEnv.Clients.WHbarContract.BalanceOf(&bind.CallOpts{}, wTokenReceiverAddress)
	} else {
		wrappedBalanceAfter, err = setupEnv.Clients.WTokenContract.BalanceOf(&bind.CallOpts{}, wTokenReceiverAddress)
	}

	if err != nil {
		t.Fatal(err)
	}

	expectedBalance := new(big.Int).Add(wrappedBalanceBefore, mintAmount)

	if wrappedBalanceAfter.Cmp(expectedBalance) != 0 {
		t.Fatalf("Incorrect token balance. Expected to be [%s], but was [%s].", expectedBalance, wrappedBalanceAfter)
	}
}

func verifyTransferFromValidatorAPI(setupEnv *setup.Setup, txResponse hedera.TransactionResponse, tokenID string, expectedSendAmount int64, t *testing.T) (*service.TransferData, *common.Address) {
	tokenAddress, err := setup.WrappedAsset(setupEnv.Clients.RouterContract, tokenID)
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
	if transactionData.Recipient != setupEnv.EthReceiver.String() {
		t.Fatalf("Receiver address mismatch: Expected [%s], but was [%s]", setupEnv.EthReceiver.String(), transactionData.Recipient)
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

func verifyTransferRecordAndSignatures(dbValidation *database.Service, expectedRecord *entity.Transfer, mintAmount string, signatures []string, t *testing.T) {
	exist, err := dbValidation.VerifyTransferAndSignatureRecords(expectedRecord, mintAmount, signatures)
	if err != nil {
		t.Fatalf("[%s] - Verification of database records failed - Error: [%s].", expectedRecord.TransactionID, err)
	}
	if !exist {
		t.Fatalf("[%s] - Database does not contain expected records", expectedRecord.TransactionID)
	}
}

func verifyTransferToBridgeAccount(setup *setup.Setup, memo string, whbarReceiverAddress common.Address, t *testing.T) (hedera.TransactionResponse, *big.Int) {
	// Get the wrapped hbar balance of the receiver before the transfer
	whbarBalanceBefore, err := setup.Clients.WHbarContract.BalanceOf(&bind.CallOpts{}, whbarReceiverAddress)
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

func verifyTokenTransferToBridgeAccount(setup *setup.Setup, memo string, wTokenReceiverAddress common.Address, t *testing.T) (hedera.TransactionResponse, *big.Int) {
	// Get the wrapped hts token balance of the receiver before the transfer
	wrappedBalanceBefore, err := setup.Clients.WTokenContract.BalanceOf(&bind.CallOpts{}, wTokenReceiverAddress)
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

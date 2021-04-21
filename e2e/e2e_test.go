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
	"github.com/ethereum/go-ethereum/crypto"
	mirror_node "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
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
	tinyBarAmount     int64 = 1000000000
	hBarReceiveAmount int64 = 100
	hBarSendAmount          = hedera.HbarFromTinybar(tinyBarAmount)
	hbarRemovalAmount       = hedera.HbarFromTinybar(-tinyBarAmount)
	now                     = time.Now()
)

const (
	expectedValidatorsCount = 3
)

func Test_Ethereum_Hedera_HBAR(t *testing.T) {
	setupEnv := setup.Load()
	now = time.Now()
	accountBalanceBefore := util.GetHederaAccountBalance(setupEnv.Clients.Hedera, setupEnv.HederaReceiver, t)

	// 1. Submit burn transaction to the bridge contract
	submittedTx, expectedRouterBurn := sendEthTransaction(setupEnv, t)

	// 2. Validate burn event
	expectedId := validateBurnEvent(setupEnv, submittedTx, expectedRouterBurn, t)

	// 3. Validate that the tx went through and emitted the correct events
	mirrorNodeScheduledTransaction := validateScheduledTx(setupEnv, t)

	// 4. Validate that the balance of the receiver account (hedera) was changed with the correct amount
	validateReceiverAccountBalance(setupEnv, accountBalanceBefore, t)

	// 5. Prepare Expected Database Record
	expectedBurnEventRecord := util.PrepareExpectedBurnEventRecord(
		mirrorNodeScheduledTransaction.TransactionID,
		hBarReceiveAmount,
		setupEnv.HederaReceiver,
		expectedId)

	// 6. Validate Database Record
	verifyBurnEventRecord(setupEnv.DbValidator, expectedBurnEventRecord, t)
}

func Test_HBAR(t *testing.T) {
	setupEnv := setup.Load()

	memo := setupEnv.EthReceiver.String()

	// Step 1 - Verify the transfer of Hbars to the Bridge Account
	transactionResponse, wrappedBalanceBefore := verifyTransferToBridgeAccount(setupEnv, memo, setupEnv.EthReceiver, t)

	// Step 2 - Verify the submitted topic messages
	receivedSignatures := verifyTopicMessages(setupEnv, transactionResponse, t)

	mintAmount := calculateMintAmount(setupEnv, hBarSendAmount.AsTinybar())

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

	mintAmount := calculateMintAmount(setupEnv, tinyBarAmount)

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

func validateReceiverAccountBalance(setup *setup.Setup, beforeHBARBalance hedera.AccountBalance, t *testing.T) {
	beforeTransfer := beforeHBARBalance.Hbars.AsTinybar()
	afterTransfer := util.GetHederaAccountBalance(setup.Clients.Hedera, setup.HederaReceiver, t).Hbars.AsTinybar()

	if afterTransfer-beforeTransfer != hBarReceiveAmount {
		t.Fatalf("Expected account balance after - [%d], but was [%d]", afterTransfer, beforeTransfer)
	}
}

func validateBurnEvent(setupEnv *setup.Setup, txHash common.Hash, expectedRouterBurn *routerContract.RouterBurn, t *testing.T) string {
	fmt.Println(fmt.Sprintf("[%s] Waiting for Burn Transaction Receipt.", txHash))
	txReceipt, err := setupEnv.Clients.EthClient.WaitForTransactionReceipt(txHash)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("[%s] Burn Transaction mined and retrieved receipt.", txHash))

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

func validateScheduledTx(setupEnv *setup.Setup, t *testing.T) *mirror_node.Transaction {
	timeLeft := 480
	for {
		transactions, err := setupEnv.Clients.MirrorNode.GetAccountCreditTransactionsAfterTimestamp(setupEnv.HederaReceiver, now.UnixNano())
		if err != nil {
			t.Fatal(err)
		}
		for _, transaction := range transactions.Transactions {
			fmt.Println(transaction.TransactionID)
			if transaction.Scheduled == true {
				// amount, "HBAR"
				_, _, err := transaction.GetIncomingTransfer(setupEnv.HederaReceiver.String())
				if err != nil {
					t.Fatal(err)
				}
				return &transaction
			}
		}

		if timeLeft > 0 {
			timeLeft -= 5
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}

	t.Fatalf("Could not find any scheduled transactions for account [%s]", setupEnv.HederaReceiver)
	return nil
}

func calculateMintAmount(setup *setup.Setup, amount int64) int64 {
	fee, remainder := setup.Clients.FeeCalculator.CalculateFee(amount)
	validFee := setup.Clients.Distributor.ValidAmount(fee)
	if validFee != fee {
		remainder += fee - validFee
	}

	return remainder
}

func submitMintTransaction(setupEnv *setup.Setup, transactionResponse hedera.TransactionResponse, transactionData *service.TransferData, tokenAddress *common.Address, t *testing.T) string {
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
	return res.Hash().String()
}

func sendEthTransaction(setupEnv *setup.Setup, t *testing.T) (common.Hash, *routerContract.RouterBurn) {
	wrappedAsset, err := setup.WrappedAsset(setupEnv.Clients.RouterContract, "HBAR")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("Parsed [%s] to ETH Token [%s]", "HBAR", wrappedAsset))

	approvedValue := big.NewInt(hBarReceiveAmount)
	setupEnv.Clients.KeyTransactor.GasPrice = big.NewInt(50000000000)
	approveTx, err := setupEnv.Clients.WHbarContract.Approve(setupEnv.Clients.KeyTransactor, setupEnv.ControllerContractAddress, approvedValue)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(fmt.Sprintf("[%s] Waiting for Approval Transaction", approveTx.Hash()))
	waitForTransaction(setupEnv, approveTx.Hash().String(), t)

	burnTx, err := setupEnv.Clients.RouterContract.Burn(setupEnv.Clients.KeyTransactor, approvedValue, setupEnv.HederaReceiver.ToBytes(), *wrappedAsset)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("[%s] Submitted Burn Transaction", burnTx.Hash()))

	expectedRouterBurn := &routerContract.RouterBurn{
		Account:      common.HexToAddress(setupEnv.Clients.Signer.Address()),
		WrappedAsset: *wrappedAsset,
		Amount:       approvedValue,
		Receiver:     setupEnv.HederaReceiver.ToBytes(),
	}
	return burnTx.Hash(), expectedRouterBurn
}

// TODO: Call only WaitForReceipt and make the function synchronous
func waitForTransaction(setupEnv *setup.Setup, txHash string, t *testing.T) {
	fmt.Println(fmt.Sprintf("Waiting for transaction: [%s] to be mined", txHash))
	c1 := make(chan bool, 1)
	onSuccess := func() {
		fmt.Println(fmt.Sprintf("Transaction [%s] mined successfully", txHash))
		c1 <- true
	}
	onRevert := func() {
		t.Fatalf("Failed to mine ethereum transaction: [%s]", txHash)
	}
	onError := func(err error) {
		t.Fatalf(fmt.Sprintf("Transaction unsuccessful, Error: [%s]", err))
	}
	setupEnv.Clients.EthClient.WaitForTransaction(txHash, onSuccess, onRevert, onError)
	<-c1
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
	exist, err := dbValidation.VerifyDatabaseRecords(expectedRecord, mintAmount, signatures)
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
		AddHbarTransfer(setup.HederaSender, hbarRemovalAmount).
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
		AddTokenTransfer(setup.TokenID, setup.HederaSender, -tinyBarAmount).
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

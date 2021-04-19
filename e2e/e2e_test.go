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
	"github.com/ethereum/go-ethereum/core/types"
	hederahelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/hedera"
	burn_event "github.com/limechain/hedera-eth-bridge-validator/app/model/burn-event"
	hederahelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"math/big"
	"strconv"
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
	amount               float64 = 400
	hBarSendAmount               = hedera.HbarFrom(amount, "hbar")
	tokensSendAmount     int64   = 1000000000
	hbarRemovalAmount            = hedera.HbarFrom(-amount, "hbar")
	precision                    = new(big.Int).SetInt64(100000)
	whbarReceiverAddress         = common.HexToAddress(receiverAddress)
)

const (
	receiverAddress         = "0x7cFae2deF15dF86CfdA9f2d25A361f1123F42eDD"
	expectedValidatorsCount = 3
)

func Test_Ethereum_Hedera_HBAR(t *testing.T) {
	setupEnv := setup.Load()
	accountBalanceBefore := getAccountBalance(setupEnv, t)

	// 1. Submit burn transaction to the bridge contract
	submittedTx := sendEthTransaction(setupEnv, t)

	// 2. Validate that the tx went through and emitted the correct events
	validateTxAndEvents(setupEnv, submittedTx, t)

	// 3. Validate that the balance of the receiver account (hedera) was changed with the correct amount
	validateBridgeAccountBalance(setupEnv, accountBalanceBefore, t)
}

func getAccountBalance(setup *setup.Setup, t *testing.T) hedera.AccountBalance {
	// Get bridge account hbar balance before transfer
	receiverBalance, err := hedera.NewAccountBalanceQuery().
		SetAccountID(setup.BridgeAccount).
		Execute(setup.Clients.Hedera)
	if err != nil {
		t.Fatalf("Unable to query the balance of the Bridge Account, Error: [%s]", err)
	}
	return receiverBalance
}

func validateBridgeAccountBalance(setup *setup.Setup, beforeHBARBalance hedera.AccountBalance, t *testing.T) {
	// Post Process HBAR Account Balance
	afterHBARBalance := getAccountBalance(setup, t)

	if afterHBARBalance.Hbars.AsTinybar()-beforeHBARBalance.Hbars.AsTinybar() != hBarSendAmount.AsTinybar() {
		t.Fatalf("Expected account balance after - [%d], but was [%d]", afterHBARBalance, beforeHBARBalance)
	}
}

func validateFees() {

}

func validateTxAndEvents(setupEnv *setup.Setup, tx *types.Transaction, t *testing.T) {
	waitForTransaction(setupEnv, tx.Hash().String(), t)
	listenBurnEvents(setupEnv, t)
	validateScheduledTransactionSubmission(setupEnv, tx, t)
}

func listenBurnEvents(setupEnv *setup.Setup, t *testing.T) {
	events := make(chan *routerContract.RouterBurn)

	sub, err := setupEnv.Clients.RouterContract.WatchBurn(nil, events, nil, nil)
	if err != nil {
		t.Fatalf("Failed to subscribe for Burn Event Logs for contract address [%s]. Error [%s].", setupEnv.RouterContractAddress, err)
	}

	for {
		select {
		case err := <-sub.Err():
			t.Fatalf("Burn Event Logs subscription failed. Error: [%s].", err)
		case eventLog := <-events:
			// TODO: Compare Event Log and Validate
			validateEvent(setupEnv, eventLog, t)
		}
	}
}

var expectedBurnEvent = burn_event.BurnEvent{
	Id:           "",
	Amount:       0,
	Recipient:    hedera.AccountID{},
	NativeToken:  "",
	WrappedToken: "",
}

func validateEvent(setupEnv *setup.Setup, eventLog *routerContract.RouterBurn, t *testing.T) {
	fmt.Println(fmt.Sprintf("[%s] - New Burn Event Log received. Waiting block confirmations", eventLog.Raw.TxHash))

	if eventLog.Raw.Removed {
		fmt.Println(fmt.Sprintf("[%s] - Uncle block transaction was removed.", eventLog.Raw.TxHash))
		return
	}

	eventAccount := string(common.TrimRightZeroes(eventLog.Receiver))
	recipientAccount, err := hedera.AccountIDFromString(eventAccount)
	if err != nil {
		t.Fatalf("[%s] - Failed to parse account [%s]. Error: [%s].", eventLog.Raw.TxHash, eventAccount, err)
	}
	nativeToken, err := setup.ParseETHToHederaToken(setupEnv.Clients.RouterContract, eventLog.WrappedToken)
	if err != nil {
		t.Fatalf("[%s] - Failed to retrieve native token of [%s]. Error: [%s].", eventLog.Raw.TxHash, eventLog.WrappedToken, err)
	}

	if nativeToken != constants.Hbar && !hederahelper.IsTokenID(nativeToken) {
		t.Fatalf("[%s] - Invalid Native Token [%s].", eventLog.Raw.TxHash, nativeToken)
		return
	}

	actualBurnEvent := burn_event.BurnEvent{
		Amount:       eventLog.Amount.Int64(),
		Id:           fmt.Sprintf("%s-%d", eventLog.Raw.TxHash, eventLog.Raw.Index),
		Recipient:    recipientAccount,
		WrappedToken: eventLog.WrappedToken.String(),
		NativeToken:  nativeToken,
	}

	// TODO: Compare
	if !equalEvents(expectedBurnEvent, actualBurnEvent) {

	}
}

func equalEvents(comparing burn_event.BurnEvent, comparable burn_event.BurnEvent) bool {
	return comparing == comparable
}

func validateScheduledTransactionSubmission(setupEnv *setup.Setup, tx *types.Transaction, t *testing.T) {
	setupEnv.Clients.MirrorNode.WaitForScheduledTransferTransaction(tx.Hash().String(), func() {}, func() {})
}

func sendEthTransaction(setupEnv *setup.Setup, t *testing.T) *types.Transaction {
	wrappedToken, err := setup.ParseHederaToETHToken(setupEnv.Clients.RouterContract, "HBAR")
	if err != nil {
		// TODO: Log properly and fatal
		t.Fatal(err)
	}

	fmt.Println(fmt.Sprintf("Parsed [%s] to ETH Token [%s]", "HBAR", wrappedToken))

	value := big.NewInt(0)
	// TODO: Which one will be the receiver address?!
	tx, err := setupEnv.Clients.RouterContract.Burn(setupEnv.Clients.KeyTransactor, value, setupEnv.BridgeAccount.ToBytes(), *wrappedToken)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(fmt.Sprintf("Submitted Burn Transaction [%s]", tx.Hash()))
	return tx
}

// TODO: Reuse or delete before PR
func commentedStuff() {
	//privateKey, err := crypto.HexToECDSA(setupEnv.EthSender.PrivateKey)
	//if err != nil {
	//	TODO: Log properly and fatal
	//}
	//
	//publicKey := privateKey.Public()
	//publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	//if !ok {
	//	// TODO: Log properly and fatal
	//}
	//
	//fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	//ethClient := setupEnv.Clients.EthClient.Client
	//nonce, err := ethClient.PendingNonceAt(context.Background(), fromAddress)
	//if err != nil {
	//	// TODO: Log properly and fatal
	//}
	//
	//value := big.NewInt(10000000)
	//gasLimit := uint64(21000)
	//gasPrice, err := ethClient.SuggestGasPrice(context.Background())
	//if err != nil {
	//	// TODO: Log properly and fatal
	//}
	//
	//toAddress := common.HexToAddress(setupEnv.RouterContractAddress)
	//
	//chainID, err := ethClient.ChainID(context.Background())
	//if err != nil {
	//	// TODO: Log properly and fatal
	//}
	//
	//var data []byte
	//
	//tx := types.NewTx(&types.AccessListTx{
	//	ChainID:    chainID,
	//	Nonce:      nonce,
	//	GasPrice:   gasPrice,
	//	Gas:        gasLimit,
	//	To:         &toAddress,
	//	Value:      value,
	//	Data:       data,
	//	AccessList: nil,
	//	V:          nil,
	//	R:          nil,
	//	S:          nil,
	//})
	//
	//err = ethClient.SendTransaction(
	//	context.Background(),
	//	tx,
	//)
	//if err != nil {
	//	// TODO: Log properly and fatal
	//}
	//
}

func Test_HBAR(t *testing.T) {
	setupEnv := setup.Load()

	memo := receiverAddress

	// Step 1 - Verify the transfer of Hbars to the Bridge Account
	transactionResponse, _ := verifyTransferToBridgeAccount(setupEnv, memo, whbarReceiverAddress, t)

	// Step 2 - Verify the submitted topic messages
	receivedSignatures := verifyTopicMessages(setupEnv, transactionResponse, t)

	// Step 3 - Verify Transfer retrieved from Validator API
	_, _ = verifyTransferFromValidatorAPI(setupEnv, transactionResponse, constants.Hbar, hBarSendAmount.AsTinybar(), t)

	// Step 4 - Prepare Comparable Expected Transfer Record
	expectedTxRecord := prepareExpectedTransfer(
		setupEnv.Clients.RouterContract,
		transactionResponse.TransactionID,
		constants.Hbar,
		strconv.FormatInt(hBarSendAmount.AsTinybar(), 10),
		database.ExpectedStatuses{
			Status:          entity_transfer.StatusCompleted,
			StatusSignature: entity_transfer.StatusSignatureMined,
		}, t)

	// Step 5 - Verify Database Records
	verifyDatabaseRecords(setupEnv.DbValidation, expectedTxRecord, receivedSignatures, t)
}

func Test_E2E_Token_Transfer(t *testing.T) {
	setupEnv := setup.Load()

	wTokenReceiverAddress := common.HexToAddress(receiverAddress)

	// Step 1 - Verify the transfer of HTS to the Bridge Account
	transactionResponse, wrappedTokenBalanceBefore := verifyTokenTransferToBridgeAccount(setupEnv, receiverAddress, wTokenReceiverAddress, t)

	// Step 2 - Verify the submitted topic messages
	receivedSignatures := verifyTopicMessages(setupEnv, transactionResponse, t)

	// Step 3 - Verify Transfer retrieved from Validator API
	transactionData, tokenAddress := verifyTransferFromValidatorAPI(setupEnv, transactionResponse, setupEnv.TokenID.String(), tokensSendAmount, t)

	// Step 4 - Submit Mint transaction
	txHash := submitMintTransaction(setupEnv, transactionResponse, transactionData, tokenAddress, t)

	// Step 5 - Wait for transaction to be mined
	waitForTransaction(setupEnv, txHash, t)

	// Step 6 - Validate Token balances
	validateTokenBalance(setupEnv, wrappedTokenBalanceBefore, wTokenReceiverAddress, t)

	// Step 7 - Verify Database records
	expectedTxRecord := prepareExpectedTransfer(
		setupEnv.Clients.RouterContract,
		transactionResponse.TransactionID,
		setupEnv.TokenID.String(),
		strconv.FormatInt(tokensSendAmount, 10),
		database.ExpectedStatuses{
			Status:          entity_transfer.StatusCompleted,
			StatusSignature: entity_transfer.StatusSignatureMined,
		}, t)
	verifyDatabaseRecords(setupEnv.DbValidation, expectedTxRecord, receivedSignatures, t)
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

	res, err := setupEnv.Clients.RouterContract.Mint(
		setupEnv.Clients.KeyTransactor,
		[]byte(hederahelper.FromHederaTransactionID(&transactionResponse.TransactionID).String()),
		*tokenAddress,
		common.HexToAddress(receiverAddress),
		big.NewInt(int64(tokensSendAmount)),
		signatures,
	)

	if err != nil {
		t.Fatalf("Cannot execute transaction - Error: [%s].", err)
	}
	return res.Hash().String()
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

func validateTokenBalance(setupEnv *setup.Setup, wrappedTokenBalanceBefore *big.Int, wTokenReceiverAddress common.Address, t *testing.T) {
	wrappedTokenBalanceAfter, err := setupEnv.Clients.WTokenContract.BalanceOf(&bind.CallOpts{}, wTokenReceiverAddress)
	if err != nil {
		t.Fatal(err)
	}

	tokensAmount := big.NewInt(int64(tokensSendAmount))

	serviceFeePercentage, err := setupEnv.Clients.RouterContract.ServiceFee(nil)
	if err != nil {
		t.Fatal(err)
	}

	txFee := new(big.Int).Mul(tokensAmount, serviceFeePercentage)
	txFee = new(big.Int).Div(txFee, precision)

	newBalance := new(big.Int).Sub(wrappedTokenBalanceAfter, wrappedTokenBalanceBefore)
	expectedBalance := new(big.Int).Sub(tokensAmount, txFee)

	if newBalance.Cmp(expectedBalance) != 0 {
		t.Fatalf("Incorect token balance. Expected to be [%s], but was [%s].", expectedBalance, newBalance)
	}
}

func verifyTransferFromValidatorAPI(setupEnv *setup.Setup, txResponce hedera.TransactionResponse, tokenID string, expectedSendAmount int64, t *testing.T) (*service.TransferData, *common.Address) {
	tokenAddress, err := setup.ParseHederaToETHToken(setupEnv.Clients.RouterContract, setupEnv.TokenID.String())
	if err != nil {
		t.Fatalf("Expecting Token [%s] is not supported. - Error: [%s]", tokenID, err)
	}

	transactionData, err := setupEnv.Clients.ValidatorClient.GetTransferData(fromHederaTransactionID(&txResponce.TransactionID).String())
	if err != nil {
		t.Fatalf("Cannot fetch transaction data - Error: [%s].", err)
	}
	if transactionData.Amount != fmt.Sprint(expectedSendAmount) {
		t.Fatalf("Transaction data mismatch: Expected [%d], but was [%s]", expectedSendAmount, transactionData.Amount)
	}
	if transactionData.NativeToken != tokenID {
		t.Fatalf("Native Token mismatch: Expected [%s], but was [%s]", setupEnv.TokenID.String(), transactionData.NativeToken)
	}
	if transactionData.Recipient != receiverAddress {
		t.Fatalf("Receiver address mismatch: Expected [%s], but was [%s]", receiverAddress, transactionData.Recipient)
	}
	if transactionData.WrappedToken != tokenAddress.String() {
		t.Fatalf("Token address mismatch: Expected [%s], but was [%s]", tokenAddress.String(), transactionData.WrappedToken)
	}

	return transactionData, tokenAddress
}

func verifyDatabaseRecords(dbValidation *database.Service, expectedRecord *entity.Transfer, signatures []string, t *testing.T) {
	exist, err := dbValidation.VerifyDatabaseRecords(expectedRecord, signatures)
	if err != nil {
		t.Fatalf("[%s] - Verification of database records failed - Error: [%s].", expectedRecord.TransactionID, err)
	}
	if !exist {
		t.Fatalf("[%s] - Database does not contain expected records", expectedRecord.TransactionID)
	}
}

func prepareExpectedTransfer(routerContract *routerContract.Router, transactionID hedera.TransactionID, nativeToken, amount string, statuses database.ExpectedStatuses, t *testing.T) *entity.Transfer {
	expectedTxId := hederahelper.FromHederaTransactionID(&transactionID)

	wrappedToken, err := setup.ParseHederaToETHToken(routerContract, nativeToken)
	if err != nil {
		t.Fatalf("Expecting Token [%s] is not supported. - Error: [%s]", nativeToken, err)
	}
	return &entity.Transfer{
		TransactionID:      expectedTxId.String(),
		Receiver:           receiverAddress,
		NativeToken:        nativeToken,
		WrappedToken:       wrappedToken.String(),
		Amount:             amount,
		Status:             statuses.Status,
		SignatureMsgStatus: statuses.StatusSignature,
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
	receiverBalance, err := hedera.NewAccountBalanceQuery().
		SetAccountID(setup.BridgeAccount).
		Execute(setup.Clients.Hedera)
	if err != nil {
		t.Fatalf("Unable to query the balance of the Bridge Account, Error: [%s]", err)
	}

	fmt.Println(fmt.Sprintf("Bridge account balance HBAR balance before transaction: [%d]", receiverBalance.Hbars.AsTinybar()))

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
	receiverBalanceNew, err := hedera.NewAccountBalanceQuery().
		SetAccountID(setup.BridgeAccount).
		Execute(setup.Clients.Hedera)
	if err != nil {
		t.Fatalf("Unable to query the balance of the Bridge Account. Error: [%s]", err)
	}

	fmt.Println(fmt.Sprintf("Bridge Account HBAR balance after transaction: [%d]", receiverBalanceNew.Hbars.AsTinybar()))

	// Verify that the custodial address has received exactly the amount sent
	amount := receiverBalanceNew.Hbars.AsTinybar() - receiverBalance.Hbars.AsTinybar()
	// Verify that the bridge account has received exactly the amount sent
	if amount != hBarSendAmount.AsTinybar() {
		t.Fatalf("Expected to receive the exact transfer amount of hbar: [%v], but was [%v]", hBarSendAmount.AsTinybar(), amount)
	}

	return *transactionResponse, whbarBalanceBefore
}

func verifyTokenTransferToBridgeAccount(setup *setup.Setup, memo string, wTokenReceiverAddress common.Address, t *testing.T) (hedera.TransactionResponse, *big.Int) {
	// Get the wrapped hts token balance of the receiver before the transfer
	wrappedTokenBalanceBefore, err := setup.Clients.WTokenContract.BalanceOf(&bind.CallOpts{}, wTokenReceiverAddress)
	if err != nil {
		t.Fatalf("Unable to query the token balance of the receiver account. Error: [%s]", err)
	}

	fmt.Println(fmt.Sprintf("Token balance before transaction: [%s]", wrappedTokenBalanceBefore))
	// Get bridge account token balance before transfer
	receiverBalance, err := hedera.NewAccountBalanceQuery().
		SetAccountID(setup.BridgeAccount).
		Execute(setup.Clients.Hedera)
	if err != nil {
		t.Fatalf("Unable to query the token balance of the Bridge Account, Error: [%s]", err)
	}
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
	receiverBalanceNew, err := hedera.NewAccountBalanceQuery().
		SetAccountID(setup.BridgeAccount).
		Execute(setup.Clients.Hedera)
	if err != nil {
		t.Fatalf("Unable to query the token balance of the Bridge Account, Error: [%s]", err)
	}

	fmt.Println(fmt.Sprintf("Bridge Account Token balance after transaction: [%d]", receiverBalanceNew.Token[setup.TokenID]))

	// Verify that the custodial address has received exactly the amount sent
	amount := receiverBalanceNew.Token[setup.TokenID] - receiverBalance.Token[setup.TokenID]
	// Verify that the bridge account has received exactly the amount sent
	if amount != uint64(tokensSendAmount) {
		t.Fatalf("Expected to receive the exact transfer amount of hbar: [%v], but received: [%v]", hBarSendAmount.AsTinybar(), amount)
	}

	return *transactionResponse, wrappedTokenBalanceBefore
}

func sendHbarsToBridgeAccount(setup *setup.Setup, memo string) (*hedera.TransactionResponse, error) {
	fmt.Println(fmt.Sprintf("Sending [%v] Hbars through the Bridge. Transaction Memo: [%s]", hBarSendAmount, memo))

	res, err := hedera.NewTransferTransaction().
		AddHbarTransfer(setup.SenderAccount, hbarRemovalAmount).
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
	fmt.Println(fmt.Sprintf("Sending [%v] Tokens to the Bridge. Transaction Memo: [%s]", tokensSendAmount, memo))

	res, err := hedera.NewTransferTransaction().
		SetTransactionMemo(memo).
		AddTokenTransfer(setup.TokenID, setup.SenderAccount, -int64(tokensSendAmount)).
		AddTokenTransfer(setup.TokenID, setup.BridgeAccount, int64(tokensSendAmount)).
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

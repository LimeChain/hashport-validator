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
	"errors"
	"fmt"
	"log"
	"math/big"
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
	incrementFloat, _            = new(big.Int).SetString("1", 10)
	amount               float64 = 400
	hBarSendAmount               = hedera.HbarFrom(amount, "hbar")
	tokensSendAmount             = 1000000000
	hbarRemovalAmount            = hedera.HbarFrom(-amount, "hbar")
	precision                    = new(big.Int).SetInt64(100000)
	whbarReceiverAddress         = common.HexToAddress(receiverAddress)
)

const (
	receiverAddress = "0x7cFae2deF15dF86CfdA9f2d25A361f1123F42eDD"
	gasPriceGwei    = "100"
	gasPrice        = "100000000000" // the Gas Price from above converted to WEI

	expectedValidatorsCount = 3
)

func Test_HBAR(t *testing.T) {
	setupEnv := setup.Load()

	metadataResponse, err := setupEnv.Clients.ValidatorClient.GetMetadata(gasPriceGwei)
	if err != nil {
		t.Fatal(err)
	}

	txFee, err := txFeeToBigInt(metadataResponse.TransactionFee)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Estimated TX TxReimbursementFee: [%s]\n", txFee)

	memo := fmt.Sprintf("%s-%s-%s", receiverAddress, txFee, gasPriceGwei)

	serviceFeePercentage, err := setupEnv.Clients.RouterContract.ServiceFee(nil)
	if err != nil {
		t.Fatal(err)
	}

	expectedWHbarAmount, err := calculateWHBarAmount(txFee, serviceFeePercentage)
	if err != nil {
		t.Fatal(err)
	}

	// Step 1 - Verify the transfer of Hbars to the Bridge Account
	transactionResponse, whbarBalanceBefore := verifyTransferToBridgeAccount(setupEnv, memo, whbarReceiverAddress, t)

	// Step 2 - Verify the submitted topic messages
	ethTransactionHash, receivedSignatures := verifyTopicMessages(setupEnv, transactionResponse, 1, t)

	// Step 3 - Verify the Ethereum Transaction execution
	verifyEthereumTXExecution(setupEnv, ethTransactionHash, whbarReceiverAddress, expectedWHbarAmount.Int64(), whbarBalanceBefore, t)

	expectedTxRecord := prepareExpectedTransfer(
		setupEnv.Clients.RouterContract,
		transactionResponse.TransactionID,
		"HBAR",
		txFee,
		gasPrice,
		database.ExpectedStatuses{
			Status:          entity_transfer.StatusCompleted,
			StatusSignature: entity_transfer.StatusSignatureMined,
			StatusEthTx:     entity_transfer.StatusEthTxMined,
			StatusEthTxMsg:  entity_transfer.StatusEthTxMsgMined,
		},
		ethTransactionHash,
		true, t)

	// Step 4 - Verify Database Records
	verifyDatabaseRecords(setupEnv.DbValidation, expectedTxRecord, receivedSignatures, t)
}

func Test_HBAR_No_Ethereum_TX_Submission(t *testing.T) {
	setupEnv := setup.Load()

	memo := fmt.Sprintf("%s-0-0", receiverAddress)

	// Step 1 - Verify the transfer of HTS Token to the Bridge Account
	transactionResponse, _ := verifyTransferToBridgeAccount(setupEnv, memo, whbarReceiverAddress, t)

	// Step 2 - Verify the submitted topic messages
	ethTransactionHash, receivedSignatures := verifyTopicMessages(setupEnv, transactionResponse, 0, t)

	expectedTxRecord := prepareExpectedTransfer(
		setupEnv.Clients.RouterContract,
		transactionResponse.TransactionID,
		"HBAR",
		"0",
		"0",
		database.ExpectedStatuses{
			Status:          entity_transfer.StatusCompleted,
			StatusSignature: entity_transfer.StatusSignatureMined,
		},
		ethTransactionHash,
		false, t)

	// Step 4 - Verify Database Records
	verifyDatabaseRecords(setupEnv.DbValidation, expectedTxRecord, receivedSignatures, t)
}

func Test_E2E_Token_Transfer(t *testing.T) {
	setupEnv := setup.Load()

	memo := fmt.Sprintf("%s-0-0", receiverAddress)

	wTokenReceiverAddress := common.HexToAddress(receiverAddress)

	// Step 1 - Verify the transfer of HTS to the Bridge Account
	transactionResponse, wrappedTokenBalanceBefore := verifyTokenTransferToBridgeAccount(setupEnv, memo, wTokenReceiverAddress, t)

	// Step 2 - Verify the submitted topic messages
	_, receivedSignatures := verifyTopicMessages(setupEnv, transactionResponse, 0, t)

	// Step 3 - Verify Transfer retrieved from Validator API
	transactionData, tokenAddress := verifyTransferFromValidatorAPI(setupEnv, transactionResponse, receivedSignatures, t)

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
		"0",
		"0",
		database.ExpectedStatuses{
			Status:          entity_transfer.StatusCompleted,
			StatusSignature: entity_transfer.StatusSignatureMined,
		},
		txHash,
		false, t)
	verifyDatabaseRecords(setupEnv.DbValidation, expectedTxRecord, receivedSignatures, t)
}

func submitMintTransaction(setupEnv *setup.Setup, transactionResponse hedera.TransactionResponse, transactionData *service.TransferData, tokenAddress *common.Address, t *testing.T) string {
	var signatures [][]byte
	for i := 0; i < len(transactionData.Signatures); i++ {
		signature, _ := hex.DecodeString(transactionData.Signatures[i])
		signatures = append(signatures, signature)
	}

	res, err := setupEnv.Clients.RouterContract.Mint(
		setupEnv.Clients.KeyTransactor,
		[]byte(fromHederaTransactionID(&transactionResponse.TransactionID).String()),
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

func waitForTransaction(setupEnv *setup.Setup, txHash string, t *testing.T) {
	fmt.Printf("Waiting for transaction: [%s] to be mined\n", txHash)
	c1 := make(chan bool, 1)
	onSuccess := func() {
		fmt.Printf("Transaction [%s] mined successfully\n", txHash)
		c1 <- true
	}
	onRevert := func() {
		t.Fatalf(`Failed to mine successfully ethereum transaction: [%s]`, txHash)
	}
	onError := func(err error) {
		fmt.Println(fmt.Sprintf(`Transaction unsuccessful, Error: [%s]`, err))
		t.Fatal(err)
	}
	setupEnv.Clients.EthClient.WaitForTransaction(txHash, onSuccess, onRevert, onError)
	<-c1
}

func validateTokenBalance(setupEnv *setup.Setup, wrappedTokenBalanceBefore *big.Int, wTokenReceiverAddress common.Address, t *testing.T) {
	wrappedTokenBalanceAfter, err := setupEnv.Clients.WTokenContract.BalanceOf(&bind.CallOpts{}, wTokenReceiverAddress)
	if err != nil {
		t.Fatal(err)
	}

	tokensAmount := new(big.Int).SetInt64(int64(tokensSendAmount))

	serviceFeePercentage, err := setupEnv.Clients.RouterContract.ServiceFee(nil)
	if err != nil {
		t.Fatal(err)
	}

	txFee := new(big.Int).Mul(tokensAmount, serviceFeePercentage)
	txFee = new(big.Int).Div(txFee, precision)

	expectedBalance := new(big.Int).Sub(wrappedTokenBalanceAfter, wrappedTokenBalanceBefore)
	mintAmount := new(big.Int).Sub(tokensAmount, txFee)

	if expectedBalance.Cmp(mintAmount) != 0 {
		t.Fatalf("Incorect token balance")
	}
}

func verifyTransferFromValidatorAPI(setupEnv *setup.Setup, txResponce hedera.TransactionResponse, signatures []string, t *testing.T) (*service.TransferData, *common.Address) {
	tokenAddress, _ := setup.ParseToken(setupEnv.Clients.RouterContract, setupEnv.TokenID.String())

	transactionData, err := setupEnv.Clients.ValidatorClient.GetTransferData(fromHederaTransactionID(&txResponce.TransactionID).String())
	if err != nil {
		t.Fatalf("Cannot fetch transaction data - Error: [%s].", err)
	}
	if transactionData.Amount != fmt.Sprint(tokensSendAmount) {
		t.Fatal("Transaction data mismatch")
	}
	if transactionData.NativeToken != setupEnv.TokenID.String() {
		t.Fatal("Native Token mismatch")
	}
	if transactionData.Recipient != receiverAddress {
		t.Fatal("Receiver address mismatch")
	}
	if transactionData.WrappedToken != tokenAddress.String() {
		t.Fatal("Token address mismatch")
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

func prepareExpectedTransfer(routerContract *routerContract.Router, transactionID hedera.TransactionID, nativeToken, txFee, gasPriceWei string, statuses database.ExpectedStatuses, ethTransactionHash string, shouldExecuteEthTx bool, t *testing.T) *entity.Transfer {
	expectedTxId := fromHederaTransactionID(&transactionID)

	wrappedToken, err := setup.ParseToken(routerContract, nativeToken)
	if err != nil {
		t.Fatalf("Expecting Token [%s] is not supported. - Error: [%s]", expectedTxId, err)
	}

	amount := strconv.FormatInt(hBarSendAmount.AsTinybar(), 10)

	return &entity.Transfer{
		TransactionID:         expectedTxId.String(),
		Receiver:              receiverAddress,
		NativeToken:           nativeToken,
		WrappedToken:          wrappedToken.String(),
		Amount:                amount,
		TxReimbursement:       txFee,
		GasPrice:              gasPriceWei,
		Status:                statuses.Status,
		SignatureMsgStatus:    statuses.StatusSignature,
		EthTxMsgStatus:        statuses.StatusEthTxMsg,
		EthTxStatus:           statuses.StatusEthTx,
		EthTxHash:             ethTransactionHash,
		ExecuteEthTransaction: shouldExecuteEthTx,
	}
}

func calculateWHBarAmount(txFee string, percentage *big.Int) (*big.Int, error) {
	bnTxFee, ok := new(big.Int).SetString(txFee, 10)
	if !ok {
		return nil, errors.New(fmt.Sprintf("could not parse txn fee [%s].", txFee))
	}

	bnHbarAmount := new(big.Int).SetInt64(hBarSendAmount.AsTinybar())
	bnAmount := new(big.Int).Sub(bnHbarAmount, bnTxFee)
	serviceFee := new(big.Int).Mul(bnAmount, percentage)
	precisionedServiceFee := new(big.Int).Div(serviceFee, precision)

	return new(big.Int).Sub(bnAmount, precisionedServiceFee), nil
}

func txFeeToBigInt(transactionFee string) (string, error) {
	amount := new(big.Float)
	amount, ok := amount.SetString(transactionFee)
	if !ok {
		return "", errors.New(fmt.Sprintf("Cannot parse amount value [%s] to big.Float", transactionFee))
	}

	bnAmount := new(big.Int)
	bnAmount, accuracy := amount.Int(bnAmount)
	if accuracy == big.Below {
		bnAmount = bnAmount.Add(bnAmount, incrementFloat)
	}

	return bnAmount.String(), nil
}

func verifyTransferToBridgeAccount(setup *setup.Setup, memo string, whbarReceiverAddress common.Address, t *testing.T) (hedera.TransactionResponse, *big.Int) {
	// Get the wrapped hbar balance of the receiver before the transfer
	whbarBalanceBefore, err := setup.Clients.WHbarContract.BalanceOf(&bind.CallOpts{}, whbarReceiverAddress)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("WHBAR balance before transaction: [%s]\n", whbarBalanceBefore)
	// Get bridge account hbar balance before transfer
	receiverBalance, err := hedera.NewAccountBalanceQuery().
		SetAccountID(setup.BridgeAccount).
		Execute(setup.Clients.Hedera)
	if err != nil {
		fmt.Println(`Unable to query the balance of the Bridge Account`)
		t.Fatal(err)
	}

	fmt.Println(fmt.Sprintf(`Bridge account balance HBAR balance before transaction: [%d]`, receiverBalance.Hbars.AsTinybar()))

	// Get the transaction receipt to verify the transaction was executed
	transactionResponse, err := sendHbarsToBridgeAccount(setup, memo)
	if err != nil {
		fmt.Println(fmt.Sprintf(`Unable to send HBARs to Bridge Account, Error: [%s]`, err))
		t.Fatal(err)
	}
	transactionReceipt, err := transactionResponse.GetReceipt(setup.Clients.Hedera)
	if err != nil {
		fmt.Println(fmt.Sprintf(`Transaction unsuccessful, Error: [%s]`, err))
		t.Fatal(err)
	}

	fmt.Println(fmt.Sprintf(`Successfully sent HBAR to bridge account, Status: [%s]`, transactionReceipt.Status))

	// Get bridge account hbar balance after transfer
	receiverBalanceNew, err := hedera.NewAccountBalanceQuery().
		SetAccountID(setup.BridgeAccount).
		Execute(setup.Clients.Hedera)
	if err != nil {
		fmt.Println("Unable to query the balance of the Bridge Account")
		t.Fatal(err)
	}

	fmt.Println(fmt.Sprintf(`Bridge Account HBAR balance after transaction: [%d]`, receiverBalanceNew.Hbars.AsTinybar()))

	// Verify that the custodial address has received exactly the amount sent
	amount := receiverBalanceNew.Hbars.AsTinybar() - receiverBalance.Hbars.AsTinybar()
	// Verify that the bridge account has received exactly the amount sent
	if amount != hBarSendAmount.AsTinybar() {
		t.Fatalf(`Expected to recieve the exact transfer amount of hbar: [%v]`, hBarSendAmount.AsTinybar())
	}

	return *transactionResponse, whbarBalanceBefore
}

func verifyTokenTransferToBridgeAccount(setup *setup.Setup, memo string, wTokenReceiverAddress common.Address, t *testing.T) (hedera.TransactionResponse, *big.Int) {
	// Get the wrapped hts token balance of the receiver before the transfer
	wrappedTokenBalanceBefore, err := setup.Clients.WTokenContract.BalanceOf(&bind.CallOpts{}, wTokenReceiverAddress)
	if err != nil {
		fmt.Println(`Unable to query the token balance of the  receiver account`)
		t.Fatal(err)
	}

	fmt.Printf("Token balance before transaction: [%s]\n", wrappedTokenBalanceBefore)
	// Get bridge account token balance before transfer
	receiverBalance, err := hedera.NewAccountBalanceQuery().
		SetAccountID(setup.BridgeAccount).
		Execute(setup.Clients.Hedera)
	if err != nil {
		fmt.Println(`Unable to query the token balance of the Bridge Account`)
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf(`Bridge account Token balance before transaction: [%d]`, receiverBalance.Token[setup.TokenID]))

	// Get the transaction receipt to verify the transaction was executed
	transactionResponse, err := sendTokensToBridgeAccount(setup, memo)
	if err != nil {
		fmt.Println(fmt.Sprintf(`Unable to send Tokens to Bridge Account, Error: [%s]`, err))
		t.Fatal(err)
	}
	transactionReceipt, err := transactionResponse.GetReceipt(setup.Clients.Hedera)
	if err != nil {
		fmt.Println(fmt.Sprintf(`Transaction unsuccessful, Error: [%s]`, err))
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf(`Successfully sent Tokens to bridge account, Status: [%s]`, transactionReceipt.Status))

	// Get bridge account HTS token balance after transfer
	receiverBalanceNew, err := hedera.NewAccountBalanceQuery().
		SetAccountID(setup.BridgeAccount).
		Execute(setup.Clients.Hedera)
	if err != nil {
		fmt.Println(`Unable to query the token balance of the Bridge Account`)
		t.Fatal(err)
	}

	fmt.Println(fmt.Sprintf(`Bridge Account Token balance after transaction: [%d]`, receiverBalanceNew.Token[setup.TokenID]))

	// Verify that the custodial address has received exactly the amount sent
	amount := receiverBalanceNew.Token[setup.TokenID] - receiverBalance.Token[setup.TokenID]
	// Verify that the bridge account has received exactly the amount sent
	if amount != uint64(tokensSendAmount) {
		t.Fatalf(`Expected to recieve the exact transfer amount of hbar: [%v]`, hBarSendAmount.AsTinybar())
	}

	return *transactionResponse, wrappedTokenBalanceBefore
}

func sendHbarsToBridgeAccount(setup *setup.Setup, memo string) (*hedera.TransactionResponse, error) {
	fmt.Println(fmt.Sprintf(`Sending [%v] Hbars through the Bridge. Transaction Memo: [%s]`, hBarSendAmount, memo))

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

	fmt.Println(fmt.Sprintf(`TX broadcasted. ID [%s], Status: [%s]`, res.TransactionID, rec.Status))
	time.Sleep(1 * time.Second)

	return &res, err
}

func sendTokensToBridgeAccount(setup *setup.Setup, memo string) (*hedera.TransactionResponse, error) {
	fmt.Println(fmt.Sprintf(`Sending [%v] Tokens to the Bridge. Transaction Memo: [%s]`, tokensSendAmount, memo))

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

	fmt.Println(fmt.Sprintf(`TX broadcasted. ID [%s], Status: [%s]`, res.TransactionID, rec.Status))
	time.Sleep(1 * time.Second)

	return &res, err
}

func verifyTopicMessages(setup *setup.Setup, transactionResponse hedera.TransactionResponse, expectedEthTxMessageCount int, t *testing.T) (string, []string) {
	ethSignaturesCollected := 0
	ethTransMsgCollected := 0
	ethTransactionHash := ""
	var receivedSignatures []string

	fmt.Println(fmt.Sprintf(`Waiting for Signatures & TX Hash to be published to Topic [%v]`, setup.TopicID.String()))

	// Subscribe to Topic
	_, err := hedera.NewTopicMessageQuery().
		SetStartTime(time.Unix(0, time.Now().UnixNano())).
		SetTopicID(setup.TopicID).
		Subscribe(
			setup.Clients.Hedera,
			func(response hedera.TopicMessage) {
				msg := &validatorproto.TopicMessage{}
				err := proto.Unmarshal(response.Contents, msg)
				if err != nil {
					t.Fatal(err)
				}

				if msg.GetType() == validatorproto.TopicMessageType_EthSignature {
					//Verify that all the submitted messages have signed the same transaction
					topicSubmissionMessageSign := fromHederaTransactionID(&transactionResponse.TransactionID)
					if msg.GetTopicSignatureMessage().TransferID != topicSubmissionMessageSign.String() {
						fmt.Println(fmt.Sprintf(`Expected signature message to contain the transaction id: [%s]`, topicSubmissionMessageSign.String()))
					} else {
						receivedSignatures = append(receivedSignatures, msg.GetTopicSignatureMessage().Signature)
						ethSignaturesCollected++
						fmt.Println(fmt.Sprintf("Received Auth Signature [%s]", msg.GetTopicSignatureMessage().Signature))
					}
				}

				if msg.GetType() == validatorproto.TopicMessageType_EthTransaction {
					//Verify that the eth transaction message has been submitted
					topicSubmissionMessageTrans := fromHederaTransactionID(&transactionResponse.TransactionID)
					if msg.GetTopicEthTransactionMessage().TransferID != topicSubmissionMessageTrans.String() {
						t.Fatalf(`Expected ethereum transaction message to contain the transaction id: [%s]`, topicSubmissionMessageTrans.String())
					}
					ethTransactionHash = msg.GetTopicEthTransactionMessage().GetEthTxHash()
					ethTransMsgCollected++
					fmt.Println(fmt.Sprintf("Received Ethereum Transaction Hash [%s]", msg.GetTopicEthTransactionMessage().EthTxHash))
				}
			},
		)
	if err != nil {
		t.Fatalf(`Unable to subscribe to Topic [%s]`, setup.TopicID)
	}

	select {
	case <-time.After(60 * time.Second):
		if ethSignaturesCollected != expectedValidatorsCount {
			t.Fatalf(`Expected the count of collected signatures to equal the number of validators: [%v], but was: [%v]`, expectedValidatorsCount, ethSignaturesCollected)
		}
		if ethTransMsgCollected != expectedEthTxMessageCount {
			t.Fatalf(`Expected to submit exactly [%v] ethereum transaction in topic, but was: [%v]`, expectedEthTxMessageCount, ethTransMsgCollected)
		}
		return ethTransactionHash, receivedSignatures
	}
	// Not possible end-case
	return "", nil
}

func verifyEthereumTXExecution(setup *setup.Setup, ethTransactionHash string, whbarReceiverAddress common.Address, expectedWHBarAmount int64, whbarBalanceBefore *big.Int, t *testing.T) {
	fmt.Printf("Waiting for transaction [%s] to succeed...\n", ethTransactionHash)
	// Make a blocking channel waiting for Ethereum TX success
	c1 := make(chan bool, 1)
	onSuccess := func() {
		fmt.Printf("Transaction [%s] mined successfully\n", ethTransactionHash)
		// Get the wrapped hbar balance of the receiver after the transfer
		whbarBalanceAfter, err := setup.Clients.WHbarContract.BalanceOf(&bind.CallOpts{}, whbarReceiverAddress)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("WHBAR balance after transaction: [%s]\n", whbarBalanceAfter)
		// Verify that the ethereum address has received the exact transfer amount of WHBARs
		amount := new(big.Int).Sub(whbarBalanceAfter, whbarBalanceBefore)
		if strings.Compare(amount.String(), strconv.FormatInt(expectedWHBarAmount, 10)) != 0 {
			t.Fatalf(`Expected to receive [%v] WHBAR, but got [%v].`, expectedWHBarAmount, amount)
		}
		c1 <- true
	}
	onRevert := func() {
		t.Fatalf(`Expected to mine successfully the broadcasted ethereum transaction: [%s]`, ethTransactionHash)
	}

	onError := func(err error) {
		if err != nil {
			fmt.Println(fmt.Sprintf(`Transaction unsuccessful, Error: [%s]`, err))
			t.Fatal(err)
		}
	}
	setup.Clients.EthClient.WaitForTransaction(ethTransactionHash, onSuccess, onRevert, onError)
	<-c1
}

type hederaTxId struct {
	AccountId string
	Seconds   string
	Nanos     string
}

func fromHederaTransactionID(id *hedera.TransactionID) hederaTxId {
	stringTxId := id.String()
	split := strings.Split(stringTxId, "@")
	accId := split[0]

	split = strings.Split(split[1], ".")

	return hederaTxId{
		AccountId: accId,
		Seconds:   fmt.Sprintf("%09s", split[0]),
		Nanos:     fmt.Sprintf("%09s", split[1]),
	}
}

func (txId hederaTxId) String() string {
	return fmt.Sprintf("%s-%s-%s", txId.AccountId, txId.Seconds, txId.Nanos)
}

func (txId hederaTxId) Timestamp() string {
	return fmt.Sprintf("%s.%s", txId.Seconds, txId.Nanos)
}

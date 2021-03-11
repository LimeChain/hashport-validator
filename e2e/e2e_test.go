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
	"errors"
	"fmt"
	"github.com/limechain/hedera-eth-bridge-validator/e2e/setup"
	"log"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashgraph/hedera-sdk-go"
	tx "github.com/limechain/hedera-eth-bridge-validator/app/process/model/transaction"
	validatorproto "github.com/limechain/hedera-eth-bridge-validator/proto"
	"google.golang.org/protobuf/proto"
)

var (
	incrementFloat, _    = new(big.Int).SetString("1", 10)
	hBarAmount           = hedera.HbarFrom(400, "hbar")
	precision            = new(big.Int).SetInt64(100000)
	whbarReceiverAddress = common.HexToAddress(receiverAddress)
)

const (
	receiverAddress         = "0x7cFae2deF15dF86CfdA9f2d25A361f1123F42eDD"
	gasPriceGwei            = "100"
	expectedValidatorsCount = 3
)

func Test_E2E(t *testing.T) {
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

	serviceFeePercentage, err := setupEnv.Clients.BridgeContract.ServiceFee(nil)
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
	ethTransactionHash := verifyTopicMessages(setupEnv, transactionResponse, t)

	// Step 3 - Verify the Ethereum Transaction execution
	verifyEthereumTXExecution(setupEnv, ethTransactionHash, whbarReceiverAddress, expectedWHbarAmount.Int64(), whbarBalanceBefore, t)

}

func calculateWHBarAmount(txFee string, percentage *big.Int) (*big.Int, error) {
	bnTxFee, ok := new(big.Int).SetString(txFee, 10)
	if !ok {
		return nil, errors.New(fmt.Sprintf("could not parse txn fee [%s].", txFee))
	}

	bnHbarAmount := new(big.Int).SetInt64(hBarAmount.AsTinybar())
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
		log.Fatal(err)
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

	// Verify that the custodial address has receive exactly the amount sent
	amount := receiverBalanceNew.Hbars.AsTinybar() - receiverBalance.Hbars.AsTinybar()
	// Verify that the bridge account has received exactly the amount sent
	if amount != hBarAmount.AsTinybar() {
		t.Fatalf(`Expected to recieve the exact transfer amount of hbar: [%v]`, hBarAmount.AsTinybar())
	}

	return *transactionResponse, whbarBalanceBefore
}

func sendHbarsToBridgeAccount(setup *setup.Setup, memo string) (*hedera.TransactionResponse, error) {
	fmt.Println(fmt.Sprintf(`Sending [%v] Hbars through the Bridge. Transaction Memo: [%s]`, hBarAmount, memo))

	res, err := hedera.NewTransferTransaction().AddHbarSender(setup.SenderAccount, hBarAmount).
		AddHbarRecipient(setup.BridgeAccount, hBarAmount).
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

func verifyTopicMessages(setup *setup.Setup, transactionResponse hedera.TransactionResponse, t *testing.T) string {
	ethSignaturesCollected := 0
	ethTransMsgCollected := 0
	ethTransactionHash := ""

	fmt.Println(fmt.Sprintf(`Waiting for Signatures & TX Hash to be published to Topic [%v]`, setup.TopicID.String()))

	c1 := make(chan bool, 1)
	// Subscribe to Topic
	_, err := hedera.NewTopicMessageQuery().
		SetStartTime(time.Unix(0, time.Now().UnixNano())).
		SetTopicID(setup.TopicID).
		Subscribe(
			setup.Clients.Hedera,
			func(response hedera.TopicMessage) {
				msg := &validatorproto.TopicSubmissionMessage{}
				err := proto.Unmarshal(response.Contents, msg)
				if err != nil {
					t.Fatal(err)
				}

				if msg.GetType() == validatorproto.TopicSubmissionType_EthSignature {
					//Verify that all the submitted messages have signed the same transaction
					topicSubmissionMessageSign := tx.FromHederaTransactionID(&transactionResponse.TransactionID)
					if msg.GetTopicSignatureMessage().TransactionId != topicSubmissionMessageSign.String() {
						t.Fatalf(`Expected signature message to contain the transaction id: [%s]`, topicSubmissionMessageSign.String())
					}
					ethSignaturesCollected++
					fmt.Println(fmt.Sprintf("Received Auth Signature [%s]", msg.GetTopicSignatureMessage().Signature))
				}

				if msg.GetType() == validatorproto.TopicSubmissionType_EthTransaction {
					//Verify that the eth transaction message has been submitted
					topicSubmissionMessageTrans := tx.FromHederaTransactionID(&transactionResponse.TransactionID)
					if msg.GetTopicEthTransactionMessage().TransactionId != topicSubmissionMessageTrans.String() {
						t.Fatalf(`Expected ethereum transaction message to contain the transaction id: [%s]`, topicSubmissionMessageTrans.String())
					}
					ethTransactionHash = msg.GetTopicEthTransactionMessage().GetEthTxHash()
					ethTransMsgCollected++
					fmt.Println(fmt.Sprintf("Received Ethereum Transaction Hash [%s]", msg.GetTopicEthTransactionMessage().EthTxHash))
				}

				// Check whether we collected everything
				if expectedValidatorsCount == ethSignaturesCollected && ethTransMsgCollected == 1 {
					c1 <- true
				}
			},
		)
	if err != nil {
		t.Fatalf(`Unable to subscribe to Topic [%s]`, setup.TopicID)
	}

	select {
	case _ = <-c1:
		return ethTransactionHash
	case <-time.After(60 * time.Second):
		if ethSignaturesCollected != expectedValidatorsCount {
			t.Fatalf(`Expected the count of collected signatures to equal the number of validators: [%v], but was: [%v]`, expectedValidatorsCount, ethSignaturesCollected)
		}
		if ethTransMsgCollected != 1 {
			t.Fatal(`Expected to submit exactly 1 ethereum transaction in topic`)
		}
	}
	// Not possible end-case
	return ""
}

func verifyEthereumTXExecution(setup *setup.Setup, ethTransactionHash string, whbarReceiverAddress common.Address, expectedWHBarAmount int64, whbarBalanceBefore *big.Int, t *testing.T) {
	fmt.Printf("Waiting for transaction [%s] to succeed...\n", ethTransactionHash)

	success, err := setup.Clients.EthClient.WaitForTransaction(common.HexToHash(ethTransactionHash))

	// Verify that the eth transaction has been mined and succeeded
	if success == false {
		t.Fatalf(`Expected to mine successfully the broadcasted ethereum transaction: [%s]`, ethTransactionHash)
	}

	if err != nil {
		fmt.Println(fmt.Sprintf(`Transaction unsuccessful, Error: [%s]`, err))
		t.Fatal(err)
	}

	fmt.Printf("Transaction [%s] mined successfully\n", ethTransactionHash)

	// Get the wrapped hbar balance of the receiver after the transfer
	whbarBalanceAfter, err := setup.Clients.WHbarContract.BalanceOf(&bind.CallOpts{}, whbarReceiverAddress)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("WHBAR balance after transaction: [%s]\n", whbarBalanceAfter)

	// Verify that the ethereum address has received the exact transfer amount of WHBARs
	amount := whbarBalanceAfter.Int64() - whbarBalanceBefore.Int64()
	if amount != expectedWHBarAmount {
		t.Fatalf(`Expected to receive [%v] WHBAR, but got [%v].`, expectedWHBarAmount, amount)
	}
}

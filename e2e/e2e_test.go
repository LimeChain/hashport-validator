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
	"fmt"
	"github.com/limechain/hedera-eth-bridge-validator/e2e/config"
	"log"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	common "github.com/ethereum/go-ethereum/common"
	"github.com/hashgraph/hedera-sdk-go"
	tx "github.com/limechain/hedera-eth-bridge-validator/app/process/model/transaction"
	validatorproto "github.com/limechain/hedera-eth-bridge-validator/proto"
	"google.golang.org/protobuf/proto"
)

func Test_E2E(t *testing.T) {
	setup := config.LoadE2EConfig()

	memo := "0x7cFae2deF15dF86CfdA9f2d25A361f1123F42eDD-600000000-1"
	whbarReceiverAddress := common.HexToAddress("0x7cFae2deF15dF86CfdA9f2d25A361f1123F42eDD")

	hBarAmount := float64(10)
	validatorsCount := 3

	// Step 1 - Verify the transfer of Hbars to the Bridge Account
	transactionResponse, whbarBalanceBefore := verifyTransferToBridgeAccount(setup, memo, hBarAmount, whbarReceiverAddress, t)
	// Step 2 - Verify the submitted topic messages
	ethTransactionHash := verifyTopicMessages(setup, transactionResponse, validatorsCount, t)
	// Step 3 - Verify the Ethereum Transaction execution
	verifyEthereumTXExecution(setup, ethTransactionHash, whbarReceiverAddress, hBarAmount, whbarBalanceBefore, t)

}

func verifyTransferToBridgeAccount(setup *config.Setup, memo string, hBarAmount float64, whbarReceiverAddress common.Address, t *testing.T) (hedera.TransactionResponse, *big.Int) {
	// Get the wrapped hbar balance of the receiver before the transfer
	whbarBalanceBefore, err := setup.Clients.WHbarContract.BalanceOf(&bind.CallOpts{}, whbarReceiverAddress)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("WHBAR balance before transaction: [%s]\n", whbarBalanceBefore)
	// Get custodian hbar balance before transfer
	receiverBalance, err := hedera.NewAccountBalanceQuery().
		SetAccountID(setup.BridgeAccount).
		Execute(setup.Clients.Hedera)
	if err != nil {
		fmt.Println(`Unable to query the balance of the Bridge Account`)
		t.Fatal(err)
	}

	fmt.Println(fmt.Sprintf(`HBAR custodian balance before transaction: [%d]`, receiverBalance.Hbars.AsTinybar()))

	// Get the transaction receipt to verify the transaction was executed
	transactionResponse, err := sendHbarsToBridgeAccount(setup.SenderAccount, setup.BridgeAccount, memo, hBarAmount, setup.Clients.Hedera)
	if err != nil {
		fmt.Println(fmt.Sprintf(`Unable to send HBARs to Bridge Account, Error: [%s]`, err))
		t.Fatal(err)
	}
	transactionReceipt, err := transactionResponse.GetReceipt(setup.Clients.Hedera)

	if err != nil {
		fmt.Println(fmt.Sprintf(`Transaction unsuccessful, Error: [%s]`, err))
		t.Fatal(err)
	}

	fmt.Println(fmt.Sprintf(`Successfully sent HBAR to custodian address, Status: [%s]`, transactionReceipt.Status))

	// Get custodian hbar balance after transfer
	receiverBalanceNew, err := hedera.NewAccountBalanceQuery().
		SetAccountID(setup.BridgeAccount).
		Execute(setup.Clients.Hedera)
	if err != nil {
		fmt.Println("Unable to query the balance of the Bridge Account")
		t.Fatal(err)
	}

	fmt.Println(fmt.Sprintf(`HBAR custodian balance after transaction: [%d]`, receiverBalanceNew.Hbars.AsTinybar()))

	// Verify that the custodial address has received exactly the amount sent
	if (receiverBalanceNew.Hbars.AsTinybar() - receiverBalance.Hbars.AsTinybar()) != hedera.HbarFrom(hBarAmount, "hbar").AsTinybar() {
		t.Fatalf(`Expected to recieve the exact transfer amount of hbar: [%v]`, hedera.HbarFrom(hBarAmount, "hbar").AsTinybar())
	}

	return transactionResponse, whbarBalanceBefore
}

func sendHbarsToBridgeAccount(senderAccount hedera.AccountID, custodialAccount hedera.AccountID, memo string, hBarAmount float64, client *hedera.Client) (hedera.TransactionResponse, error) {
	fmt.Println(fmt.Sprintf(`Sending [%v] Hbars through the Bridge. Transaction Memo: [%s]`, hBarAmount, memo))

	res, _ := hedera.NewTransferTransaction().AddHbarSender(senderAccount, hedera.HbarFrom(hBarAmount, "hbar")).
		AddHbarRecipient(custodialAccount, hedera.HbarFrom(hBarAmount, "hbar")).
		SetTransactionMemo(memo).
		Execute(client)
	rec, err := res.GetReceipt(client)

	fmt.Println(fmt.Sprintf(`TX broadcasted. ID [%s], Status: [%s]`, res.TransactionID, rec.Status))
	time.Sleep(1 * time.Second)

	return res, err
}

func verifyTopicMessages(setup *config.Setup, transactionResponse hedera.TransactionResponse, validatorsCount int, t *testing.T) string {
	ethSignaturesCollected := 0
	ethTransMsgCollected := 0
	ethTransactionHash := ""

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
				}

				if msg.GetType() == validatorproto.TopicSubmissionType_EthTransaction {
					//Verify that the eth transaction message has been submitted
					topicSubmissionMessageTrans := tx.FromHederaTransactionID(&transactionResponse.TransactionID)
					if msg.GetTopicEthTransactionMessage().TransactionId != topicSubmissionMessageTrans.String() {
						t.Fatalf(`Expected ethereum transaction message to contain the transaction id: [%s]`, topicSubmissionMessageTrans.String())
					}
					ethTransactionHash = msg.GetTopicEthTransactionMessage().GetEthTxHash()
					ethTransMsgCollected++
				}
			},
		)
	if err != nil {
		t.Fatalf(`Unable to subscribe to Topic [%s]`, setup.TopicID)
	}

	// Wait for topic consensus messages to arrive
	time.Sleep(60 * time.Second)

	// Check that all the validators have submitted a message with authorisation signature
	if ethSignaturesCollected != validatorsCount {
		t.Fatalf(`Expected the count of collected signatures to equal the number of validators: [%v], but was: [%v]`, validatorsCount, ethSignaturesCollected)
	}

	// Verify the exactly on eth transaction hash has been submitted
	if ethTransMsgCollected != 1 {
		t.Fatal(`Expected to submit exactly 1 ethereum transaction in topic`)
	}

	return ethTransactionHash
}

func verifyEthereumTXExecution(setup *config.Setup, ethTransactionHash string, whbarReceiverAddress common.Address, hBarAmount float64, whbarBalanceBefore *big.Int, t *testing.T) {
	fmt.Printf("Waiting for transaction [%s] to succeed...\n", ethTransactionHash)

	success, err := setup.Clients.EthClient.WaitForTransactionSuccess(common.HexToHash(ethTransactionHash))

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
	if (whbarBalanceAfter.Int64() - whbarBalanceBefore.Int64()) != hedera.HbarFrom(hBarAmount, "hbar").AsTinybar() {
		t.Fatalf(`Expected to recieve the exact transfer amount of WHBAR: [%v]`, hedera.HbarFrom(hBarAmount, "hbar").AsTinybar())
	}
}

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
	"log"
	"math/big"
	"testing"
	"time"

	whbar "github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum/contracts/whbar"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	common "github.com/ethereum/go-ethereum/common"
	"github.com/hashgraph/hedera-sdk-go"
	ethclient "github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum"
	tx "github.com/limechain/hedera-eth-bridge-validator/app/process/model/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	validatorproto "github.com/limechain/hedera-eth-bridge-validator/proto"
	"google.golang.org/protobuf/proto"
)

func Test_E2E(t *testing.T) {
	configuration := config.LoadTestConfig()

	memo := "0x7cFae2deF15dF86CfdA9f2d25A361f1123F42eDD1126221237211"
	whbarReceiverAddress := common.HexToAddress("0x7cFae2deF15dF86CfdA9f2d25A361f1123F42eDD")

	hBarAmount := 0.0001
	validatorsCount := 3

	whbarContractAddress := common.HexToAddress(configuration.Hedera.Eth.WhbarContractAddress)
	acc, _ := hedera.AccountIDFromString(configuration.Hedera.Client.Operator.AccountId)
	receiving, _ := hedera.AccountIDFromString(configuration.Hedera.Watcher.CryptoTransfer.Accounts[0].Id)
	topicID, _ := hedera.TopicIDFromString(configuration.Hedera.Watcher.ConsensusMessage.Topics[0].Id)
	ethConfig := configuration.Hedera.Eth

	accID, _ := hedera.AccountIDFromString(configuration.Hedera.Client.Operator.AccountId)
	pK, _ := hedera.PrivateKeyFromString(configuration.Hedera.Client.Operator.PrivateKey)

	client := initClient(accID, pK)
	ethClient := ethclient.NewEthereumClient(ethConfig)
	whbarInstance, err := whbar.NewWhbar(whbarContractAddress, ethClient.Client)

	if err != nil {
		t.Fatal(err)
	}

	transactionResponse, whbarBalanceBefore := verifyCryptoTransfer(memo, acc, receiving, hBarAmount, whbarInstance, whbarReceiverAddress, client, t)

	ethTransactionHash := verifyTopicMessages(topicID, client, transactionResponse, validatorsCount, t)

	verifyEthereumTXExecution(ethTransactionHash, whbarInstance, whbarReceiverAddress, hBarAmount, whbarBalanceBefore, ethClient, t)

}

func sendTransactionToCustodialAccount(senderAccount hedera.AccountID, custodialAccount hedera.AccountID, memo string, hBarAmount float64, client *hedera.Client) (hedera.TransactionResponse, error) {
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

func initClient(accID hedera.AccountID, pK hedera.PrivateKey) *hedera.Client {
	client := hedera.ClientForTestnet()
	client.SetOperator(accID, pK)

	return client
}

func verifyCryptoTransfer(memo string, acc hedera.AccountID, receiving hedera.AccountID, hBarAmount float64, whbarInstance *whbar.Whbar, whbarReceiverAddress common.Address, client *hedera.Client, t *testing.T) (hedera.TransactionResponse, *big.Int) {
	// Get the wrapped hbar balance of the receiver before the transfer
	whbarBalanceBefore, err := whbarInstance.BalanceOf(&bind.CallOpts{}, whbarReceiverAddress)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("WHBAR balance before transaction: [%s]\n", whbarBalanceBefore)
	// Get custodian hbar balance before transfer
	receiverBalance, err := hedera.NewAccountBalanceQuery().
		SetAccountID(receiving).
		Execute(client)

	fmt.Println(fmt.Sprintf(`HBAR custodian balance before transaction: [%d]`, receiverBalance.Hbars.AsTinybar()))

	// Get the transaction receipt to verify the transaction was executed
	transactionResponse, err := sendTransactionToCustodialAccount(acc, receiving, memo, hBarAmount, client)
	transactionReceipt, err := transactionResponse.GetReceipt(client)

	if err != nil {
		fmt.Println(fmt.Sprintf(`Transaction unsuccessful, Error: [%s]`, err))
		t.Fatal(err)
	}

	fmt.Println(fmt.Sprintf(`Successfully sent HBAR to custodian adress, Status: [%s]`, transactionReceipt.Status))

	// Get custodian hbar balance after transfer
	receiverBalanceNew, err := hedera.NewAccountBalanceQuery().
		SetAccountID(receiving).
		Execute(client)

	fmt.Println(fmt.Sprintf(`HBAR custodian balance after transaction: [%d]`, receiverBalanceNew.Hbars.AsTinybar()))

	// Verify that the custodial address has recieved exactly the amount sent
	if (receiverBalanceNew.Hbars.AsTinybar() - receiverBalance.Hbars.AsTinybar()) != hedera.HbarFrom(hBarAmount, "hbar").AsTinybar() {
		t.Fatalf(`Expected to recieve the exact transfer amount of hbar: [%v]`, hedera.HbarFrom(hBarAmount, "hbar").AsTinybar())
	}

	return transactionResponse, whbarBalanceBefore
}

func verifyTopicMessages(topicID hedera.TopicID, client *hedera.Client, transactionResponse hedera.TransactionResponse, validatorsCount int, t *testing.T) string {
	ethSignaturesCollected := 0
	ethTransMsgCollected := 0
	ethTransactionHash := ""

	// Subscribe to Topic
	hedera.NewTopicMessageQuery().
		SetStartTime(time.Unix(0, time.Now().UnixNano())).
		SetTopicID(topicID).
		Subscribe(
			client,
			func(response hedera.TopicMessage) {
				msg := &validatorproto.TopicSubmissionMessage{}
				proto.Unmarshal(response.Contents, msg)

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

	// Wait for topic consensus messages to arrive
	time.Sleep(60 * time.Second)

	// Check that all the validators have submitted a message with authorisation signature
	if ethSignaturesCollected != validatorsCount {
		t.Fatalf(`Expected the count of collected signatures to equal the number of validators: [%v]`, validatorsCount)
	}

	// Verify the exactly on eth transaction hash has been submitted
	if ethTransMsgCollected != 1 {
		t.Fatal(`Expected to submit exactly 1 ethereum transaction in topic`)
	}

	return ethTransactionHash
}

func verifyEthereumTXExecution(ethTransactionHash string, whbarInstance *whbar.Whbar, whbarReceiverAddress common.Address, hBarAmount float64, whbarBalanceBefore *big.Int, ethClient *ethclient.EthereumClient, t *testing.T) {
	fmt.Printf("Waiting for transaction [%s] to succeed...\n", ethTransactionHash)

	success, err := ethClient.WaitForTransactionSuccess(common.HexToHash(ethTransactionHash))

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
	whbarBalanceAfter, err := whbarInstance.BalanceOf(&bind.CallOpts{}, whbarReceiverAddress)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("WHBAR balance after transaction: [%s]\n", whbarBalanceAfter)

	// Verify that the ethereum address hass recieved the exact transfer amount of WHBARs
	if (whbarBalanceAfter.Int64() - whbarBalanceBefore.Int64()) != hedera.HbarFrom(hBarAmount, "hbar").AsTinybar() {
		t.Fatalf(`Expected to recieve the exact transfer amount of WHBAR: [%v]`, hedera.HbarFrom(hBarAmount, "hbar").AsTinybar())
	}
}

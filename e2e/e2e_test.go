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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashgraph/hedera-sdk-go"
	ethclient "github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum/contracts/bridge"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum/contracts/whbar"
	tx "github.com/limechain/hedera-eth-bridge-validator/app/process/model/transaction"
	apiresponse "github.com/limechain/hedera-eth-bridge-validator/app/router/response"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	validatorproto "github.com/limechain/hedera-eth-bridge-validator/proto"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"testing"
	"time"
)

var (
	incrementFloat, _    = new(big.Int).SetString("1", 10)
	hBarAmount           = hedera.HbarFrom(100, "hbar")
	precision            = new(big.Int).SetInt64(100000)
	whbarReceiverAddress = common.HexToAddress(receiverAddress)
)

const (
	receiverAddress         = "0x7cFae2deF15dF86CfdA9f2d25A361f1123F42eDD"
	gasPriceGwei            = "1"
	apiMetadataUrl          = "http://localhost:%s/api/v1/metadata?gasPriceGwei=%s"
	expectedValidatorsCount = 3
)

func Test_E2E(t *testing.T) {
	configuration := config.LoadTestConfig()

	whbarContractAddress := common.HexToAddress(configuration.Hedera.Eth.WhbarContractAddress)
	bridgeContractAddress := common.HexToAddress(configuration.Hedera.Eth.BridgeContractAddress)
	acc, _ := hedera.AccountIDFromString(configuration.Hedera.Client.Operator.AccountId)
	receiving, _ := hedera.AccountIDFromString(configuration.Hedera.Watcher.CryptoTransfer.Accounts[0].Id)
	topicID, _ := hedera.TopicIDFromString(configuration.Hedera.Watcher.ConsensusMessage.Topics[0].Id)
	ethConfig := configuration.Hedera.Eth

	accID, _ := hedera.AccountIDFromString(configuration.Hedera.Client.Operator.AccountId)
	pK, _ := hedera.PrivateKeyFromString(configuration.Hedera.Client.Operator.PrivateKey)

	client := initClient(accID, pK)
	ethClient := ethclient.NewEthereumClient(ethConfig)
	whbarInstance, err := whbar.NewWhbar(whbarContractAddress, ethClient.Client)

	metadataResponse, err := getMetadata(
		fmt.Sprintf(
			apiMetadataUrl,
			configuration.Hedera.Validator.Port,
			gasPriceGwei))
	if err != nil {
		t.Fatal(err)
	}

	txFee, err := txFeeToBigInt(metadataResponse.TransactionFee)
	if err != nil {
		t.Fatal(err)
	}

	memo := fmt.Sprintf("%s-%s-%s", receiverAddress, txFee, gasPriceGwei)

	bridgeInstance, err := bridge.NewBridge(bridgeContractAddress, ethClient.Client)
	if err != nil {
		t.Fatal(err)
	}

	serviceFeePercentage, err := bridgeInstance.ServiceFee(nil)
	if err != nil {
		t.Fatal(err)
	}

	whbarAmount, err := calculateWHBarAmount(txFee, serviceFeePercentage)
	if err != nil {
		t.Fatal(err)
	}

	transactionResponse, whbarBalanceBefore := verifyCryptoTransfer(memo, acc, receiving, whbarInstance, whbarReceiverAddress, client, t)

	ethTransactionHash := verifyTopicMessages(topicID, client, transactionResponse, expectedValidatorsCount, t)

	verifyEthereumTXExecution(ethTransactionHash, whbarInstance, whbarReceiverAddress, whbarAmount.Int64(), whbarBalanceBefore, ethClient, t)
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

func getMetadata(url string) (*apiresponse.MetadataResponse, error) {
	httpClient := http.Client{}

	response, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("Get Metadata resolved with status [%d].", response.StatusCode))
	}

	bodyBytes, err := ioutil.ReadAll(response.Body)

	var metadataResponse *apiresponse.MetadataResponse
	err = json.Unmarshal(bodyBytes, &metadataResponse)
	if err != nil {
		return nil, err
	}

	return metadataResponse, nil
}

func sendTransactionToCustodialAccount(senderAccount hedera.AccountID, custodialAccount hedera.AccountID, memo string, client *hedera.Client) (hedera.TransactionResponse, error) {
	fmt.Println(fmt.Sprintf(`Sending [%v] Hbars through the Bridge. Transaction Memo: [%s]`, hBarAmount, memo))

	res, _ := hedera.NewTransferTransaction().AddHbarSender(senderAccount, hBarAmount).
		AddHbarRecipient(custodialAccount, hBarAmount).
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

func verifyCryptoTransfer(memo string, acc hedera.AccountID, receiving hedera.AccountID, whbarInstance *whbar.Whbar, whbarReceiverAddress common.Address, client *hedera.Client, t *testing.T) (hedera.TransactionResponse, *big.Int) {
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
	transactionResponse, err := sendTransactionToCustodialAccount(acc, receiving, memo, client)
	transactionReceipt, err := transactionResponse.GetReceipt(client)

	if err != nil {
		fmt.Println(fmt.Sprintf(`Transaction unsuccessful, Error: [%s]`, err))
		t.Fatal(err)
	}

	fmt.Println(fmt.Sprintf(`Successfully sent HBAR to custodian address, Status: [%s]`, transactionReceipt.Status))

	// Get custodian hbar balance after transfer
	receiverBalanceNew, err := hedera.NewAccountBalanceQuery().
		SetAccountID(receiving).
		Execute(client)

	fmt.Println(fmt.Sprintf(`HBAR custodian balance after transaction: [%d]`, receiverBalanceNew.Hbars.AsTinybar()))

	// Verify that the custodial address has receive exactly the amount sent
	amount := receiverBalanceNew.Hbars.AsTinybar() - receiverBalance.Hbars.AsTinybar()
	if amount != hBarAmount.AsTinybar() {
		t.Fatalf(`Expected to receive [%v] transfer amount of hbar, but got [%v].`, hBarAmount.AsTinybar(), amount)
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

	fmt.Println("Sleeping 60s...")
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

func verifyEthereumTXExecution(ethTransactionHash string, whbarInstance *whbar.Whbar, whbarReceiverAddress common.Address, wHbarAmount int64, whbarBalanceBefore *big.Int, ethClient *ethclient.EthereumClient, t *testing.T) {
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

	// Verify that the ethereum address has received the exact transfer amount of WHBARs
	amount := whbarBalanceAfter.Int64() - whbarBalanceBefore.Int64()
	if amount != wHbarAmount {
		t.Fatalf(`Expected to receive [%v] WHBAR, but got [%v].`, wHbarAmount, amount)
	}
}

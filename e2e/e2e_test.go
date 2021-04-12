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
	routerContract "github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	entity_transfer "github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/e2e/service/database"
	"github.com/limechain/hedera-eth-bridge-validator/e2e/setup"
	"log"
	"math/big"
	"strconv"
	"strings"
	"testing"
	"time"

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
	hbarRemovalAmount            = hedera.HbarFrom(-amount, "hbar")
	precision                    = new(big.Int).SetInt64(100000)
	whbarReceiverAddress         = common.HexToAddress(receiverAddress)
)

const (
	receiverAddress         = "0x7cFae2deF15dF86CfdA9f2d25A361f1123F42eDD"
	expectedValidatorsCount = 3
)

func Test_HBAR(t *testing.T) {
	setupEnv := setup.Load()

	memo := fmt.Sprintf("%s-0", receiverAddress)

	// Step 1 - Verify the transfer of Hbars to the Bridge Account
	transactionResponse, _ := verifyTransferToBridgeAccount(setupEnv, memo, whbarReceiverAddress, t)

	// Step 2 - Verify the submitted topic messages
	ethTransactionHash, receivedSignatures := verifyTopicMessages(setupEnv, transactionResponse, 0, t)

	expectedTxRecord := prepareExpectedTransfer(
		setupEnv.Clients.RouterContract,
		transactionResponse.TransactionID,
		"HBAR",
		"0",
		database.ExpectedStatuses{
			Status:          entity_transfer.StatusCompleted,
			StatusSignature: entity_transfer.StatusSignatureMined,
		},
		ethTransactionHash, t)

	// Step 4 - Verify Database Records
	verifyDatabaseRecords(setupEnv.DbValidation, expectedTxRecord, receivedSignatures, t)
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

func prepareExpectedTransfer(routerContract *routerContract.Router, transactionID hedera.TransactionID, nativeToken, txFee string, statuses database.ExpectedStatuses, ethTransactionHash string, t *testing.T) *entity.Transfer {
	expectedTxId := fromHederaTransactionID(&transactionID)

	wrappedToken, err := setup.ParseToken(routerContract, nativeToken)
	if err != nil {
		t.Fatalf("Expecting Token [%s] is not supported. - Error: [%s]", expectedTxId, err)
	}

	amount := strconv.FormatInt(hBarSendAmount.AsTinybar(), 10)

	return &entity.Transfer{
		TransactionID:      expectedTxId.String(),
		Receiver:           receiverAddress,
		NativeToken:        nativeToken,
		WrappedToken:       wrappedToken.String(),
		Amount:             amount,
		TxReimbursement:    txFee,
		Status:             statuses.Status,
		SignatureMsgStatus: statuses.StatusSignature,
		EthTxMsgStatus:     statuses.StatusEthTxMsg,
		EthTxStatus:        statuses.StatusEthTx,
		EthTxHash:          ethTransactionHash,
	}
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

	// Verify that the custodial address has received exactly the amount sent
	amount := receiverBalanceNew.Hbars.AsTinybar() - receiverBalance.Hbars.AsTinybar()
	// Verify that the bridge account has received exactly the amount sent
	if amount != hBarSendAmount.AsTinybar() {
		t.Fatalf(`Expected to recieve the exact transfer amount of hbar: [%v]`, hBarSendAmount.AsTinybar())
	}

	return *transactionResponse, whbarBalanceBefore
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

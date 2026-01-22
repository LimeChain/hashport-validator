/*
 * Copyright 2022 LimeChain Ltd.
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

package submit

import (
	"fmt"
	"time"

	"github.com/hashgraph/hedera-sdk-go/v2"
)

func HbarToBridgeAccount(hederaClient *hedera.Client, bridgeAccount hedera.AccountID, memo string, amount int64) (*hedera.TransactionResponse, error) {
	hbarSendAmount := hedera.HbarFromTinybar(amount)
	hbarRemovalAmount := hedera.HbarFromTinybar(-amount)
	fmt.Printf("Sending [%v] Hbars through the Bridge. Transaction Memo: [%s]\n", hbarSendAmount, memo)

	res, err := hedera.NewTransferTransaction().
		AddHbarTransfer(hederaClient.GetOperatorAccountID(), hbarRemovalAmount).
		AddHbarTransfer(bridgeAccount, hbarSendAmount).
		SetTransactionMemo(memo).
		Execute(hederaClient)
	if err != nil {
		return nil, err
	}
	rec, err := res.GetReceipt(hederaClient)
	if err != nil {
		return nil, err
	}

	fmt.Printf("TX broadcasted. ID [%s], Status: [%s]\n", res.TransactionID, rec.Status)
	time.Sleep(1 * time.Second)

	return &res, err
}

func TokensToBridgeAccount(hederaClient *hedera.Client, bridgeAccount hedera.AccountID, tokenID hedera.TokenID, memo string, amount int64) (*hedera.TransactionResponse, error) {
	fmt.Printf("Sending [%v] Tokens to the Bridge. Transaction Memo: [%s]\n", amount, memo)

	res, err := hedera.NewTransferTransaction().
		SetTransactionMemo(memo).
		AddTokenTransfer(tokenID, hederaClient.GetOperatorAccountID(), -amount).
		AddTokenTransfer(tokenID, bridgeAccount, amount).
		Execute(hederaClient)
	if err != nil {
		return nil, err
	}
	rec, err := res.GetReceipt(hederaClient)
	if err != nil {
		return nil, err
	}

	fmt.Printf("TX broadcasted. ID [%s], Status: [%s]\n", res.TransactionID, rec.Status)
	time.Sleep(1 * time.Second)

	return &res, err
}

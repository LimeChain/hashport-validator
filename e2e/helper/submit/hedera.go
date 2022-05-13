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
	"github.com/limechain/hedera-eth-bridge-validator/e2e/setup"
)

func HbarToBridgeAccount(setup *setup.Setup, memo string, amount int64) (*hedera.TransactionResponse, error) {
	hbarSendAmount := hedera.HbarFromTinybar(amount)
	hbarRemovalAmount := hedera.HbarFromTinybar(-amount)
	fmt.Println(fmt.Sprintf("Sending [%v] Hbars through the Bridge. Transaction Memo: [%s]", hbarSendAmount, memo))

	res, err := hedera.NewTransferTransaction().
		AddHbarTransfer(setup.Clients.Hedera.GetOperatorAccountID(), hbarRemovalAmount).
		AddHbarTransfer(setup.BridgeAccount, hbarSendAmount).
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

func TokensToBridgeAccount(setup *setup.Setup, tokenID hedera.TokenID, memo string, amount int64) (*hedera.TransactionResponse, error) {
	fmt.Println(fmt.Sprintf("Sending [%v] Tokens to the Bridge. Transaction Memo: [%s]", amount, memo))

	res, err := hedera.NewTransferTransaction().
		SetTransactionMemo(memo).
		AddTokenTransfer(tokenID, setup.Clients.Hedera.GetOperatorAccountID(), -amount).
		AddTokenTransfer(tokenID, setup.BridgeAccount, amount).
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

func NFTWithFeeToBridgeAccount(setup *setup.Setup, memo string, token string, serialNum int64, fee int64) (*hedera.TransactionResponse, error) {
	hbarSendAmount := hedera.HbarFromTinybar(fee)
	hbarRemovalAmount := hedera.HbarFromTinybar(-fee)

	fmt.Println(fmt.Sprintf("Sending NFT [%s], Serial num [%d] through the Portal. Transaction Memo: [%s]", token, serialNum, memo))
	nftID, err := hedera.NftIDFromString(fmt.Sprintf("%d@%s", serialNum, token))
	if err != nil {
		return nil, err
	}

	res, err := hedera.NewTransferTransaction().
		AddNftTransfer(nftID, setup.Clients.Hedera.GetOperatorAccountID(), setup.BridgeAccount).
		AddHbarTransfer(setup.Clients.Hedera.GetOperatorAccountID(), hbarRemovalAmount).
		AddHbarTransfer(setup.BridgeAccount, hbarSendAmount).
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

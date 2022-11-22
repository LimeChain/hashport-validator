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

package transaction

import (
	mirrorNodeErr "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/error"
	timestampHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	transaction                  Transaction
	latestTransaction            Transaction
	transferAccountId            = "0.0.444444"
	tokenAccountId               = "0.0.555555"
	amount                       = int64(10)
	token                        = "0.0.111111"
	nftSenderAccountId           = "0.0.666666"
	nftReceiverAccountId         = "0.0.777777"
	serialNumber                 = int64(1234)
	olderConsensusTimestamp      = "1631092490.303966000"
	moreRecentConsensusTimestamp = "1631092491.483966000"
	nonExistingAccount           = "0.0.non-existing"
	response                     Response
)

func Test_GetHBARTransfer(t *testing.T) {
	setup()

	actualAmount, isFound := transaction.GetHBARTransfer(transferAccountId)

	assert.True(t, isFound)
	assert.Equal(t, amount, actualAmount)
}

func Test_GetHBARTransfer_NotFound(t *testing.T) {
	setup()

	actualAmount, isFound := transaction.GetHBARTransfer(nonExistingAccount)

	assert.False(t, isFound)
	assert.Equal(t, int64(0), actualAmount)
}

func Test_GetIncomingTransfer_Transfer(t *testing.T) {
	setup()

	parsedTransfer, err := transaction.GetIncomingTransfer(transferAccountId)

	assert.Nil(t, err)
	assert.False(t, parsedTransfer.IsNft)
	assert.Equal(t, constants.Hbar, parsedTransfer.Asset)
	assert.Equal(t, amount, parsedTransfer.AmountOrSerialNum)
}

func Test_GetIncomingTransfer_TokenTransfer(t *testing.T) {
	setup()

	parsedTransfer, err := transaction.GetIncomingTransfer(tokenAccountId)

	assert.Nil(t, err)
	assert.False(t, parsedTransfer.IsNft)
	assert.Equal(t, token, parsedTransfer.Asset)
	assert.Equal(t, amount, parsedTransfer.AmountOrSerialNum)
}

func Test_GetIncomingTransfer_NftTransfer(t *testing.T) {
	setup()

	parsedTransfer, err := transaction.GetIncomingTransfer(nftReceiverAccountId)

	assert.Nil(t, err)
	assert.True(t, parsedTransfer.IsNft)
	assert.Equal(t, token, parsedTransfer.Asset)
	assert.Equal(t, serialNumber, parsedTransfer.AmountOrSerialNum)
}

func Test_GetIncomingTransfer_NonExisting(t *testing.T) {
	setup()

	parsedTransfer, err := transaction.GetIncomingTransfer(nonExistingAccount)

	assert.Error(t, err)
	assert.Equal(t, ParsedTransfer{}, parsedTransfer)
}

func Test_IsNotFound(t *testing.T) {
	setup()

	isNotFound := response.IsNotFound()

	assert.True(t, isNotFound)
}

func Test_IsNotFound_False(t *testing.T) {
	setup()
	response.Status.Messages[0] = mirrorNodeErr.ErrorMessage{}

	isNotFound := response.IsNotFound()

	response.Status.Messages[0] = mirrorNodeErr.ErrorMessage{Message: mirrorNodeErr.NotFoundMsg}

	assert.False(t, isNotFound)
}

func Test_GetLatestTxnConsensusTime(t *testing.T) {
	setup()
	response.Transactions[1].ConsensusTimestamp = "invalid"

	consensusTimestamp, err := response.GetLatestTxnConsensusTime()

	response.Transactions[1].ConsensusTimestamp = moreRecentConsensusTimestamp

	assert.Error(t, err)
	assert.Equal(t, int64(0), consensusTimestamp)
}

func Test_GetLatestTxnConsensusTime_Err(t *testing.T) {
	setup()
	expectedTimestamp, _ := timestampHelper.FromString(moreRecentConsensusTimestamp)

	consensusTimestamp, err := response.GetLatestTxnConsensusTime()

	assert.Nil(t, err)
	assert.Equal(t, expectedTimestamp, consensusTimestamp)
}

func setup() {

	transaction = Transaction{
		TokenTransfers: []Transfer{
			{
				Account: tokenAccountId,
				Amount:  amount,
				Token:   token,
			},
		},
		Transfers: []Transfer{
			{
				Account: transferAccountId,
				Amount:  amount,
				Token:   constants.Hbar,
			},
		},
		NftTransfers: []NftTransfer{
			{
				SenderAccountID:   nftSenderAccountId,
				ReceiverAccountID: nftReceiverAccountId,
				SerialNumber:      serialNumber,
				Token:             token,
			},
		},
		ConsensusTimestamp: olderConsensusTimestamp,
	}

	latestTransaction = Transaction{
		Transfers: []Transfer{
			{
				Account: transferAccountId,
				Amount:  amount,
				Token:   constants.Hbar,
			},
		},
		ConsensusTimestamp: moreRecentConsensusTimestamp,
	}

	response = Response{
		Transactions: []Transaction{
			latestTransaction,
			transaction,
		},
		Status: mirrorNodeErr.Status{
			Messages: []mirrorNodeErr.ErrorMessage{
				{Message: mirrorNodeErr.NotFoundMsg},
			},
		},
	}
}

func Test_GetTokenTransfer(t *testing.T) {
	setup()

	actualAmount, isFound := transaction.GetTokenTransfer(tokenAccountId)

	assert.True(t, isFound)
	assert.Equal(t, amount, actualAmount)
}

func Test_GetTokenTransfer_NotFound(t *testing.T) {
	setup()

	actualAmount, isFound := transaction.GetTokenTransfer(nonExistingAccount)

	assert.False(t, isFound)
	assert.Equal(t, int64(0), actualAmount)
}

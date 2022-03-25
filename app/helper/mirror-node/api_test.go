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
 * imitations under the License.
 */

package mirror_node

import (
	mirrorNodeModel "github.com/limechain/hedera-eth-bridge-validator/app/model/mirror-node"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	parsedTransactionResponse = mirrorNodeModel.TransactionsResponse{
		Transactions: []mirrorNodeModel.Transaction{
			{
				Bytes:                    nil,
				ChargedTxFee:             0,
				ConsensusTimestamp:       "1648206035.257921608",
				EntityId:                 "0.0.112",
				MaxFee:                   "200000000",
				MemoBase64:               "Y3VycmVudFJhdGUgOiAwLjIyNTksIG5leHRSYXRlIDogMC4yMjY0LCBtaWRuaWdodC1jdXJyZW50UmF0ZSA6IDAuMjIwNSBtaWRuaWdodC1uZXh0UmF0ZSA6IDAuMjIwMA==",
				Name:                     "FILEUPDATE",
				Node:                     "0.0.4",
				Nonce:                    0,
				ParentConsensusTimestamp: "",
				Result:                   "SUCCESS",
				Scheduled:                false,
				TransactionHash:          "zSMXPzZVq1czqAoNmpXjvBybW5K9RHdgGSDoBAI+nmAV/viN7tCCBp7gmbe+vuki",
				TransactionId:            "0.0.57-1648206026-390263003",
				Transfers:                make([]interface{}, 0),
				ValidDurationSeconds:     "120",
				ValidStartTimestamp:      "1648206026.390263003",
			},
		},
		Links: map[string]string{
			"next": "/api/v1/transactions?account.id=0.0.57&transactiontype=fileupdate&limit=1&timestamp=lt:1648206035.257921608",
		},
	}

	expectedCurrentRate = decimal.NewFromFloat(0.2259)
	expectedNextRate    = decimal.NewFromFloat(0.2264)

	logger = config.GetLoggerFor("Test")
)

func Test_GetUpdatedFileRateFromParsedResponseForHBARPrice(t *testing.T) {
	response, err := GetUpdatedFileRateFromParsedResponseForHBARPrice(parsedTransactionResponse, logger)

	assert.Equal(t, expectedCurrentRate, response.CurrentRate)
	assert.Equal(t, expectedNextRate, response.NextRate)
	assert.Nil(t, err)
}

func Test_GetUpdatedFileRateFromParsedResponseForHBARPrice_ErrorEmptyTransactions(t *testing.T) {

	parsedTransactionResponseWithBrokenMemo := parsedTransactionResponse
	parsedTransactionResponseWithBrokenMemo.Transactions = make([]mirrorNodeModel.Transaction, 0)
	_, err := GetUpdatedFileRateFromParsedResponseForHBARPrice(parsedTransactionResponseWithBrokenMemo, logger)

	assert.Error(t, err, "No transactions received from HBAR Price Hedera Response.")
}

func Test_GetUpdatedFileRateFromParsedResponseForHBARPrice_ErrorWithBrokenMemoBase64(t *testing.T) {

	parsedTransactionResponseWithBrokenMemo := parsedTransactionResponse
	parsedTransactionResponseWithBrokenMemo.Transactions[0].MemoBase64 = "..."
	_, err := GetUpdatedFileRateFromParsedResponseForHBARPrice(parsedTransactionResponseWithBrokenMemo, logger)

	assert.Error(t, err, "illegal base64 data at input byte 0")
}

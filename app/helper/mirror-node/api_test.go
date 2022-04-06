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

package mirror_node

import (
	mirrorNodeModel "github.com/limechain/hedera-eth-bridge-validator/app/model/mirror-node"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	testConstants "github.com/limechain/hedera-eth-bridge-validator/test/constants"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	logger = config.GetLoggerFor("Test")
)

func Test_GetUpdatedFileRateFromParsedResponseForHBARPrice(t *testing.T) {
	response, err := GetUpdatedFileRateFromParsedResponseForHBARPrice(testConstants.ParsedTransactionResponse, logger)

	assert.Equal(t, testConstants.ParsedTransactionResponseCurrentRate, response.CurrentRate)
	assert.Equal(t, testConstants.ParsedTransactionResponseNextRate, response.NextRate)
	assert.Nil(t, err)
}

func Test_GetUpdatedFileRateFromParsedResponseForHBARPrice_ErrorEmptyTransactions(t *testing.T) {

	parsedTransactionResponseWithBrokenMemo := testConstants.ParsedTransactionResponse
	parsedTransactionResponseWithBrokenMemo.Transactions = make([]mirrorNodeModel.Transaction, 0)
	_, err := GetUpdatedFileRateFromParsedResponseForHBARPrice(parsedTransactionResponseWithBrokenMemo, logger)

	assert.Error(t, err, "No transactions received from HBAR Price Hedera Response.")
}

func Test_GetUpdatedFileRateFromParsedResponseForHBARPrice_ErrorWithBrokenMemoBase64(t *testing.T) {

	parsedTransactionResponseWithBrokenMemo := testConstants.ParsedTransactionResponse
	parsedTransactionResponseWithBrokenMemo.Transactions[0].MemoBase64 = "..."
	_, err := GetUpdatedFileRateFromParsedResponseForHBARPrice(parsedTransactionResponseWithBrokenMemo, logger)

	assert.Error(t, err, "illegal base64 data at input byte 0")
}

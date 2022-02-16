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

package hedera

import (
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	transactionID         = "0.0.9401@1598924675.82525000"
	expectedTransactionID = "0.0.9401-1598924675-082525000"
	expectedAccountID     = "0.0.9401"
	expectedSeconds       = "1598924675"
	expectedNanos         = "082525000"
	expectedTimestamp     = "1598924675.082525000"
)

func Test_ToMirrorNodeTransactionID(t *testing.T) {
	res := ToMirrorNodeTransactionID(transactionID)
	assert.Equal(t, expectedTransactionID, res)
}

func Test_FromHederaTransactionID(t *testing.T) {
	hederaTransactionID, err := hedera.TransactionIdFromString(transactionID)

	res := FromHederaTransactionID(hederaTransactionID)
	assert.Equal(t, expectedAccountID, res.AccountId)
	assert.Equal(t, expectedSeconds, res.Seconds)
	assert.Equal(t, expectedNanos, res.Nanos)
	assert.Equal(t, expectedTransactionID, res.String())
	assert.Equal(t, expectedTimestamp, res.Timestamp())
	assert.Nil(t, err)
}

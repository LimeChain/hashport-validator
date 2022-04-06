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

package fee

import (
	"github.com/hashgraph/hedera-sdk-go/v2"
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	transfers []model.Hedera
)

func setup() {
	acc1, acc2, _ := accountIds()
	transfers = []model.Hedera{
		{
			AccountID: acc1,
			Amount:    -5000,
		},
		{
			AccountID: acc2,
			Amount:    5000,
		},
	}
}

func accountIds() (hedera.AccountID, hedera.AccountID, hedera.AccountID) {
	accId1, _ := hedera.AccountIDFromString("0.0.1")
	accId2, _ := hedera.AccountIDFromString("0.0.2")
	accId3, _ := hedera.AccountIDFromString("0.0.3")
	return accId1, accId2, accId3
}

func Test_TotalFeeFromTransfers(t *testing.T) {
	setup()
	_, receiver, _ := accountIds()

	expectedReceiverFound := true
	expectedFee := "0"

	actualFee, actualReceiverFound := TotalFeeFromTransfers(transfers, receiver)
	assert.Equal(t, expectedFee, actualFee)
	assert.Equal(t, expectedReceiverFound, actualReceiverFound)
}

func Test_TotalFeeFromTransfers_ReceiverNotFound(t *testing.T) {
	setup()
	_, _, receiver2 := accountIds()

	expectedReceiverFound := false
	expectedFee := "5000"

	actualFee, actualReceiverFound := TotalFeeFromTransfers(transfers, receiver2)
	assert.Equal(t, expectedFee, actualFee)
	assert.Equal(t, expectedReceiverFound, actualReceiverFound)
}

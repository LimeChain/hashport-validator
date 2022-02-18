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

package distributor

import (
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_SplitTransfersBelowTotal(t *testing.T) {
	length := 6
	positiveAccountAmounts := make([]transfer.Hedera, length)
	for i := 0; i < length; i++ {
		positiveAccountAmounts[i] = transfer.Hedera{
			AccountID: hedera.AccountID{
				Shard:   0,
				Realm:   0,
				Account: uint64(i),
			},
			Amount: 1,
		}
	}
	negativeAccountAmount := transfer.Hedera{
		AccountID: hedera.AccountID{
			Shard:   0,
			Realm:   0,
			Account: 7,
		},
		Amount: int64(-length),
	}
	expected := [][]transfer.Hedera{
		append(positiveAccountAmounts, negativeAccountAmount),
	}

	// when:
	result := SplitAccountAmounts(positiveAccountAmounts, negativeAccountAmount)

	// then:
	assert.Equal(t, 1, len(result))
	assert.Equal(t, expected, result)
	assert.Equal(t, 7, len(result[0]))
}

func Test_SplitTransfersExactLength(t *testing.T) {
	length := 9
	positiveAccountAmounts := make([]transfer.Hedera, length)
	for i := 0; i < length; i++ {
		positiveAccountAmounts[i] = transfer.Hedera{
			AccountID: hedera.AccountID{
				Shard:   0,
				Realm:   0,
				Account: uint64(i),
			},
			Amount: 1,
		}
	}
	negativeAccountAmount := transfer.Hedera{
		AccountID: hedera.AccountID{Shard: 0, Realm: 0, Account: 7},
		Amount:    int64(-length),
	}
	expected := [][]transfer.Hedera{
		append(positiveAccountAmounts, negativeAccountAmount),
	}

	// when:
	result := SplitAccountAmounts(positiveAccountAmounts, negativeAccountAmount)

	// then:
	assert.Equal(t, 1, len(result))
	assert.Equal(t, expected, result)
	assert.Equal(t, 10, len(result[0]))
}

func Test_SplitTransfersAboveTotalTransfersPerTransaction(t *testing.T) {
	length := 11
	positiveAccountAmounts := make([]transfer.Hedera, length)
	for i := 0; i < length; i++ {
		positiveAccountAmounts[i] = transfer.Hedera{
			AccountID: hedera.AccountID{Shard: 0, Realm: 0, Account: uint64(i)},
			Amount:    1,
		}
	}
	negativeAccountAmount := transfer.Hedera{
		AccountID: hedera.AccountID{Shard: 0, Realm: 0, Account: 7},
		Amount:    int64(-length),
	}
	expectedSplit := (length + TotalPositiveTransfersPerTransaction - 1) / TotalPositiveTransfersPerTransaction
	expectedChunkOneLength := 10
	expectedChunkTwoLength := 3

	// when:
	result := SplitAccountAmounts(positiveAccountAmounts, negativeAccountAmount)

	// then:
	assert.Equal(t, expectedSplit, len(result))
	assert.Equal(t, expectedChunkOneLength, len(result[0]))
	// and:
	for i := 0; i < expectedChunkOneLength-1; i++ {
		assert.Equal(t, positiveAccountAmounts[i], result[0][i])
	}
	assert.Equal(t, transfer.Hedera{
		AccountID: negativeAccountAmount.AccountID,
		Amount:    int64(-9),
	}, result[0][expectedChunkOneLength-1])

	// and:
	assert.Equal(t, expectedChunkTwoLength, len(result[1]))
	for i := 0; i < expectedChunkTwoLength-1; i++ {
		assert.Equal(t, positiveAccountAmounts[TotalPositiveTransfersPerTransaction+i], result[1][i])
	}

	assert.Equal(t, transfer.Hedera{
		AccountID: negativeAccountAmount.AccountID,
		Amount:    int64(-2),
	}, result[1][expectedChunkTwoLength-1])
}

func Test_SplitTransfersAboveTotalTransfersEquallyDivided(t *testing.T) {
	length := 18
	positiveAccountAmounts := make([]transfer.Hedera, length)
	for i := 0; i < length; i++ {
		positiveAccountAmounts[i] = transfer.Hedera{
			AccountID: hedera.AccountID{Shard: 0, Realm: 0, Account: uint64(i)},
			Amount:    1,
		}
	}
	negativeAccountAmount := transfer.Hedera{
		AccountID: hedera.AccountID{Shard: 0, Realm: 0, Account: 7},
		Amount:    int64(-length),
	}
	expectedSplit := (length + TotalPositiveTransfersPerTransaction - 1) / TotalPositiveTransfersPerTransaction
	expectedChunkOneLength := 10
	expectedChunkTwoLength := 10

	// when:
	result := SplitAccountAmounts(positiveAccountAmounts, negativeAccountAmount)

	// then:
	assert.Equal(t, expectedSplit, len(result))
	assert.Equal(t, expectedChunkOneLength, len(result[0]))
	// and:
	for i := 0; i < expectedChunkOneLength-1; i++ {
		assert.Equal(t, positiveAccountAmounts[i], result[0][i])
	}
	assert.Equal(t, transfer.Hedera{
		AccountID: negativeAccountAmount.AccountID,
		Amount:    int64(-9),
	}, result[0][expectedChunkOneLength-1])

	// and:
	assert.Equal(t, expectedChunkTwoLength, len(result[1]))
	for i := 0; i < expectedChunkTwoLength-1; i++ {
		assert.Equal(t, positiveAccountAmounts[TotalPositiveTransfersPerTransaction+i], result[1][i])
	}

	assert.Equal(t, transfer.Hedera{
		AccountID: negativeAccountAmount.AccountID,
		Amount:    int64(-9),
	}, result[1][expectedChunkTwoLength-1])
}

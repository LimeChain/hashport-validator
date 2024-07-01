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
	"testing"

	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func Test_ValidAmounts(t *testing.T) {
	accountID1, _ := hedera.AccountIDFromString("0.0.1")
	accountID2, _ := hedera.AccountIDFromString("0.0.2")
	accountID3, _ := hedera.AccountIDFromString("0.0.3")
	tests := []struct {
		name               string
		service            Service
		amount             int64
		expectedTreasury   int64
		expectedValidators int64
	}{
		{
			name: "Case 1: Valid distribution",
			service: Service{
				rewardPercentages: map[string]int{
					TreasuryReward:  10,
					ValidatorReward: 90,
				},
				accountIDs: []hedera.AccountID{
					accountID1,
					accountID2,
					accountID3,
				},
			},
			amount:             1111,
			expectedTreasury:   111,
			expectedValidators: 999,
		},
		{
			name: "Case 2: Single account",
			service: Service{
				rewardPercentages: map[string]int{
					TreasuryReward:  85,
					ValidatorReward: 15,
				},
				accountIDs: []hedera.AccountID{
					accountID1,
				},
			},
			amount:             2000,
			expectedTreasury:   1700,
			expectedValidators: 300,
		},
		{
			name: "Case 3: No Amount",
			service: Service{
				rewardPercentages: map[string]int{
					TreasuryReward:  50,
					ValidatorReward: 50,
				},
				accountIDs: []hedera.AccountID{
					accountID1,
					accountID2,
					accountID3,
				},
			},
			amount:             0,
			expectedTreasury:   0,
			expectedValidators: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			treasuryAmount, validatorsAmount := tt.service.ValidAmounts(tt.amount)
			if treasuryAmount != tt.expectedTreasury {
				t.Errorf("expected treasury amount %d, got %d", tt.expectedTreasury, treasuryAmount)
			}
			if validatorsAmount != tt.expectedValidators {
				t.Errorf("expected validators amount %d, got %d", tt.expectedValidators, validatorsAmount)
			}
		})
	}
}

func Test_CalculateMemberDistribution(t *testing.T) {
	mockLogger := common.MockLogger{}
	accountID1, _ := hedera.AccountIDFromString("0.0.1")
	accountID2, _ := hedera.AccountIDFromString("0.0.2")
	accountID3, _ := hedera.AccountIDFromString("0.0.3")
	treasuryID, _ := hedera.AccountIDFromString("0.0.4")

	tests := []struct {
		name                  string
		service               Service
		validTreasuryFee      int64
		validValidatorFee     int64
		expectedTransfers     []transfer.Hedera
		expectedError         string
		expectedLoggerEntries []string
	}{
		{
			name: "Case 1: Valid distribution",
			service: Service{
				accountIDs: []hedera.AccountID{
					accountID1,
					accountID2,
					accountID3,
				},
				treasuryID: treasuryID,
				logger:     &mockLogger,
			},
			validTreasuryFee:  100,
			validValidatorFee: 300,
			expectedTransfers: []transfer.Hedera{
				{AccountID: accountID1, Amount: 100},
				{AccountID: accountID2, Amount: 100},
				{AccountID: accountID3, Amount: 100},
				{AccountID: treasuryID, Amount: 100},
			},
			expectedError: "",
		},
		{
			name: "Case 2: Non-divisible validator fee",
			service: Service{
				accountIDs: []hedera.AccountID{accountID1, accountID2},
				treasuryID: treasuryID,
				logger:     &mockLogger,
			},
			validTreasuryFee:  50,
			validValidatorFee: 101,
			expectedTransfers: nil,
			expectedError:     "amount not divisible",
			expectedLoggerEntries: []string{
				"Provided validator fee [%d] is not divisible.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger.On("Errorf", "Provided validator fee [%d] is not divisible.", mock.Anything).Maybe()
			transfers, err := tt.service.CalculateMemberDistribution(tt.validTreasuryFee, tt.validValidatorFee)

			if err != nil && err.Error() != tt.expectedError {
				t.Errorf("expected error '%s', got '%s'", tt.expectedError, err.Error())
			}

			if err == nil && tt.expectedError != "" {
				t.Errorf("expected error '%s', got nil", tt.expectedError)
			}

			if len(transfers) != len(tt.expectedTransfers) {
				t.Errorf("expected transfers length %d, got %d", len(tt.expectedTransfers), len(transfers))
			}

			loggerEntries := tt.service.logger.(*common.MockLogger).Entries
			if len(loggerEntries) != len(tt.expectedLoggerEntries) {
				t.Errorf("expected logger entries length %d, got %d", len(tt.expectedLoggerEntries), len(loggerEntries))
			}

			for i, transfer := range transfers {
				if transfer.AccountID != tt.expectedTransfers[i].AccountID || transfer.Amount != tt.expectedTransfers[i].Amount {
					t.Errorf("expected transfer %v, got %v", tt.expectedTransfers[i], transfer)
				}
			}
		})
	}
}

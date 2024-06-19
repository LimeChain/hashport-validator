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
	"errors"

	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	accountIDs        []hedera.AccountID
	treasuryID        hedera.AccountID
	rewardPercentages map[string]int
	logger            *log.Entry
}

const (
	TreasuryReward  = "treasury"
	ValidatorReward = "validator"
	totalPercentage = 100
)

const TotalPositiveTransfersPerTransaction = 9

func New(members []string, treasuryID string, treasuryRewardPercentage int, validatorRewardPercentage int) *Service {
	if len(members) == 0 {
		log.Fatal("No members accounts provided")
	}
	if treasuryRewardPercentage+validatorRewardPercentage != totalPercentage {
		log.Fatalf("Rewards percentage total must be %d", totalPercentage)
	}
	var accountIDs []hedera.AccountID
	for _, v := range members {
		accountID, err := hedera.AccountIDFromString(v)
		if err != nil {
			log.Fatalf("Invalid members account: [%s].", v)
		}
		accountIDs = append(accountIDs, accountID)
	}

	rewardPercentages := map[string]int{
		ValidatorReward: validatorRewardPercentage,
		TreasuryReward:  treasuryRewardPercentage,
	}

	treasuryAccountID, err := hedera.AccountIDFromString(treasuryID)
	if err != nil {
		log.Fatalf("Invalid treasury account: [%s].", treasuryID)
	}
	return &Service{
		accountIDs:        accountIDs,
		treasuryID:        treasuryAccountID,
		rewardPercentages: rewardPercentages,
		logger:            config.GetLoggerFor("Fee Service")}
}

// CalculateMemberDistribution Returns the transactions to the members and the treasury
func (s Service) CalculateMemberDistribution(validTreasuryFee int64, validValdiatorFee int64) ([]transfer.Hedera, error) {
	feePerAccount := validValdiatorFee / int64(len(s.accountIDs))

	totalValidatorAmount := feePerAccount * int64(len(s.accountIDs))
	if totalValidatorAmount != validValdiatorFee {
		s.logger.Errorf("Provided validator fee [%d] is not divisible.", validValdiatorFee)
		return nil, errors.New("amount not divisible")
	}

	var transfers []transfer.Hedera
	for _, a := range s.accountIDs {
		transfers = append(transfers, transfer.Hedera{
			AccountID: a,
			Amount:    feePerAccount,
		})
	}

	treasuryTransfer := transfer.Hedera{
		AccountID: s.treasuryID,
		Amount:    validTreasuryFee,
	}
	transfers = append(transfers, treasuryTransfer)

	return transfers, nil
}

// SplitAccountAmounts splits account amounts to a chunks of TotalPositiveTransfersPerTransaction + 1
// (1 comes from the negative account amount, opposite to the sum of the positive account amounts)
// It is necessary, because at this given moment, Hedera does not support a transfer transaction with
// a transfer list exceeding (TotalPositiveTransfersPerTransaction + 1)
func SplitAccountAmounts(positiveAccountAmounts []transfer.Hedera, negativeAccountAmount transfer.Hedera) [][]transfer.Hedera {
	totalLength := len(positiveAccountAmounts)

	if totalLength <= TotalPositiveTransfersPerTransaction {
		transfers := append(positiveAccountAmounts, negativeAccountAmount)

		return [][]transfer.Hedera{transfers}
	} else {
		splits := (totalLength + TotalPositiveTransfersPerTransaction - 1) / TotalPositiveTransfersPerTransaction
		result := make([][]transfer.Hedera, splits)

		previous := 0
		for i := 0; previous < totalLength; i++ {
			next := previous + TotalPositiveTransfersPerTransaction
			if next > totalLength {
				next = totalLength
			}
			transfers := make([]transfer.Hedera, next-previous)
			copy(transfers, positiveAccountAmounts[previous:next])
			transfers = append(transfers, transfer.Hedera{AccountID: negativeAccountAmount.AccountID, Amount: calculateOppositeNegative(transfers)})
			result[i] = transfers
			previous = next
		}

		return result
	}
}

func (s Service) PrepareTransfers(amount int64, token string) ([]transaction.Transfer, error) {
	feePerAccount := amount / int64(len(s.accountIDs))

	totalAmount := feePerAccount * int64(len(s.accountIDs))
	if totalAmount != amount {
		s.logger.Errorf("Provided fee [%d] is not divisible.", amount)
		return nil, errors.New("amount not divisible")
	}

	var transfers []transaction.Transfer
	for _, a := range s.accountIDs {
		if token == constants.Hbar {
			transfers = append(transfers, transaction.Transfer{
				Account: a.String(),
				Amount:  feePerAccount,
			})
		} else {
			transfers = append(transfers, transaction.Transfer{
				Account: a.String(),
				Amount:  feePerAccount,
				Token:   token,
			})
		}
	}

	return transfers, nil
}

// ValidAmounts Returns the closest amounts, which can be equally divided to members and treasury
func (s Service) ValidAmounts(amount int64) (int64, int64) {
	treasuryAmount := (amount * int64(s.rewardPercentages[TreasuryReward])) / 100
	feeForMembers := (amount * int64(s.rewardPercentages[ValidatorReward])) / 100

	feePerMember := feeForMembers / int64(len(s.accountIDs))
	validatorsAmount := feePerMember * int64(len(s.accountIDs))
	return treasuryAmount, validatorsAmount
}

// Sums the amounts and returns the opposite
func calculateOppositeNegative(transfers []transfer.Hedera) int64 {
	negatedValue := int64(0)
	for _, transfer := range transfers {
		negatedValue += transfer.Amount
	}

	return -negatedValue
}

/*
 * Copyright 2021 LimeChain Ltd.
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
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	accountIDs []hedera.AccountID
	logger     *log.Entry
}

const TotalPositiveTransfersPerTransaction = 9

func New(members []string) *Service {
	if len(members) == 0 {
		log.Fatal("No members accounts provided")
	}

	var accountIDs []hedera.AccountID
	for _, v := range members {
		accountID, err := hedera.AccountIDFromString(v)
		if err != nil {
			log.Fatalf("Invalid members account: [%s].", v)
		}
		accountIDs = append(accountIDs, accountID)
	}

	return &Service{
		accountIDs: accountIDs,
		logger:     config.GetLoggerFor("Fee Service")}
}

// CalculateMemberDistribution Returns an equally divided to each member
func (s Service) CalculateMemberDistribution(amount int64) ([]transfer.Hedera, error) {
	feePerAccount := amount / int64(len(s.accountIDs))

	totalAmount := feePerAccount * int64(len(s.accountIDs))
	if totalAmount != amount {
		s.logger.Errorf("Provided fee [%d] is not divisible.", amount)
		return nil, errors.New("amount not divisible")
	}

	var transfers []transfer.Hedera
	for _, a := range s.accountIDs {
		transfers = append(transfers, transfer.Hedera{
			AccountID: a,
			Amount:    feePerAccount,
		})
	}

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

func (s Service) PrepareTransfers(amount int64, token string) ([]model.Transfer, error) {
	feePerAccount := amount / int64(len(s.accountIDs))

	totalAmount := feePerAccount * int64(len(s.accountIDs))
	if totalAmount != amount {
		s.logger.Errorf("Provided fee [%d] is not divisible.", amount)
		return nil, errors.New("amount not divisible")
	}

	var transfers []model.Transfer
	for _, a := range s.accountIDs {
		if token == constants.Hbar {
			transfers = append(transfers, model.Transfer{
				Account: a.String(),
				Amount:  feePerAccount,
			})
		} else {
			transfers = append(transfers, model.Transfer{
				Account: a.String(),
				Amount:  feePerAccount,
				Token:   token,
			})
		}
	}

	return transfers, nil
}

// ValidAmount Returns the closest amount, which can be equally divided to members
func (s Service) ValidAmount(amount int64) int64 {
	feePerAccount := amount / int64(len(s.accountIDs))

	totalAmount := feePerAccount * int64(len(s.accountIDs))
	if totalAmount != amount {
		return totalAmount
	}

	return amount
}

// Sums the amounts and returns the opposite
func calculateOppositeNegative(transfers []transfer.Hedera) int64 {
	negatedValue := int64(0)
	for _, transfer := range transfers {
		negatedValue += transfer.Amount
	}

	return -negatedValue
}

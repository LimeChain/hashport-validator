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

package service

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/stretchr/testify/mock"
)

type MockDistrubutorService struct {
	mock.Mock
}

func (mds *MockDistrubutorService) PrepareTransfers(amount int64, token string) ([]model.Transfer, error) {
	args := mds.Called(amount, token)
	if args.Get(1) == nil {
		return args.Get(0).([]model.Transfer), nil
	}
	return nil, args.Get(1).(error)
}

func (mds *MockDistrubutorService) CalculateMemberDistribution(validFee int64) ([]transfer.Hedera, error) {
	args := mds.Called(validFee)
	if args.Get(1) == nil {
		return args.Get(0).([]transfer.Hedera), nil
	}
	return nil, args.Get(1).(error)
}

func (mds *MockDistrubutorService) ValidAmount(amount int64) int64 {
	args := mds.Called(amount)
	return args.Get(0).(int64)
}

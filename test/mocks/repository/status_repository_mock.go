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

package repository

import "github.com/stretchr/testify/mock"

type MockStatusRepository struct {
	mock.Mock
}

func (msr *MockStatusRepository) GetLastFetchedTimestamp(entityID string) (int64, error) {
	args := msr.Called(entityID)
	if args[1] == nil {
		return args[0].(int64), nil
	}
	return args[0].(int64), args[1].(error)
}

func (msr *MockStatusRepository) UpdateLastFetchedTimestamp(entityID string, timestamp int64) error {
	args := msr.Called(entityID, timestamp)
	if args[0] == nil {
		return nil
	}
	return args[0].(error)
}

func (msr *MockStatusRepository) CreateTimestamp(entityID string, timestamp int64) error {
	args := msr.Called(entityID, timestamp)
	if args[0] == nil {
		return nil
	}
	return args[0].(error)
}

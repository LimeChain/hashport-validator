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

package main

import (
	burn_event "github.com/limechain/hedera-eth-bridge-validator/app/persistence/burn-event"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/fee"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/status"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
)

func TestPrepareRepositories(t *testing.T) {
	mocks.Setup()
	mocks.MDatabase.On("GetConnection").Return(&gorm.DB{})
	repositories := PrepareRepositories(mocks.MDatabase)

	assert.IsType(t, &fee.Repository{}, repositories.fee)
	assert.IsType(t, &message.Repository{}, repositories.message)
	assert.IsType(t, &status.Repository{}, repositories.messageStatus)
	assert.IsType(t, &burn_event.Repository{}, repositories.burnEvent)
	assert.IsType(t, &transfer.Repository{}, repositories.transfer)
	assert.IsType(t, &status.Repository{}, repositories.transferStatus)

	assert.NotEmpty(t, repositories)

	assert.NotEmpty(t, repositories.fee)
	assert.NotEmpty(t, repositories.message)
	assert.NotEmpty(t, repositories.messageStatus)
	assert.NotEmpty(t, repositories.burnEvent)
	assert.NotEmpty(t, repositories.transfer)
	assert.NotEmpty(t, repositories.transferStatus)

}

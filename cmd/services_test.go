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
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	tc "github.com/limechain/hedera-eth-bridge-validator/test/test-config"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
)

func TestPrepareServices(t *testing.T) {
	client := PrepareClients(tc.TestConfig.Validator.Clients)

	mocks.Setup()
	mocks.MDatabase.On("GetConnection").Return(&gorm.DB{})
	repositories := PrepareRepositories(mocks.MDatabase)

	res := PrepareServices(tc.TestConfig, *client, *repositories)
	assert.NotEmpty(t, res)
}

func TestPrepareApiOnlyServices(t *testing.T) {
	client := PrepareClients(tc.TestConfig.Validator.Clients)
	res := PrepareApiOnlyServices(tc.TestConfig, *client)
	assert.NotEmpty(t, res)
}

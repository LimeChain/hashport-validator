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
	"fmt"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"testing"
)

var (
	testConfig = config.Database{
		Host:     "localhost",
		Name:     "hedera_validator",
		Password: "validator_pass",
		Port:     "5432",
		Username: "validator",
	}
)

func TestPrepareRepositories(t *testing.T) {
	mocks.Setup()
	res := PrepareRepositories(testConfig)
	fmt.Println(res)
}

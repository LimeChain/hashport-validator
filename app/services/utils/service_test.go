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

package utils

import (
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	svc *utilsService
)

func setup() {
	mocks.Setup()
}

func Test_ConvertEvmTxIdToHederaTxId(t *testing.T) {
	setup()
	in := "0xa83be7d95c58f57e11f5c27dedd963217d47bdeab897bc98f2f5410d9f6c0026"
	expected := "0xa83be7d95c58f57e11f5c27dedd963217d47bdeab897bc98f2f5410d9f6c0026-4"

	actual, err := svc.ConvertEvmHashToBridgeTxId(in, 80001)

	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
}

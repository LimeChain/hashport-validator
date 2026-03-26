/*
 * Copyright 2026 LimeChain Ltd.
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

package gas_price

import (
	"context"
	"errors"
	"math/big"
	"testing"

	mocksClient "github.com/limechain/hedera-eth-bridge-validator/test/mocks/client"
	"github.com/stretchr/testify/assert"
)

var (
	gasPriceClient *Client
	mockEVM        *mocksClient.MockEVM
)

func setup() {
	mockEVM = &mocksClient.MockEVM{}
	gasPriceClient = NewClient(mockEVM)
}

func TestGetGasPrice_Success(t *testing.T) {
	setup()
	expected := big.NewInt(20_000_000_000)
	mockEVM.On("SuggestGasPrice", context.Background()).Return(expected, nil)

	result, err := gasPriceClient.GetGasPrice(context.Background())

	assert.Nil(t, err)
	assert.Equal(t, expected, result)
	mockEVM.AssertExpectations(t)
}

func TestGetGasPrice_SuggestGasPriceFails(t *testing.T) {
	setup()
	mockEVM.On("SuggestGasPrice", context.Background()).Return((*big.Int)(nil), errors.New("rpc error"))

	result, err := gasPriceClient.GetGasPrice(context.Background())

	assert.Nil(t, result)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "rpc error")
	mockEVM.AssertExpectations(t)
}

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

package evm

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"math/big"
	"testing"
	"time"
)

var (
	c       *Client
	address = "0x0000000000000000000000000000000000000001"
)

func setup() {
	mocks.Setup()
	c = &Client{
		Core:   mocks.MEVMCoreClient,
		logger: config.GetLoggerFor("EVM Client"),
	}
}

func Test_ChainID(t *testing.T) {
	setup()
	mocks.MEVMCoreClient.On("ChainID", context.Background()).Return(big.NewInt(1), nil)
	chain, err := c.ChainID(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, big.NewInt(1), chain)
}

func Test_ValidateContractDeployedAt(t *testing.T) {
	setup()

	var nilBlockNumber *big.Int = nil
	mocks.MEVMCoreClient.On("CodeAt", context.Background(), common.HexToAddress(address), nilBlockNumber).Return([]byte{0x1}, nil)

	_, err := c.ValidateContractDeployedAt(address)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ValidateContractDeployedAt_CodeAtFails(t *testing.T) {
	setup()

	var nilBlockNumber *big.Int = nil
	mocks.MEVMCoreClient.On("CodeAt", context.Background(), common.HexToAddress(address), nilBlockNumber).Return(nil, errors.New("some-error"))

	result, err := c.ValidateContractDeployedAt(address)
	assert.NotNil(t, err)
	assert.Nil(t, result)
}

func Test_ValidateContractDeployedAt_NotASmartContract(t *testing.T) {
	setup()

	var nilBlockNumber *big.Int = nil
	mocks.MEVMCoreClient.On("CodeAt", context.Background(), common.HexToAddress(address), nilBlockNumber).Return([]byte{}, nil)

	result, err := c.ValidateContractDeployedAt(address)
	assert.NotNil(t, err)
	assert.Nil(t, result)
}

func Test_GetClient(t *testing.T) {
	setup()
	assert.Equal(t, c.Core, c.GetClient())
}

func Test_GetBlockTimestamp(t *testing.T) {
	setup()
	now := uint64(time.Now().Unix())
	blockNumber := big.NewInt(1)
	mocks.MEVMCoreClient.On("BlockByNumber", context.Background(), blockNumber).Return(types.NewBlockWithHeader(
		&types.Header{Time: now},
	), nil)
	ts := c.GetBlockTimestamp(blockNumber)
	assert.Equal(t, now, ts)
}

func Test_GetBlockTimestamp_Fails(t *testing.T) {
	setup()
	blockNumber := big.NewInt(1)
	now := uint64(time.Now().Unix())
	mocks.MEVMCoreClient.On("BlockByNumber", context.Background(), blockNumber).Return(nil, errors.New("some-error")).Once()
	mocks.MEVMCoreClient.On("BlockByNumber", context.Background(), blockNumber).Return(types.NewBlockWithHeader(
		&types.Header{Time: now},
	), nil)
	res := c.GetBlockTimestamp(blockNumber)
	assert.Equal(t, now, res)
}

func Test_CheckTransactionReceipt(t *testing.T) {
	setup()
	onSuccess := func() {
		fmt.Println("Successful.")
	}
	onRevert := func() {
		fmt.Println("Reverted.")
	}
	onError := func(err error) {
		fmt.Println("Error.", err)
	}

	hash := common.HexToHash(address)
	mocks.MEVMCoreClient.On("TransactionByHash", context.Background(), hash).Return(nil, false, nil)
	mocks.MEVMCoreClient.On("TransactionReceipt", context.Background(), hash).Return(&types.Receipt{Status: 1}, nil)
	c.checkTransactionReceipt(address, onSuccess, onRevert, onError)
}

func Test_WaitForTransactionReceipt_NotFound(t *testing.T) {
	setup()

	hash := common.HexToHash(address)
	mocks.MEVMCoreClient.On("TransactionByHash", context.Background(), hash).Return(nil, false, ethereum.NotFound)

	receipt, err := c.WaitForTransactionReceipt(hash)
	assert.Error(t, ethereum.NotFound, err)
	assert.Nil(t, receipt)
}

func Test_CheckTransactionReceipt_Reverted(t *testing.T) {
	setup()
	onSuccess := func() {
		fmt.Println("Successful.")
	}
	onRevert := func() {
		fmt.Println("Reverted.")
	}
	onError := func(err error) {
		fmt.Println("Error.", err)
	}

	hash := common.HexToHash(address)
	mocks.MEVMCoreClient.On("TransactionByHash", context.Background(), hash).Return(nil, false, nil)
	mocks.MEVMCoreClient.On("TransactionReceipt", context.Background(), hash).Return(&types.Receipt{Status: 2}, nil)
	c.checkTransactionReceipt(address, onSuccess, onRevert, onError)
}

func Test_CheckTransactionReceipt_Fails(t *testing.T) {
	setup()
	onSuccess := func() {
		fmt.Println("Successful.")
	}
	onRevert := func() {
		fmt.Println("Reverted.")
	}
	onError := func(err error) {
		fmt.Println("Error.", err)
	}

	hash := common.HexToHash(address)
	mocks.MEVMCoreClient.On("TransactionByHash", context.Background(), hash).Return(nil, false, nil)
	mocks.MEVMCoreClient.On("TransactionReceipt", context.Background(), hash).Return(nil, errors.New("some-error"))
	c.checkTransactionReceipt(address, onSuccess, onRevert, onError)
}

func Test_GetPrivateKey(t *testing.T) {
	setup()
	assert.Equal(t, c.config.PrivateKey, c.GetPrivateKey())
}

func Test_WaitForConfirmations(t *testing.T) {
	setup()

	log := types.Log{
		BlockNumber: 20,
	}

	mocks.MEVMCoreClient.On("BlockNumber", context.Background()).Return(uint64(20), nil)
	mocks.MEVMCoreClient.On("TransactionReceipt", context.Background(), log.TxHash).Return(&types.Receipt{
		BlockNumber: big.NewInt(20),
	}, nil)

	err := c.WaitForConfirmations(log)
	assert.Nil(t, err)
}

func Test_WaitForConfirmations_MovedFromOriginalBlock(t *testing.T) {
	setup()

	log := types.Log{
		BlockNumber: 19,
	}

	mocks.MEVMCoreClient.On("BlockNumber", context.Background()).Return(uint64(20), nil)
	mocks.MEVMCoreClient.On("TransactionReceipt", context.Background(), log.TxHash).Return(&types.Receipt{
		BlockNumber: big.NewInt(20),
	}, nil)

	err := c.WaitForConfirmations(log)
	assert.Error(t, errors.New("moved from original block"), err)
}

func Test_WaitForConfirmations_TransactionReceipt_EthereumNotFound(t *testing.T) {
	setup()

	log := types.Log{}

	mocks.MEVMCoreClient.On("BlockNumber", context.Background()).Return(uint64(20), nil)
	mocks.MEVMCoreClient.On("TransactionReceipt", context.Background(), log.TxHash).Return(&types.Receipt{}, ethereum.NotFound)

	err := c.WaitForConfirmations(log)
	assert.Error(t, ethereum.NotFound, err)
}

func Test_WaitForConfirmations_TransactionReceipt_OtherError(t *testing.T) {
	setup()

	log := types.Log{}

	mocks.MEVMCoreClient.On("BlockNumber", context.Background()).Return(uint64(20), nil)
	mocks.MEVMCoreClient.On("TransactionReceipt", context.Background(), log.TxHash).Return(&types.Receipt{}, errors.New("some-error"))

	err := c.WaitForConfirmations(log)
	assert.Error(t, errors.New("some-error"), err)
}

func Test_WaitForConfirmations_BlockNumberFails(t *testing.T) {
	setup()

	mocks.MEVMCoreClient.On("BlockNumber", context.Background()).Return(uint64(0), errors.New("some-error"))

	err := c.WaitForConfirmations(types.Log{})
	assert.NotNil(t, err)
	mocks.MEVMCoreClient.AssertNotCalled(t, "TransactionReceipt", context.Background(), mock.Anything)
}

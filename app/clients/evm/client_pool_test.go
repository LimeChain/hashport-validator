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
	"github.com/ethereum/go-ethereum/common"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"

	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"

	"github.com/limechain/hedera-eth-bridge-validator/config"

	"errors"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/stretchr/testify/mock"

	"time"
)

var (
	cp      *ClientPool
	retries = 3
)

func setupCP() {
	setup()
	evmList := make([]client.EVM, 0)
	evmList = append(evmList, c)

	cp = &ClientPool{
		clients: evmList,
		retries: retries,
	}
}

func TestNewClientPool(t *testing.T) {
	nodeUrls := []string{"http://localhost:8545", "http://localhost:8546"}
	configEvmPool := config.EvmPool{
		BlockConfirmations: 3,
		NodeUrls:           nodeUrls,
		PrivateKey:         "0x000000000",
		StartBlock:         88,
		PollingInterval:    5,
		MaxLogsBlocks:      10,
	}

	configEvm := config.Evm{
		BlockConfirmations: configEvmPool.BlockConfirmations,
		NodeUrl:            configEvmPool.NodeUrls[0],
		PrivateKey:         configEvmPool.PrivateKey,
		StartBlock:         configEvmPool.StartBlock,
		PollingInterval:    configEvmPool.PollingInterval,
		MaxLogsBlocks:      configEvmPool.MaxLogsBlocks,
	}

	client := NewClient(configEvm, 256)
	clientPool := NewClientPool(configEvmPool, 256)
	assert.Equal(t, 6, clientPool.retries)

	assert.Equal(t, client.GetChainID(), clientPool.GetChainID())
	assert.Equal(t, client.GetPrivateKey(), clientPool.GetPrivateKey())
}

func TestClientPool_SetChainID(t *testing.T) {
	setupCP()
	cp.SetChainID(chainId)
	assert.Equal(t, chainId, cp.GetChainID())
}

func TestClientPool_GetChainID(t *testing.T) {
	setupCP()
	cp.SetChainID(chainId)
	assert.Equal(t, chainId, cp.GetChainID())
}

func TestClientPool_ChainID(t *testing.T) {
	setupCP()
	mocks.MEVMCoreClient.On("ChainID", context.Background()).Return(big.NewInt(1), nil)
	chain, err := cp.ChainID(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, big.NewInt(1), chain)
}

func TestClientPool_ValidateContractDeployedAt(t *testing.T) {
	setupCP()

	var nilBlockNumber *big.Int = nil
	mocks.MEVMCoreClient.On("CodeAt", context.Background(), common.HexToAddress(address), nilBlockNumber).Return([]byte{0x1}, nil)

	_, err := cp.ValidateContractDeployedAt(address)
	if err != nil {
		t.Fatal(err)
	}
}

func TestClientPool_ValidateContractDeployedAt_CodeAtFails(t *testing.T) {
	setupCP()

	var nilBlockNumber *big.Int = nil
	mocks.MEVMCoreClient.On("CodeAt", context.Background(), common.HexToAddress(address), nilBlockNumber).Return(nil, errors.New("some-error"))

	result, err := cp.ValidateContractDeployedAt(address)
	assert.NotNil(t, err)
	assert.Nil(t, result)
}

func TestClientPool_ValidateContractDeployedAt_NotASmartContract(t *testing.T) {
	setupCP()

	var nilBlockNumber *big.Int = nil
	mocks.MEVMCoreClient.On("CodeAt", context.Background(), common.HexToAddress(address), nilBlockNumber).Return([]byte{}, nil)

	result, err := cp.ValidateContractDeployedAt(address)
	assert.NotNil(t, err)
	assert.Nil(t, result)
}

func TestClientPool_GetClient(t *testing.T) {
	setupCP()
	// assert.Equal(t, cp.Core, cp.GetClient())
	assert.Equal(t, c.Core, cp.GetClient())
}

func TestClientPool_GetBlockTimestamp(t *testing.T) {
	setupCP()
	now := uint64(time.Now().Unix())
	blockNumber := big.NewInt(1)
	mocks.MEVMCoreClient.On("HeaderByNumber", context.Background(), blockNumber).Return(&types.Header{Time: now}, nil)
	ts := cp.GetBlockTimestamp(blockNumber)
	assert.Equal(t, now, ts)
}

func TestClientPool_GetBlockTimestamp_Fails(t *testing.T) {
	setupCP()
	blockNumber := big.NewInt(1)
	now := uint64(time.Now().Unix())
	mocks.MEVMCoreClient.On("HeaderByNumber", context.Background(), blockNumber).Return(nil, errors.New("some-error")).Once()
	mocks.MEVMCoreClient.On("HeaderByNumber", context.Background(), blockNumber).Return(&types.Header{Time: now}, nil)
	res := cp.GetBlockTimestamp(blockNumber)
	assert.Equal(t, now, res)
}

func TestClientPool_WaitForTransactionReceipt_NotFound(t *testing.T) {
	setupCP()

	hash := common.HexToHash(address)
	mocks.MEVMCoreClient.On("TransactionByHash", context.Background(), hash).Return(nil, false, ethereum.NotFound)

	receipt, err := cp.WaitForTransactionReceipt(hash)
	assert.Error(t, ethereum.NotFound, err)
	assert.Nil(t, receipt)
}

func TestClientPool_GetPrivateKey(t *testing.T) {
	setupCP()
	assert.Equal(t, c.config.PrivateKey, cp.GetPrivateKey())
}

func TestClientPool_WaitForConfirmations(t *testing.T) {
	setupCP()

	log := types.Log{
		BlockNumber: 20,
	}

	mocks.MEVMCoreClient.On("BlockNumber", context.Background()).Return(uint64(20), nil)
	mocks.MEVMCoreClient.On("TransactionReceipt", context.Background(), log.TxHash).Return(&types.Receipt{
		BlockNumber: big.NewInt(20),
	}, nil)

	err := cp.WaitForConfirmations(log)
	assert.Nil(t, err)
}

func TestClientPool_WaitForConfirmations_MovedFromOriginalBlock(t *testing.T) {
	setupCP()

	log := types.Log{
		BlockNumber: 19,
	}

	mocks.MEVMCoreClient.On("BlockNumber", context.Background()).Return(uint64(20), nil)
	mocks.MEVMCoreClient.On("TransactionReceipt", context.Background(), log.TxHash).Return(&types.Receipt{
		BlockNumber: big.NewInt(20),
	}, nil)

	err := cp.WaitForConfirmations(log)
	assert.Error(t, errors.New("moved from original block"), err)
}

func TestClientPool_WaitForConfirmations_TransactionReceipt_EthereumNotFound(t *testing.T) {
	setupCP()

	log := types.Log{}

	mocks.MEVMCoreClient.On("BlockNumber", context.Background()).Return(uint64(20), nil)
	mocks.MEVMCoreClient.On("TransactionReceipt", context.Background(), log.TxHash).Return(&types.Receipt{}, ethereum.NotFound)

	err := cp.WaitForConfirmations(log)
	assert.Error(t, ethereum.NotFound, err)
}

func TestClientPool_WaitForConfirmations_TransactionReceipt_OtherError(t *testing.T) {
	setupCP()

	log := types.Log{}

	mocks.MEVMCoreClient.On("BlockNumber", context.Background()).Return(uint64(20), nil)
	mocks.MEVMCoreClient.On("TransactionReceipt", context.Background(), log.TxHash).Return(&types.Receipt{}, errors.New("some-error"))

	err := cp.WaitForConfirmations(log)
	assert.Error(t, errors.New("some-error"), err)
}

func TestClientPool_WaitForConfirmations_BlockNumberFails(t *testing.T) {
	setupCP()

	mocks.MEVMCoreClient.On("BlockNumber", context.Background()).Return(uint64(0), errors.New("some-error"))

	err := cp.WaitForConfirmations(types.Log{})
	assert.NotNil(t, err)
	mocks.MEVMCoreClient.AssertNotCalled(t, "TransactionReceipt", context.Background(), mock.Anything)
}

func TestClientPool_CodeAt(t *testing.T) {
	setupCP()
	ctx := context.TODO()
	contractAddress := common.HexToAddress(address)
	blockNumber := big.NewInt(1)
	code := []byte{1, 2, 3, 4}
	mocks.MEVMCoreClient.On("CodeAt", ctx, contractAddress, blockNumber).Return([]byte(nil), errors.New("error")).Once().
		On("CodeAt", ctx, contractAddress, blockNumber).Return([]byte(nil), errors.New("error")).Once().
		On("CodeAt", ctx, contractAddress, blockNumber).Return(code, nil).Once()

	res, err := cp.CodeAt(ctx, contractAddress, blockNumber)

	assert.NoError(t, err)
	assert.Equal(t, code, res)
	mocks.MEVMCoreClient.AssertNumberOfCalls(t, "CodeAt", 3)
}

func TestClientPool_HeaderByNumber(t *testing.T) {
	setupCP()
	ctx := context.TODO()
	number := big.NewInt(1)
	header := &types.Header{}
	mocks.MEVMCoreClient.On("HeaderByNumber", ctx, number).Return((*types.Header)(nil), errors.New("error")).Once().
		On("HeaderByNumber", ctx, number).Return((*types.Header)(nil), errors.New("error")).Once().
		On("HeaderByNumber", ctx, number).Return(header, nil).Once()

	res, err := cp.HeaderByNumber(ctx, number)

	assert.NoError(t, err)
	assert.Equal(t, header, res)
	mocks.MEVMCoreClient.AssertNumberOfCalls(t, "HeaderByNumber", 3)
}

func TestClientPool_SuggestGasPrice(t *testing.T) {
	setupCP()
	ctx := context.TODO()
	price := big.NewInt(1)
	mocks.MEVMCoreClient.On("SuggestGasPrice", ctx).Return(nil, errors.New("error")).Once().
		On("SuggestGasPrice", ctx).Return(nil, errors.New("error")).Once().
		On("SuggestGasPrice", ctx).Return(price, nil).Once()

	res, err := cp.SuggestGasPrice(ctx)

	assert.NoError(t, err)
	assert.Equal(t, price, res)
	mocks.MEVMCoreClient.AssertNumberOfCalls(t, "SuggestGasPrice", 3)
}

func TestClientPool_SuggestGasTipCap(t *testing.T) {
	setupCP()
	ctx := context.TODO()
	cap := big.NewInt(1)
	mocks.MEVMCoreClient.On("SuggestGasTipCap", ctx).Return(nil, errors.New("error")).Once().
		On("SuggestGasTipCap", ctx).Return(nil, errors.New("error")).Once().
		On("SuggestGasTipCap", ctx).Return(cap, nil).Once()

	res, err := cp.SuggestGasTipCap(ctx)

	assert.NoError(t, err)
	assert.Equal(t, cap, res)
	mocks.MEVMCoreClient.AssertNumberOfCalls(t, "SuggestGasTipCap", 3)
}

func TestClientPool_EstimateGas(t *testing.T) {
	setupCP()
	ctx := context.TODO()
	call := ethereum.CallMsg{}
	gas := uint64(1)
	mocks.MEVMCoreClient.On("EstimateGas", ctx, call).Return(uint64(0), errors.New("error")).Once().
		On("EstimateGas", ctx, call).Return(uint64(0), errors.New("error")).Once().
		On("EstimateGas", ctx, call).Return(gas, nil).Once()

	res, err := cp.EstimateGas(ctx, call)

	assert.NoError(t, err)
	assert.Equal(t, gas, res)
	mocks.MEVMCoreClient.AssertNumberOfCalls(t, "EstimateGas", 3)
}

func TestClientPool_SendTransaction(t *testing.T) {
	setupCP()
	ctx := context.TODO()
	tx := &types.Transaction{}
	mocks.MEVMCoreClient.On("SendTransaction", ctx, tx).Return(errors.New("error")).Once().
		On("SendTransaction", ctx, tx).Return(errors.New("error")).Once().
		On("SendTransaction", ctx, tx).Return(nil).Once()

	err := cp.SendTransaction(ctx, tx)

	assert.NoError(t, err)
	mocks.MEVMCoreClient.AssertNumberOfCalls(t, "SendTransaction", 3)
}

func TestClientPool_FilterLogs(t *testing.T) {
	setupCP()
	ctx := context.TODO()
	query := ethereum.FilterQuery{}
	logs := []types.Log{{}}
	mocks.MEVMCoreClient.On("FilterLogs", ctx, query).Return(nil, errors.New("error")).Once().
		On("FilterLogs", ctx, query).Return(nil, errors.New("error")).Once().
		On("FilterLogs", ctx, query).Return(logs, nil).Once()

	res, err := cp.FilterLogs(ctx, query)

	assert.NoError(t, err)
	assert.Equal(t, logs, res)
	mocks.MEVMCoreClient.AssertNumberOfCalls(t, "FilterLogs", 3)
}

func TestClientPool_SubscribeFilterLogs(t *testing.T) {
	setupCP()
	ctx := context.TODO()
	query := ethereum.FilterQuery{}
	ch := make(chan<- types.Log)
	sub := new(MockSubscription)
	mocks.MEVMCoreClient.On("SubscribeFilterLogs", ctx, query, ch).Return(nil, errors.New("error")).Once().
		On("SubscribeFilterLogs", ctx, query, ch).Return(nil, errors.New("error")).Once().
		On("SubscribeFilterLogs", ctx, query, ch).Return(sub, nil).Once()

	res, err := cp.SubscribeFilterLogs(ctx, query, ch)

	assert.NoError(t, err)
	assert.Equal(t, sub, res)
	mocks.MEVMCoreClient.AssertNumberOfCalls(t, "SubscribeFilterLogs", 3)
}

func TestClientPool_BlockNumber(t *testing.T) {
	setupCP()
	ctx := context.TODO()
	number := uint64(1)
	mocks.MEVMCoreClient.On("BlockNumber", ctx).Return(uint64(0), errors.New("error")).Once().
		On("BlockNumber", ctx).Return(uint64(0), errors.New("error")).Once().
		On("BlockNumber", ctx).Return(number, nil).Once()

	res, err := cp.BlockNumber(ctx)

	assert.NoError(t, err)
	assert.Equal(t, number, res)
	mocks.MEVMCoreClient.AssertNumberOfCalls(t, "BlockNumber", 3)
}

type MockSubscription struct {
	mock.Mock
}

func (m *MockSubscription) Unsubscribe() {
	m.Called()
}

func (m *MockSubscription) Err() <-chan error {
	args := m.Called()
	return args.Get(0).(chan error)
}

func TestClientPool_PendingCodeAt(t *testing.T) {
	setupCP()
	ctx := context.TODO()
	account := common.HexToAddress("0x123")
	expectedCode := []byte{1, 2, 3}
	mocks.MEVMCoreClient.On("PendingCodeAt", ctx, account).Return(nil, errors.New("error")).Once().
		On("PendingCodeAt", ctx, account).Return(nil, errors.New("error")).Once().
		On("PendingCodeAt", ctx, account).Return(expectedCode, nil).Once()

	actualCode, err := cp.PendingCodeAt(ctx, account)

	assert.NoError(t, err)
	assert.Equal(t, expectedCode, actualCode)
	mocks.MEVMCoreClient.AssertNumberOfCalls(t, "PendingCodeAt", 3)
}

func TestClientPool_PendingNonceAt(t *testing.T) {
	setupCP()
	ctx := context.TODO()
	account := common.HexToAddress("0x123")
	expectedNonce := uint64(3)
	mocks.MEVMCoreClient.On("PendingNonceAt", ctx, account).Return(uint64(0), errors.New("error")).Once().
		On("PendingNonceAt", ctx, account).Return(uint64(0), errors.New("error")).Once().
		On("PendingNonceAt", ctx, account).Return(expectedNonce, nil).Once()

	actualNonce, err := cp.PendingNonceAt(ctx, account)

	assert.NoError(t, err)
	assert.Equal(t, expectedNonce, actualNonce)
	mocks.MEVMCoreClient.AssertNumberOfCalls(t, "PendingNonceAt", 3)
}

func TestClientPool_CallContract(t *testing.T) {
	setupCP()
	ctx := context.TODO()
	call := ethereum.CallMsg{}
	blockNumber := big.NewInt(1)
	expectedResult := []byte{1, 2, 3}
	mocks.MEVMCoreClient.On("CallContract", ctx, call, blockNumber).Return(nil, errors.New("error")).Once().
		On("CallContract", ctx, call, blockNumber).Return(nil, errors.New("error")).Once().
		On("CallContract", ctx, call, blockNumber).Return(expectedResult, nil).Once()

	actualResult, err := cp.CallContract(ctx, call, blockNumber)

	assert.NoError(t, err)
	assert.Equal(t, expectedResult, actualResult)
	mocks.MEVMCoreClient.AssertNumberOfCalls(t, "CallContract", 3)
}
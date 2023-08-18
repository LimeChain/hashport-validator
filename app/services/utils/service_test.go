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
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/retry"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

var (
	svc              *utilsService
	lockHash         common.Hash
	burnHash         common.Hash
	burnErc721       common.Hash
	someHash         common.Hash
	evmTx            = "0xa83be7d95c58f57e11f5c27dedd963217d47bdeab897bc98f2f5410d9f6c0026"
	evmTxHash        = common.HexToHash(evmTx)
	expectedBridgeTx = "0.0.1-123-123"
	expectedResult   = &service.BridgeTxId{
		BridgeTxId: expectedBridgeTx,
	}
	mockReceipt = &types.Receipt{
		Logs: []*types.Log{
			{
				Topics: []common.Hash{
					someHash,
				},
				Index: 1,
			},
			{
				Topics: []common.Hash{
					someHash,
				},
				Index: 2,
			},
			{
				Topics: []common.Hash{
					someHash,
				},
				Index: 3,
			},
			{
				Topics: []common.Hash{
					someHash,
				},
				Index: 4,
			},
		},
	}
	mockErr = errors.New("some error")
)

func setup() {
	mocks.Setup()

	routerAbi, _ := abi.JSON(strings.NewReader(router.RouterABI))
	burnHash = routerAbi.Events["Burn"].ID
	lockHash = routerAbi.Events["Lock"].ID
	burnErc721 = routerAbi.Events["BurnERC721"].ID

	svc = &utilsService{
		evmClients: map[uint64]client.EVM{
			80001: mocks.MEVMClient,
		},
		burnEvt:        mocks.MBurnService,
		burnHash:       burnHash,
		burnErc721Hash: burnErc721,
		lockHash:       lockHash,
		log:            config.GetLoggerFor("Utils Service"),
	}
}

func Test_New(t *testing.T) {
	setup()

	actual := New(
		map[uint64]client.EVM{
			80001: mocks.MEVMClient,
		},
		mocks.MBurnService)

	assert.Equal(t, svc, actual)
}

func Test_ConvertEvmHashToBridgeTxId_LockEvent(t *testing.T) {
	setup()
	mockReceipt.Logs[3].Topics[0] = lockHash
	mocks.MEVMClient.On("WaitForTransactionReceipt", evmTxHash).Return(mockReceipt, nil)
	mocks.MBurnService.On("TransactionID", fmt.Sprintf("%s-4", evmTx)).Return(expectedBridgeTx, nil)

	actual, err := svc.ConvertEvmHashToBridgeTxId(evmTx, 80001)

	assert.Nil(t, err)
	assert.Equal(t, expectedResult, actual)
}

func Test_ConvertEvmHashToBridgeTxId_BurnEvent(t *testing.T) {
	setup()
	mockReceipt.Logs[3].Topics[0] = burnHash
	mocks.MEVMClient.On("WaitForTransactionReceipt", evmTxHash).Return(mockReceipt, nil)
	mocks.MBurnService.On("TransactionID", fmt.Sprintf("%s-4", evmTx)).Return(expectedBridgeTx, nil)

	actual, err := svc.ConvertEvmHashToBridgeTxId(evmTx, 80001)

	assert.Nil(t, err)
	assert.Equal(t, expectedResult, actual)
}

func Test_ConvertEvmHashToBridgeTxId_BurnErc721Event(t *testing.T) {
	setup()
	mockReceipt.Logs[3].Topics[0] = burnErc721
	mocks.MEVMClient.On("WaitForTransactionReceipt", evmTxHash).Return(mockReceipt, nil)
	mocks.MBurnService.On("TransactionID", fmt.Sprintf("%s-4", evmTx)).Return(expectedBridgeTx, nil)

	actual, err := svc.ConvertEvmHashToBridgeTxId(evmTx, 80001)

	assert.Nil(t, err)
	assert.Equal(t, expectedResult, actual)
}

func Test_ConvertEvmHashToBridgeTxId_WithErrorFromWaitForTransactionReceipt(t *testing.T) {
	setup()
	mocks.MEVMClient.On("WaitForTransactionReceipt", evmTxHash).Return(nil, mockErr)

	actual, err := svc.ConvertEvmHashToBridgeTxId(evmTx, 80001)

	assert.NotNil(t, err)
	assert.Nil(t, actual)
}

func Test_ConvertEvmHashToBridgeTxId_WithErrorFromTransactionID(t *testing.T) {
	setup()
	mocks.MEVMClient.On("WaitForTransactionReceipt", evmTxHash).Return(mockReceipt, nil)
	mocks.MBurnService.On("TransactionID", fmt.Sprintf("%s-4", evmTx)).Return(nil, mockErr)

	actual, err := svc.ConvertEvmHashToBridgeTxId(evmTx, 80001)

	assert.NotNil(t, err)
	assert.Nil(t, actual)
}

func Test_ConvertEvmHashToBridgeTxId_WithNotFoundEventLog(t *testing.T) {
	setup()
	mockReceipt.Logs[3].Topics[0] = someHash
	mocks.MEVMClient.On("WaitForTransactionReceipt", evmTxHash).Return(mockReceipt, nil)

	actual, err := svc.ConvertEvmHashToBridgeTxId(evmTx, 80001)

	assert.NotNil(t, err)
	assert.Equal(t, service.ErrNotFound, err)
	assert.Nil(t, actual)
}

func Test_ConvertEvmHashToBridgeTxId_WithInvalidChainId(t *testing.T) {
	setup()
	delete(svc.evmClients, 80001)

	actual, err := svc.ConvertEvmHashToBridgeTxId(evmTx, 80001)

	assert.NotNil(t, err)
	assert.Nil(t, actual)
}

func Test_Retry_HappyPath(t *testing.T) {
	setup()

	expectedValue := 1

	res, err := service.Retry(func(ctx context.Context) retry.Result {
		select {
		case <-time.After(1 * time.Second): // Simulate work
		case <-ctx.Done():
			return retry.Result{
				Value: nil,
				Error: ctx.Err(),
			}
		}
		return retry.Result{
			Value: expectedValue,
			Error: nil,
		}
	}, 1)

	require.NoError(t, err)
	require.Equal(t, expectedValue, res)
}

func Test_Retry_SuccessAfterRetry(t *testing.T) {
	setup()

	expectedValue := 1
	currentRun := 0

	res, err := service.Retry(func(ctx context.Context) retry.Result {
		currentRun++

		waitTime := 10 * time.Second
		if currentRun > 1 {
			waitTime = 1 * time.Second
		}

		select {
		case <-time.After(waitTime): // Simulate work
		case <-ctx.Done():
			return retry.Result{
				Value: nil,
				Error: ctx.Err(),
			}
		}
		return retry.Result{
			Value: expectedValue,
			Error: nil,
		}
	}, 2)

	require.NoError(t, err)
	require.Equal(t, expectedValue, res)
	require.Equal(t, currentRun, 2)
}

func Test_Retry_Timeout(t *testing.T) {
	setup()

	expectedValue := 1

	res, err := service.Retry(func(ctx context.Context) retry.Result {
		waitTime := 10 * time.Second

		select {
		case <-time.After(waitTime): // Simulate work
		case <-ctx.Done():
			return retry.Result{
				Value: nil,
				Error: ctx.Err(),
			}
		}
		return retry.Result{
			Value: expectedValue,
			Error: nil,
		}
	}, 1)

	require.Error(t, err)
	require.Nil(t, res)
	require.ErrorIs(t, err, service.ErrTooManyRetires)
}

func Test_Retry_ReturnError(t *testing.T) {
	setup()

	currentRun := 0
	res, err := service.Retry(func(ctx context.Context) retry.Result {
		currentRun++
		return retry.Result{
			Value: nil,
			Error: errors.New("some error"),
		}
	}, 3)

	require.Error(t, err)
	require.Nil(t, res)
	require.Equal(t, err.Error(), "some error")
	require.Equal(t, currentRun, 1)
}

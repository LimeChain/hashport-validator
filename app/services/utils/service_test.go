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
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
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

	actual := New(map[uint64]client.EVM{
		80001: mocks.MEVMClient,
	}, mocks.MBurnService)

	assert.Equal(t, svc, actual)
}

func Test_ConvertEvmTxIdToHederaTxId_LockEvent(t *testing.T) {
	setup()
	mockReceipt.Logs[3].Topics[0] = lockHash
	mocks.MEVMClient.On("WaitForTransactionReceipt", evmTxHash).Return(mockReceipt, nil)
	mocks.MBurnService.On("TransactionID", fmt.Sprintf("%s-4", evmTx)).Return(expectedBridgeTx, nil)

	actual, err := svc.ConvertEvmHashToBridgeTxId(evmTx, 80001)

	assert.Nil(t, err)
	assert.Equal(t, expectedResult, actual)
}

func Test_ConvertEvmTxIdToHederaTxId_BurnEvent(t *testing.T) {
	setup()
	mockReceipt.Logs[3].Topics[0] = burnHash
	mocks.MEVMClient.On("WaitForTransactionReceipt", evmTxHash).Return(mockReceipt, nil)
	mocks.MBurnService.On("TransactionID", fmt.Sprintf("%s-4", evmTx)).Return(expectedBridgeTx, nil)

	actual, err := svc.ConvertEvmHashToBridgeTxId(evmTx, 80001)

	assert.Nil(t, err)
	assert.Equal(t, expectedResult, actual)
}

func Test_ConvertEvmTxIdToHederaTxId_BurnErc721Event(t *testing.T) {
	setup()
	mockReceipt.Logs[3].Topics[0] = burnErc721
	mocks.MEVMClient.On("WaitForTransactionReceipt", evmTxHash).Return(mockReceipt, nil)
	mocks.MBurnService.On("TransactionID", fmt.Sprintf("%s-4", evmTx)).Return(expectedBridgeTx, nil)

	actual, err := svc.ConvertEvmHashToBridgeTxId(evmTx, 80001)

	assert.Nil(t, err)
	assert.Equal(t, expectedResult, actual)
}

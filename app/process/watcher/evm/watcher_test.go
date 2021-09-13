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

package evm

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/core/queue"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/mock"
	"math/big"
	"testing"
)

var (
	w       = &Watcher{}
	lockLog = &router.RouterLock{
		TargetChain: big.NewInt(0),
		Token:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
		Receiver:    hederaAcc.ToBytes(),
		Amount:      big.NewInt(1),
	}
	hederaAcc, _ = hedera.AccountIDFromString("0.0.123456")
	hederaBytes  = hederaAcc.ToBytes()
)

func Test_HandleLockLog_Removed_Fails(t *testing.T) {
	setup()

	lockLog.Raw.Removed = true
	w.handleLockLog(lockLog, mocks.MQueue)
	lockLog.Raw.Removed = false

	mocks.MEVMClient.AssertNotCalled(t, "WaitForConfirmations", lockLog.Raw)
	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything)
}

func Test_HandleLockLog_EmptyReceiver_Fails(t *testing.T) {
	setup()

	lockLog.Receiver = []byte{}
	w.handleLockLog(lockLog, mocks.MQueue)
	lockLog.Receiver = hederaBytes

	mocks.MEVMClient.AssertNotCalled(t, "WaitForConfirmations", lockLog.Raw)
	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything)
}

func Test_HandleLockLog_InvalidReceiver_Fails(t *testing.T) {
	setup()

	lockLog.Receiver = []byte{1}
	w.handleLockLog(lockLog, mocks.MQueue)
	lockLog.Receiver = hederaBytes

	mocks.MEVMClient.AssertNotCalled(t, "WaitForConfirmations", lockLog.Raw)
	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything)
}

func Test_HandleLockLog_EmptyWrappedAsset_Fails(t *testing.T) {
	setup()
	mocks.MEVMClient.On("ChainID").Return(big.NewInt(2))

	w.handleLockLog(lockLog, mocks.MQueue)

	mocks.MEVMClient.AssertNotCalled(t, "WaitForConfirmations", lockLog.Raw)
	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything)
}

func Test_HandleLockLog_WaitingForConfirmations_Fails(t *testing.T) {
	setup()
	mocks.MEVMClient.On("ChainID").Return(big.NewInt(33))
	mocks.MEVMClient.On("WaitForConfirmations", lockLog.Raw).Return(errors.New("some-error"))

	w.handleLockLog(lockLog, mocks.MQueue)

	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything)
}

func Test_HandleLockLog_HappyPath(t *testing.T) {
	setup()
	mocks.MEVMClient.On("ChainID").Return(big.NewInt(33))
	mocks.MEVMClient.On("WaitForConfirmations", lockLog.Raw).Return(nil)
	parsedLockLog := &transfer.Transfer{
		TransactionId: fmt.Sprintf("%s-%d", lockLog.Raw.TxHash, lockLog.Raw.Index),
		SourceChainId: int64(33),
		TargetChainId: lockLog.TargetChain.Int64(),
		NativeChainId: int64(33),
		SourceAsset:   lockLog.Token.String(),
		TargetAsset:   constants.Hbar,
		NativeAsset:   lockLog.Token.String(),
		Receiver:      hederaAcc.String(),
		Amount:        lockLog.Amount.String(),
		RouterAddress: "",
	}

	mocks.MStatusRepository.On("Update", mocks.MBridgeContractService.Address().String(), int64(0)).Return(nil)
	mocks.MQueue.On("Push", &queue.Message{Payload: parsedLockLog, Topic: constants.HederaMintHtsTransfer}).Return()

	w.handleLockLog(lockLog, mocks.MQueue)
}

func setup() {
	mocks.Setup()

	networks := map[int64]*parser.Network{
		2: {
			Tokens: map[string]parser.Token{
				"0x0000000000000000000000000000000000000000": {
					Networks: map[int64]string{
						0: "",
					},
				},
			},
		},
		3: {
			Tokens: map[string]parser.Token{
				"0x0000000000000000000000000000000000000000": {
					Networks: map[int64]string{
						0: "",
					},
				},
			},
		},
		32: {
			Tokens: map[string]parser.Token{
				"0x0000000000000000000000000000000000000000": {
					Networks: map[int64]string{
						0: "",
					},
				},
			},
		},
		33: {
			Tokens: map[string]parser.Token{
				"0x0000000000000000000000000000000000000000": {
					Networks: map[int64]string{
						0: constants.Hbar,
					},
				},
			}},
	}

	mocks.MStatusRepository.On("Get", mock.Anything).Return(int64(0), nil)
	w = &Watcher{
		repository: mocks.MStatusRepository,
		contracts:  mocks.MBridgeContractService,
		evmClient:  mocks.MEVMClient,
		logger:     config.GetLoggerFor("EVM Watcher"),
		mappings:   config.LoadAssets(networks),
		validator:  true,
	}
}

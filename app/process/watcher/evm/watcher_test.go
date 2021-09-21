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
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/core/queue"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
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
	burnLog = &router.RouterBurn{
		TargetChain: big.NewInt(0),
		Token:       common.HexToAddress("0x0000000000000000000000000000000000000001"),
		Receiver:    hederaAcc.ToBytes(),
		Amount:      big.NewInt(1),
	}
	hederaAcc, _ = hedera.AccountIDFromString("0.0.123456")
	hederaBytes  = hederaAcc.ToBytes()

	networks = map[int64]*parser.Network{
		0: {
			Tokens: map[string]parser.Token{
				constants.Hbar: {
					Networks: map[int64]string{
						33: "0x0000000000000000000000000000000000000001",
					},
				},
			},
		},
		1: {
			Tokens: map[string]parser.Token{
				"0xsomeethaddress": {
					Networks: map[int64]string{
						33: "0x0000000000000000000000000000000000000123",
					},
				},
			},
		},
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
						1: "0xsome-other-eth-address",
					},
				},
			}},
	}
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
	mocks.MEVMClient.On("ChainID", context.Background()).Return(big.NewInt(1), nil)

	lockLog.Receiver = []byte{1}
	w.handleLockLog(lockLog, mocks.MQueue)
	lockLog.Receiver = hederaBytes

	mocks.MEVMClient.AssertNotCalled(t, "WaitForConfirmations", lockLog.Raw)
	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything)
}

func Test_HandleLockLog_EmptyWrappedAsset_Fails(t *testing.T) {
	setup()
	mocks.MEVMClient.On("ChainID", context.Background()).Return(big.NewInt(2), nil)

	w.handleLockLog(lockLog, mocks.MQueue)

	mocks.MEVMClient.AssertNotCalled(t, "WaitForConfirmations", lockLog.Raw)
	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything)
}

func Test_HandleLockLog_WaitingForConfirmations_Fails(t *testing.T) {
	setup()
	mocks.MBridgeContractService.On("RemoveDecimals", lockLog.Amount, lockLog.Token).Return(lockLog.Amount, nil)
	mocks.MEVMClient.On("ChainID", context.Background()).Return(big.NewInt(33), nil)
	mocks.MEVMClient.On("WaitForConfirmations", lockLog.Raw).Return(errors.New("some-error"))

	w.handleLockLog(lockLog, mocks.MQueue)

	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything)
}

func Test_HandleLockLog_HappyPath(t *testing.T) {
	setup()
	mocks.MEVMClient.On("ChainID", context.Background()).Return(big.NewInt(33), nil)
	mocks.MBridgeContractService.On("RemoveDecimals", lockLog.Amount, lockLog.Token).Return(lockLog.Amount, nil)
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
	}

	mocks.MStatusRepository.On("Update", mocks.MBridgeContractService.Address().String(), int64(0)).Return(nil)
	mocks.MQueue.On("Push", &queue.Message{Payload: parsedLockLog, Topic: constants.HederaMintHtsTransfer}).Return()

	w.handleLockLog(lockLog, mocks.MQueue)
}

func Test_HandleLockLog_ReadOnlyHederaMintHtsTransfer(t *testing.T) {
	mocks.Setup()
	mocks.MBridgeContractService.On("RemoveDecimals", lockLog.Amount, lockLog.Token).Return(lockLog.Amount, nil)
	mocks.MEVMClient.On("GetBlockTimestamp", big.NewInt(0)).Return(uint64(1), nil)
	mocks.MStatusRepository.On("Get", mock.Anything).Return(int64(0), nil)

	w = &Watcher{
		repository: mocks.MStatusRepository,
		contracts:  mocks.MBridgeContractService,
		evmClient:  mocks.MEVMClient,
		logger:     config.GetLoggerFor("EVM Router Watcher [0x0000000000000000000000000000000000000000]"),
		mappings:   config.LoadAssets(networks),
		validator:  false,
	}

	mocks.MEVMClient.On("ChainID", context.Background()).Return(big.NewInt(33), nil)
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
		Timestamp:     "1",
	}

	mocks.MStatusRepository.On("Update", mocks.MBridgeContractService.Address().String(), int64(0)).Return(nil)
	mocks.MQueue.On("Push", &queue.Message{Payload: parsedLockLog, Topic: constants.ReadOnlyHederaMintHtsTransfer}).Return()

	w.handleLockLog(lockLog, mocks.MQueue)
}

func Test_HandleLockLog_ReadOnlyTransferSave(t *testing.T) {
	mocks.Setup()
	mocks.MEVMClient.On("GetBlockTimestamp", big.NewInt(0)).Return(uint64(1), nil)
	mocks.MStatusRepository.On("Get", mock.Anything).Return(int64(0), nil)

	lockLog.TargetChain = big.NewInt(1)
	w = &Watcher{
		repository: mocks.MStatusRepository,
		contracts:  mocks.MBridgeContractService,
		evmClient:  mocks.MEVMClient,
		logger:     config.GetLoggerFor("EVM Router Watcher [0x0000000000000000000000000000000000000000]"),
		mappings:   config.LoadAssets(networks),
		validator:  false,
	}

	mocks.MEVMClient.On("ChainID", context.Background()).Return(big.NewInt(33), nil)
	mocks.MEVMClient.On("WaitForConfirmations", lockLog.Raw).Return(nil)
	parsedLockLog := &transfer.Transfer{
		TransactionId: fmt.Sprintf("%s-%d", lockLog.Raw.TxHash, lockLog.Raw.Index),
		SourceChainId: int64(33),
		TargetChainId: lockLog.TargetChain.Int64(),
		NativeChainId: int64(33),
		SourceAsset:   lockLog.Token.String(),
		TargetAsset:   "0xsome-other-eth-address",
		NativeAsset:   lockLog.Token.String(),
		Receiver:      common.BytesToAddress(hederaAcc.ToBytes()).String(),
		Amount:        lockLog.Amount.String(),
		Timestamp:     "1",
	}

	mocks.MStatusRepository.On("Update", mocks.MBridgeContractService.Address().String(), int64(0)).Return(nil)
	mocks.MQueue.On("Push", &queue.Message{Payload: parsedLockLog, Topic: constants.ReadOnlyTransferSave}).Return()

	w.handleLockLog(lockLog, mocks.MQueue)
	lockLog.TargetChain = big.NewInt(0)
}

func Test_HandleLockLog_UpdateFails(t *testing.T) {
	setup()
	mocks.MBridgeContractService.On("RemoveDecimals", lockLog.Amount, lockLog.Token).Return(lockLog.Amount, nil)
	mocks.MEVMClient.On("ChainID", context.Background()).Return(big.NewInt(33), nil)
	mocks.MEVMClient.On("WaitForConfirmations", lockLog.Raw).Return(nil)
	mocks.MStatusRepository.On("Update", mocks.MBridgeContractService.Address().String(), int64(0)).Return(errors.New("some-error"))

	w.handleLockLog(lockLog, mocks.MQueue)

	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything)
}

func Test_HandleLockLog_TopicMessageSubmission(t *testing.T) {
	setup()
	mocks.MEVMClient.On("ChainID", context.Background()).Return(big.NewInt(33), nil)
	mocks.MEVMClient.On("WaitForConfirmations", lockLog.Raw).Return(nil)

	lockLog.TargetChain = big.NewInt(1)
	parsedLockLog := &transfer.Transfer{
		TransactionId: fmt.Sprintf("%s-%d", lockLog.Raw.TxHash, lockLog.Raw.Index),
		SourceChainId: int64(33),
		TargetChainId: lockLog.TargetChain.Int64(),
		NativeChainId: int64(33),
		SourceAsset:   lockLog.Token.String(),
		TargetAsset:   "0xsome-other-eth-address",
		NativeAsset:   lockLog.Token.String(),
		Receiver:      common.BytesToAddress(hederaAcc.ToBytes()).String(),
		Amount:        lockLog.Amount.String(),
	}

	mocks.MStatusRepository.On("Update", mocks.MBridgeContractService.Address().String(), int64(0)).Return(nil)
	mocks.MQueue.On("Push", &queue.Message{Payload: parsedLockLog, Topic: constants.TopicMessageSubmission}).Return()

	w.handleLockLog(lockLog, mocks.MQueue)
	lockLog.TargetChain = big.NewInt(0)
}

func Test_HandleBurnLog_HappyPath(t *testing.T) {
	setup()
	mocks.MBridgeContractService.On("RemoveDecimals", burnLog.Amount, burnLog.Token).Return(lockLog.Amount, nil)
	mocks.MEVMClient.On("ChainID", context.Background()).Return(big.NewInt(33), nil)
	mocks.MEVMClient.On("WaitForConfirmations", burnLog.Raw).Return(nil)
	parsedBurnLog := &transfer.Transfer{
		TransactionId: fmt.Sprintf("%s-%d", burnLog.Raw.TxHash, burnLog.Raw.Index),
		SourceChainId: int64(33),
		TargetChainId: burnLog.TargetChain.Int64(),
		NativeChainId: int64(0),
		SourceAsset:   burnLog.Token.String(),
		TargetAsset:   constants.Hbar,
		NativeAsset:   constants.Hbar,
		Receiver:      hederaAcc.String(),
		Amount:        burnLog.Amount.String(),
	}

	mocks.MStatusRepository.On("Update", mocks.MBridgeContractService.Address().String(), int64(0)).Return(nil)
	mocks.MQueue.On("Push", &queue.Message{Payload: parsedBurnLog, Topic: constants.HederaFeeTransfer}).Return()

	w.handleBurnLog(burnLog, mocks.MQueue)
}

func Test_HandleBurnLog_UpdateFails(t *testing.T) {
	setup()
	mocks.MBridgeContractService.On("RemoveDecimals", burnLog.Amount, burnLog.Token).Return(lockLog.Amount, nil)
	mocks.MEVMClient.On("ChainID", context.Background()).Return(big.NewInt(33), nil)
	mocks.MEVMClient.On("WaitForConfirmations", burnLog.Raw).Return(nil)
	mocks.MStatusRepository.On("Update", mocks.MBridgeContractService.Address().String(), int64(0)).Return(errors.New("some-error"))

	w.handleBurnLog(burnLog, mocks.MQueue)

	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything)
}

func Test_HandleBurnLog_WaitForConfirmationFails(t *testing.T) {
	setup()
	mocks.MBridgeContractService.On("RemoveDecimals", burnLog.Amount, burnLog.Token).Return(lockLog.Amount, nil)
	mocks.MEVMClient.On("ChainID", context.Background()).Return(big.NewInt(33), nil)
	mocks.MEVMClient.On("WaitForConfirmations", burnLog.Raw).Return(errors.New("some-error"))
	w.handleBurnLog(burnLog, mocks.MQueue)
}

func Test_HandleBurnLog_InvalidHederaRecipient(t *testing.T) {
	setup()
	defaultReceiver := burnLog.Receiver
	burnLog.Receiver = []byte{1, 2, 3, 4}
	mocks.MEVMClient.On("ChainID", context.Background()).Return(big.NewInt(33), nil)
	w.handleBurnLog(burnLog, mocks.MQueue)
	burnLog.Receiver = defaultReceiver
}

func Test_HandleBurnLog_TopicMessageSubmission(t *testing.T) {
	setup()
	mocks.MEVMClient.On("ChainID", context.Background()).Return(big.NewInt(33), nil)
	mocks.MEVMClient.On("WaitForConfirmations", burnLog.Raw).Return(nil)

	burnLog.TargetChain = big.NewInt(1)
	defaultToken := burnLog.Token
	burnLog.Token = common.HexToAddress("0x123")
	receiver := common.BytesToAddress(burnLog.Receiver).String()
	parsedBurnLog := &transfer.Transfer{
		TransactionId: fmt.Sprintf("%s-%d", burnLog.Raw.TxHash, burnLog.Raw.Index),
		SourceChainId: int64(33),
		TargetChainId: 1,
		NativeChainId: int64(1),
		SourceAsset:   burnLog.Token.String(),
		TargetAsset:   "0xsomeethaddress",
		NativeAsset:   "0xsomeethaddress",
		Receiver:      receiver,
		Amount:        burnLog.Amount.String(),
	}

	mocks.MStatusRepository.On("Update", mocks.MBridgeContractService.Address().String(), int64(0)).Return(nil)
	mocks.MQueue.On("Push", &queue.Message{Payload: parsedBurnLog, Topic: constants.TopicMessageSubmission}).Return()

	w.handleBurnLog(burnLog, mocks.MQueue)
	burnLog.TargetChain = big.NewInt(0)
	burnLog.Token = defaultToken
}

func Test_HandleBurnLog_ReadOnlyTransferSave(t *testing.T) {
	mocks.Setup()
	mocks.MEVMClient.On("GetBlockTimestamp", big.NewInt(0)).Return(uint64(1), nil)
	mocks.MStatusRepository.On("Get", mock.Anything).Return(int64(0), nil)

	burnLog.TargetChain = big.NewInt(1)
	w = &Watcher{
		repository: mocks.MStatusRepository,
		contracts:  mocks.MBridgeContractService,
		evmClient:  mocks.MEVMClient,
		logger:     config.GetLoggerFor("EVM Router Watcher [0x0000000000000000000000000000000000000000]"),
		mappings:   config.LoadAssets(networks),
		validator:  false,
	}

	mocks.MEVMClient.On("ChainID", context.Background()).Return(big.NewInt(33), nil)
	mocks.MEVMClient.On("WaitForConfirmations", burnLog.Raw).Return(nil)

	burnLog.TargetChain = big.NewInt(1)
	defaultToken := burnLog.Token
	burnLog.Token = common.HexToAddress("0x123")
	receiver := common.BytesToAddress(burnLog.Receiver).String()
	parsedBurnLog := &transfer.Transfer{
		TransactionId: fmt.Sprintf("%s-%d", burnLog.Raw.TxHash, burnLog.Raw.Index),
		SourceChainId: int64(33),
		TargetChainId: 1,
		NativeChainId: int64(1),
		SourceAsset:   burnLog.Token.String(),
		TargetAsset:   "0xsomeethaddress",
		NativeAsset:   "0xsomeethaddress",
		Receiver:      receiver,
		Amount:        burnLog.Amount.String(),
		Timestamp:     "1",
	}

	mocks.MStatusRepository.On("Update", mocks.MBridgeContractService.Address().String(), int64(0)).Return(nil)
	mocks.MQueue.On("Push", &queue.Message{Payload: parsedBurnLog, Topic: constants.ReadOnlyTransferSave}).Return()

	w.handleBurnLog(burnLog, mocks.MQueue)
	burnLog.TargetChain = big.NewInt(0)
	burnLog.Token = defaultToken
}

func Test_HandleBurnLog_ReadOnlyHederaTransfer(t *testing.T) {
	mocks.Setup()
	mocks.MBridgeContractService.On("RemoveDecimals", burnLog.Amount, burnLog.Token).Return(lockLog.Amount, nil)
	mocks.MEVMClient.On("GetBlockTimestamp", big.NewInt(0)).Return(uint64(1), nil)
	mocks.MStatusRepository.On("Get", mock.Anything).Return(int64(0), nil)
	w = &Watcher{
		repository: mocks.MStatusRepository,
		contracts:  mocks.MBridgeContractService,
		evmClient:  mocks.MEVMClient,
		logger:     config.GetLoggerFor("EVM Router Watcher [0x0000000000000000000000000000000000000000]"),
		mappings:   config.LoadAssets(networks),
		validator:  false,
	}

	mocks.MEVMClient.On("ChainID", context.Background()).Return(big.NewInt(33), nil)
	mocks.MEVMClient.On("WaitForConfirmations", burnLog.Raw).Return(nil)
	parsedBurnLog := &transfer.Transfer{
		TransactionId: fmt.Sprintf("%s-%d", burnLog.Raw.TxHash, burnLog.Raw.Index),
		SourceChainId: int64(33),
		TargetChainId: 0,
		NativeChainId: int64(0),
		SourceAsset:   burnLog.Token.String(),
		TargetAsset:   constants.Hbar,
		NativeAsset:   constants.Hbar,
		Receiver:      hederaAcc.String(),
		Amount:        burnLog.Amount.String(),
		Timestamp:     "1",
	}

	mocks.MStatusRepository.On("Update", mocks.MBridgeContractService.Address().String(), int64(0)).Return(nil)
	mocks.MQueue.On("Push", &queue.Message{Payload: parsedBurnLog, Topic: constants.ReadOnlyHederaTransfer}).Return()

	w.handleBurnLog(burnLog, mocks.MQueue)
}

func Test_HandleBurnLog_GetBlockTimestamp_Fails(t *testing.T) {
	mocks.Setup()
	mocks.MEVMClient.On("GetBlockTimestamp", big.NewInt(0)).Return(uint64(0), errors.New("some-error"))
	mocks.MBridgeContractService.On("RemoveDecimals", burnLog.Amount, burnLog.Token).Return(lockLog.Amount, nil)
	mocks.MStatusRepository.On("Get", mock.Anything).Return(int64(0), nil)
	w = &Watcher{
		repository: mocks.MStatusRepository,
		contracts:  mocks.MBridgeContractService,
		evmClient:  mocks.MEVMClient,
		logger:     config.GetLoggerFor("EVM Router Watcher [0x0000000000000000000000000000000000000000]"),
		mappings:   config.LoadAssets(networks),
		validator:  false,
	}

	mocks.MEVMClient.On("ChainID", context.Background()).Return(big.NewInt(33), nil)
	mocks.MEVMClient.On("WaitForConfirmations", burnLog.Raw).Return(nil)
	mocks.MStatusRepository.On("Update", mocks.MBridgeContractService.Address().String(), int64(0)).Return(nil)

	w.handleBurnLog(burnLog, mocks.MQueue)

	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything)
}

func Test_HandleLockLog_GetBlockTimestamp_Fails(t *testing.T) {
	mocks.Setup()
	mocks.MBridgeContractService.On("RemoveDecimals", lockLog.Amount, lockLog.Token).Return(lockLog.Amount, nil)
	mocks.MEVMClient.On("GetBlockTimestamp", big.NewInt(0)).Return(uint64(0), errors.New("some-error"))
	mocks.MStatusRepository.On("Get", mock.Anything).Return(int64(0), nil)
	w = &Watcher{
		repository: mocks.MStatusRepository,
		contracts:  mocks.MBridgeContractService,
		evmClient:  mocks.MEVMClient,
		logger:     config.GetLoggerFor("EVM Router Watcher [0x0000000000000000000000000000000000000000]"),
		mappings:   config.LoadAssets(networks),
		validator:  false,
	}

	mocks.MEVMClient.On("ChainID", context.Background()).Return(big.NewInt(33), nil)
	mocks.MEVMClient.On("WaitForConfirmations", lockLog.Raw).Return(nil)
	mocks.MStatusRepository.On("Update", mocks.MBridgeContractService.Address().String(), int64(0)).Return(nil)

	w.handleLockLog(lockLog, mocks.MQueue)

	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything)
}

func Test_ListenForEvents_BurnSubscription_Fails(t *testing.T) {
	setup()
	var opts *bind.WatchOpts = nil
	mocks.MBridgeContractService.On("WatchBurnEventLogs", opts, mock.Anything).Return(nil, errors.New("some-error"))
	w.listenForEvents(mocks.MQueue)
}

func Test_ListenForEvents_LockSubscription_Fails(t *testing.T) {
	setup()
	var opts *bind.WatchOpts = nil
	mocks.MBridgeContractService.On("WatchBurnEventLogs", opts, mock.Anything).Return(nil, nil)
	mocks.MBridgeContractService.On("WatchLockEventLogs", opts, mock.Anything).Return(nil, errors.New("some-error"))
	w.listenForEvents(mocks.MQueue)
}

func Test_HandleBurnLog_Token_Not_Supported(t *testing.T) {
	setup()
	mocks.MEVMClient.On("ChainID", context.Background()).Return(big.NewInt(33), nil)

	defaultToken := burnLog.Token
	burnLog.Token = common.HexToAddress("0x0123123")
	w.handleBurnLog(burnLog, mocks.MQueue)
	mocks.MStatusRepository.AssertNotCalled(t, "Update", mocks.MBridgeContractService.Address().String(), int64(0))
	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything)
	burnLog.Token = defaultToken
}

func Test_HandleBurnLog_WrappedToWrapped_Not_Supported(t *testing.T) {
	setup()
	mocks.MEVMClient.On("ChainID", context.Background()).Return(big.NewInt(33), nil)

	defaultTargetChain := burnLog.TargetChain
	burnLog.TargetChain = big.NewInt(1)
	w.handleBurnLog(burnLog, mocks.MQueue)
	mocks.MStatusRepository.AssertNotCalled(t, "Update", mocks.MBridgeContractService.Address().String(), int64(0))
	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything)
	burnLog.TargetChain = defaultTargetChain
}

func Test_HandleBurnLog_Raw_Removed(t *testing.T) {
	setup()
	burnLog.Raw.Removed = true

	w.handleBurnLog(burnLog, mocks.MQueue)

	mocks.MStatusRepository.AssertNotCalled(t, "Update", mocks.MBridgeContractService.Address().String(), int64(0))
	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything)
	burnLog.Raw.Removed = false
}

func Test_HandleBurnLog_No_Receivers(t *testing.T) {
	setup()
	receiver := burnLog.Receiver
	burnLog.Receiver = []byte{}

	w.handleBurnLog(burnLog, mocks.MQueue)

	mocks.MStatusRepository.AssertNotCalled(t, "Update", mocks.MBridgeContractService.Address().String(), int64(0))
	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything)
	burnLog.Receiver = receiver
}

func TestNewWatcher(t *testing.T) {
	mocks.Setup()

	mocks.MStatusRepository.On("Get", mock.Anything).Return(int64(0), nil)
	mocks.MEVMClient.On("BlockNumber", mock.Anything).Return(uint64(0), nil)

	assets := config.LoadAssets(networks)
	w = &Watcher{
		repository: mocks.MStatusRepository,
		contracts:  mocks.MBridgeContractService,
		evmClient:  mocks.MEVMClient,
		logger:     config.GetLoggerFor("EVM Router Watcher [0x0000000000000000000000000000000000000000]"),
		mappings:   assets,
		validator:  true,
	}

	assert.EqualValues(t, w, NewWatcher(mocks.MStatusRepository, mocks.MBridgeContractService, mocks.MEVMClient, assets, 0, true))
}

// TODO: Test_NewWatcher_Fails

func Test_ProcessPastLogs_ParseBurnLogFails(t *testing.T) {
	setup()

	burnHash := common.HexToHash("97715804dcd62a721835eaba4356dc90eaf6d442a12fe944f01bbf5f8c0b8992")
	lockHash := common.HexToHash("aa3a3bc72b8c754ca6ee8425a5531bafec37569ec012d62d5f682ca909ae06f1")
	topics := [][]common.Hash{
		{
			burnHash,
			lockHash,
		},
	}
	query := &ethereum.FilterQuery{
		FromBlock: new(big.Int).SetInt64(0),
		Addresses: []common.Address{
			common.HexToAddress("0x0000000000000000000000000000000000000000"),
		},
		Topics: topics,
	}

	mocks.MEVMClient.On("FilterLogs", context.Background(), *query).
		Return([]types.Log{
			{
				Topics: []common.Hash{
					burnHash,
				},
			},
		}, nil)

	mocks.MBridgeContractService.On("ParseBurnLog", types.Log{
		Topics: []common.Hash{
			burnHash,
		},
	}).Return(burnLog, errors.New("some-error"))
	w.processPastLogs(mocks.MQueue)
}

func Test_ProcessPastLogs_ParseLockLogFails(t *testing.T) {
	setup()

	burnHash := common.HexToHash("97715804dcd62a721835eaba4356dc90eaf6d442a12fe944f01bbf5f8c0b8992")
	lockHash := common.HexToHash("aa3a3bc72b8c754ca6ee8425a5531bafec37569ec012d62d5f682ca909ae06f1")
	topics := [][]common.Hash{
		{
			burnHash,
			lockHash,
		},
	}
	query := &ethereum.FilterQuery{
		FromBlock: new(big.Int).SetInt64(0),
		Addresses: []common.Address{
			common.HexToAddress("0x0000000000000000000000000000000000000000"),
		},
		Topics: topics,
	}

	mocks.MEVMClient.On("FilterLogs", context.Background(), *query).
		Return([]types.Log{
			{
				Topics: []common.Hash{
					lockHash,
				},
			},
		}, nil)

	mocks.MBridgeContractService.On("ParseLockLog", types.Log{
		Topics: []common.Hash{
			lockHash,
		},
	}).Return(lockLog, errors.New("some-error"))
	w.processPastLogs(mocks.MQueue)
}

func Test_ProcessPastLogs_FilterLogsFails(t *testing.T) {
	setup()

	burnHash := common.HexToHash("97715804dcd62a721835eaba4356dc90eaf6d442a12fe944f01bbf5f8c0b8992")
	lockHash := common.HexToHash("aa3a3bc72b8c754ca6ee8425a5531bafec37569ec012d62d5f682ca909ae06f1")
	topics := [][]common.Hash{
		{
			burnHash,
			lockHash,
		},
	}
	query := &ethereum.FilterQuery{
		FromBlock: new(big.Int).SetInt64(0),
		Addresses: []common.Address{
			common.HexToAddress("0x0000000000000000000000000000000000000000"),
		},
		Topics: topics,
	}

	mocks.MEVMClient.On("FilterLogs", context.Background(), *query).
		Return([]types.Log{}, errors.New("some-error"))

	w.processPastLogs(mocks.MQueue)
}

func Test_ProcessPastLogs_RepoGetFails(t *testing.T) {
	mocks.Setup()
	mocks.MStatusRepository.On("Get", mock.Anything).Return(int64(0), errors.New("some-error"))
	w = &Watcher{
		repository: mocks.MStatusRepository,
		contracts:  mocks.MBridgeContractService,
		evmClient:  mocks.MEVMClient,
		logger:     config.GetLoggerFor("EVM Router Watcher [0x0000000000000000000000000000000000000000]"),
		mappings:   config.LoadAssets(networks),
		validator:  true,
	}
	w.processPastLogs(mocks.MQueue)
}

func setup() {
	mocks.Setup()

	mocks.MStatusRepository.On("Get", mock.Anything).Return(int64(0), nil)
	w = &Watcher{
		repository: mocks.MStatusRepository,
		contracts:  mocks.MBridgeContractService,
		evmClient:  mocks.MEVMClient,
		logger:     config.GetLoggerFor("EVM Router Watcher [0x0000000000000000000000000000000000000000]"),
		mappings:   config.LoadAssets(networks),
		validator:  true,
	}
}

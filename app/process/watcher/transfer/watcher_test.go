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

package cryptotransfer

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model"
	iservice "github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"math/big"
	"testing"
	"time"
)

var (
	tx = model.Transaction{
		TokenTransfers: []model.Transfer{
			{
				Account: "0.0.444444",
				Amount:  10,
				Token:   "0.0.111111",
			},
		},
		ConsensusTimestamp: "1631092491.483966000",
	}
	networks = map[int64]*parser.Network{
		0: {
			Tokens: map[string]parser.Token{
				"0.0.111111": {
					Networks: map[int64]string{
						3: "0x0000000000000000000000000000000000000001",
					},
				},
			},
		},
	}
	assets = config.LoadAssets(networks)
)

func Test_NewMemo_MissingWrappedCorrelation(t *testing.T) {
	w := initializeWatcher()
	mocks.MTransferService.On("SanityCheckTransfer", mock.Anything).Return(int64(0), "0xevmaddress", nil)

	w.processTransaction(tx, mocks.MQueue)
	mocks.MTransferService.AssertCalled(t, "SanityCheckTransfer", tx)
	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything)
}

// TODO: Uncomment after unit test infrastructure refactor
//func Test_NewMemo_CorrectCorrelation(t *testing.T) {
//	w := initializeWatcher()
//	mocks.MTransferService.On("SanityCheckTransfer", mock.Anything).Return(int64(3), "0xevmaddress", nil)
//	mocks.MQueue.On("Push", mock.Anything).Return()
//
//	w.processTransaction(tx, mocks.MQueue)
//	mocks.MTransferService.AssertCalled(t, "SanityCheckTransfer", tx)
//	mocks.MQueue.AssertCalled(t, "Push", mock.Anything)
//}

//func Test_NewMemo_CorrectCorrelation_OnlyWrappedAssets(t *testing.T) {
//	w := initializeWatcher()
//	mocks.MTransferService.On("SanityCheckTransfer", mock.Anything).Return(int64(3), "0xevmaddress", nil)
//	mocks.MQueue.On("Push", mock.Anything).Return()
//
//	w.processTransaction(tx, mocks.MQueue)
//	mocks.MTransferService.AssertCalled(t, "SanityCheckTransfer", tx)
//	mocks.MQueue.AssertCalled(t, "Push", mock.Anything)
//}

// TODO: uncomment when log.Fatalf is defered properly
//func Test_NewWatcher_RecordNotFound_Fails(t *testing.T) {
//	mocks.Setup()
//	mocks.MStatusRepository.On("Get", mock.Anything).Return(int64(0), gorm.ErrRecordNotFound)
//	mocks.MStatusRepository.On("Create", mock.Anything, mock.Anything).Return(errors.New("some-error"))
//
//	NewWatcher(
//		mocks.MTransferService,
//		mocks.MHederaMirrorClient,
//		"0.0.444444",
//		5,
//		mocks.MStatusRepository,
//		0,
//		map[int64]iservice.Contracts{3: mocks.MBridgeContractService, 0: mocks.MBridgeContractService},
//		assets,
//		true)
//
//	mocks.MStatusRepository.AssertCalled(t, "Create", "0.0.444444", mock.Anything)
//}
//
//func Test_NewWatcher_GetError_Fails(t *testing.T) {
//	mocks.Setup()
//	mocks.MStatusRepository.On("Get", mock.Anything).Return(int64(0), errors.New("some-error"))
//
//	NewWatcher(
//		mocks.MTransferService,
//		mocks.MHederaMirrorClient,
//		"0.0.444444",
//		5,
//		mocks.MStatusRepository,
//		0,
//		map[int64]iservice.Contracts{3: mocks.MBridgeContractService, 0: mocks.MBridgeContractService},
//		assets,
//		true)
//}

func Test_NewWatcher_RecordNotFound_Creates(t *testing.T) {
	mocks.Setup()
	mocks.MStatusRepository.On("Get", mock.Anything).Return(int64(0), gorm.ErrRecordNotFound)
	mocks.MStatusRepository.On("Create", mock.Anything, mock.Anything).Return(nil)

	NewWatcher(
		mocks.MTransferService,
		mocks.MHederaMirrorClient,
		"0.0.444444",
		5,
		mocks.MStatusRepository,
		0,
		map[int64]iservice.Contracts{3: mocks.MBridgeContractService, 0: mocks.MBridgeContractService},
		assets,
		true)

	mocks.MStatusRepository.AssertCalled(t, "Create", "0.0.444444", mock.Anything)
}

func Test_NewWatcher_NotNilTS_Works(t *testing.T) {
	mocks.Setup()
	mocks.MStatusRepository.On("Update", "0.0.444444", mock.Anything).Return(nil)

	NewWatcher(
		mocks.MTransferService,
		mocks.MHederaMirrorClient,
		"0.0.444444",
		5,
		mocks.MStatusRepository,
		1,
		map[int64]iservice.Contracts{3: mocks.MBridgeContractService, 0: mocks.MBridgeContractService},
		assets,
		true)

	mocks.MStatusRepository.AssertCalled(t, "Update", "0.0.444444", mock.Anything)
}

//TODO: same
//func Test_NewWatcher_NotNilTS_Update_Fails(t *testing.T) {
//	mocks.Setup()
//	mocks.MStatusRepository.On("Update", "0.0.444444", mock.Anything).Return(errors.New("some-error"))
//
//	NewWatcher(
//		mocks.MTransferService,
//		mocks.MHederaMirrorClient,
//		"0.0.444444",
//		5,
//		mocks.MStatusRepository,
//		1,
//		map[int64]iservice.Contracts{3: mocks.MBridgeContractService, 0: mocks.MBridgeContractService},
//		assets,
//		true)
//
//	mocks.MStatusRepository.AssertCalled(t, "Update", "0.0.444444", mock.Anything)
//}

func Test_Watch_AccountNotExist(t *testing.T) {
	w := initializeWatcher()
	hederaAcc := hedera.AccountID{
		Shard:   0,
		Realm:   0,
		Account: 444444,
	}
	mocks.MHederaMirrorClient.On("AccountExists", hederaAcc).Return(false)
	w.Watch(mocks.MQueue)
}

//TODO: SAME
//func Test_BeginWatching(t *testing.T) {
//	w := initializeWatcher()
//	hederaAcc := hedera.AccountID{
//		Shard:   0,
//		Realm:   0,
//		Account: 444444,
//	}
//	transactions := &mirror_node.Response{
//		Transactions: []mirror_node.Transaction{},
//		Status:       mirror_node.Status{},
//	}
//	mocks.MHederaMirrorClient.On("GetAccountCreditTransactionsAfterTimestamp", hederaAcc, mock.Anything).
//		Return(transactions, nil)
//	w.beginWatching(mocks.MQueue)
//}

func Test_ProcessTransaction(t *testing.T) {
	w := initializeWatcher()
	mocks.MTransferService.On("SanityCheckTransfer", tx).Return(int64(3), "0xaiskdjakdjakl", nil)
	mocks.MQueue.On("Push", mock.Anything).Return()
	mocks.MBridgeContractService.On("AddDecimals", big.NewInt(10), common.HexToAddress("0x000001")).Return(big.NewInt(10), nil)
	w.processTransaction(tx, mocks.MQueue)
}

func Test_ProcessTransaction_WithTS(t *testing.T) {
	w := initializeWatcher()
	anotherTx := tx
	anotherTx.ConsensusTimestamp = fmt.Sprintf("%d.0", time.Now().Add(time.Hour).Unix())
	mocks.MTransferService.On("SanityCheckTransfer", anotherTx).Return(int64(3), "0xaiskdjakdjakl", nil)
	mocks.MBridgeContractService.On("AddDecimals", big.NewInt(10), common.HexToAddress("0x000001")).Return(big.NewInt(10), nil)
	mocks.MQueue.On("Push", mock.Anything).Return()
	w.processTransaction(anotherTx, mocks.MQueue)
}

func Test_UpdateStatusTimestamp_Works(t *testing.T) {
	w := initializeWatcher()
	mocks.MStatusRepository.On("Update", "0.0.444444", int64(100)).Return(nil)
	w.updateStatusTimestamp(100)
}

func Test_ProcessTransaction_SanityCheckTransfer_Fails(t *testing.T) {
	w := initializeWatcher()
	mocks.MTransferService.On("SanityCheckTransfer", tx).Return(int64(0), "", errors.New("some-error"))

	w.processTransaction(tx, mocks.MQueue)

	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything)
}

func Test_ProcessTransaction_GetIncomingTransfer_Fails(t *testing.T) {
	w := initializeWatcher()
	anotherTx := tx
	anotherTx.Transfers = []model.Transfer{}
	anotherTx.TokenTransfers = []model.Transfer{}
	w.processTransaction(anotherTx, mocks.MQueue)

	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything)
	mocks.MTransferService.AssertNotCalled(t, "SanityCheckTransfer", mock.Anything)
}

func Test_ConsensusTimestamp_Fails(t *testing.T) {
	w := initializeWatcher()
	anotherTx := tx
	anotherTx.ConsensusTimestamp = "asd"
	mocks.MTransferService.On("SanityCheckTransfer", anotherTx).Return(int64(3), "0xaiskdjakdjakl", nil)
	mocks.MBridgeContractService.On("AddDecimals", big.NewInt(10), common.HexToAddress("0x000001")).Return(big.NewInt(10), nil)
	mocks.MQueue.On("Push", mock.Anything).Return()
	w.processTransaction(anotherTx, mocks.MQueue)
}

func initializeWatcher() *Watcher {
	mocks.Setup()

	mocks.MStatusRepository.On("Get", mock.Anything).Return(int64(0), nil)

	return NewWatcher(
		mocks.MTransferService,
		mocks.MHederaMirrorClient,
		"0.0.444444",
		5,
		mocks.MStatusRepository,
		0,
		map[int64]iservice.Contracts{3: mocks.MBridgeContractService, 0: mocks.MBridgeContractService},
		assets,
		true)
}

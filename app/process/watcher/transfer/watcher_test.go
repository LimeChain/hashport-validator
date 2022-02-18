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

package cryptotransfer

import (
	"errors"
	"fmt"
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
	networks = map[uint64]*parser.Network{
		0: {
			Tokens: parser.Tokens{
				Fungible: map[string]parser.Token{
					"0.0.111111": {
						Networks: map[uint64]string{
							3: "0x0000000000000000000000000000000000000001",
						},
					},
				},
			},
		},
	}
	assets = config.LoadAssets(networks)
)

func Test_NewMemo_MissingWrappedCorrelation(t *testing.T) {
	w := initializeWatcher()
	mocks.MHederaMirrorClient.On("GetSuccessfulTransaction", tx.TransactionID).Return(tx, nil)
	mocks.MTransferService.On("SanityCheckTransfer", mock.Anything).Return(uint64(0), "0xevmaddress", nil)

	w.processTransaction(tx.TransactionID, mocks.MQueue)
	mocks.MTransferService.AssertCalled(t, "SanityCheckTransfer", tx)
	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything)
}

func Test_NewWatcher_RecordNotFound_Creates(t *testing.T) {
	setup()
	mocks.MStatusRepository.On("Get", mock.Anything).Return(int64(0), gorm.ErrRecordNotFound)
	mocks.MStatusRepository.On("Create", mock.Anything, mock.Anything).Return(nil)

	NewWatcher(
		mocks.MTransferService,
		mocks.MHederaMirrorClient,
		"0.0.444444",
		5,
		mocks.MStatusRepository,
		0,
		map[uint64]iservice.Contracts{3: mocks.MBridgeContractService, 0: mocks.MBridgeContractService},
		assets,
		map[string]int64{},
		true,
		mocks.MPrometheusService)

	mocks.MStatusRepository.AssertCalled(t, "Create", "0.0.444444", mock.Anything)
}

func Test_NewWatcher_NotNilTS_Works(t *testing.T) {
	setup()
	mocks.MStatusRepository.On("Update", "0.0.444444", mock.Anything).Return(nil)

	NewWatcher(
		mocks.MTransferService,
		mocks.MHederaMirrorClient,
		"0.0.444444",
		5,
		mocks.MStatusRepository,
		1,
		map[uint64]iservice.Contracts{3: mocks.MBridgeContractService, 0: mocks.MBridgeContractService},
		assets,
		map[string]int64{},
		true,
		mocks.MPrometheusService)

	mocks.MStatusRepository.AssertCalled(t, "Update", "0.0.444444", mock.Anything)
}

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

func Test_ProcessTransaction(t *testing.T) {
	w := initializeWatcher()
	mocks.MHederaMirrorClient.On("GetSuccessfulTransaction", tx.TransactionID).Return(tx, nil)
	mocks.MTransferService.On("SanityCheckTransfer", tx).Return(uint64(3), "0xaiskdjakdjakl", nil)
	mocks.MQueue.On("Push", mock.Anything).Return()
	mocks.MBridgeContractService.On("AddDecimals", big.NewInt(10), "0x0000000000000000000000000000000000000001").Return(big.NewInt(10), nil)
	w.processTransaction(tx.TransactionID, mocks.MQueue)
}

func Test_ProcessTransaction_WithTS(t *testing.T) {
	w := initializeWatcher()
	anotherTx := tx
	anotherTx.ConsensusTimestamp = fmt.Sprintf("%d.0", time.Now().Add(time.Hour).Unix())
	mocks.MHederaMirrorClient.On("GetSuccessfulTransaction", anotherTx.TransactionID).Return(anotherTx, nil)
	mocks.MTransferService.On("SanityCheckTransfer", anotherTx).Return(uint64(3), "0xaiskdjakdjakl", nil)
	mocks.MBridgeContractService.On("AddDecimals", big.NewInt(10), "0x0000000000000000000000000000000000000001").Return(big.NewInt(10), nil)
	mocks.MQueue.On("Push", mock.Anything).Return()
	w.processTransaction(anotherTx.TransactionID, mocks.MQueue)
}

func Test_UpdateStatusTimestamp_Works(t *testing.T) {
	w := initializeWatcher()
	mocks.MStatusRepository.On("Update", "0.0.444444", int64(100)).Return(nil)
	w.updateStatusTimestamp(100)
}

func Test_ProcessTransaction_SanityCheckTransfer_Fails(t *testing.T) {
	w := initializeWatcher()
	mocks.MHederaMirrorClient.On("GetSuccessfulTransaction", tx.TransactionID).Return(tx, nil)
	mocks.MTransferService.On("SanityCheckTransfer", tx).Return(uint64(0), "", errors.New("some-error"))

	w.processTransaction(tx.TransactionID, mocks.MQueue)

	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything)
}

func Test_ProcessTransaction_GetIncomingTransfer_Fails(t *testing.T) {
	w := initializeWatcher()
	anotherTx := tx
	anotherTx.Transfers = []model.Transfer{}
	anotherTx.TokenTransfers = []model.Transfer{}
	mocks.MHederaMirrorClient.On("GetSuccessfulTransaction", anotherTx.TransactionID).Return(anotherTx, nil)
	w.processTransaction(anotherTx.TransactionID, mocks.MQueue)

	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything)
	mocks.MTransferService.AssertNotCalled(t, "SanityCheckTransfer", mock.Anything)
}

func Test_ConsensusTimestamp_Fails(t *testing.T) {
	w := initializeWatcher()
	anotherTx := tx
	anotherTx.ConsensusTimestamp = "asd"
	mocks.MHederaMirrorClient.On("GetSuccessfulTransaction", anotherTx.TransactionID).Return(anotherTx, nil)
	mocks.MTransferService.On("SanityCheckTransfer", anotherTx).Return(uint64(3), "0xaiskdjakdjakl", nil)
	mocks.MBridgeContractService.On("AddDecimals", big.NewInt(10), "0x0000000000000000000000000000000000000001").Return(big.NewInt(10), nil)
	mocks.MQueue.On("Push", mock.Anything).Return()
	w.processTransaction(anotherTx.TransactionID, mocks.MQueue)
}

func setup() {
	mocks.Setup()
	mocks.MPrometheusService.On("GetIsMonitoringEnabled").Return(false)
}

func initializeWatcher() *Watcher {
	setup()

	mocks.MStatusRepository.On("Get", mock.Anything).Return(int64(0), nil)

	return NewWatcher(
		mocks.MTransferService,
		mocks.MHederaMirrorClient,
		"0.0.444444",
		5,
		mocks.MStatusRepository,
		0,
		map[uint64]iservice.Contracts{3: mocks.MBridgeContractService, 0: mocks.MBridgeContractService},
		assets,
		map[string]int64{},
		true,
		mocks.MPrometheusService)
}

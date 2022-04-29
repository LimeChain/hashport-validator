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
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/transaction"
	iservice "github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/asset"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/pricing"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"math/big"
	"testing"
	"time"
)

var (
	nativeTokenAddressNetwork0  = "0.0.111111"
	wrappedTokenAddressNetwork3 = "0x0000000000000000000000000000000000000001"
	network0                    = constants.HederaNetworkId
	network3                    = uint64(3)
	emptyString                 = ""
	nilNativeAsset              *asset.NativeAsset
	nativeAssetNetwork0         = &asset.NativeAsset{ChainId: constants.HederaNetworkId, Asset: nativeTokenAddressNetwork0}
	fungibleAssetInfoNetwork0   = &asset.FungibleAssetInfo{Decimals: 8}
	fungibleAssetInfoNetwork3   = &asset.FungibleAssetInfo{Decimals: 18}
	tokenPriceInfo              = pricing.TokenPriceInfo{decimal.NewFromFloat(20), big.NewInt(10000)}

	tx = transaction.Transaction{
		TokenTransfers: []transaction.Transfer{
			{
				Account: "0.0.444444",
				Amount:  10,
				Token:   nativeTokenAddressNetwork0,
			},
		},
		ConsensusTimestamp: "1631092491.483966000",
	}

	networks = map[uint64]*parser.Network{
		network0: {
			Tokens: parser.Tokens{
				Fungible: map[string]parser.Token{
					nativeTokenAddressNetwork0: {
						Networks: map[uint64]string{
							network3: wrappedTokenAddressNetwork3,
						},
					},
				},
			},
		},
	}
)

func Test_NewMemo_MissingWrappedCorrelation(t *testing.T) {
	w := initializeWatcher()
	mocks.MHederaMirrorClient.On("GetSuccessfulTransaction", tx.TransactionID).Return(tx, nil)
	mocks.MTransferService.On("SanityCheckTransfer", mock.Anything).Return(network3, emptyString, nil)
	mocks.MPrometheusService.On("GetIsMonitoringEnabled").Return(false)
	mocks.MAssetsService.On("NativeToWrapped", nativeTokenAddressNetwork0, network0, network3).Return(emptyString)
	mocks.MAssetsService.On("WrappedToNative", nativeTokenAddressNetwork0, network0).Return(nilNativeAsset)

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
		mocks.MAssetsService,
		map[string]int64{},
		true,
		mocks.MPrometheusService,
		mocks.MPricingService)

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
		mocks.MAssetsService,
		map[string]int64{},
		true,
		mocks.MPrometheusService,
		mocks.MPricingService)

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
	mocks.MPrometheusService.On("GetIsMonitoringEnabled").Return(false)
	mocks.MAssetsService.On("NativeToWrapped", nativeTokenAddressNetwork0, network0, network3).Return(wrappedTokenAddressNetwork3)
	mocks.MAssetsService.On("FungibleNativeAsset", network0, nativeTokenAddressNetwork0).Return(nativeAssetNetwork0)
	mocks.MPricingService.On("GetTokenPriceInfo", network0, nativeTokenAddressNetwork0).Return(tokenPriceInfo, true)
	mocks.MAssetsService.On("FungibleAssetInfo", network0, nativeTokenAddressNetwork0).Return(fungibleAssetInfoNetwork0, true)
	mocks.MAssetsService.On("FungibleAssetInfo", network3, wrappedTokenAddressNetwork3).Return(fungibleAssetInfoNetwork3, true)

	w.processTransaction(tx.TransactionID, mocks.MQueue)
}

func Test_ProcessTransaction_WithTS(t *testing.T) {
	w := initializeWatcher()
	anotherTx := tx
	anotherTx.ConsensusTimestamp = fmt.Sprintf("%d.0", time.Now().Add(time.Hour).Unix())
	mocks.MHederaMirrorClient.On("GetSuccessfulTransaction", anotherTx.TransactionID).Return(anotherTx, nil)
	mocks.MTransferService.On("SanityCheckTransfer", anotherTx).Return(uint64(3), "0xaiskdjakdjakl", nil)
	mocks.MPrometheusService.On("GetIsMonitoringEnabled").Return(false)
	mocks.MAssetsService.On("NativeToWrapped", nativeTokenAddressNetwork0, network0, network3).Return(wrappedTokenAddressNetwork3)
	mocks.MAssetsService.On("FungibleNativeAsset", network0, nativeTokenAddressNetwork0).Return(nativeAssetNetwork0)
	mocks.MPricingService.On("GetTokenPriceInfo", network0, nativeTokenAddressNetwork0).Return(tokenPriceInfo, true)
	mocks.MAssetsService.On("FungibleAssetInfo", network0, nativeTokenAddressNetwork0).Return(fungibleAssetInfoNetwork0, true)
	mocks.MAssetsService.On("FungibleAssetInfo", network3, wrappedTokenAddressNetwork3).Return(fungibleAssetInfoNetwork3, true)

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
	anotherTx.Transfers = []transaction.Transfer{}
	anotherTx.TokenTransfers = []transaction.Transfer{}
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
	mocks.MQueue.On("Push", mock.Anything).Return()
	mocks.MPrometheusService.On("GetIsMonitoringEnabled").Return(false)
	mocks.MAssetsService.On("NativeToWrapped", nativeTokenAddressNetwork0, network0, network3).Return(wrappedTokenAddressNetwork3)
	mocks.MAssetsService.On("FungibleNativeAsset", network0, nativeTokenAddressNetwork0).Return(nativeAssetNetwork0)
	mocks.MPricingService.On("GetTokenPriceInfo", network0, nativeTokenAddressNetwork0).Return(tokenPriceInfo, true)
	mocks.MAssetsService.On("FungibleAssetInfo", network0, nativeTokenAddressNetwork0).Return(fungibleAssetInfoNetwork0, true)
	mocks.MAssetsService.On("FungibleAssetInfo", network3, wrappedTokenAddressNetwork3).Return(fungibleAssetInfoNetwork3, true)

	w.processTransaction(anotherTx.TransactionID, mocks.MQueue)
}

func setup() {
	mocks.Setup()
	mocks.MPrometheusService.On("GetIsMonitoringEnabled").Return(false)
}

func initializeWatcher() *Watcher {
	setup()
	mocks.Setup()
	mocks.MStatusRepository.On("Get", mock.Anything).Return(int64(0), nil)

	return NewWatcher(
		mocks.MTransferService,
		mocks.MHederaMirrorClient,
		"0.0.444444",
		5,
		mocks.MStatusRepository,
		0,
		map[uint64]iservice.Contracts{3: mocks.MBridgeContractService, 0: mocks.MBridgeContractService},
		mocks.MAssetsService,
		map[string]int64{},
		true,
		mocks.MPrometheusService,
		mocks.MPricingService)
}

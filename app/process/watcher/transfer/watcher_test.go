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
	mirror_node "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	service2 "github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/mock"
	"testing"
)

var (
	tx = mirror_node.Transaction{
		TokenTransfers: []mirror_node.Transfer{
			{
				Account: "0.0.444444",
				Amount:  10,
				Token:   "0.0.111111",
			},
		},
	}
	onlyNativeToWrapped = config.AssetMappings{
		NativeToWrappedByNetwork: map[int64]config.Network{
			0: {
				Native: map[string]config.Mappings{
					"0.0.111111": {
						FeePercentage: 10000,
						Assets: map[int64]string{
							3: "0xevmaddress",
						},
					},
				},
			},
		},
	}
	onlyWrappedToNative = config.AssetMappings{
		NativeToWrappedByNetwork: map[int64]config.Network{
			3: {
				Native: map[string]config.Mappings{
					"0xevmaddress": {
						Assets: map[int64]string{
							0: "0.0.11111",
						},
					},
				},
			},
		},
		WrappedToNativeByNetwork: map[int64]map[string]*config.NativeAsset{
			0: {
				"0.0.111111": &config.NativeAsset{
					ChainId: 3,
					Asset:   "0xevmaddress",
				},
			},
		},
	}
)

func Test_NewMemo_MissingWrappedCorrelation(t *testing.T) {
	w := initializeWatcher()
	mocks.MTransferService.On("SanityCheckTransfer", mock.Anything).Return(int64(3), "0xevmaddress", nil)

	w.processTransaction(tx, mocks.MQueue)
	mocks.MTransferService.AssertCalled(t, "SanityCheckTransfer", tx)
	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything)
}

func Test_NewMemo_CorrectCorrelation(t *testing.T) {
	w := initializeWatcher()
	mocks.MTransferService.On("SanityCheckTransfer", mock.Anything).Return(int64(3), "0xevmaddress", nil)
	mocks.MQueue.On("Push", mock.Anything).Return()

	w.mappings = onlyNativeToWrapped

	w.processTransaction(tx, mocks.MQueue)
	mocks.MTransferService.AssertCalled(t, "SanityCheckTransfer", tx)
	mocks.MQueue.AssertCalled(t, "Push", mock.Anything)
}

func Test_NewMemo_CorrectCorrelation_OnlyWrappedAssets(t *testing.T) {
	w := initializeWatcher()
	mocks.MTransferService.On("SanityCheckTransfer", mock.Anything).Return(int64(3), "0xevmaddress", nil)
	mocks.MQueue.On("Push", mock.Anything).Return()

	w.mappings = onlyWrappedToNative

	w.processTransaction(tx, mocks.MQueue)
	mocks.MTransferService.AssertCalled(t, "SanityCheckTransfer", tx)
	mocks.MQueue.AssertCalled(t, "Push", mock.Anything)
}

func initializeWatcher() *Watcher {
	mocks.Setup()

	return NewWatcher(
		mocks.MTransferService,
		mocks.MHederaMirrorClient,
		"0.0.444444",
		5,
		mocks.MStatusRepository,
		0,
		map[int64]service2.Contracts{3: mocks.MBridgeContractService, 0: mocks.MBridgeContractService},
		config.AssetMappings{})
}

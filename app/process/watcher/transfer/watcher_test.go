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
				NativeAssets: map[string]map[int64]string{
					"0.0.111111": {
						3: "0xevmaddress",
					},
				},
			},
		},
	}
	onlyWrappedToNative = config.AssetMappings{
		WrappedToNativeByNetwork: map[int64]map[string]map[int64]string{
			0: {
				"0.0.111111": {
					3: "0xevmaddress",
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
		map[int64]service2.Contracts{3: mocks.MBridgeContractService},
		config.AssetMappings{})
}

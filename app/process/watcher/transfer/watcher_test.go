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
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	testConstants "github.com/limechain/hedera-eth-bridge-validator/test/constants"
	"github.com/stretchr/testify/assert"

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
)

var (
	nativeTokenAddressNetwork0  = "0.0.111111"
	wrappedTokenAddressNetwork3 = "0x0000000000000000000000000000000000000001"
	network0                    = constants.HederaNetworkId
	network3                    = uint64(3)
	evmAddress                  = "0xaiskdjakdjakl"
	emptyString                 = ""
	nilNativeAsset              *asset.NativeAsset
	nativeAssetNetwork0         = &asset.NativeAsset{ChainId: constants.HederaNetworkId, Asset: nativeTokenAddressNetwork0}
	fungibleAssetInfoNetwork0   = &asset.FungibleAssetInfo{Decimals: 8}
	fungibleAssetInfoNetwork3   = &asset.FungibleAssetInfo{Decimals: 18}
	tokenPriceInfo              = pricing.TokenPriceInfo{decimal.NewFromFloat(20), big.NewInt(10000), big.NewInt(10000)}
	txAccountId                 = "0.0.444444"
	txAmount                    = int64(10)

	tx = transaction.Transaction{
		TokenTransfers: []transaction.Transfer{
			{
				Account: txAccountId,
				Amount:  txAmount,
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
	mocks.MTransferService.On("SanityCheckTransfer", mock.Anything).Return(transfer.SanityCheckResult{ChainId: network3, EvmAddress: emptyString})
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
	blacklist := []string{"0.0.333", "0.0.444"}

	NewWatcher(
		mocks.MTransferService,
		mocks.MHederaMirrorClient,
		txAccountId,
		5,
		mocks.MStatusRepository,
		0,
		map[uint64]iservice.Contracts{3: mocks.MBridgeContractService, 0: mocks.MBridgeContractService},
		mocks.MAssetsService,
		true,
		mocks.MPrometheusService,
		mocks.MPricingService,
		blacklist,
	)

	mocks.MStatusRepository.AssertCalled(t, "Create", txAccountId, mock.Anything)
}

func Test_NewWatcher_NotNilTS_Works(t *testing.T) {
	setup()
	mocks.MStatusRepository.On("Update", txAccountId, mock.Anything).Return(nil)
	blacklist := []string{"0.0.333", "0.0.444"}

	NewWatcher(
		mocks.MTransferService,
		mocks.MHederaMirrorClient,
		txAccountId,
		5,
		mocks.MStatusRepository,
		1,
		map[uint64]iservice.Contracts{3: mocks.MBridgeContractService, 0: mocks.MBridgeContractService},
		mocks.MAssetsService,
		true,
		mocks.MPrometheusService,
		mocks.MPricingService,
		blacklist,
	)

	mocks.MStatusRepository.AssertCalled(t, "Update", txAccountId, mock.Anything)
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
	mocks.MTransferService.On("SanityCheckTransfer", tx).Return(transfer.SanityCheckResult{ChainId: network3, EvmAddress: evmAddress})
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
	mocks.MTransferService.On("SanityCheckTransfer", anotherTx).Return(transfer.SanityCheckResult{ChainId: network3, EvmAddress: evmAddress})
	mocks.MPrometheusService.On("GetIsMonitoringEnabled").Return(false)
	mocks.MAssetsService.On("NativeToWrapped", nativeTokenAddressNetwork0, network0, network3).Return(wrappedTokenAddressNetwork3)
	mocks.MAssetsService.On("FungibleNativeAsset", network0, nativeTokenAddressNetwork0).Return(nativeAssetNetwork0)
	mocks.MPricingService.On("GetTokenPriceInfo", network0, nativeTokenAddressNetwork0).Return(tokenPriceInfo, true)
	mocks.MAssetsService.On("FungibleAssetInfo", network0, nativeTokenAddressNetwork0).Return(fungibleAssetInfoNetwork0, true)
	mocks.MAssetsService.On("FungibleAssetInfo", network3, wrappedTokenAddressNetwork3).Return(fungibleAssetInfoNetwork3, true)

	mocks.MQueue.On("Push", mock.Anything).Return()
	w.processTransaction(anotherTx.TransactionID, mocks.MQueue)
}

func Test_ProcessTransaction_SanityCheckTransfer_Fails(t *testing.T) {
	w := initializeWatcher()
	mocks.MHederaMirrorClient.On("GetSuccessfulTransaction", tx.TransactionID).Return(tx, nil)
	mocks.MTransferService.On("SanityCheckTransfer", tx).Return(transfer.SanityCheckResult{ChainId: network0, EvmAddress: "", Err: errors.New("some-error")})

	w.processTransaction(tx.TransactionID, mocks.MQueue)

	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything)
}

func Test_ProcessTransaction_Blacklist_Fails(t *testing.T) {
	w := initializeWatcher()
	w.blacklistedAccounts = append(w.blacklistedAccounts, tx.TokenTransfers[0].Account)
	mocks.MHederaMirrorClient.On("GetSuccessfulTransaction", tx.TransactionID).Return(tx, nil)

	w.processTransaction(tx.TransactionID, mocks.MQueue)

	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything)
}

func Test_ProcessNFTTransaction_Blacklist_Fails(t *testing.T) {
	tx_blacklist := transaction.Transaction{
		ConsensusTimestamp: "1631092491.483966000",
		TransactionID:      "0.0.111-1631092491-483966000",
	}

	nftTransfer := transaction.NftTransfer{
		ReceiverAccountID: "0.0.111",
		SenderAccountID:   "0.0.333",
		SerialNumber:      1,
		Token:             "0.0.21241241",
	}

	tx_blacklist.NftTransfers = []transaction.NftTransfer{
		nftTransfer,
	}
	tx_blacklist.Transfers = []transaction.Transfer{}
	tx_blacklist.TokenTransfers = []transaction.Transfer{}

	w := initializeWatcher()
	mocks.MHederaMirrorClient.On("GetSuccessfulTransaction", tx_blacklist.TransactionID).Return(tx_blacklist, nil)

	w.processTransaction(tx_blacklist.TransactionID, mocks.MQueue)

	mocks.MQueue.AssertNotCalled(t, "Push", mock.Anything)
}

func Test_ProcessHbarTransaction_Blacklist_Fails(t *testing.T) {
	tx_blacklist := transaction.Transaction{
		ConsensusTimestamp: "1631092491.483966000",
		TransactionID:      "0.0.111-1631092491-483966000",
	}

	transfer := transaction.Transfer{
		Account: "0.0.333",
		Amount:  303030303030303030,
		Token:   "HBAR",
	}

	tx_blacklist.Transfers = []transaction.Transfer{
		transfer,
	}
	tx_blacklist.NftTransfers = []transaction.NftTransfer{}
	tx_blacklist.TokenTransfers = []transaction.Transfer{}

	w := initializeWatcher()
	mocks.MHederaMirrorClient.On("GetSuccessfulTransaction", tx_blacklist.TransactionID).Return(tx_blacklist, nil)

	w.processTransaction(tx_blacklist.TransactionID, mocks.MQueue)

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

func Test_validateNFTFeeSent_ShouldNotValidateFee(t *testing.T) {
	w := initializeWatcher()

	mocks.MPricingService.On("GetHederaNftFee", testConstants.NetworkHederaNonFungibleNativeToken).Return(int64(0), false)

	feeForValidators, ok := w.validateNFTFeeSent(
		testConstants.NetworkHederaNonFungibleNativeToken,
		tx,
		"",
		testConstants.NonFungibleAssetInfos[constants.HederaNetworkId][testConstants.NetworkHederaNonFungibleNativeToken],
		10,
	)
	assert.Equal(t, int64(0), feeForValidators)
	assert.False(t, ok)
}

func Test_validateNFTFeeSent_ShouldValidateFee(t *testing.T) {
	w := initializeWatcher()

	mocks.MPricingService.On("GetHederaNftFee", testConstants.NetworkHederaNonFungibleNativeToken).Return(int64(10), true)
	mocks.MPricingService.On("GetHederaNftPrevFee", testConstants.NetworkHederaNonFungibleNativeToken).Return(int64(20), true)

	feeForValidators, ok := w.validateNFTFeeSent(
		testConstants.NetworkHederaNonFungibleNativeToken,
		tx,
		"",
		testConstants.NonFungibleAssetInfos[constants.HederaNetworkId][testConstants.NetworkHederaNonFungibleNativeToken],
		10,
	)
	assert.Equal(t, int64(10), feeForValidators)
	assert.True(t, ok)
}

func Test_validateNFTFeeSent_ShouldValidatePrevFee(t *testing.T) {
	w := initializeWatcher()

	mocks.MPricingService.On("GetHederaNftFee", testConstants.NetworkHederaNonFungibleNativeToken).Return(int64(10), true)
	mocks.MPricingService.On("GetHederaNftPrevFee", testConstants.NetworkHederaNonFungibleNativeToken).Return(int64(20), true)

	feeForValidators, ok := w.validateNFTFeeSent(
		testConstants.NetworkHederaNonFungibleNativeToken,
		tx,
		"",
		testConstants.NonFungibleAssetInfos[constants.HederaNetworkId][testConstants.NetworkHederaNonFungibleNativeToken],
		20,
	)
	assert.Equal(t, int64(20), feeForValidators)
	assert.True(t, ok)
}

func Test_validateNFTFeeSent_ShouldNotValidateAnyFee(t *testing.T) {
	w := initializeWatcher()

	mocks.MPricingService.On("GetHederaNftFee", testConstants.NetworkHederaNonFungibleNativeToken).Return(int64(0), false)
	mocks.MPricingService.On("GetHederaNftPrevFee", testConstants.NetworkHederaNonFungibleNativeToken).Return(int64(0), false)

	feeForValidators, ok := w.validateNFTFeeSent(
		testConstants.NetworkHederaNonFungibleNativeToken,
		tx,
		"",
		testConstants.NonFungibleAssetInfos[constants.HederaNetworkId][testConstants.NetworkHederaNonFungibleNativeToken],
		20,
	)
	assert.Equal(t, int64(0), feeForValidators)
	assert.False(t, ok)
}

func Test_validateNFTFeeSent_ShouldNotValidateFeeWithOriginator(t *testing.T) {
	w := initializeWatcher()

	mocks.MPricingService.On("GetHederaNftFee", testConstants.NetworkHederaNonFungibleNativeToken).Return(int64(10), true)
	mocks.MPricingService.On("GetHederaNftPrevFee", testConstants.NetworkHederaNonFungibleNativeToken).Return(int64(20), true)

	feeForValidators, ok := w.validateNFTFeeSent(
		testConstants.NetworkHederaNonFungibleNativeToken,
		tx,
		"different originator",
		testConstants.NonFungibleAssetInfos[constants.HederaNetworkId][testConstants.NetworkHederaNonFungibleNativeToken],
		10,
	)
	assert.Equal(t, int64(0), feeForValidators)
	assert.False(t, ok)
}

func Test_validateNFTFeeSent_ShouldValidateLargerFee(t *testing.T) {
	w := initializeWatcher()

	mocks.MPricingService.On("GetHederaNftFee", testConstants.NetworkHederaNonFungibleNativeToken).Return(int64(10), true)
	mocks.MPricingService.On("GetHederaNftPrevFee", testConstants.NetworkHederaNonFungibleNativeToken).Return(int64(20), true)

	feeForValidators, ok := w.validateNFTFeeSent(
		testConstants.NetworkHederaNonFungibleNativeToken,
		tx,
		"",
		testConstants.NonFungibleAssetInfos[constants.HederaNetworkId][testConstants.NetworkHederaNonFungibleNativeToken],
		30,
	)
	assert.Equal(t, int64(30), feeForValidators)
	assert.True(t, ok)
}

func Test_UpdateStatusTimestamp_Works(t *testing.T) {
	w := initializeWatcher()
	mocks.MStatusRepository.On("Update", txAccountId, int64(100)).Return(nil)
	w.updateStatusTimestamp(100)
}

func Test_ConsensusTimestamp_Fails(t *testing.T) {
	w := initializeWatcher()
	anotherTx := tx
	anotherTx.ConsensusTimestamp = "asd"
	mocks.MHederaMirrorClient.On("GetSuccessfulTransaction", anotherTx.TransactionID).Return(anotherTx, nil)
	mocks.MTransferService.On("SanityCheckTransfer", anotherTx).Return(transfer.SanityCheckResult{ChainId: network3, EvmAddress: evmAddress})
	mocks.MQueue.On("Push", mock.Anything).Return()
	mocks.MPrometheusService.On("GetIsMonitoringEnabled").Return(false)
	mocks.MAssetsService.On("NativeToWrapped", nativeTokenAddressNetwork0, network0, network3).Return(wrappedTokenAddressNetwork3)
	mocks.MAssetsService.On("FungibleNativeAsset", network0, nativeTokenAddressNetwork0).Return(nativeAssetNetwork0)
	mocks.MPricingService.On("GetTokenPriceInfo", network0, nativeTokenAddressNetwork0).Return(tokenPriceInfo, true)
	mocks.MAssetsService.On("FungibleAssetInfo", network0, nativeTokenAddressNetwork0).Return(fungibleAssetInfoNetwork0, true)
	mocks.MAssetsService.On("FungibleAssetInfo", network3, wrappedTokenAddressNetwork3).Return(fungibleAssetInfoNetwork3, true)

	w.processTransaction(anotherTx.TransactionID, mocks.MQueue)
}

func Test_validateNftTokenCustomFees(t *testing.T) {
	w := initializeWatcher()

	ok := w.validateNftTokenCustomFees(testConstants.NonFungibleAssetInfos[constants.HederaNetworkId][testConstants.NetworkHederaNonFungibleNativeToken], tx, testConstants.NetworkHederaNonFungibleNativeToken)

	assert.True(t, ok)
}

func Test_validateNftTokenCustomFees_ErrOnTransferForAccountNotFound(t *testing.T) {
	w := initializeWatcher()
	tx.TokenTransfers[0].Account = txAccountId + "1"

	ok := w.validateNftTokenCustomFees(testConstants.NonFungibleAssetInfos[constants.HederaNetworkId][testConstants.NetworkHederaNonFungibleNativeToken], tx, testConstants.NetworkHederaNonFungibleNativeToken)
	tx.TokenTransfers[0].Account = txAccountId

	assert.False(t, ok)
}

func Test_validateNftTokenCustomFees_ErrOnTransferForFeeLessThanExpected(t *testing.T) {
	w := initializeWatcher()
	tx.TokenTransfers[0].Amount = txAmount - 1

	ok := w.validateNftTokenCustomFees(testConstants.NonFungibleAssetInfos[constants.HederaNetworkId][testConstants.NetworkHederaNonFungibleNativeToken], tx, testConstants.NetworkHederaNonFungibleNativeToken)
	tx.TokenTransfers[0].Amount = txAmount

	assert.False(t, ok)
}

func Test_createFungiblePayload(t *testing.T) {
	w := initializeWatcher()

	transactionID := "0.0.111-1-1"
	receiver := "0.0.111"
	sourceAsset := nativeTokenAddressNetwork0
	assetNative := nativeAssetNetwork0
	amount := int64(10000)
	targetChainId := network3
	targetChainAsset := wrappedTokenAddressNetwork3

	fungibleAssetInfoNetwork0 := &asset.FungibleAssetInfo{Decimals: 8}
	fungibleAssetInfoNetwork3 := &asset.FungibleAssetInfo{Decimals: 8}

	mocks.MAssetsService.On("FungibleNativeAsset", network0, nativeTokenAddressNetwork0).Return(nativeAssetNetwork0)
	mocks.MAssetsService.On("FungibleAssetInfo", network0, nativeTokenAddressNetwork0).Return(fungibleAssetInfoNetwork0, true)
	mocks.MAssetsService.On("FungibleAssetInfo", network3, wrappedTokenAddressNetwork3).Return(fungibleAssetInfoNetwork3, true)
	mocks.MPricingService.On("GetTokenPriceInfo", network0, nativeTokenAddressNetwork0).Return(tokenPriceInfo, true)

	payload, err := w.createFungiblePayload(
		transactionID,
		receiver,
		sourceAsset,
		*assetNative,
		amount,
		targetChainId,
		targetChainAsset,
	)

	assert.NoError(t, err)
	assert.Equal(t, transactionID, payload.TransactionId)
	assert.Equal(t, strconv.FormatInt(amount, 10), payload.Amount)
	assert.Equal(t, receiver, payload.Receiver)
}

func Test_createFungiblePayload_ErrorWrongDecimals(t *testing.T) {
	w := initializeWatcher()

	transactionID := "0.0.111-1-1"
	receiver := "0.0.111"
	sourceAsset := nativeTokenAddressNetwork0
	assetNative := nativeAssetNetwork0
	amount := int64(10000)
	targetChainId := network3
	targetChainAsset := wrappedTokenAddressNetwork3

	fungibleAssetInfoNetwork0 := &asset.FungibleAssetInfo{Decimals: 8}
	fungibleAssetInfoNetwork3 := &asset.FungibleAssetInfo{Decimals: 18}

	mocks.MAssetsService.On("FungibleNativeAsset", network0, nativeTokenAddressNetwork0).Return(nativeAssetNetwork0)
	mocks.MAssetsService.On("FungibleAssetInfo", network0, nativeTokenAddressNetwork0).Return(fungibleAssetInfoNetwork0, true)
	mocks.MAssetsService.On("FungibleAssetInfo", network3, wrappedTokenAddressNetwork3).Return(fungibleAssetInfoNetwork3, true)
	mocks.MPricingService.On("GetTokenPriceInfo", network0, nativeTokenAddressNetwork0).Return(tokenPriceInfo, true)

	_, err := w.createFungiblePayload(
		transactionID,
		receiver,
		sourceAsset,
		*assetNative,
		amount,
		targetChainId,
		targetChainAsset,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "decimals of source asset")
}

// insufficient amount
func Test_createFungiblePayload_ErrorInsufficientAmount(t *testing.T) {
	w := initializeWatcher()

	transactionID := "0.0.111-1-1"
	receiver := "0.0.111"
	sourceAsset := nativeTokenAddressNetwork0
	assetNative := nativeAssetNetwork0
	amount := int64(256)
	targetChainId := network3
	targetChainAsset := wrappedTokenAddressNetwork3

	fungibleAssetInfoNetwork0 := &asset.FungibleAssetInfo{Decimals: 8}
	fungibleAssetInfoNetwork3 := &asset.FungibleAssetInfo{Decimals: 8}

	mocks.MAssetsService.On("FungibleNativeAsset", network0, nativeTokenAddressNetwork0).Return(nativeAssetNetwork0)
	mocks.MAssetsService.On("FungibleAssetInfo", network0, nativeTokenAddressNetwork0).Return(fungibleAssetInfoNetwork0, true)
	mocks.MAssetsService.On("FungibleAssetInfo", network3, wrappedTokenAddressNetwork3).Return(fungibleAssetInfoNetwork3, true)
	mocks.MPricingService.On("GetTokenPriceInfo", network0, nativeTokenAddressNetwork0).Return(tokenPriceInfo, true)

	_, err := w.createFungiblePayload(
		transactionID,
		receiver,
		sourceAsset,
		*assetNative,
		amount,
		targetChainId,
		targetChainAsset,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is less than Minimum Amount")
}

func setup() {
	mocks.Setup()
	mocks.MPrometheusService.On("GetIsMonitoringEnabled").Return(false)
}

func initializeWatcher() *Watcher {
	setup()
	mocks.Setup()
	mocks.MStatusRepository.On("Get", mock.Anything).Return(int64(0), nil)
	blacklist := []string{"0.0.333", "0.0.444"}

	return NewWatcher(
		mocks.MTransferService,
		mocks.MHederaMirrorClient,
		txAccountId,
		5,
		mocks.MStatusRepository,
		0,
		map[uint64]iservice.Contracts{3: mocks.MBridgeContractService, 0: mocks.MBridgeContractService},
		mocks.MAssetsService,
		true,
		mocks.MPrometheusService,
		mocks.MPricingService,
		blacklist,
	)
}

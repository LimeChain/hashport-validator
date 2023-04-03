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

package submit

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	evmSetup "github.com/limechain/hedera-eth-bridge-validator/e2e/setup/evm"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/werc721"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/wtoken"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
)

func MintTransaction(t *testing.T, evm evmSetup.Utils, txId string, transactionData *service.FungibleTransferData, tokenAddress common.Address) common.Hash {
	t.Helper()
	var signatures [][]byte
	for i := 0; i < len(transactionData.Signatures); i++ {
		signature, err := hex.DecodeString(transactionData.Signatures[i])
		if err != nil {
			t.Fatalf("Failed to decode signature with error: [%s]", err)
		}
		signatures = append(signatures, signature)
	}
	mintAmount, ok := new(big.Int).SetString(transactionData.Amount, 10)
	if !ok {
		t.Fatalf("Could not convert mint amount [%s] to big int", transactionData.Amount)
	}

	res, err := evm.RouterContract.Mint(
		evm.KeyTransactor,
		new(big.Int).SetUint64(transactionData.SourceChainId),
		[]byte(txId),
		tokenAddress,
		evm.Receiver,
		mintAmount,
		signatures,
	)

	if err != nil {
		t.Fatalf("Cannot execute transaction - Error: [%s].", err)
	}
	return res.Hash()
}

func MintERC721Transaction(t *testing.T, evm evmSetup.Utils, txId string, transactionData *service.NonFungibleTransferData) common.Hash {
	t.Helper()
	var signatures [][]byte
	for i := 0; i < len(transactionData.Signatures); i++ {
		signature, err := hex.DecodeString(transactionData.Signatures[i])
		if err != nil {
			t.Fatalf("Failed to decode signature with error: [%s]", err)
		}
		signatures = append(signatures, signature)
	}

	res, err := evm.RouterContract.MintERC721(
		evm.KeyTransactor,
		new(big.Int).SetUint64(transactionData.SourceChainId),
		[]byte(txId),
		common.HexToAddress(transactionData.TargetAsset),
		big.NewInt(transactionData.TokenId),
		transactionData.Metadata,
		evm.Receiver,
		signatures,
	)

	if err != nil {
		t.Fatalf("Cannot execute transaction - Error: [%s].", err)
	}
	return res.Hash()
}

// Depricated. Used only for test Test_EVM_Wrapped_to_EVM_Token
func UnlockTransaction(t *testing.T, evm evmSetup.Utils, txId string, transactionData *service.FungibleTransferData, tokenAddress common.Address) common.Hash {
	t.Helper()
	var signatures [][]byte
	for i := 0; i < len(transactionData.Signatures); i++ {
		signature, err := hex.DecodeString(transactionData.Signatures[i])
		if err != nil {
			t.Fatalf("Failed to decode signature with error: [%s]", err)
		}
		signatures = append(signatures, signature)
	}
	mintAmount, ok := new(big.Int).SetString(transactionData.Amount, 10)
	if !ok {
		t.Fatalf("Could not convert mint amount [%s] to big int", transactionData.Amount)
	}

	res, err := evm.RouterContract.Unlock(
		evm.KeyTransactor,
		new(big.Int).SetUint64(transactionData.SourceChainId),
		[]byte(txId),
		tokenAddress,
		mintAmount,
		evm.Receiver,
		signatures,
	)

	if err != nil {
		t.Fatalf("Cannot execute transaction - Error: [%s].", err)
	}
	return res.Hash()
}

func BurnEthTransaction(t *testing.T, assetsService service.Assets, evm evmSetup.Utils, asset string, sourceChainId, targetChainId uint64, receiver []byte, amount int64) (*types.Receipt, *router.RouterBurn) {
	t.Helper()
	wrappedAsset, err := evmSetup.NativeToWrappedAsset(assetsService, sourceChainId, targetChainId, asset)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Parsed [%s] to ETH Token [%s]\n", asset, wrappedAsset)

	approvedValue := big.NewInt(amount)

	instance, err := evmSetup.InitAssetContract(wrappedAsset, evm.EVMClient)
	if err != nil {
		t.Fatal(err)
	}

	approveTx, err := instance.Approve(evm.KeyTransactor, evm.RouterAddress, approvedValue)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("[%s] Waiting for Approval Transaction\n", approveTx.Hash())
	WaitForTransaction(t, evm, approveTx.Hash())

	burnTx, err := evm.RouterContract.Burn(evm.KeyTransactor, new(big.Int).SetUint64(sourceChainId), common.HexToAddress(wrappedAsset), approvedValue, receiver)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("[%s] Submitted Burn Transaction\n", burnTx.Hash())

	expectedRouterBurn := &router.RouterBurn{
		Amount:   approvedValue,
		Receiver: receiver,
	}

	burnTxHash := burnTx.Hash()

	fmt.Printf("[%s] Waiting for Burn Transaction Receipt\n", burnTxHash)
	burnTxReceipt, err := evm.EVMClient.WaitForTransactionReceipt(burnTxHash)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("[%s] Burn Transaction mined and retrieved receipt.\n", burnTxHash)

	return burnTxReceipt, expectedRouterBurn
}

func LockEthTransaction(t *testing.T, evm evmSetup.Utils, asset string, targetChainId uint64, receiver []byte, amount int64) (*types.Receipt, *router.RouterLock) {
	t.Helper()
	approvedValue := big.NewInt(amount)

	instance, err := evmSetup.InitAssetContract(asset, evm.EVMClient)
	if err != nil {
		t.Fatal(err)
	}
	approveTx, err := instance.Approve(evm.KeyTransactor, evm.RouterAddress, approvedValue)
	if err != nil {
		t.Fatal(err)
	}

	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("[%s] Waiting for Approval Transaction\n", approveTx.Hash())
	WaitForTransaction(t, evm, approveTx.Hash())

	lockTx, err := evm.RouterContract.Lock(
		evm.KeyTransactor,
		new(big.Int).SetUint64(targetChainId),
		common.HexToAddress(asset),
		approvedValue,
		receiver,
	)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("[%s] Submitted Lock Transaction\n", lockTx.Hash())

	expectedRouterLock := &router.RouterLock{
		TargetChain: new(big.Int).SetUint64(targetChainId),
		Token:       common.HexToAddress(asset),
		Receiver:    receiver,
		Amount:      approvedValue,
		Raw:         types.Log{},
	}

	lockTxHash := lockTx.Hash()

	fmt.Printf("[%s] Waiting for Lock Transaction Receipt\n", lockTxHash)
	lockTxReceipt, err := evm.EVMClient.WaitForTransactionReceipt(lockTxHash)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("[%s] Lock Transaction mined and retrieved receipt\n", lockTxHash)

	return lockTxReceipt, expectedRouterLock
}

func BurnERC721Transaction(t *testing.T, evm evmSetup.Utils, wrappedToken string, targetChainId uint64, receiver []byte, serialNumber int64) (*types.Receipt, *router.RouterBurnERC721) {
	t.Helper()
	wrappedAddress := common.HexToAddress(wrappedToken)

	paymentToken, err := evm.RouterContract.Erc721Payment(nil, wrappedAddress)
	if err != nil {
		t.Fatal(err)
	}

	fee, err := evm.RouterContract.Erc721Fee(nil, wrappedAddress)
	if err != nil {
		t.Fatal(err)
	}

	erc20Contract, err := wtoken.NewWtoken(paymentToken, evm.EVMClient.GetClient())
	if err != nil {
		t.Fatal(err)
	}

	approveERC20Tx, err := erc20Contract.Approve(evm.KeyTransactor, evm.RouterAddress, fee)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("[%s] Waiting for ERC-20 Approval Transaction\n", approveERC20Tx.Hash())
	WaitForTransaction(t, evm, approveERC20Tx.Hash())

	erc721Contract, err := werc721.NewWerc721(wrappedAddress, evm.EVMClient)
	if err != nil {
		t.Fatal(err)
	}
	tokenId := big.NewInt(serialNumber)

	approveERC721Tx, err := erc721Contract.Approve(evm.KeyTransactor, evm.RouterAddress, tokenId)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("[%s] Waiting for ERC-721 Approval Transaction\n", approveERC721Tx.Hash())
	WaitForTransaction(t, evm, approveERC721Tx.Hash())
	targetChainIdBigInt := new(big.Int).SetUint64(targetChainId)
	burnTx, err := evm.RouterContract.BurnERC721(evm.KeyTransactor, targetChainIdBigInt, wrappedAddress, tokenId, paymentToken, fee, receiver)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("[%s] Submitted Burn Transaction\n", burnTx.Hash())

	expectedRouterBurn := &router.RouterBurnERC721{
		TargetChain:  targetChainIdBigInt,
		WrappedToken: common.HexToAddress(wrappedToken),
		TokenId:      tokenId,
		Receiver:     receiver,
	}

	burnTxHash := burnTx.Hash()

	fmt.Printf("[%s] Waiting for Burn ERC-721 Transaction Receipt\n", burnTxHash)
	burnTxReceipt, err := evm.EVMClient.WaitForTransactionReceipt(burnTxHash)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("[%s] Burn ERC-721 Transaction mined and retrieved receipt\n", burnTxHash)

	return burnTxReceipt, expectedRouterBurn
}

func WaitForTransaction(t *testing.T, evm evmSetup.Utils, txHash common.Hash) {
	t.Helper()
	receipt, err := evm.EVMClient.WaitForTransactionReceipt(txHash)
	if err != nil {
		t.Fatalf("[%s] - Error occurred while monitoring. Error: [%s]", txHash, err)
		return
	}

	if receipt.Status == 1 {
		fmt.Printf("TX [%s] was successfully mined\n", txHash)
	} else {
		t.Fatalf("TX [%s] reverted", txHash)
	}
}

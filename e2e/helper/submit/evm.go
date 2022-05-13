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

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/werc721"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/wtoken"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/e2e/setup"
)

func MintTransaction(t *testing.T, evm setup.EVMUtils, txId string, transactionData *service.FungibleTransferData, tokenAddress common.Address) common.Hash {
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

func MintERC721Transaction(t *testing.T, evm setup.EVMUtils, txId string, transactionData *service.NonFungibleTransferData) common.Hash {
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

func UnlockTransaction(t *testing.T, evm setup.EVMUtils, txId string, transactionData *service.FungibleTransferData, tokenAddress common.Address) common.Hash {
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

func BurnEthTransaction(t *testing.T, assetsService service.Assets, evm setup.EVMUtils, asset string, sourceChainId, targetChainId uint64, receiver []byte, amount int64) (*types.Receipt, *router.RouterBurn) {
	t.Helper()
	wrappedAsset, err := setup.NativeToWrappedAsset(assetsService, sourceChainId, targetChainId, asset)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("Parsed [%s] to ETH Token [%s]", asset, wrappedAsset))

	approvedValue := big.NewInt(amount)

	instance, err := setup.InitAssetContract(wrappedAsset, evm.EVMClient)
	if err != nil {
		t.Fatal(err)
	}

	approveTx, err := instance.Approve(evm.KeyTransactor, evm.RouterAddress, approvedValue)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(fmt.Sprintf("[%s] Waiting for Approval Transaction", approveTx.Hash()))
	WaitForTransaction(t, evm, approveTx.Hash())

	burnTx, err := evm.RouterContract.Burn(evm.KeyTransactor, new(big.Int).SetUint64(sourceChainId), common.HexToAddress(wrappedAsset), approvedValue, receiver)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("[%s] Submitted Burn Transaction", burnTx.Hash()))

	expectedRouterBurn := &router.RouterBurn{
		//Account:      common.HexToAddress(evm.Signer.Address()),
		//WrappedAsset: *wrappedAsset,
		Amount:   approvedValue,
		Receiver: receiver,
	}

	burnTxHash := burnTx.Hash()

	fmt.Println(fmt.Sprintf("[%s] Waiting for Burn Transaction Receipt.", burnTxHash))
	burnTxReceipt, err := evm.EVMClient.WaitForTransactionReceipt(burnTxHash)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("[%s] Burn Transaction mined and retrieved receipt.", burnTxHash))

	return burnTxReceipt, expectedRouterBurn
}

func LockEthTransaction(t *testing.T, evm setup.EVMUtils, asset string, targetChainId uint64, receiver []byte, amount int64) (*types.Receipt, *router.RouterLock) {
	t.Helper()
	approvedValue := big.NewInt(amount)

	instance, err := setup.InitAssetContract(asset, evm.EVMClient)
	if err != nil {
		t.Fatal(err)
	}
	approveTx, err := instance.Approve(evm.KeyTransactor, evm.RouterAddress, approvedValue)
	if err != nil {
		t.Fatal(err)
	}

	feeData, err := evm.RouterContract.TokenFeeData(nil, common.HexToAddress(asset))
	if err != nil {
		t.Fatal(err)
	}

	precision, err := evm.RouterContract.ServiceFeePrecision(nil)
	if err != nil {
		t.Fatal(err)
	}

	multiplied := new(big.Int).Mul(approvedValue, feeData.ServiceFeePercentage)
	serviceFee := new(big.Int).Div(multiplied, precision)

	fmt.Println(fmt.Sprintf("[%s] Waiting for Approval Transaction", approveTx.Hash()))
	WaitForTransaction(t, evm, approveTx.Hash())

	lockTx, err := evm.RouterContract.Lock(evm.KeyTransactor, new(big.Int).SetUint64(targetChainId), common.HexToAddress(asset), approvedValue, receiver)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("[%s] Submitted Lock Transaction", lockTx.Hash()))

	expectedRouterLock := &router.RouterLock{
		TargetChain: new(big.Int).SetUint64(targetChainId),
		Token:       common.HexToAddress(asset),
		Receiver:    receiver,
		Amount:      approvedValue,
		ServiceFee:  serviceFee,
		Raw:         types.Log{},
	}

	lockTxHash := lockTx.Hash()

	fmt.Println(fmt.Sprintf("[%s] Waiting for Lock Transaction Receipt.", lockTxHash))
	lockTxReceipt, err := evm.EVMClient.WaitForTransactionReceipt(lockTxHash)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("[%s] Lock Transaction mined and retrieved receipt.", lockTxHash))

	return lockTxReceipt, expectedRouterLock
}

func BurnERC721Transaction(t *testing.T, evm setup.EVMUtils, wrappedToken string, targetChainId uint64, receiver []byte, serialNumber int64) (*types.Receipt, *router.RouterBurnERC721) {
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

	fmt.Println(fmt.Sprintf("[%s] Waiting for ERC-20 Approval Transaction", approveERC20Tx.Hash()))
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

	fmt.Println(fmt.Sprintf("[%s] Waiting for ERC-721 Approval Transaction", approveERC721Tx.Hash()))
	WaitForTransaction(t, evm, approveERC721Tx.Hash())
	targetChainIdBigInt := new(big.Int).SetUint64(targetChainId)
	burnTx, err := evm.RouterContract.BurnERC721(evm.KeyTransactor, targetChainIdBigInt, wrappedAddress, tokenId, paymentToken, fee, receiver)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("[%s] Submitted Burn Transaction", burnTx.Hash()))

	expectedRouterBurn := &router.RouterBurnERC721{
		TargetChain:  targetChainIdBigInt,
		WrappedToken: common.HexToAddress(wrappedToken),
		TokenId:      tokenId,
		Receiver:     receiver,
	}

	burnTxHash := burnTx.Hash()

	fmt.Println(fmt.Sprintf("[%s] Waiting for Burn ERC-721 Transaction Receipt.", burnTxHash))
	burnTxReceipt, err := evm.EVMClient.WaitForTransactionReceipt(burnTxHash)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fmt.Sprintf("[%s] Burn ERC-721 Transaction mined and retrieved receipt.", burnTxHash))

	return burnTxReceipt, expectedRouterBurn
}

func WaitForTransaction(t *testing.T, evm setup.EVMUtils, txHash common.Hash) {
	t.Helper()
	receipt, err := evm.EVMClient.WaitForTransactionReceipt(txHash)
	if err != nil {
		t.Fatalf("[%s] - Error occurred while monitoring. Error: [%s]", txHash, err)
		return
	}

	if receipt.Status == 1 {
		fmt.Println(fmt.Sprintf("TX [%s] was successfully mined", txHash))
	} else {
		t.Fatalf("TX [%s] reverted", txHash)
	}
}

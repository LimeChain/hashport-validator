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

package verify

import (
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"testing"

	evmSetup "github.com/limechain/hedera-eth-bridge-validator/e2e/setup/evm"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/werc721"
)

func BurnEvent(t *testing.T, txReceipt *types.Receipt, expectedRouterBurn *router.RouterBurn) string {
	t.Helper()
	parsedAbi, err := abi.JSON(strings.NewReader(router.RouterABI))
	if err != nil {
		t.Fatal(err)
	}

	routerBurn := router.RouterBurn{}
	eventSignature := []byte("Burn(uint256,address,uint256,bytes)")
	eventSignatureHash := crypto.Keccak256Hash(eventSignature)
	for _, log := range txReceipt.Logs {
		if log.Topics[0] != eventSignatureHash {
			continue
		}

		err := parsedAbi.UnpackIntoInterface(&routerBurn, "Burn", log.Data)
		if err != nil {
			t.Fatal(err)
		}

		if routerBurn.Amount.String() != expectedRouterBurn.Amount.String() {
			t.Fatalf("Expected Burn Event Amount [%v], but actually was [%v]", expectedRouterBurn.Amount, routerBurn.Amount)
		}

		if !reflect.DeepEqual(routerBurn.Receiver, expectedRouterBurn.Receiver) {
			t.Fatalf("Expected Burn Event Receiver [%v], but actually was [%v]", expectedRouterBurn.Receiver, routerBurn.Receiver)
		}

		expectedId := fmt.Sprintf("%s-%d", log.TxHash, log.Index)
		return expectedId
	}

	t.Fatal("Could not retrieve valid Burn Event Log information.")
	return ""
}

func LockEvent(t *testing.T, txReceipt *types.Receipt, expectedRouterLock *router.RouterLock) string {
	t.Helper()
	parsedAbi, err := abi.JSON(strings.NewReader(router.RouterABI))
	if err != nil {
		t.Fatal(err)
	}

	routerLock := router.RouterLock{}
	eventSignature := []byte("Lock(uint256,address,bytes,uint256,uint256)")
	eventSignatureHash := crypto.Keccak256Hash(eventSignature)
	for _, log := range txReceipt.Logs {
		if log.Topics[0] != eventSignatureHash {
			continue
		}

		err := parsedAbi.UnpackIntoInterface(&routerLock, "Lock", log.Data)
		if err != nil {
			t.Fatal(err)
		}

		if routerLock.Amount.String() != expectedRouterLock.Amount.String() {
			t.Fatalf("Expected Lock Event Amount [%v], actual [%v]", expectedRouterLock.Amount, routerLock.Amount)
		}

		if routerLock.TargetChain.String() != expectedRouterLock.TargetChain.String() {
			t.Fatalf("Expected Lock Event Target Chain [%v], actual [%v]", expectedRouterLock.TargetChain, routerLock.TargetChain)
		}

		if routerLock.Token.String() != expectedRouterLock.Token.String() {
			t.Fatalf("Expected Lock Event Token [%v], actual [%v]", expectedRouterLock.Token, routerLock.Token)
		}

		if routerLock.ServiceFee.String() != expectedRouterLock.ServiceFee.String() {
			t.Fatalf("Expected Lock Event Service Fee [%v], actual [%s]", expectedRouterLock.ServiceFee, routerLock.ServiceFee)
		}

		if !reflect.DeepEqual(routerLock.Receiver, expectedRouterLock.Receiver) {
			t.Fatalf("Expected Lock Event Receiver [%v], actual [%v]", expectedRouterLock.Receiver, routerLock.Receiver)
		}

		expectedId := fmt.Sprintf("%s-%d", log.TxHash, log.Index)
		return expectedId
	}

	t.Fatal("Could not retrieve valid Burn Event Log information.")
	return ""
}

func BurnERC721Event(t *testing.T, txReceipt *types.Receipt, expectedRouterLock *router.RouterBurnERC721) string {
	t.Helper()
	parsedAbi, err := abi.JSON(strings.NewReader(router.RouterABI))
	if err != nil {
		t.Fatal(err)
	}

	event := router.RouterBurnERC721{}
	eventSignature := parsedAbi.Events["BurnERC721"].ID
	for _, log := range txReceipt.Logs {
		if log.Topics[0] != eventSignature {
			continue
		}

		err := parsedAbi.UnpackIntoInterface(&event, "BurnERC721", log.Data)
		if err != nil {
			t.Fatal(err)
		}

		if event.TokenId.String() != expectedRouterLock.TokenId.String() {
			t.Fatalf("Expected Burn-ERC721 Event TokenId [%v], actual [%v]", expectedRouterLock.TokenId, event.TokenId)
		}

		if event.TargetChain.String() != expectedRouterLock.TargetChain.String() {
			t.Fatalf("Expected Burn-ERC721 Event Target Chain [%v], actual [%v]", expectedRouterLock.TargetChain, event.TargetChain)
		}

		if !reflect.DeepEqual(event.WrappedToken, expectedRouterLock.WrappedToken) {
			t.Fatalf("Expected Burn-ERC721 Event Token [%v], actual [%v]", expectedRouterLock.WrappedToken, event.WrappedToken)
		}

		if !reflect.DeepEqual(event.Receiver, expectedRouterLock.Receiver) {
			t.Fatalf("Expected Burn-ERC721 Event Receiver [%v], actual [%v]", expectedRouterLock.Receiver, event.Receiver)
		}

		expectedId := fmt.Sprintf("%s-%d", log.TxHash, log.Index)
		return expectedId
	}

	t.Fatal("Could not retrieve valid Burn Event Log information.")
	return ""
}

func WrappedAssetBalance(t *testing.T, evm evmSetup.Utils, nativeAsset string, mintAmount *big.Int, wrappedBalanceBefore *big.Int, wTokenReceiverAddress common.Address) {
	t.Helper()
	instance, err := evmSetup.InitAssetContract(nativeAsset, evm.EVMClient)
	if err != nil {
		t.Fatal(err)
	}

	wrappedBalanceAfter, err := instance.BalanceOf(&bind.CallOpts{}, wTokenReceiverAddress)
	if err != nil {
		t.Fatal(err)
	}

	if err != nil {
		t.Fatal(err)
	}

	expectedBalance := new(big.Int).Add(wrappedBalanceBefore, mintAmount)

	if wrappedBalanceAfter.Cmp(expectedBalance) != 0 {
		t.Fatalf("Incorrect token balance. Expected to be [%s], but was [%s].", expectedBalance, wrappedBalanceAfter)
	}
}

func ERC721TokenId(t *testing.T, evm *evm.Client, wrappedToken string, serialNumber int64, receiver string, expectedMetadata string) {
	t.Helper()
	contract, err := werc721.NewWerc721(common.HexToAddress(wrappedToken), evm)
	if err != nil {
		t.Fatal(err)
	}

	owner, err := contract.OwnerOf(nil, big.NewInt(serialNumber))
	if err != nil {
		t.Fatal(err)
	}

	if owner.String() != receiver {
		t.Fatalf("Invalid owner. Expected owner for serial number [%d] to be [%s], but was [%s]", serialNumber, receiver, owner.String())
	}

	tokenURI, err := contract.TokenURI(nil, big.NewInt(serialNumber))
	if err != nil {
		t.Fatal(err)
	}

	if expectedMetadata != tokenURI {
		t.Fatalf("Invalid token URI. Expected token URI for serial number [%d] to be [%s], but was [%s]", serialNumber, expectedMetadata, tokenURI)
	}
}

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

package evm

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/wtoken"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	evm_signer "github.com/limechain/hedera-eth-bridge-validator/app/services/signer/evm"
)

type Utils struct {
	EVMClient               *evm.Client
	EVMFungibleTokenClients map[string]client.EvmFungibleToken
	EVMNftClients           map[string]client.EvmNft
	RouterContract          *router.Router
	KeyTransactor           *bind.TransactOpts
	Signer                  *evm_signer.Signer
	Receiver                common.Address
	RouterAddress           common.Address
	WTokenContractAddress   string
}

func RouterAndEVMTokenClientsFromEVMUtils(evmUtils map[uint64]Utils) (
	routerClients map[uint64]client.DiamondRouter,
	evmFungibleTokenClients map[uint64]map[string]client.EvmFungibleToken,
	evmNftClients map[uint64]map[string]client.EvmNft,
) {
	routerClients = make(map[uint64]client.DiamondRouter)
	evmFungibleTokenClients = make(map[uint64]map[string]client.EvmFungibleToken)
	evmNftClients = make(map[uint64]map[string]client.EvmNft)
	for networkId, evmUtil := range evmUtils {
		routerClients[networkId] = evmUtil.RouterContract

		evmFungibleTokenClients[networkId] = make(map[string]client.EvmFungibleToken)
		for tokenAddress, evmTokenClient := range evmUtil.EVMFungibleTokenClients {
			evmFungibleTokenClients[networkId][tokenAddress] = evmTokenClient
		}

		evmNftClients[networkId] = make(map[string]client.EvmNft)
		for tokenAddress, evmTokenClient := range evmUtil.EVMNftClients {
			evmNftClients[networkId][tokenAddress] = evmTokenClient
		}
	}

	return routerClients, evmFungibleTokenClients, evmNftClients
}

func InitAssetContract(asset string, evmClient *evm.Client) (*wtoken.Wtoken, error) {
	return wtoken.NewWtoken(common.HexToAddress(asset), evmClient.GetClient())
}

func NativeToWrappedAsset(assetsService service.Assets, sourceChain, targetChain uint64, nativeAsset string) (string, error) {
	wrappedAsset := assetsService.NativeToWrapped(nativeAsset, sourceChain, targetChain)

	if wrappedAsset == "" {
		return "", fmt.Errorf("EvmFungibleToken [%s] is not supported", nativeAsset)
	}

	return wrappedAsset, nil
}

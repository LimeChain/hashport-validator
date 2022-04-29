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

package assets

import (
	"fmt"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/account"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	qi "github.com/limechain/hedera-eth-bridge-validator/app/domain/queue"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	log "github.com/sirupsen/logrus"
	"math/big"
	"time"
)

var (
	sleepTime = 10 * time.Minute
)

type Watcher struct {
	mirrorNode                 client.MirrorNode
	EvmFungibleTokenClients    map[uint64]map[string]client.EvmFungibleToken
	EvmNonFungibleTokenClients map[uint64]map[string]client.EvmNft
	configuration              config.Config
	logger                     *log.Entry
	assetsService              service.Assets
}

func NewWatcher(
	mirrorNode client.MirrorNode,
	configuration config.Config,
	EvmFungibleTokenClients map[uint64]map[string]client.EvmFungibleToken,
	EvmNonFungibleTokenClients map[uint64]map[string]client.EvmNft,
	assetsService service.Assets,
) *Watcher {

	return &Watcher{
		mirrorNode:                 mirrorNode,
		EvmFungibleTokenClients:    EvmFungibleTokenClients,
		EvmNonFungibleTokenClients: EvmNonFungibleTokenClients,
		configuration:              configuration,
		logger:                     config.GetLoggerFor(fmt.Sprintf("Assets Watcher on interval [%v]", sleepTime)),
		assetsService:              assetsService,
	}
}
func (pw *Watcher) Watch(q qi.Queue) {

	// there will be no handler, so the q is to implement the interface
	go func() {
		pw.watchIteration()
		time.Sleep(sleepTime)
	}()
}

func (pw *Watcher) watchIteration() {
	bridgeAccount, err := pw.getAccount(pw.configuration.Bridge.Hedera.BridgeAccount)
	if err != nil {
		return
	}

	hederaTokenBalances := bridgeAccount.Balance.GetAccountTokenBalancesByAddress()
	fungibleAssets := pw.assetsService.FungibleNetworkAssets()
	nonFungibleAssets := pw.assetsService.NonFungibleNetworkAssets()
	pw.updateAssetInfos(hederaTokenBalances, fungibleAssets, true)
	pw.updateAssetInfos(hederaTokenBalances, nonFungibleAssets, false)
}

func (pw *Watcher) updateAssetInfos(hederaTokenBalances map[string]int, assets map[uint64][]string, isFungible bool) {
	for networkId, networkAssets := range assets {
		for _, assetAddress := range networkAssets {
			IsNative := pw.assetsService.IsNative(networkId, assetAddress)
			pw.updateAssetInfo(networkId, assetAddress, hederaTokenBalances, isFungible, IsNative)
		}
	}
}

func (pw *Watcher) getAccount(accountId string) (*account.AccountsResponse, error) {
	account, e := pw.mirrorNode.GetAccount(accountId)
	if e != nil {
		pw.logger.Errorf("Hedera Mirror Node for Account ID [%s] method GetAccount - Error: [%s]", accountId, e)
		return nil, e
	}
	return account, nil
}

func (pw *Watcher) updateAssetInfo(networkId uint64, assetId string, hederaTokenBalances map[string]int, isFungible bool, isNative bool) {
	var (
		reserveAmount *big.Int
	)

	var err error
	if networkId == constants.HederaNetworkId {
		reserveAmount, err = pw.assetsService.FetchHederaTokenReserveAmount(assetId, pw.mirrorNode, isNative, hederaTokenBalances)
	} else {
		if isFungible { // Fungible
			reserveAmount, err = pw.assetsService.FetchEvmFungibleReserveAmount(
				networkId,
				assetId,
				isNative,
				pw.EvmFungibleTokenClients[networkId][assetId],
				pw.configuration.Bridge.EVMs[networkId].RouterContractAddress,
			)
		} else { // Non-Fungible
			reserveAmount, err = pw.assetsService.FetchEvmNonFungibleReserveAmount(
				networkId,
				assetId,
				isNative,
				pw.EvmNonFungibleTokenClients[networkId][assetId],
				pw.configuration.Bridge.EVMs[networkId].RouterContractAddress,
			)
		}
	}

	if err != nil {
		pw.logger.Errorf("error while fetching reserve amount for token id: %s, skipping update ...", assetId)
		return
	}

	if isFungible {
		assetInfo, ok := pw.assetsService.FungibleAssetInfo(networkId, assetId)
		if ok {
			assetInfo.ReserveAmount = reserveAmount
		}
	} else {
		assetInfo, ok := pw.assetsService.NonFungibleAssetInfo(networkId, assetId)
		if ok {
			assetInfo.ReserveAmount = reserveAmount
		}
	}
}

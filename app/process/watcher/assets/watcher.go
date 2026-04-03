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
	"errors"
	"fmt"
	"github.com/gookit/event"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/account"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	qi "github.com/limechain/hedera-eth-bridge-validator/app/domain/queue"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	bridge_config_event "github.com/limechain/hedera-eth-bridge-validator/app/model/bridge-config-event"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	log "github.com/sirupsen/logrus"
	"math/big"
	"time"
)

var (
	sleepTime       = 10 * time.Minute
	pausedSleepTime = 10 * time.Second
)

type Watcher struct {
	mirrorNode                 client.MirrorNode
	evmRegularTokenClients    map[uint64]map[string]client.EvmRegularToken
	bridgeCfg                  *config.Bridge
	assetsService              service.Assets
	paused                     bool
	logger                     *log.Entry
}

func NewWatcher(
	mirrorNode client.MirrorNode,
	bridgeCfg *config.Bridge,
	EvmRegularTokenClients map[uint64]map[string]client.EvmRegularToken,
	assetsService service.Assets,
) *Watcher {

	instance := &Watcher{
		mirrorNode:                 mirrorNode,
		evmRegularTokenClients:    EvmRegularTokenClients,
		bridgeCfg:                  bridgeCfg,
		logger:                     config.GetLoggerFor(fmt.Sprintf("Assets Watcher on interval [%v]", sleepTime)),
		assetsService:              assetsService,
	}

	event.On(constants.EventBridgeConfigUpdate, event.ListenerFunc(func(e event.Event) error {
		instance.paused = true
		res := bridgeCfgUpdateEventHandler(e, instance)
		instance.paused = false

		return res
	}), constants.WatcherEventPriority)

	return instance
}

func (pw *Watcher) Watch(q qi.Queue) {

	// there will be no handler, so the q is to implement the interface
	go func() {
		for {
			if !pw.paused {
				pw.watchIteration()
				time.Sleep(sleepTime)
			} else {
				time.Sleep(pausedSleepTime)
			}
		}
	}()
}

func (pw *Watcher) watchIteration() {
	bridgeAccount, err := pw.getAccount(pw.bridgeCfg.Hedera.BridgeAccount)
	if err != nil {
		return
	}

	hederaTokenBalances := bridgeAccount.Balance.GetAccountTokenBalancesByAddress()
	fungibleAssets := pw.assetsService.FungibleNetworkAssets()
	pw.updateAssetInfos(hederaTokenBalances, fungibleAssets)
}

func (pw *Watcher) updateAssetInfos(hederaTokenBalances map[string]int, assets map[uint64][]string) {
	for networkId, networkAssets := range assets {
		for _, assetAddress := range networkAssets {
			IsNative := pw.assetsService.IsNative(networkId, assetAddress)
			pw.updateAssetInfo(networkId, assetAddress, hederaTokenBalances, IsNative)
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

func (pw *Watcher) updateAssetInfo(networkId uint64, assetId string, hederaTokenBalances map[string]int, isNative bool) {
	var (
		reserveAmount *big.Int
		err           error
	)

	if networkId == constants.HederaNetworkId {
		reserveAmount, err = pw.assetsService.FetchHederaTokenReserveAmount(assetId, pw.mirrorNode, isNative, hederaTokenBalances)
	} else {
		reserveAmount, err = pw.assetsService.FetchEvmRegularReserveAmount(
			networkId,
			assetId,
			isNative,
			pw.evmRegularTokenClients[networkId][assetId],
			pw.bridgeCfg.EVMs[networkId].RouterContractAddress,
		)
	}

	if err != nil {
		pw.logger.Errorf("error while fetching reserve amount for token id: %s, skipping update ...", assetId)
		return
	}

	assetInfo, ok := pw.assetsService.FungibleAssetInfo(networkId, assetId)
	if ok {
		assetInfo.ReserveAmount = reserveAmount
	}
}

func bridgeCfgUpdateEventHandler(e event.Event, instance *Watcher) error {
	params, ok := e.Get(constants.BridgeConfigUpdateEventParamsKey).(*bridge_config_event.Params)
	if !ok {
		errMsg := fmt.Sprintf("failed to cast params from event [%s]", constants.EventBridgeConfigUpdate)
		log.Errorf(errMsg)
		return errors.New(errMsg)
	}
	instance.evmRegularTokenClients = params.EvmRegularTokenClients
	instance.bridgeCfg = params.Bridge
	instance.watchIteration()

	return nil
}

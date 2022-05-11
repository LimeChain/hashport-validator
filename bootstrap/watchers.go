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

package bootstrap

import (
	"time"

	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	aw "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/assets"
	cmw "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/message"
	pw "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/prometheus"
	tw "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

func createTransferWatcher(configuration *config.Config,
	bridgeService service.Transfers,
	assetsService service.Assets,
	mirrorNode client.MirrorNode,
	repository *repository.Status,
	contractServices map[uint64]service.Contracts,
	prometheusService service.Prometheus,
	pricingService service.Pricing,
) *tw.Watcher {
	account := configuration.Bridge.Hedera.BridgeAccount

	log.Debugf("Added Transfer Watcher for account [%s]", account)
	return tw.NewWatcher(
		bridgeService,
		mirrorNode,
		account,
		configuration.Node.Clients.MirrorNode.PollingInterval,
		*repository,
		configuration.Node.Clients.Hedera.StartTimestamp,
		contractServices,
		assetsService,
		configuration.Bridge.Hedera.NftFees,
		configuration.Bridge.Hedera.NftDynamicFees,
		configuration.Node.Validator,
		prometheusService,
		pricingService,
	)
}

func createConsensusTopicWatcher(configuration *config.Config,
	client client.MirrorNode,
	repository repository.Status,
) *cmw.Watcher {
	topic := configuration.Bridge.TopicId
	log.Debugf("Added Topic Watcher for topic [%s]\n", topic)
	return cmw.NewWatcher(client,
		topic,
		repository,
		configuration.Node.Clients.MirrorNode.PollingInterval,
		configuration.Node.Clients.Hedera.StartTimestamp)
}

func createAssetsWatcher(
	mirrorNode client.MirrorNode,
	configuration config.Config,
	evmFungibleTokenClients map[uint64]map[string]client.EvmFungibleToken,
	evmNonFungibleTokenClients map[uint64]map[string]client.EvmNft,
	assetsService service.Assets,

) *aw.Watcher {
	log.Debugf("Added Assets Watcher")
	return aw.NewWatcher(
		mirrorNode,
		configuration,
		evmFungibleTokenClients,
		evmNonFungibleTokenClients,
		assetsService)
}

func createPrometheusWatcher(
	dashboardPolling time.Duration,
	mirrorNode client.MirrorNode,
	configuration config.Config,
	prometheusService service.Prometheus,
	evmFungibleTokenClients map[uint64]map[string]client.EvmFungibleToken,
	evmNonFungibleTokenClients map[uint64]map[string]client.EvmNft,
	assetsService service.Assets,

) *pw.Watcher {
	log.Debugf("Added Prometheus Watcher for dashboard metrics")
	return pw.NewWatcher(
		dashboardPolling,
		mirrorNode,
		configuration,
		prometheusService,
		evmFungibleTokenClients,
		evmNonFungibleTokenClients,
		assetsService)
}

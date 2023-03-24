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
	apirouter "github.com/limechain/hedera-eth-bridge-validator/app/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/assets"
	burn_event "github.com/limechain/hedera-eth-bridge-validator/app/router/burn-event"
	config_bridge "github.com/limechain/hedera-eth-bridge-validator/app/router/config-bridge"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/fees"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/healthcheck"
	min_amounts "github.com/limechain/hedera-eth-bridge-validator/app/router/min-amounts"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/utils"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func InitializeAPIRouter(services *Services, bridgeConfig *parser.Bridge) *apirouter.APIRouter {
	apiRouter := apirouter.NewAPIRouter()
	apiRouter.AddV1Router(healthcheck.Route, healthcheck.NewRouter())
	apiRouter.AddV1Router(transfer.Route, transfer.NewRouter(services.transfers))
	apiRouter.AddV1Router(burn_event.Route, burn_event.NewRouter(services.BurnEvents))
	apiRouter.AddV1Router(constants.PrometheusMetricsEndpoint, promhttp.Handler())
	apiRouter.AddV1Router(config_bridge.Route, config_bridge.NewRouter(bridgeConfig))
	apiRouter.AddV1Router(min_amounts.Route, min_amounts.NewRouter(services.Pricing))
	apiRouter.AddV1Router(assets.Route, assets.NewRouter(bridgeConfig, services.Assets, services.Pricing))
	apiRouter.AddV1Router(utils.Route, utils.NewRouter(services.Utils))
	apiRouter.AddV1Router(fees.Route, fees.NewRouter(services.Pricing, services.Fees))
	return apiRouter
}

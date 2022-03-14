package bootstrap

import (
	apirouter "github.com/limechain/hedera-eth-bridge-validator/app/router"
	burn_event "github.com/limechain/hedera-eth-bridge-validator/app/router/burn-event"
	config_bridge "github.com/limechain/hedera-eth-bridge-validator/app/router/config-bridge"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/healthcheck"
	min_amounts "github.com/limechain/hedera-eth-bridge-validator/app/router/min-amounts"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config/parser"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func InitializeAPIRouter(services *Services, bridgeConfig parser.Bridge) *apirouter.APIRouter {
	apiRouter := apirouter.NewAPIRouter()
	apiRouter.AddV1Router(healthcheck.Route, healthcheck.NewRouter())
	apiRouter.AddV1Router(transfer.Route, transfer.NewRouter(services.transfers))
	apiRouter.AddV1Router(burn_event.Route, burn_event.NewRouter(services.BurnEvents))
	apiRouter.AddV1Router(constants.PrometheusMetricsEndpoint, promhttp.Handler())
	apiRouter.AddV1Router(config_bridge.Route, config_bridge.NewRouter(bridgeConfig))
	apiRouter.AddV1Router(min_amounts.Route, min_amounts.NewRouter(services.Pricing))

	return apiRouter
}

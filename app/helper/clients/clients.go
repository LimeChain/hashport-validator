package clients

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

func InitRouterClients(bridgeEVMsCfgs map[uint64]config.BridgeEvm, evmClients map[uint64]client.EVM, log *log.Entry) map[uint64]*router.Router {
	routers := make(map[uint64]*router.Router)
	for networkId, bridgeEVMsCfg := range bridgeEVMsCfgs {
		contractAddress, err := evmClients[networkId].ValidateContractDeployedAt(bridgeEVMsCfg.RouterContractAddress)
		if err != nil {
			log.Fatal(err)
		}

		contractInstance, err := router.NewRouter(*contractAddress, evmClients[networkId].GetClient())
		if err != nil {
			log.Fatalf("Failed to initialize Router Contract Instance at [%s]. Error [%s]", bridgeEVMsCfg.RouterContractAddress, err)
		}
		routers[networkId] = contractInstance
	}

	return routers
}

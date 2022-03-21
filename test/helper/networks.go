package helper

import (
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	testConstants "github.com/limechain/hedera-eth-bridge-validator/test/constants"
)

func SetupNetworks() {
	for networkId, networkInfo := range testConstants.Networks {
		constants.NetworksById[networkId] = networkInfo.Name
		constants.NetworksByName[networkInfo.Name] = networkId
	}
}

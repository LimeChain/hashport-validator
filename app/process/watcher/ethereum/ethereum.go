package ethereum

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethClient "github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum"
	eth "github.com/limechain/hedera-eth-bridge-validator/app/helper/ethereum"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-watcher-sdk/queue"
	log "github.com/sirupsen/logrus"
)

type EthWatcher struct {
	client          *ethClient.EthereumClient
	config          config.EthereumWatcher
	contractAddress *common.Address
}

func (ew *EthWatcher) Watch(queue *queue.Queue) {
	log.Infof("[Ethereum Watcher] - Start listening for events for contract address [%s].", ew.contractAddress.String())
	go ew.listenForEvents(queue)
}

func (ew *EthWatcher) listenForEvents(q *queue.Queue) {
	sub, logs, err := ew.client.SubscribeToEventLogs(*ew.contractAddress)
	if err != nil {
		log.Errorf("Failed to subscribe for events for contract address [%s]. Error [%s].", ew.contractAddress, err)
	}

	for {
		select {
		case err := <-sub.Err():
			log.Errorf("Event subscription failed with error [%s].", err)
		case vLog := <-logs:
			ew.handleLog(vLog, q)
		}
	}
}

func (ew *EthWatcher) handleLog(eventLog types.Log, q *queue.Queue) {
	if len(eventLog.Topics) == 0 {
		// TODO:
		return
	}

	switch eventLog.Topics[0].Hex() {
	case eth.LogEventBridgeEthBurnHash.Hex():
		log.Infof("Found a new [%s] event.", eth.LogEventBridgeEthBurn)
		// TODO:
	}
}

func NewEthereumWatcher(ethClient *ethClient.EthereumClient, config config.EthereumWatcher) *EthWatcher {
	contractAddress, err := ethClient.ValidateContractAddress(config.ContractAddress)
	if err != nil {
		log.Fatal(err)
	}

	return &EthWatcher{
		config:          config,
		client:          ethClient,
		contractAddress: contractAddress,
	}
}

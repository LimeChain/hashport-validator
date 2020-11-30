package ethereum

import (
	ethClient "github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum"
	bridgecontract "github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum/contracts/bridge"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/ethereum/bridge"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-watcher-sdk/queue"
	log "github.com/sirupsen/logrus"
)

type EthWatcher struct {
	config          config.Ethereum
	contractService *bridge.BridgeContractService
	client          *ethClient.EthereumClient
}

func (ew *EthWatcher) Watch(queue *queue.Queue) {
	log.Infof("[Ethereum Watcher] - Start listening for events for contract address [%s].", ew.config.BridgeContractAddress)
	go ew.listenForEvents(queue)
}

func (ew *EthWatcher) listenForEvents(q *queue.Queue) {
	events := make(chan *bridgecontract.BridgeBurn)
	sub, err := ew.contractService.WatchBurnEventLogs(nil, events)
	if err != nil {
		log.Errorf("Failed to subscribe for Burn Event Logs for contract address [%s]. Error [%s].", ew.config.BridgeContractAddress, err)
	}

	for {
		select {
		case err := <-sub.Err():
			log.Errorf("Burn Event Logs subscription failed. Error [%s].", err)
		case eventLog := <-events:
			ew.handleLog(eventLog, q)
		}
	}
}

func (ew *EthWatcher) handleLog(eventLog *bridgecontract.BridgeBurn, q *queue.Queue) {
	log.Infof("New Burn Event Log for [%s], Amount [%s], Receiver Address [%s] has been found.",
		eventLog.Account.Hex(),
		eventLog.Amount.String(),
		eventLog.ReceiverAddress)
	// TODO: push to queue with message type, corresponding to ETH Handler
}

func NewEthereumWatcher(ethClient *ethClient.EthereumClient, config config.Ethereum) *EthWatcher {
	bridgeContractService := bridge.NewBridgeContractService(ethClient, config)

	return &EthWatcher{
		config:          config,
		contractService: bridgeContractService,
		client:          ethClient,
	}
}

package ethereum

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

// Ethereum Node Client
type EthereumClient struct {
	client *ethclient.Client
	config config.Ethereum
}

func (ec *EthereumClient) SubscribeToEventLogs(contractAddress common.Address) (ethereum.Subscription, chan types.Log, error) {
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
	}
	logs := make(chan types.Log)

	sub, err := ec.client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		return nil, nil, err
	}

	return sub, logs, nil
}

func (ec *EthereumClient) ValidateContractAddress(contractAddress string) (*common.Address, error) {
	address := common.HexToAddress(contractAddress)

	bytecode, err := ec.client.CodeAt(context.Background(), address, nil)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to Get Code for contract address [%s].", contractAddress))
	}

	if len(bytecode) == 0 {
		return nil, errors.New(fmt.Sprintf("Provided address [%s] is not an Ethereum smart contract.", contractAddress))
	}

	return &address, nil
}

func NewEthereumClient(config config.Ethereum) *EthereumClient {
	client, err := ethclient.Dial(config.InfuraUrl)
	if err != nil {
		log.Fatalf("Failed to initialize EthereumClient. Error [%s]", err)
	}

	return &EthereumClient{
		config: config,
		client: client,
	}
}

package ethereum

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
	"math/big"
	"time"
)

// Ethereum Node Client
type EthereumClient struct {
	Client *ethclient.Client
	config config.Ethereum
}

func (ec *EthereumClient) ValidateContractAddress(contractAddress string) (*common.Address, error) {
	address := common.HexToAddress(contractAddress)

	bytecode, err := ec.Client.CodeAt(context.Background(), address, nil)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to Get Code for contract address [%s].", contractAddress))
	}

	if len(bytecode) == 0 {
		return nil, errors.New(fmt.Sprintf("Provided address [%s] is not an Ethereum smart contract.", contractAddress))
	}

	return &address, nil
}

func (ec *EthereumClient) WaitForTransactionSuccess(hash common.Hash) (isSuccessful bool, err error) {
	receipt, err := ec.waitForTransactionReceipt(hash)
	if err != nil {
		return false, err
	}

	// 1 == success
	return receipt.Status == 1, nil
}

func (ec *EthereumClient) waitForTransactionReceipt(hash common.Hash) (txReceipt *types.Receipt, err error) {
	for {
		_, isPending, err := ec.Client.TransactionByHash(context.Background(), hash)
		if err != nil {
			return nil, err
		}
		if !isPending {
			break
		}
		time.Sleep(5 * time.Second)
	}

	return ec.Client.TransactionReceipt(context.Background(), hash)
}

func NewEthereumClient(config config.Ethereum) *EthereumClient {
	client, err := ethclient.Dial(config.NodeUrl)
	if err != nil {
		log.Fatalf("Failed to initialize EthereumClient. Error [%s]", err)
	}

	ethereumClient := &EthereumClient{
		Client: client,
		config: config,
	}

	return ethereumClient
}

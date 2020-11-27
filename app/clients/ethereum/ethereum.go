package ethereum

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum/contracts/bridge"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

// Ethereum Node Client
type EthereumClient struct {
	client           *ethclient.Client
	config           config.Ethereum
	contractInstance *bridge.Bridge
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

func (ec *EthereumClient) SubmitSignatures(opts *bind.TransactOpts, ctm *proto.CryptoTransferMessage, signatures [][]byte) (*types.Transaction, error) {
	amountBn, err := helper.ToBigInt(strconv.Itoa(int(ctm.Amount)))
	if err != nil {
		return nil, err
	}

	feeBn, err := helper.ToBigInt(ctm.Fee)
	if err != nil {
		return nil, err
	}

	return ec.contractInstance.Mint(
		opts,
		[]byte(ctm.TransactionId),
		common.HexToAddress(ctm.EthAddress),
		amountBn,
		feeBn,
		signatures)
}

func (ec *EthereumClient) WaitForTransactionStatus(hash common.Hash) (isSuccessful bool, err error) {
	receipt, err := ec.waitForTransactionReceipt(hash)
	if err != nil {
		return false, err
	}

	// 1 == success
	return receipt.Status == 1, nil
}

func (ec *EthereumClient) waitForTransactionReceipt(hash common.Hash) (txReceipt *types.Receipt, err error) {
	for {
		_, isPending, err := ec.client.TransactionByHash(context.Background(), hash)
		if err != nil {
			return nil, err
		}
		if !isPending {
			break
		}
		time.Sleep(5 * time.Second)
	}

	return ec.client.TransactionReceipt(context.Background(), hash)
}

func NewEthereumClient(config config.Ethereum) *EthereumClient {
	client, err := ethclient.Dial(config.InfuraUrl)
	if err != nil {
		log.Fatalf("Failed to initialize EthereumClient. Error [%s]", err)
	}

	ethereumClient := &EthereumClient{
		config: config,
		client: client,
	}

	bridgeContractAddress, err := ethereumClient.ValidateContractAddress(config.BridgeContractAddress)
	if err != nil {
		log.Fatal(err)
	}

	bridgeContractInstance, err := bridge.NewBridge(*bridgeContractAddress, ethereumClient.client)
	if err != nil {
		log.Fatalf("Failed to initialize Bridge Contract Instance at [%s]. Error [%s].", config.BridgeContractAddress, err)
	}

	ethereumClient.contractInstance = bridgeContractInstance

	return ethereumClient
}

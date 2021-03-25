/*
 * Copyright 2021 LimeChain Ltd.
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
	"math/big"
	"time"
)

// Client Ethereum JSON RPC Client
type Client struct {
	chainId *big.Int
	config  config.Ethereum
	*ethclient.Client
	logger *log.Entry
}

// NewClient creates new instance of an Ethereum client
func NewClient(c config.Ethereum) *Client {
	logger := config.GetLoggerFor(fmt.Sprintf("Ethereum Client"))
	client, err := ethclient.Dial(c.NodeUrl)
	if err != nil {
		logger.Fatalf("Failed to initialize Client. Error [%s]", err)
	}

	chainId, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatalf("Failed to get ChainID. Error [%s]", err)
	}

	return &Client{
		chainId,
		c,
		client,
		logger,
	}
}

func (ec *Client) ChainID() *big.Int {
	return ec.chainId
}

// GetClients returns the instance of a ethclient already established connection to a JSON RPC Ethereum Node
func (ec *Client) GetClient() *ethclient.Client {
	return ec.Client
}

// ValidateContractDeployedAt performs validation that a smart contract is deployed at the provided address
func (ec *Client) ValidateContractDeployedAt(contractAddress string) (*common.Address, error) {
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

// WaitForTransaction waits for transaction receipt and depending on receipt status calls one of the provided functions
// onSuccess is called once the TX is successfully mined
// onRevert is called once the TX is mined but it reverted
// onError is called if an error occurs while waiting for TX to go into one of the other 2 states
func (ec *Client) WaitForTransaction(hex string, onSuccess, onRevert func(), onError func(err error)) {
	go func() {
		receipt, err := ec.waitForTransactionReceipt(common.HexToHash(hex))
		if err != nil {
			ec.logger.Errorf("Error occurred while monitoring TX [%s]. Error: %s", hex, err)
			onError(err)
			return
		}

		if receipt.Status == 1 {
			ec.logger.Debugf("TX [%s] was successfully mined", hex)
			onSuccess()
		} else {
			ec.logger.Debugf("TX [%s] reverted", hex)
			onRevert()
		}
		return
	}()
	ec.logger.Debugf("Added new Transaction [%s] for monitoring", hex)
}

// waitForTransactionReceipt Polls the provided hash every 5 seconds until the transaction mined (either successfully or reverted)
func (ec *Client) waitForTransactionReceipt(hash common.Hash) (txReceipt *types.Receipt, err error) {
	for {
		_, isPending, err := ec.Client.TransactionByHash(context.Background(), hash)

		// try again mechanism in case transaction is not validated for tx mempool yet
		if errors.Is(ethereum.NotFound, err) {
			time.Sleep(5 * time.Second)
			_, isPending, err = ec.Client.TransactionByHash(context.Background(), hash)
		}

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

func (ec *Client) WaitBlocks(numberOfBlocks uint64) error {
	if numberOfBlocks < 1 {
		return errors.New("numberOfBlocks should be a positive number")
	}

	blockNumber, err := ec.BlockNumber(context.Background())
	if err != nil {
		return err
	}

	target := blockNumber + numberOfBlocks
	for {
		blockNumber, err = ec.BlockNumber(context.Background())
		if err != nil {
			return err
		}

		if blockNumber == target {
			return nil
		}
	}
}

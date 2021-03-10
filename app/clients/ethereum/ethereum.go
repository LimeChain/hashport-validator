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
}

// NewClient creates new instance of an Ethereum client
func NewClient(config config.Ethereum) *Client {
	client, err := ethclient.Dial(config.NodeUrl)
	if err != nil {
		log.Fatalf("Failed to initialize Client. Error [%s]", err)
	}

	chainId, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatalf("Failed to get ChainID. Error [%s]", err)
	}

	return &Client{
		chainId,
		config,
		client,
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

// WaitForTransactionSuccess polls the JSON RPC node every 5 seconds for any updates (whether TX is mined) for the provided Hash
func (ec *Client) WaitForTransactionSuccess(hash common.Hash) (isSuccessful bool, err error) {
	receipt, err := ec.waitForTransactionReceipt(hash)
	if err != nil {
		return false, err
	}

	// 1 == success
	return receipt.Status == 1, nil
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

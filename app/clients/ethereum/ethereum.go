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
	if c.BlockConfirmations < 1 {
		logger.Fatalf("BlockConfirmations should be a positive number")
	}

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

func (ec *Client) WaitForConfirmations(raw types.Log) error {
	target := raw.BlockNumber + ec.config.BlockConfirmations
	for {
		currentBlockNumber, err := ec.BlockNumber(context.Background())
		if err != nil {
			ec.logger.Errorf("[%s] Failed retrieving block number.", raw.TxHash.String())
			return err
		}

		if target <= currentBlockNumber {
			receipt, err := ec.TransactionReceipt(context.Background(), raw.TxHash)
			if errors.Is(ethereum.NotFound, err) {
				ec.logger.Infof("[%s] Ethereum TX went into an uncle block.", raw.TxHash.String())
				return err
			}
			if err != nil {
				ec.logger.Infof("[%s] Failed to get Transaction receipt - Error: %s", raw.TxHash.String(), err)
				return err
			}

			if receipt.Status == 1 {
				ec.logger.Debugf("[%s] Transaction received [%d] block confirmations", raw.TxHash.String(), ec.config.BlockConfirmations)
				return nil
			} else {
				ec.logger.Debugf("[%s] Transaction reverted", raw.TxHash.String())
				return errors.New("reverted")
			}
		}
		time.Sleep(time.Second * 5)
	}
}

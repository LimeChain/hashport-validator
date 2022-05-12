/*
 * Copyright 2022 LimeChain Ltd.
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

package evm

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/retry"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

// Used as a maximum amount of retries that need to be done when executing
// RetryBlockNumber, RetryFilterLogs
const executionRetries = 10

// Client EVM JSON RPC Client
type Client struct {
	config config.Evm
	client.Core
	logger  *log.Entry
	chainId uint64
}

// NewClient creates new instance of an EVM client
func NewClient(c config.Evm, chainId uint64) *Client {
	logger := config.GetLoggerFor(fmt.Sprintf("EVM Client"))
	if c.BlockConfirmations < 1 {
		logger.Fatalf("BlockConfirmations should be a positive number")
	}

	var client client.Core
	client, err := ethclient.Dial(c.NodeUrl)
	if err != nil {
		logger.Fatalf("Failed to initialize Client with Chain Id [%v]. Error [%s]", chainId, err)
	}

	return &Client{
		c,
		client,
		logger,
		chainId,
	}
}
func (ec *Client) GetChainID() uint64 {
	return ec.chainId
}

func (ec *Client) SetChainID(chainId uint64) {
	ec.chainId = chainId
}

func (ec *Client) ChainID(ctx context.Context) (*big.Int, error) {
	return ec.Core.ChainID(ctx)
}

// GetClient returns the instance of an ethclient already established connection to a JSON RPC EVM Node
func (ec *Client) GetClient() client.Core {
	return ec.Core
}

// ValidateContractDeployedAt performs validation that a smart contract is deployed at the provided address
func (ec *Client) ValidateContractDeployedAt(contractAddress string) (*common.Address, error) {
	address := common.HexToAddress(contractAddress)

	bytecode, err := ec.CodeAt(context.Background(), address, nil)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to Get Code for contract address [%s].", contractAddress))
	}

	if len(bytecode) == 0 {
		return nil, errors.New(fmt.Sprintf("Provided address [%s] is not an EVM smart contract.", contractAddress))
	}

	return &address, nil
}

// GetBlockTimestamp retrieves the timestamp of the given block
func (ec *Client) GetBlockTimestamp(blockNumber *big.Int) uint64 {
	block, err := ec.BlockByNumber(context.Background(), blockNumber)

	if err != nil {
		ec.logger.Errorf("Failed to get block [%s]. Error: [%s]. Retrying...", blockNumber, err)
		time.Sleep(5 * time.Second)
		return ec.GetBlockTimestamp(blockNumber)
	}

	return block.Time()
}

// WaitForTransactionCallback waits for transaction receipt and depending on receipt status calls one of the provided functions
// onSuccess is called once the TX is successfully mined
// onRevert is called once the TX is mined but it reverted
// onError is called if an error occurs while waiting for TX to go into one of the other 2 states
func (ec *Client) WaitForTransactionCallback(hex string, onSuccess, onRevert func(), onError func(err error)) {
	go ec.checkTransactionReceipt(hex, onSuccess, onRevert, onError)
	ec.logger.Debugf("Added new Transaction [%s] for monitoring", hex)
}

func (ec *Client) checkTransactionReceipt(hex string, onSuccess, onRevert func(), onError func(err error)) {
	receipt, err := ec.WaitForTransactionReceipt(common.HexToHash(hex))
	if err != nil {
		ec.logger.Errorf("[%s] - Error occurred while monitoring. Error: [%s]", hex, err)
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
}

// WaitForTransactionReceipt Polls the provided hash every 5 seconds until the transaction mined (either successfully or reverted)
func (ec *Client) WaitForTransactionReceipt(hash common.Hash) (txReceipt *types.Receipt, err error) {
	for {
		_, isPending, err := ec.TransactionByHash(context.Background(), hash)

		// try again mechanism in case transaction is not validated for tx mempool yet
		if errors.Is(ethereum.NotFound, err) {
			time.Sleep(5 * time.Second)
			_, isPending, err = ec.TransactionByHash(context.Background(), hash)
		}

		if err != nil {
			return nil, err
		}
		if !isPending {
			break
		}
		time.Sleep(5 * time.Second)
	}

	return ec.TransactionReceipt(context.Background(), hash)
}

func (ec *Client) GetPrivateKey() string {
	return ec.config.PrivateKey
}

func (ec Client) BlockConfirmations() uint64 {
	return ec.config.BlockConfirmations
}

// RetryBlockNumber returns the most recent block number
// Uses a retry mechanism in case the filter query is stuck
func (ec Client) RetryBlockNumber() (uint64, error) {
	blockNumberFunc := func() <-chan retry.Result {
		r := make(chan retry.Result)
		go func() {
			defer close(r)

			block, err := ec.BlockNumber(context.Background())
			r <- retry.Result{
				Value: block,
				Error: err,
			}
		}()

		return r
	}

	result, err := service.Retry(blockNumberFunc, executionRetries)
	if err != nil {
		return 0, err
	}

	block, ok := result.(uint64)
	if !ok {
		return 0, errors.New(fmt.Sprintf("failed to cast block [%v]", result))
	}

	return block, nil
}

// RetryFilterLogs returns the logs from the input query
// Uses a retry mechanism in case the filter query is stuck
func (ec Client) RetryFilterLogs(query ethereum.FilterQuery) ([]types.Log, error) {
	filterLogsFunc := func() <-chan retry.Result {
		r := make(chan retry.Result)
		go func() {
			defer close(r)

			logs, err := ec.FilterLogs(context.Background(), query)
			r <- retry.Result{
				Value: logs,
				Error: err,
			}
		}()

		return r
	}

	result, err := service.Retry(filterLogsFunc, executionRetries)
	if err != nil {
		return nil, err
	}

	logs, ok := result.([]types.Log)
	if !ok {
		return nil, errors.New(fmt.Sprintf("failed to cast logs [%v]", result))
	}

	return logs, nil
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
				ec.logger.Infof("[%s] EVM TX went into an uncle block.", raw.TxHash.String())
				return err
			}
			if err != nil {
				ec.logger.Infof("[%s] Failed to get Transaction receipt - Error: [%s]", raw.TxHash.String(), err)
				return err
			}

			if receipt.BlockNumber.Uint64() != raw.BlockNumber {
				ec.logger.Debugf("[%s] has been moved from original block", raw.TxHash.String())
				return errors.New("moved from original block")
			}

			return nil
		}
		time.Sleep(time.Second * 5)
	}
}

func (ec *Client) WaitForTransaction(hash common.Hash) (*types.Transaction, error) {
	res, err := service.Retry(func() <-chan retry.Result {
		r := make(chan retry.Result)
		go func() {
			defer close(r)

			tx, _, err := ec.TransactionByHash(context.Background(), hash)
			if err != nil {
				r <- retry.Result{
					Value: nil,
					Error: err,
				}
				return
			}

			r <- retry.Result{
				Value: tx,
				Error: nil,
			}
		}()

		return r
	}, executionRetries)
	if err != nil {
		return nil, err
	}
	return res.(*types.Transaction), nil
}

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
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/config"
)

type ClientPool struct {
	clients []client.EVM
	retries int
}

func NewClientPool(c config.EvmPool, chainId uint64) *ClientPool {
	nodeURLs := c.NodeUrls
	clients := make([]client.EVM, 0, len(nodeURLs))
	for _, nodeURL := range nodeURLs {
		configEvm := config.Evm{
			BlockConfirmations: c.BlockConfirmations,
			NodeUrl:            nodeURL,
			PrivateKey:         c.PrivateKey,
			StartBlock:         c.StartBlock,
			PollingInterval:    c.PollingInterval,
			MaxLogsBlocks:      c.MaxLogsBlocks,
		}
		clients = append(clients, NewClient(configEvm, chainId))
	}

	retry := len(clients) * 3

	return &ClientPool{
		clients: clients,
		retries: retry,
	}
}

func (cp *ClientPool) getClient(idx int) client.EVM {
	return cp.clients[idx%len(cp.clients)]
}

func (cp *ClientPool) retryOperation(operation func(client.EVM) (interface{}, error)) (interface{}, error) {
	var err error
	for i := 0; i < cp.retries; i++ {
		client := cp.getClient(i)
		result, e := operation(client)
		if e == nil {
			return result, nil
		}
		err = e
	}

	return nil, err
}

func (cp *ClientPool) ChainID(ctx context.Context) (*big.Int, error) {
	operation := func(c client.EVM) (interface{}, error) {
		return c.ChainID(ctx)
	}

	result, err := cp.retryOperation(operation)
	if err != nil {
		return nil, err
	}

	return result.(*big.Int), nil
}

func (cp *ClientPool) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	operation := func(c client.EVM) (interface{}, error) {
		return c.CodeAt(ctx, contract, blockNumber)
	}

	result, err := cp.retryOperation(operation)
	if err != nil {
		return nil, err
	}

	return result.([]byte), nil
}

func (cp *ClientPool) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	operation := func(c client.EVM) (interface{}, error) {
		return c.CallContract(ctx, call, blockNumber)
	}

	result, err := cp.retryOperation(operation)
	if err != nil {
		return nil, err
	}

	return result.([]byte), nil
}

func (cp *ClientPool) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	operation := func(c client.EVM) (interface{}, error) {
		return c.HeaderByNumber(ctx, number)
	}

	result, err := cp.retryOperation(operation)
	if err != nil {
		return nil, err
	}

	return result.(*types.Header), nil
}

func (cp *ClientPool) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	operation := func(c client.EVM) (interface{}, error) {
		return c.PendingCodeAt(ctx, account)
	}

	result, err := cp.retryOperation(operation)
	if err != nil {
		return nil, err
	}

	return result.([]byte), nil
}

func (cp *ClientPool) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	operation := func(c client.EVM) (interface{}, error) {
		return c.PendingNonceAt(ctx, account)
	}

	result, err := cp.retryOperation(operation)
	if err != nil {
		return 0, err
	}

	return result.(uint64), nil
}

func (cp *ClientPool) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	operation := func(c client.EVM) (interface{}, error) {
		return c.SuggestGasPrice(ctx)
	}

	result, err := cp.retryOperation(operation)
	if err != nil {
		return nil, err
	}

	return result.(*big.Int), nil
}

func (cp *ClientPool) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	operation := func(c client.EVM) (interface{}, error) {
		return c.SuggestGasTipCap(ctx)
	}

	result, err := cp.retryOperation(operation)
	if err != nil {
		return nil, err
	}

	return result.(*big.Int), nil
}

func (cp *ClientPool) EstimateGas(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error) {
	operation := func(c client.EVM) (interface{}, error) {
		return c.EstimateGas(ctx, call)
	}

	result, err := cp.retryOperation(operation)
	if err != nil {
		return 0, err
	}

	return result.(uint64), nil
}

func (cp *ClientPool) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	operation := func(c client.EVM) (interface{}, error) {
		return nil, c.SendTransaction(ctx, tx)
	}

	_, err := cp.retryOperation(operation)
	return err
}

func (cp *ClientPool) FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error) {
	operation := func(c client.EVM) (interface{}, error) {
		return c.FilterLogs(ctx, query)
	}

	result, err := cp.retryOperation(operation)
	if err != nil {
		return nil, err
	}

	return result.([]types.Log), nil
}

func (cp *ClientPool) SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	operation := func(c client.EVM) (interface{}, error) {
		return c.SubscribeFilterLogs(ctx, query, ch)
	}

	result, err := cp.retryOperation(operation)
	if err != nil {
		return nil, err
	}

	return result.(ethereum.Subscription), nil
}

func (cp *ClientPool) BlockNumber(ctx context.Context) (uint64, error) {
	operation := func(c client.EVM) (interface{}, error) {
		return c.BlockNumber(ctx)
	}

	result, err := cp.retryOperation(operation)
	if err != nil {
		return 0, err
	}

	return result.(uint64), nil
}

func (cp *ClientPool) ValidateContractDeployedAt(contractAddress string) (*common.Address, error) {
	operation := func(c client.EVM) (interface{}, error) {
		return c.ValidateContractDeployedAt(contractAddress)
	}

	result, err := cp.retryOperation(operation)
	if err != nil {
		return nil, err
	}

	return result.(*common.Address), nil
}

func (cp *ClientPool) RetryBlockNumber() (uint64, error) {
	operation := func(c client.EVM) (interface{}, error) {
		return c.RetryBlockNumber()
	}

	result, err := cp.retryOperation(operation)
	if err != nil {
		return 0, err
	}

	return result.(uint64), nil
}

func (cp *ClientPool) RetryFilterLogs(query ethereum.FilterQuery) ([]types.Log, error) {
	operation := func(c client.EVM) (interface{}, error) {
		return c.RetryFilterLogs(query)
	}

	result, err := cp.retryOperation(operation)
	if err != nil {
		return nil, err
	}

	return result.([]types.Log), nil
}

func (cp *ClientPool) WaitForTransactionReceipt(hash common.Hash) (*types.Receipt, error) {
	operation := func(c client.EVM) (interface{}, error) {
		return c.WaitForTransactionReceipt(hash)
	}

	result, err := cp.retryOperation(operation)
	if err != nil {
		return nil, err
	}

	return result.(*types.Receipt), nil
}

func (cp *ClientPool) RetryTransactionByHash(hash common.Hash) (*types.Transaction, error) {
	operation := func(c client.EVM) (interface{}, error) {
		return c.RetryTransactionByHash(hash)
	}

	result, err := cp.retryOperation(operation)
	if err != nil {
		return nil, err
	}

	return result.(*types.Transaction), nil
}

func (cp *ClientPool) WaitForTransactionCallback(hex string, onSuccess, onRevert func(), onError func(err error)) {
	cp.clients[0].WaitForTransactionCallback(hex, onSuccess, onRevert, onError)
}

func (cp *ClientPool) WaitForConfirmations(raw types.Log) error {
	operation := func(c client.EVM) (interface{}, error) {
		return nil, c.WaitForConfirmations(raw)
	}

	_, err := cp.retryOperation(operation)
	return err
}

func (cp *ClientPool) GetChainID() uint64 {
	return cp.clients[0].GetChainID()
}

func (cp *ClientPool) SetChainID(chainID uint64) {
	for _, client := range cp.clients {
		client.SetChainID(chainID)
	}
}

func (cp *ClientPool) GetClient() client.Core {
	return cp.clients[0].GetClient()
}

func (cp *ClientPool) GetPrivateKey() string {
	return cp.clients[0].GetPrivateKey()
}

func (cp *ClientPool) BlockConfirmations() uint64 {
	return cp.clients[0].BlockConfirmations()
}

func (cp *ClientPool) GetBlockTimestamp(blockNumber *big.Int) uint64 {
	return cp.clients[0].GetBlockTimestamp(blockNumber)
}

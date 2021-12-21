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

package client

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

type EVM interface {
	ChainID(ctx context.Context) (*big.Int, error)
	GetClient() Core
	BlockNumber(ctx context.Context) (uint64, error)
	FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error)
	GetBlockTimestamp(blockNumber *big.Int) uint64
	ValidateContractDeployedAt(contractAddress string) (*common.Address, error)
	// WaitForTransaction waits for transaction receipt and depending on receipt status calls one of the provided functions
	// onSuccess is called once the TX is successfully mined
	// onRevert is called once the TX is mined but it reverted
	// onError is called if an error occurs while waiting for TX to go into one of the other 2 states
	WaitForTransaction(hex string, onSuccess, onRevert func(), onError func(err error))
	// WaitForConfirmations starts a loop which ends either when we reach the target block number or an error occurs with block number retrieval
	WaitForConfirmations(raw types.Log) error
	// GetPrivateKey retrieves private key used for the specific EVM Client
	GetPrivateKey() string
	BlockConfirmations() uint64
	// RetryBlockNumber returns the most recent block number
	// Uses a retry mechanism in case the filter query is stuck
	RetryBlockNumber() (uint64, error)
	// RetryFilterLogs returns the logs from the input query
	// Uses a retry mechanism in case the filter query is stuck
	RetryFilterLogs(query ethereum.FilterQuery) ([]types.Log, error)
}

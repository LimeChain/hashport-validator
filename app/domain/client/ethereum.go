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
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
)

type Ethereum interface {
	ChainID() *big.Int
	GetClient() *ethclient.Client
	ValidateContractDeployedAt(contractAddress string) (*common.Address, error)
	// WaitForTransaction waits for transaction receipt and depending on receipt status calls one of the provided functions
	// onSuccess is called once the TX is successfully mined
	// onRevert is called once the TX is mined but it reverted
	// onError is called if an error occurs while waiting for TX to go into one of the other 2 states
	WaitForTransaction(hex string, onSuccess, onRevert func(), onError func(err error))
	// WaitBlocks starts a loop which ends either when we reach the target block number or an error occurs with block number retrieval
	WaitBlocks(numberOfBlocks uint64) error
}

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
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	tc "github.com/limechain/hedera-eth-bridge-validator/test/test-config"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"math/big"
	"sync"
	"testing"
)

var (
	onExecution = "Pending"
	wg          sync.WaitGroup
	nodeClient  = NewClient(tc.TestConfig.Validator.Clients.Ethereum)
)

func TestNewClient(t *testing.T) {
	assert.NotNil(t, nodeClient)
	assert.Equal(t, nodeClient.config.PrivateKey, tc.TestConfig.Validator.Clients.Ethereum.PrivateKey)
	assert.Equal(t, nodeClient.config.NodeUrl, tc.TestConfig.Validator.Clients.Ethereum.NodeUrl)
	assert.Equal(t, nodeClient.config.BlockConfirmations, tc.TestConfig.Validator.Clients.Ethereum.BlockConfirmations)
	assert.Equal(t, nodeClient.config.RouterContractAddress, tc.TestConfig.Validator.Clients.Ethereum.RouterContractAddress)
}

func TestGetClient(t *testing.T) {
	assert.NotNil(t, nodeClient)

	client := nodeClient.GetClient()
	assert.NotNil(t, client)
}

func TestNewClientConfirmationsError(t *testing.T) {
	tc.TestConfig.Validator.Clients.Ethereum.BlockConfirmations = 0
	defer func() { tc.TestConfig.Validator.Clients.Ethereum.BlockConfirmations = 5 }()
	defer func() { log.StandardLogger().ExitFunc = nil }()

	var fatal bool
	log.StandardLogger().ExitFunc = func(int) { fatal = true }

	fatal = false
	NewClient(tc.TestConfig.Validator.Clients.Ethereum)
	assert.Equal(t, true, fatal)
}

func TestNewClientChainIDError(t *testing.T) {
	tc.TestConfig.Validator.Clients.Ethereum.NodeUrl = "http://127.0.0.1"
	defer func() { log.StandardLogger().ExitFunc = nil }()
	defer func() {
		tc.TestConfig.Validator.Clients.Ethereum.NodeUrl = "wss://ropsten.infura.io/ws/v3/8b64d65996d24dc0aae2e0c6029e5a9b"
	}()

	var fatal bool
	log.StandardLogger().ExitFunc = func(int) { fatal = true }

	fatal = false
	NewClient(tc.TestConfig.Validator.Clients.Ethereum)
	assert.Equal(t, true, fatal)

}

func TestNewClientDialError(t *testing.T) {
	tc.TestConfig.Validator.Clients.Ethereum.NodeUrl = "wss://ropsten.infura.io/ws/v3/"

	defer func() { log.StandardLogger().ExitFunc = nil }()
	defer func() {
		tc.TestConfig.Validator.Clients.Ethereum.NodeUrl = "wss://ropsten.infura.io/ws/v3/8b64d65996d24dc0aae2e0c6029e5a9b"
	}()
	var fatal bool
	log.StandardLogger().ExitFunc = func(int) { fatal = true }

	fatal = false

	assert.Panics(t, func() {
		NewClient(tc.TestConfig.Validator.Clients.Ethereum)
	})
	assert.Equal(t, true, fatal)
}

func TestChainID(t *testing.T) {
	chainID := nodeClient.ChainID()
	expectedChainID := big.NewInt(3)
	assert.Equal(t, expectedChainID, chainID)
}

func TestValidateContractDeployedAt(t *testing.T) {
	contractAddress := "0xbaA7610B498f7527D58bDced6046a8e7202180FD"
	address, err := nodeClient.ValidateContractDeployedAt(contractAddress)
	assert.Equal(t, *address, common.HexToAddress(contractAddress))
	assert.Nil(t, err)
}

func TestValidateContractNonContractAddress(t *testing.T) {
	contractAddress := "0xbaA7610B498f7527D58bDced6046a8e7202180Ff"
	address, err := nodeClient.ValidateContractDeployedAt(contractAddress)
	assert.NotNil(t, err)
	assert.Nil(t, address)
}

func TestWaitForTransaction(t *testing.T) {
	transactionHash := "0x7c39eecd9af35f2a59bda7c0a600fea0cc99b26444de5418e7eba0285b2a99e8"

	wg.Add(1)
	nodeClient.WaitForTransaction(transactionHash, onSuccess, onRevert, onError)
	wg.Wait()

	assert.Equal(t, "Success", onExecution)
}

func TestWaitForTransactionRevert(t *testing.T) {
	transactionHash := "0x9634c0496039bee73555688efd5943a5cf86caa0d4826edb0a7d7642fc8e0392"
	wg.Add(1)
	nodeClient.WaitForTransaction(transactionHash, onSuccess, onRevert, onError)
	wg.Wait()

	assert.Equal(t, "Revert", onExecution)
}

func TestWaitForTransactionError(t *testing.T) {
	transactionHash := "0x9634c0496039bee73555688efd5943a5cf86caa0d4826edb0a7d7642fc8e0390"
	wg.Add(1)
	nodeClient.WaitForTransaction(transactionHash, onSuccess, onRevert, onError)
	wg.Wait()

	assert.Equal(t, "Error", onExecution)
}

func TestWaitForTransactionReceipt(t *testing.T) {
	transactionHash := "0x7c39eecd9af35f2a59bda7c0a600fea0cc99b26444de5418e7eba0285b2a99e8"
	receipt, err := nodeClient.WaitForTransactionReceipt(common.HexToHash(transactionHash))
	expectedStatus := uint64(1)
	assert.Equal(t, expectedStatus, receipt.Status)
	assert.Nil(t, err)
}

func TestWaitForTransactionReceiptError(t *testing.T) {
	transactionHash := "0x9634c0496039bee73555688efd5943a5cf86caa0d4826edb0a7d7642fc8e0390"
	receipt, err := nodeClient.WaitForTransactionReceipt(common.HexToHash(transactionHash))
	assert.Nil(t, receipt)
	assert.Error(t, err)
}

func TestWaitForConfirmations(t *testing.T) {
	raw := types.Log{
		Address: common.HexToAddress("0xbaA7610B498f7527D58bDced6046a8e7202180FD"),
		Topics: []common.Hash{
			common.HexToHash("asd"),
		},
		Data:        []byte{0},
		BlockNumber: 10351812,
		TxHash:      common.HexToHash("0x7c39eecd9af35f2a59bda7c0a600fea0cc99b26444de5418e7eba0285b2a99e8"),
		TxIndex:     2,
		BlockHash:   common.HexToHash("0x1e3c56de9b4060e9fa2b83e6100aec92e0edeaeb14247ec8217fba96994bf6e3"),
		Index:       1,
		Removed:     false,
	}

	err := nodeClient.WaitForConfirmations(raw)
	assert.Nil(t, err)
}

func TestWaitForConfirmationsBlockError(t *testing.T) {
	raw := types.Log{
		Address: common.HexToAddress("0xbaA7610B498f7527D58bDced6046a8e7202180FD"),
		Topics: []common.Hash{
			common.HexToHash("asd"),
		},
		Data:        []byte{0},
		BlockNumber: 10351,
		TxHash:      common.HexToHash("0x7c39eecd9af35f2a59bda7c0a600fea0cc99b26444de5418e7eba0285b2a99e8"),
		TxIndex:     2,
		BlockHash:   common.HexToHash("0x1e3c56de9b4060e9fa2b83e6100aec92e0edeaeb14247ec8217fba96994bf6e3"),
		Index:       1,
		Removed:     false,
	}

	err := nodeClient.WaitForConfirmations(raw)
	assert.Error(t, err)
}

func onSuccess() {
	onExecution = "Success"
	defer wg.Done()
}

func onRevert() {
	onExecution = "Revert"
	defer wg.Done()
}

func onError(err error) {
	onExecution = "Error"
	defer wg.Done()
}

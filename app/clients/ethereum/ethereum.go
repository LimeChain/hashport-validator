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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"math/big"
	"net/http"
	"time"
)

// Ethereum Node Client
type EthereumClient struct {
	Client     *ethclient.Client
	httpClient *http.Client
	config     config.Ethereum
}

func NewEthereumClient(config config.Ethereum) *EthereumClient {
	client, err := ethclient.Dial(config.NodeUrl)
	if err != nil {
		log.Fatalf("Failed to initialize EthereumClient. Error [%s]", err)
	}

	ethereumClient := &EthereumClient{
		httpClient: &http.Client{},
		Client:     client,
		config:     config,
	}

	return ethereumClient
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

func (ec *EthereumClient) GetSlowGasPrice() (uint64, error) {
	apiURL := "https://api.etherscan.io/api?module=gastracker&action=gasoracle&apikey=2TX7CUPTGNKIFYC4V12GTJSQJ73EKZ5Z21"

	response, err := ec.httpClient.Get(apiURL)

	bodyBytes, err := readResponseBody(response)
	if err != nil {
		return 0, err
	}

	var data map[string]interface{}
	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		return 0, err
	}

	result := data["result"]
	pricesData := result.(map[string]interface{})
	slowGasPrice := pricesData["SafeGasPrice"].(uint64)
	if err != nil {
		return 0, err
	}

	return slowGasPrice, nil
}

func readResponseBody(response *http.Response) ([]byte, error) {
	defer response.Body.Close()

	return ioutil.ReadAll(response.Body)
}

func (ec *EthereumClient) EstimateGas(amount uint64) (uint64, error) {
	slowGasPrice, err := ec.GetSlowGasPrice()
	if err != nil {
		return 0, err
	}

	msg := ethereum.CallMsg{
		From:     common.HexToAddress("0x37EA216A28628eCC1de4d982cEc46569203a9F4a"),
		To:       nil,
		Gas:      0, // Gas limit: in app.yml -> base gas usage + X * gas usage per signature, X >= 50% of validator signatures ... get count of operators from smart contract
		GasPrice: new(big.Int).SetUint64(slowGasPrice),
		Value:    new(big.Int).SetUint64(amount),
		Data:     nil, // Should send necessary data (signatures, etc.)
	}

	gasEstimation, err := ec.Client.EstimateGas(context.Background(), msg)
	if err != nil {
		return 0, err
	}

	return gasEstimation, nil
}

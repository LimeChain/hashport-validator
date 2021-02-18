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

package bridge

import (
<<<<<<< HEAD
	"strconv"
=======
>>>>>>> b54f2c1996e016dddf90856cc82e1a589a49d604
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	ethclient "github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum/contracts/bridge"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
	log "github.com/sirupsen/logrus"
)

type BridgeContractService struct {
	contractInstance *bridge.Bridge
	Client           *ethclient.EthereumClient
	mutex            sync.Mutex
}

func NewBridgeContractService(client *ethclient.EthereumClient, config config.Ethereum) *BridgeContractService {
	bridgeContractAddress, err := client.ValidateContractAddress(config.BridgeContractAddress)
	if err != nil {
		log.Fatal(err)
	}

	contractInstance, err := bridge.NewBridge(*bridgeContractAddress, client.Client)
	if err != nil {
		log.Fatalf("Failed to initialize Bridge Contract Instance at [%s]. Error [%s].", config.BridgeContractAddress, err)
	}

	return &BridgeContractService{
		Client:           client,
		contractInstance: contractInstance,
	}
}

func (bsc *BridgeContractService) SubmitSignatures(opts *bind.TransactOpts, ctm *proto.CryptoTransferMessage, signatures [][]byte) (*types.Transaction, error) {
	bsc.mutex.Lock()
	defer bsc.mutex.Unlock()

	amountBn, err := helper.ToBigInt(ctm.Amount)
	if err != nil {
		return nil, err
	}

	feeBn, err := helper.ToBigInt(ctm.Fee)
	if err != nil {
		return nil, err
	}

	return bsc.contractInstance.Mint(
		opts,
		[]byte(ctm.TransactionId),
		common.HexToAddress(ctm.EthAddress),
		amountBn,
		feeBn,
		signatures)
}

func (bsc *BridgeContractService) WatchBurnEventLogs(opts *bind.WatchOpts, sink chan<- *bridge.BridgeBurn) (event.Subscription, error) {
	var addresses []common.Address
	return bsc.contractInstance.WatchBurn(opts, sink, addresses, [][]byte{})
}

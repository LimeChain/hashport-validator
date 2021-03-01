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
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	ethclient "github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum/contracts/bridge"
	bridgecontract "github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum/contracts/bridge"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/model/bridge/custodians"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/model/bridge/servicefee"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
	log "github.com/sirupsen/logrus"
)

type BridgeContractService struct {
	contractInstance *bridge.Bridge
	Client           *ethclient.EthereumClient
	mutex            sync.Mutex
	custodians       custodians.Custodians
	servicefee       servicefee.Servicefee
	logger           *log.Entry
}

func NewBridgeContractService(client *ethclient.EthereumClient, c config.Ethereum) *BridgeContractService {
	bridgeContractAddress, err := client.ValidateContractAddress(c.BridgeContractAddress)
	if err != nil {
		log.Fatal(err)
	}

	contractInstance, err := bridge.NewBridge(*bridgeContractAddress, client.Client)
	if err != nil {
		log.Fatalf("Failed to initialize Bridge Contract Instance at [%s]. Error [%s].", c.BridgeContractAddress, err)
	}

	bridge := &BridgeContractService{
		Client:           client,
		contractInstance: contractInstance,
		logger:           config.GetLoggerFor("Bridge Contract Service"),
	}

	bridge.updateMembers()
	bridge.updateServiceFee()

	go bridge.listenForMemberUpdatedEvent()
	go bridge.listenForChangeFeeEvent()

	return bridge
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

func (bsc *BridgeContractService) GetCustodians() []string {
	return bsc.custodians.Get()
}

func (bsc *BridgeContractService) GetServiceFee() *big.Int {
	return bsc.servicefee.Get()
}

func (bsc *BridgeContractService) updateServiceFee() {
	newFee, err := bsc.contractInstance.ServiceFee(nil)
	if err != nil {
		bsc.logger.Fatal("Failed to get service fee.", err)
	}
	bsc.servicefee.Set(*newFee)
	bsc.logger.Infof("Updating service fee [%s]...", newFee)

}

func (bsc *BridgeContractService) updateMembers() {
	membersCount, err := bsc.contractInstance.MembersCount(nil)
	if err != nil {
		bsc.logger.Fatal("Failed to get members count.", err)
	}

	var newCustodiansArray []string
	for i := 0; i < int(membersCount.Int64()); i++ {
		addr, err := bsc.contractInstance.MemberAt(nil, big.NewInt(int64(i)))
		if err != nil {
			bsc.logger.Fatal("Failed to get member address.", err)
		}
		newCustodiansArray = append(newCustodiansArray, addr.String())
	}
	bsc.custodians.Set(newCustodiansArray)
	bsc.logger.Infof("Updating custodians list with [%s]", newCustodiansArray)

}

func (bsc *BridgeContractService) WatchBurnEventLogs(opts *bind.WatchOpts, sink chan<- *bridge.BridgeBurn) (event.Subscription, error) {
	var addresses []common.Address
	return bsc.contractInstance.WatchBurn(opts, sink, addresses, [][]byte{})
}

func (bsc *BridgeContractService) listenForChangeFeeEvent() {
	events := make(chan *bridgecontract.BridgeServiceFeeSet)
	sub, err := bsc.contractInstance.WatchServiceFeeSet(nil, events)
	if err != nil {
		bsc.logger.Fatal("Failed to subscribe for WatchServiceFeeSet Event Logs for contract. Error [%s].", err)
	}

	for {
		select {
		case err := <-sub.Err():
			bsc.logger.Errorf("ServiceFeeSet Event Logs subscription failed. Error [%s].", err)
			return
		case eventLog := <-events:
			bsc.servicefee.Set(*eventLog.NewServiceFee)
			bsc.logger.Infof("Updating service fee [%s]...", eventLog.NewServiceFee)
		}
	}
}

func (bsc *BridgeContractService) listenForMemberUpdatedEvent() {
	events := make(chan *bridgecontract.BridgeMemberUpdated)
	sub, err := bsc.contractInstance.WatchMemberUpdated(nil, events)
	if err != nil {
		bsc.logger.Fatal("Failed to subscribe for WatchMemberUpdated Event Logs for contract. Error [%s].", err)
	}

	for {
		select {
		case err := <-sub.Err():
			bsc.logger.Errorf("MemberUpdated Event Logs subscription failed. Error [%s].", err)
			return
		case <-events:
			bsc.updateMembers()
		}
	}
}

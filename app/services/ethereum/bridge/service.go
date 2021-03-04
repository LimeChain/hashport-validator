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
	"github.com/ethereum/go-ethereum/event"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/clients"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	abi "github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum/contracts/bridge"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
	log "github.com/sirupsen/logrus"
)

type ContractService struct {
	address    common.Address
	contract   *abi.Bridge
	Client     clients.Ethereum
	mutex      sync.Mutex
	members    Members
	serviceFee ServiceFee
	logger     *log.Entry
}

func NewContractService(client clients.Ethereum, c config.Ethereum) *ContractService {
	contractAddress, err := client.ValidateContractDeployedAt(c.BridgeContractAddress)
	if err != nil {
		log.Fatal(err)
	}

	contractInstance, err := abi.NewBridge(*contractAddress, client.GetClient())
	if err != nil {
		log.Fatalf("Failed to initialize Bridge Contract Instance at [%s]. Error [%s]", c.BridgeContractAddress, err)
	}

	contractService := &ContractService{
		address:  *contractAddress,
		Client:   client,
		contract: contractInstance,
		logger:   config.GetLoggerFor("Bridge Contract ContractService"),
	}

	contractService.updateMembers()
	contractService.updateServiceFee()

	go contractService.listenForMemberUpdatedEvent()
	go contractService.listenForChangeFeeEvent()

	return contractService
}

func (bsc *ContractService) GetContractAddress() common.Address {
	return bsc.address
}

func (bsc *ContractService) SubmitSignatures(opts *bind.TransactOpts, ctm *proto.CryptoTransferMessage, signatures [][]byte) (*types.Transaction, error) {
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

	return bsc.contract.Mint(
		opts,
		[]byte(ctm.TransactionId),
		common.HexToAddress(ctm.EthAddress),
		amountBn,
		feeBn,
		signatures)
}

func (bsc *ContractService) GetMembers() []string {
	return bsc.members.Get()
}

func (bsc *ContractService) GetServiceFee() *big.Int {
	return bsc.serviceFee.Get()
}

func (bsc *ContractService) WatchBurnEventLogs(opts *bind.WatchOpts, sink chan<- *abi.BridgeBurn) (event.Subscription, error) {
	var addresses []common.Address
	return bsc.contract.WatchBurn(opts, sink, addresses, [][]byte{})
}

func (bsc *ContractService) updateServiceFee() {
	newFee, err := bsc.contract.ServiceFee(nil)
	if err != nil {
		bsc.logger.Fatal("Failed to get service fee", err)
	}

	bsc.serviceFee.Set(*newFee)
	bsc.logger.Infof("Set service fee to [%s]", newFee)

}

func (bsc *ContractService) updateMembers() {
	membersCount, err := bsc.contract.MembersCount(nil)
	if err != nil {
		bsc.logger.Fatal("Failed to get members count", err)
	}

	var membersArray []string
	for i := 0; i < int(membersCount.Int64()); i++ {
		addr, err := bsc.contract.MemberAt(nil, big.NewInt(int64(i)))
		if err != nil {
			bsc.logger.Fatal("Failed to get member address", err)
		}
		membersArray = append(membersArray, addr.String())
	}
	bsc.members.Set(membersArray)
	bsc.logger.Infof("Set members list to %s", membersArray)

}

func (bsc *ContractService) listenForChangeFeeEvent() {
	events := make(chan *abi.BridgeServiceFeeSet)
	sub, err := bsc.contract.WatchServiceFeeSet(nil, events)
	if err != nil {
		bsc.logger.Fatal("Failed to subscribe for WatchServiceFeeSet Event Logs for contract. Error ", err)
	}

	for {
		select {
		case err := <-sub.Err():
			bsc.logger.Errorf("ServiceFeeSet Event Logs subscription failed. Error [%s].", err)
			return
		case eventLog := <-events:
			bsc.serviceFee.Set(*eventLog.NewServiceFee)
			bsc.logger.Infof(`Set service fee to [%s]`, eventLog.NewServiceFee)
		}
	}
}

func (bsc *ContractService) listenForMemberUpdatedEvent() {
	events := make(chan *abi.BridgeMemberUpdated)
	sub, err := bsc.contract.WatchMemberUpdated(nil, events)
	if err != nil {
		bsc.logger.Fatal("Failed to subscribe for WatchMemberUpdated Event Logs for contract. Error ", err)
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

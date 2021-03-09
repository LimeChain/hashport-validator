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

package contracts

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	abi "github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum/contracts/bridge"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	address    common.Address
	contract   *abi.Bridge
	Client     client.Ethereum
	mutex      sync.Mutex
	members    Members
	serviceFee ServiceFee
	logger     *log.Entry
}

func (bsc *Service) GetBridgeContractAddress() common.Address {
	return bsc.address
}

// GetServiceFee returns the current service fee configured in the Bridge contract
func (bsc *Service) GetServiceFee() *big.Int {
	return bsc.serviceFee.Get()
}

// GetMembers returns the array of bridge members currently set in the Bridge contract
func (bsc *Service) GetMembers() []string {
	return bsc.members.Get()
}

// IsMember returns true/false depending on whether the provided address is a Bridge member or not
func (bsc *Service) IsMember(address string) bool {
	for _, k := range bsc.members.Get() {
		if strings.ToLower(k) == strings.ToLower(address) {
			return true
		}
	}
	return false
}

// WatchBurnEventLogs creates a subscription for Burn Events emitted in the Bridge contract
func (bsc *Service) WatchBurnEventLogs(opts *bind.WatchOpts, sink chan<- *abi.BridgeBurn) (event.Subscription, error) {
	var addresses []common.Address
	return bsc.contract.WatchBurn(opts, sink, addresses)
}

// SubmitSignatures signs and broadcasts an Ethereum TX authorising the mint operation on the Ethereum network
func (bsc *Service) SubmitSignatures(opts *bind.TransactOpts, ctm *proto.TransferMessage, signatures [][]byte) (*types.Transaction, error) {
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

func (bsc *Service) updateServiceFee() {
	newFee, err := bsc.contract.ServiceFee(nil)
	if err != nil {
		bsc.logger.Fatal("Failed to get service fee", err)
	}

	bsc.serviceFee.Set(*newFee)
	bsc.logger.Infof("Set service fee to [%s]", newFee)

}

func (bsc *Service) updateMembers() {
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

func (bsc *Service) listenForChangeFeeEvent() {
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

func (bsc *Service) listenForMemberUpdatedEvent() {
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

// NewService creates new instance of a Contract Services based on the provided configuration
func NewService(client client.Ethereum, c config.Ethereum) *Service {
	contractAddress, err := client.ValidateContractDeployedAt(c.BridgeContractAddress)
	if err != nil {
		log.Fatal(err)
	}

	contractInstance, err := abi.NewBridge(*contractAddress, client.GetClient())
	if err != nil {
		log.Fatalf("Failed to initialize Bridge Contract Instance at [%s]. Error [%s]", c.BridgeContractAddress, err)
	}

	contractService := &Service{
		address:  *contractAddress,
		Client:   client,
		contract: contractInstance,
		logger:   config.GetLoggerFor("Bridge Contract Service"),
	}

	contractService.updateMembers()
	contractService.updateServiceFee()

	go contractService.listenForMemberUpdatedEvent()
	go contractService.listenForChangeFeeEvent()

	return contractService
}

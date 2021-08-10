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
	"github.com/ethereum/go-ethereum/event"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	log "github.com/sirupsen/logrus"
)

const (
	nilErc20Address = "0x0000000000000000000000000000000000000000"
)

type Service struct {
	address  common.Address
	contract *router.Router
	Client   client.Ethereum
	mutex    sync.Mutex
	members  Members
	logger   *log.Entry
}

func (bsc *Service) WatchLockEventLogs(opts *bind.WatchOpts, sink chan<- *router.RouterLock) (event.Subscription, error) {
	return bsc.contract.WatchLock(opts, sink)
}

// Address returns the address of the contract instance
func (bsc *Service) Address() common.Address {
	return bsc.address
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
func (bsc *Service) WatchBurnEventLogs(opts *bind.WatchOpts, sink chan<- *router.RouterBurn) (event.Subscription, error) {
	return bsc.contract.WatchBurn(opts, sink)
}

func (bsc *Service) updateMembers() {
	membersCount, err := bsc.contract.NativeTokensCount(nil)
	if err != nil {
		bsc.logger.Fatal("Failed to get members count", err)
	}

	var membersArray []string
	for i := 0; i < int(membersCount.Int64()); i++ {
		addr, err := bsc.contract.NativeTokenAt(nil, big.NewInt(int64(i)))
		if err != nil {
			bsc.logger.Fatal("Failed to get member address", err)
		}
		membersArray = append(membersArray, addr.String())
	}
	bsc.members.Set(membersArray)
	bsc.logger.Infof("Set members list to %s", membersArray)

}

func (bsc *Service) listenForMemberUpdatedEvent() {
	events := make(chan *router.RouterNativeTokenUpdated)
	sub, err := bsc.contract.WatchNativeTokenUpdated(nil, events)
	if err != nil {
		bsc.logger.Fatal("Failed to subscribe for WatchMemberUpdated Event Logs for contract. Error ", err)
	}

	for {
		select {
		case err := <-sub.Err():
			bsc.logger.Errorf("MemberUpdated Event Logs subscription failed. Error [%s].", err)
			go bsc.listenForMemberUpdatedEvent()
			return
		case <-events:
			bsc.updateMembers()
		}
	}
}

// NewService creates new instance of a Contract Services based on the provided configuration
func NewService(client client.Ethereum, c config.Ethereum) *Service {
	contractAddress, err := client.ValidateContractDeployedAt(c.RouterContractAddress)
	if err != nil {
		log.Fatal(err)
	}

	contractInstance, err := router.NewRouter(*contractAddress, client.GetClient())
	if err != nil {
		log.Fatalf("Failed to initialize Router Contract Instance at [%s]. Error [%s]", c.RouterContractAddress, err)
	}

	contractService := &Service{
		address:  *contractAddress,
		Client:   client,
		contract: contractInstance,
		logger:   config.GetLoggerFor("Contract Service"),
	}

	contractService.updateMembers()

	go contractService.listenForMemberUpdatedEvent()

	return contractService
}

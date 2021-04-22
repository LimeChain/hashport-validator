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
	"errors"
	"github.com/ethereum/go-ethereum/event"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	routerAbi "github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum/contracts/router"
	log "github.com/sirupsen/logrus"
)

const (
	nilErc20Address = "0x0000000000000000000000000000000000000000"
)

type Service struct {
	address  common.Address
	contract *routerAbi.Router
	Client   client.Ethereum
	mutex    sync.Mutex
	members  Members
	logger   *log.Entry
}

func (bsc *Service) ToWrapped(nativeAsset string) (string, error) {
	wrappedAsset, err := bsc.contract.NativeToWrapped(
		nil,
		common.RightPadBytes([]byte(nativeAsset), 32),
	)
	if err != nil {
		return "", err
	}

	erc20address := wrappedAsset.String()
	if erc20address == nilErc20Address {
		return "", errors.New("token-not-supported")
	}

	return erc20address, nil
}

func (bsc *Service) ToNative(wrappedAsset common.Address) (string, error) {
	native, err := bsc.contract.WrappedToNative(nil, wrappedAsset)
	if err != nil {
		return "", err
	}

	if len(native) == 0 {
		return "", errors.New("native token not found")
	}

	return string(common.TrimRightZeroes(native)), nil
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
func (bsc *Service) WatchBurnEventLogs(opts *bind.WatchOpts, sink chan<- *routerAbi.RouterBurn) (event.Subscription, error) {
	return bsc.contract.WatchBurn(opts, sink, nil, nil)
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

func (bsc *Service) listenForMemberUpdatedEvent() {
	events := make(chan *routerAbi.RouterMemberUpdated)
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
	contractAddress, err := client.ValidateContractDeployedAt(c.RouterContractAddress)
	if err != nil {
		log.Fatal(err)
	}

	contractInstance, err := routerAbi.NewRouter(*contractAddress, client.GetClient())
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

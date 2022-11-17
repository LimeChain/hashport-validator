/*
 * Copyright 2022 LimeChain Ltd.
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
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
	"math/big"
	"strings"
	"sync"
	"time"
)

type Service struct {
	address  common.Address
	contract client.DiamondRouter
	Client   client.EVM
	mutex    sync.Mutex
	members  Members
	logger   *log.Entry
}

func (bsc *Service) GetClient() client.Core {
	return bsc.Client.GetClient()
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
		if strings.EqualFold(k, address) {
			return true
		}
	}
	return false
}

// HasValidSignaturesLength returns whether the signatures are enough for submission
func (bsc *Service) HasValidSignaturesLength(signaturesLength *big.Int) (bool, error) {
	return bsc.contract.HasValidSignaturesLength(nil, signaturesLength)
}

// ParseMintLog parses a general typed log to a RouterMint event
func (bsc *Service) ParseMintLog(log types.Log) (*router.RouterMint, error) {
	return bsc.contract.ParseMint(log)
}

// ParseBurnLog parses a general typed log to a RouterBurn event
func (bsc *Service) ParseBurnLog(log types.Log) (*router.RouterBurn, error) {
	return bsc.contract.ParseBurn(log)
}

// ParseLockLog parses a general typed log to a RouterLock event
func (bsc *Service) ParseLockLog(log types.Log) (*router.RouterLock, error) {
	return bsc.contract.ParseLock(log)
}

// ParseUnlockLog parses a general typed log to a RouterUnlock event
func (bsc *Service) ParseUnlockLog(log types.Log) (*router.RouterUnlock, error) {
	return bsc.contract.ParseUnlock(log)
}

// ParseBurnERC721Log parses a general typed log to a BurnERC721event
func (bsc *Service) ParseBurnERC721Log(log types.Log) (*router.RouterBurnERC721, error) {
	return bsc.contract.ParseBurnERC721(log)
}

// WatchBurnEventLogs creates a subscription for Burn Events emitted in the Bridge contract
func (bsc *Service) WatchBurnEventLogs(opts *bind.WatchOpts, sink chan<- *router.RouterBurn) (event.Subscription, error) {
	return bsc.contract.WatchBurn(opts, sink)
}

func (bsc *Service) ReloadMembers() {
	members, err := bsc.getMembers()
	if err != nil {
		time.Sleep(10 * time.Second)
		go bsc.ReloadMembers()
		return
	}

	bsc.members.Set(members)
	bsc.logger.Infof("Set members list to [%s].", members)
}

func (bsc *Service) getMembers() ([]string, error) {
	membersCount, err := bsc.contract.MembersCount(nil)
	if err != nil {
		bsc.logger.Errorf("Failed to get members count. Error: [%s].", err)
		return nil, err
	}

	var membersArray []string
	for i := 0; i < int(membersCount.Int64()); i++ {
		addr, err := bsc.contract.MemberAt(nil, big.NewInt(int64(i)))
		if err != nil {
			bsc.logger.Errorf("Failed to get member address [%d]. Error: [%s].", i, err)
			return nil, err
		}
		membersArray = append(membersArray, addr.String())
	}

	return membersArray, nil
}

func (bsc *Service) FeeAmountFor(_targetChain *big.Int, _userAddress common.Address, _tokenAddress common.Address, _amount *big.Int) (*big.Int, error) {
	return bsc.contract.FeeAmountFor(nil, _targetChain, _userAddress, _tokenAddress, _amount)
}

func (bsc *Service) TokenFeeData(_token common.Address) (struct {
	ServiceFeePercentage *big.Int
	FeesAccrued          *big.Int
	PreviousAccrued      *big.Int
	Accumulator          *big.Int
}, error) {
	return bsc.contract.TokenFeeData(nil, _token)
}

// NewService creates new instance of a Contract Services based on the provided configuration
func NewService(client client.EVM, address string, contractInstance client.DiamondRouter) *Service {
	contractAddress, err := client.ValidateContractDeployedAt(address)
	if err != nil {
		log.Fatal(err)
	}

	contractService := &Service{
		address:  *contractAddress,
		Client:   client,
		contract: contractInstance,
		logger:   config.GetLoggerFor(fmt.Sprintf("Contract Service [%s]", contractAddress.String())),
	}

	contractService.ReloadMembers()

	return contractService
}

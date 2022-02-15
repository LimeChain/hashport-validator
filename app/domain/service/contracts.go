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

package service

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	abi "github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"math/big"
)

// Contracts interface is implemented by the Contracts Service providing business logic access to the EVM SmartContracts and other related utility functions
type Contracts interface {
	// Address returns the address of the contract instance
	Address() common.Address
	// GetMembers returns the array of bridge members currently set in the Bridge contract
	GetMembers() []string
	// ReloadMembers triggers to fetch all the members from the Router Contract
	ReloadMembers()
	// GetClient returns the Contracts Service corresponding EVM Client
	GetClient() client.Core
	// IsMember returns true/false depending on whether the provided address is a Bridge member or not
	IsMember(address string) bool
	// HasValidSignaturesLength returns whether the signatures are enough for submission
	HasValidSignaturesLength(*big.Int) (bool, error)
	// ParseMintLog parses a general typed log to a RouterMint event
	ParseMintLog(log types.Log) (*abi.RouterMint, error)
	// ParseBurnLog parses a general typed log to a RouterBurn event
	ParseBurnLog(log types.Log) (*abi.RouterBurn, error)
	// ParseLockLog parses a general typed log to a RouterLock event
	ParseLockLog(log types.Log) (*abi.RouterLock, error)
	// ParseUnlockLog parses a general typed log to a RouterUnlock event
	ParseUnlockLog(log types.Log) (*abi.RouterUnlock, error)
	// ParseBurnERC721Log parses a general typed log to a BurnERC721event
	ParseBurnERC721Log(log types.Log) (*abi.RouterBurnERC721, error)
	// WatchBurnEventLogs creates a subscription for Burn Events emitted in the Bridge contract
	WatchBurnEventLogs(opts *bind.WatchOpts, sink chan<- *abi.RouterBurn) (event.Subscription, error)
	// WatchLockEventLogs creates a subscription for Lock Events emitted in the Bridge contract
	WatchLockEventLogs(opts *bind.WatchOpts, sink chan<- *abi.RouterLock) (event.Subscription, error)
	// AddDecimals adjusts the decimals in the native and wrapped tokens when their decimals do not match and one of them is over 8
	AddDecimals(amount *big.Int, asset string) (*big.Int, error)
	// RemoveDecimals adjusts the decimals in the native and wrapped tokens when their decimals do not match and one of them is over 8
	RemoveDecimals(amount *big.Int, asset string) (*big.Int, error)
}

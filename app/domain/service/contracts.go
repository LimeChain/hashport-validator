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
	"github.com/ethereum/go-ethereum/event"
	abi "github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum/contracts/router"
)

// Contracts interface is implemented by the Contracts Service providing business logic access to the Ethereum SmartContracts and other related utility functions
type Contracts interface {
	// Address returns the address of the contract instance
	Address() common.Address
	// GetMembers returns the array of bridge members currently set in the Bridge contract
	GetMembers() []string
	// IsMember returns true/false depending on whether the provided address is a Bridge member or not
	IsMember(address string) bool
	// WatchBurnEventLogs creates a subscription for Burn Events emitted in the Bridge contract
	WatchBurnEventLogs(opts *bind.WatchOpts, sink chan<- *abi.RouterBurn) (event.Subscription, error)
	// Check whether a specific asset has a valid bridge token address. Returns the erc20 token address if native asset is valid. Returns an empty string if not.
	ToWrapped(native string) (string, error)
	// Checks whether a specific wrapped token has a corresponding native token. Returns the native token as string
	ToNative(wrapped common.Address) (string, error)
}

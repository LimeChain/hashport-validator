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

package client

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
)

type DiamondRouter interface {
	WatchLock(opts *bind.WatchOpts, sink chan<- *router.RouterLock) (event.Subscription, error)
	HasValidSignaturesLength(opts *bind.CallOpts, _n *big.Int) (bool, error)
	ParseMint(log types.Log) (*router.RouterMint, error)
	ParseBurn(log types.Log) (*router.RouterBurn, error)
	ParseLock(log types.Log) (*router.RouterLock, error)
	ParseUnlock(log types.Log) (*router.RouterUnlock, error)
	ParseBurnERC721(log types.Log) (*router.RouterBurnERC721, error)
	WatchBurn(opts *bind.WatchOpts, sink chan<- *router.RouterBurn) (event.Subscription, error)
	MembersCount(opts *bind.CallOpts) (*big.Int, error)
	MemberAt(opts *bind.CallOpts, _index *big.Int) (common.Address, error)
	TokenFeeData(opts *bind.CallOpts, _token common.Address) (struct {
		ServiceFeePercentage *big.Int
		FeesAccrued          *big.Int
		PreviousAccrued      *big.Int
		Accumulator          *big.Int
	}, error)
	Erc721Fee(opts *bind.CallOpts, _erc721 common.Address) (*big.Int, error)
	Erc721Payment(opts *bind.CallOpts, _erc721 common.Address) (common.Address, error)
}

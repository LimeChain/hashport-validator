package client

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"math/big"
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
}

package service

import (
	mirror_node "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	"github.com/limechain/hedera-eth-bridge-validator/app/encoding"
	"github.com/limechain/hedera-eth-bridge-validator/app/encoding/memo"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transfer"
	"github.com/stretchr/testify/mock"
)

type MockTransferService struct {
	mock.Mock
}

func (mts *MockTransferService) ProcessTransfer(tm encoding.TransferMessage) error {
	args := mts.Called(tm)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (mts *MockTransferService) SanityCheckTransfer(tx mirror_node.Transaction) (*memo.Memo, error) {
	args := mts.Called(tx)
	if args.Get(0) == nil {
		return nil, args.Get(1).(error)
	}
	if args.Get(1) == nil {
		return args.Get(0).(*memo.Memo), nil
	}
	return args.Get(0).(*memo.Memo), args.Get(1).(error)
}

func (mts *MockTransferService) SaveRecoveredTxn(txId, amount, sourceAsset, targetAsset string, m memo.Memo) error {
	args := mts.Called(txId, amount, sourceAsset, targetAsset, m)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (mts *MockTransferService) InitiateNewTransfer(tm encoding.TransferMessage) (*transfer.Transfer, error) {
	args := mts.Called(tm)
	if args.Get(0) == nil {
		return nil, args.Get(1).(error)
	}
	if args.Get(1) == nil {
		return args.Get(0).(*transfer.Transfer), nil
	}
	return args.Get(0).(*transfer.Transfer), args.Get(1).(error)
}

func (mts *MockTransferService) VerifyFee(tm encoding.TransferMessage) error {
	args := mts.Called(tm)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

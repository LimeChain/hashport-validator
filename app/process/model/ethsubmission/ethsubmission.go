package ethsubmission

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
)

type Submission struct {
	TransactOps           *bind.TransactOpts
	CryptoTransferMessage *proto.CryptoTransferMessage
	Messages              []message.TransactionMessage
	Slot                  int64
}

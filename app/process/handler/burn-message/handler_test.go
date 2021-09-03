package burn_message

import (
	"errors"
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks/service"
	"testing"
)

var (
	mt = model.Transfer{
		TransactionId: "0.0.0-0000000-1234",
		Receiver:      "0x12345",
		Amount:        "10000000000",
		NativeAsset:   constants.Hbar,
		TargetAsset:   "0x45678",
	}
)

func InitializeHandler() (*Handler, *service.MockTransferService) {
	mocks.Setup()

	return NewHandler(mocks.MTransferService), mocks.MTransferService
}

func Test_Handle_ProcessWrappedTransfer_Fails(t *testing.T) {
	ctHandler, mockedService := InitializeHandler()

	tx := &entity.Transfer{
		TransactionID: mt.TransactionId,
		Receiver:      mt.Receiver,
		Amount:        mt.Amount,
		Status:        transfer.StatusInitial,
	}

	mockedService.On("InitiateNewTransfer", mt).Return(tx, nil)
	mockedService.On("ProcessWrappedTransfer", mt).Return(errors.New("some-error"))

	ctHandler.Handle(&mt)
}

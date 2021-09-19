package transfer

import (
	"errors"
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	h  *Handler
	tr = &model.Transfer{
		TransactionId: "some-tx-id",
		SourceChainId: 0,
		TargetChainId: 1,
		NativeChainId: 0,
		SourceAsset:   constants.Hbar,
		TargetAsset:   "0xsomeethaddress",
		NativeAsset:   constants.Hbar,
		Receiver:      "0xsomeotherethaddress",
		Amount:        "100",
		RouterAddress: "0xrouteraddress",
		HasFee:        true,
		Timestamp:     "1",
	}
)

func Test_NewHandler(t *testing.T) {
	setup()
	assert.Equal(t, h, NewHandler(mocks.MTransferService))
}

func Test_Handle(t *testing.T) {
	setup()
	mocks.MTransferService.On("InitiateNewTransfer", *tr).Return(&entity.Transfer{Status: transfer.StatusInitial}, nil)
	h.Handle(tr)
}

func Test_Handle_NotInitialFails(t *testing.T) {
	setup()
	mocks.MTransferService.On("InitiateNewTransfer", *tr).Return(&entity.Transfer{Status: "not-initial"}, nil)
	h.Handle(tr)
}

func Test_Handle_InvalidPayload(t *testing.T) {
	setup()
	h.Handle("invalid-payload")
	mocks.MTransferService.AssertNotCalled(t, "InitiateNewTransfer", *tr)
}

func Test_Handle_InitiateNewTransferFails(t *testing.T) {
	setup()
	mocks.MTransferService.On("InitiateNewTransfer", *tr).Return(nil, errors.New("some-error"))
	h.Handle(tr)
}

func setup() {
	mocks.Setup()
	h = &Handler{
		transfersService: mocks.MTransferService,
		logger:           config.GetLoggerFor("Hedera Mint and Transfer Handler"),
	}
}

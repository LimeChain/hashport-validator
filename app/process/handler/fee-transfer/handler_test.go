package fee_transfer

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	feeTransferHandler *Handler
)

func Test_NewHandler(t *testing.T) {
	mocks.Setup()
	h := NewHandler(mocks.MBurnService)
	assert.Equal(t, &Handler{
		burnService: mocks.MBurnService,
		logger:      config.GetLoggerFor("Hedera Fee and Schedule Transfer Handler"),
	}, h)
}

func Test_Handle_Burn(t *testing.T) {
	setup()
	someEvent := &transfer.Transfer{
		TransactionId: "",
		SourceChainId: 0,
		TargetChainId: 0,
		NativeChainId: 0,
		SourceAsset:   "",
		TargetAsset:   "",
		NativeAsset:   "",
		Receiver:      "",
		Amount:        "0",
		RouterAddress: "",
	}
	mocks.MBurnService.On("ProcessEvent", *someEvent).Return()
	feeTransferHandler.Handle(someEvent)
	mocks.MBurnService.AssertCalled(t, "ProcessEvent", *someEvent)
}

func Test_Handle_Encoding_Fails(t *testing.T) {
	setup()

	invalidTransferPayload := []byte{1, 2, 1}

	feeTransferHandler.Handle(invalidTransferPayload)

	mocks.MBurnService.AssertNotCalled(t, "ProcessEvent")
}

func setup() {
	mocks.Setup()
	feeTransferHandler = &Handler{
		burnService: mocks.MBurnService,
		logger:      config.GetLoggerFor("Hedera Fee and Schedule Transfer Handler"),
	}
}

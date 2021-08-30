package mint_hts

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	mintHtsHandler *Handler
)

func Test_NewHandler(t *testing.T) {
	mocks.Setup()
	h := NewHandler(mocks.MLockService)
	assert.Equal(t, &Handler{
		lockService: mocks.MLockService,
		logger:      config.GetLoggerFor("Hedera Mint and Transfer Handler"),
	}, h)
}

func Test_Handle_Lock(t *testing.T) {
	setup()
	tr := &transfer.Transfer{
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
	mocks.MLockService.On("ProcessEvent", *tr).Return()
	mintHtsHandler.Handle(tr)
	mocks.MLockService.AssertCalled(t, "ProcessEvent", *tr)
}

func Test_Handle_Encoding_Fails(t *testing.T) {
	setup()

	invalidTransferPayload := []byte{1, 2, 1}

	mintHtsHandler.Handle(invalidTransferPayload)

	mocks.MLockService.AssertNotCalled(t, "ProcessEvent")
}

func setup() {
	mocks.Setup()
	mintHtsHandler = &Handler{
		lockService: mocks.MLockService,
		logger:      config.GetLoggerFor("Hedera Mint and Transfer Handler"),
	}
}

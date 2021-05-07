package scheduled

import (
	"fmt"
	"github.com/hashgraph/hedera-sdk-go/v2"
	hedera2 "github.com/limechain/hedera-eth-bridge-validator/app/helper/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"testing"
	"time"
)

var (
	s = &Service{}
)

func setup() {
	mocks.Setup()
	s = &Service{
		payerAccount:     hedera.AccountID{},
		hederaNodeClient: mocks.MHederaNodeClient,
		mirrorNodeClient: mocks.MHederaMirrorClient,
		logger:           config.GetLoggerFor("Burn Event Service"),
	}
}

func Test_ExecuteHBARTransfer(t *testing.T) {
	setup()

	mockTransfers := []transfer.Hedera{}
	now := time.Now()
	TransactionID := hedera.TransactionID{
		AccountID: &hedera.AccountID{
			Shard:   0,
			Realm:   0,
			Account: 111111,
		},
		ValidStart: &now,
	}

	mockScheduleId, err := hedera.ScheduleIDFromString("0.0.10830")
	if err != nil {
		t.Fatal(err)
	}
	mockScheduledTransactionId, err := hedera.TransactionIdFromString("0.0.123213@123982.012342")
	if err != nil {
		t.Fatal(err)
	}

	mockTxReceipt := hedera.TransactionReceipt{
		Status:                 hedera.StatusSuccess,
		ScheduleID:             &mockScheduleId,
		ScheduledTransactionID: &mockScheduledTransactionId,
	}

	mockHederaClient := hedera.ClientForTestnet()
	mocks.MHederaTransactionResponse.On("GetReceipt", mockHederaClient).Return(mockTxReceipt, nil)
	mocks.MHederaNodeClient.On(
		"SubmitScheduledHbarTransferTransaction",
		mockTransfers, s.payerAccount, mockScheduledTransactionId.String()).
		Return(mocks.MHederaTransactionResponse, nil)
	mocks.MHederaNodeClient.On("GetClient").Return(mockHederaClient)
	mocks.MHederaTransactionResponse.On("GetTransactionID").Return(TransactionID)

	onSuccess := func(string, string) {
		fmt.Println("Stuff worked well done.")
	}
	onEverythingElse := func(string) {

	}

	// TODO: Find a way to assert/check all callback functions
	mocks.MHederaMirrorClient.On("WaitForScheduledTransferTransaction", hedera2.FromHederaTransactionID(&mockScheduledTransactionId).String())
	s.Execute(
		mockScheduledTransactionId.String(),
		constants.Hbar,
		[]transfer.Hedera{},
		onSuccess,
		onEverythingElse,
		onEverythingElse,
		onEverythingElse)
}

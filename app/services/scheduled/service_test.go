package scheduled

import (
	"errors"
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
	transactionId := hedera.TransactionID{
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
	mocks.MHederaTransactionResponse.On("GetTransactionID").Return(transactionId)

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

func Test_ExecuteFailOnSubmitScheduled(t *testing.T) {
	setup()

	now := time.Now()
	mockTransactionId := hedera.TransactionID{
		AccountID: &hedera.AccountID{
			Shard:   0,
			Realm:   0,
			Account: 111111,
		},
		ValidStart: &now,
	}
	mockTransfers := []transfer.Hedera{}

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
		Return(mocks.MHederaTransactionResponse, errors.New("submission-fail"))
	mocks.MHederaNodeClient.AssertNotCalled(t, "GetClient")
	mocks.MHederaTransactionResponse.AssertNotCalled(t, "GetTransactionID")
	mocks.MHederaTransactionResponse.On("GetTransactionID").Return(mockTransactionId)

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

func Test_ExecuteFailsToGetReceipt(t *testing.T) {
	setup()

	now := time.Now()
	mockTransactionId := hedera.TransactionID{
		AccountID: &hedera.AccountID{
			Shard:   0,
			Realm:   0,
			Account: 111111,
		},
		ValidStart: &now,
	}

	mockScheduledTransactionId, err := hedera.TransactionIdFromString("0.0.123213@123982.012342")
	if err != nil {
		t.Fatal(err)
	}

	mockHederaClient := hedera.ClientForTestnet()
	mocks.MHederaTransactionResponse.On("GetReceipt", mockHederaClient).Return(nil, errors.New("get-receipt-fail"))
	mocks.MHederaNodeClient.AssertNotCalled(t,
		"SubmitScheduledHbarTransferTransaction")
	mocks.MHederaNodeClient.AssertNotCalled(t, "GetClient")
	mocks.MHederaTransactionResponse.AssertNotCalled(t, "GetTransactionID")
	mocks.MHederaTransactionResponse.On("GetTransactionID").Return(mockTransactionId)

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

func Test_ExecuteTokenTransfer(t *testing.T) {
	setup()

	tokenId, err := hedera.TokenIDFromString("0.0.999999")
	if err != nil {
		t.Fatal(err)
	}

	mockTransfers := []transfer.Hedera{}
	now := time.Now()
	transactionId := hedera.TransactionID{
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
		"SubmitScheduledTokenTransferTransaction",
		tokenId, mockTransfers, s.payerAccount, mockScheduledTransactionId.String()).
		Return(mocks.MHederaTransactionResponse, nil)
	mocks.MHederaNodeClient.On("GetClient").Return(mockHederaClient)
	mocks.MHederaTransactionResponse.On("GetTransactionID").Return(transactionId)

	onSuccess := func(string, string) {
		fmt.Println("Stuff worked well done.")
	}
	onEverythingElse := func(string) {

	}

	// TODO: Find a way to assert/check all callback functions
	mocks.MHederaMirrorClient.On("WaitForScheduledTransferTransaction", hedera2.FromHederaTransactionID(&mockScheduledTransactionId).String())
	s.Execute(
		mockScheduledTransactionId.String(),
		"0.0.999999",
		[]transfer.Hedera{},
		onSuccess,
		onEverythingElse,
		onEverythingElse,
		onEverythingElse)
}

func Test_ExecuteWithInvalidTokenID(t *testing.T) {
	setup()
	mockScheduledTransactionId, err := hedera.TransactionIdFromString("0.0.123213@123982.012342")
	if err != nil {
		t.Fatal(err)
	}

	onSuccess := func(string, string) {
		fmt.Println("Stuff worked well done.")
	}
	onEverythingElse := func(string) {

	}

	s.Execute(
		mockScheduledTransactionId.String(),
		"invalid-token-id",
		[]transfer.Hedera{},
		onSuccess,
		onEverythingElse,
		onEverythingElse,
		onEverythingElse)

	mocks.MHederaTransactionResponse.AssertNotCalled(t, "GetReceipt")
	mocks.MHederaNodeClient.AssertNotCalled(t, "SubmitScheduledTokenTransferTransaction")
	mocks.MHederaNodeClient.AssertNotCalled(t, "GetClient")
	mocks.MHederaTransactionResponse.AssertNotCalled(t, "GetTransactionID")
}

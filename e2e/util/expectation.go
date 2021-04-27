package util

import (
	"github.com/hashgraph/hedera-sdk-go/v2"
	routerContract "github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum/contracts/router"
	hederahelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	burn_event "github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/burn-event"
	"github.com/limechain/hedera-eth-bridge-validator/e2e/service/database"
	"github.com/limechain/hedera-eth-bridge-validator/e2e/setup"
	"testing"
)

func PrepareExpectedBurnEventRecord(scheduleID string, amount int64, recipient hedera.AccountID, burnEventId string) *entity.BurnEvent {
	return &entity.BurnEvent{
		Id:         burnEventId,
		ScheduleID: scheduleID,
		Amount:     amount,
		Recipient:  recipient.String(),
		Status:     burn_event.StatusCompleted,
	}
}

func PrepareExpectedTransfer(routerContract *routerContract.Router, transactionID hedera.TransactionID, nativeAsset, amount, receiver string, statuses database.ExpectedStatuses, t *testing.T) *entity.Transfer {
	expectedTxId := hederahelper.FromHederaTransactionID(&transactionID)

	wrappedAsset, err := setup.WrappedAsset(routerContract, nativeAsset)
	if err != nil {
		t.Fatalf("Expecting Token [%s] is not supported. - Error: [%s]", nativeAsset, err)
	}
	return &entity.Transfer{
		TransactionID:      expectedTxId.String(),
		Receiver:           receiver,
		NativeAsset:        nativeAsset,
		WrappedAsset:       wrappedAsset.String(),
		Amount:             amount,
		Status:             statuses.Status,
		SignatureMsgStatus: statuses.StatusSignature,
	}
}

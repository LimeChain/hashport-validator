package expected

import (
	"database/sql"
	"strconv"

	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
)

func FeeRecord(transactionID, scheduleID string, amount int64, transferID string) *entity.Fee {
	fee := &entity.Fee{
		TransactionID: transactionID,
		ScheduleID:    scheduleID,
		Amount:        strconv.FormatInt(amount, 10),
		Status:        status.Completed,
	}

	if transferID != "" {
		fee.TransferID = sql.NullString{
			String: transferID,
			Valid:  true,
		}
	}

	return fee
}

func FungibleTransferRecord(
	sourceChainId,
	targetChainId,
	nativeChainId uint64,
	transactionID,
	sourceAsset,
	targetAsset,
	nativeAsset,
	amount,
	fee,
	receiver string,
	status string,
	originator string,
	timestamp entity.NanoTime) *entity.Transfer {

	return &entity.Transfer{
		TransactionID: transactionID,
		SourceChainID: sourceChainId,
		TargetChainID: targetChainId,
		NativeChainID: nativeChainId,
		Receiver:      receiver,
		SourceAsset:   sourceAsset,
		TargetAsset:   targetAsset,
		NativeAsset:   nativeAsset,
		Amount:        amount,
		Fee:		   fee,
		Status:        status,
		Originator:    originator,
		Timestamp:     timestamp,
	}
}

func ScheduleRecord(txId, scheduleId, operation string, hasReceiver bool, status string, transferId sql.NullString) *entity.Schedule {
	return &entity.Schedule{
		TransactionID: txId,
		ScheduleID:    scheduleId,
		HasReceiver:   hasReceiver,
		Operation:     operation,
		Status:        status,
		TransferID:    transferId,
	}
}

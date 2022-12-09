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
		Status:        status,
		Originator:    originator,
		Timestamp:     timestamp,
	}
}

func NonFungibleTransferRecord(
	sourceChainId,
	targetChainId,
	nativeChainId uint64,
	transactionID,
	sourceAsset,
	targetAsset,
	nativeAsset,
	receiver string,
	status string,
	fee string,
	serialNumber int64,
	metadata string,
	originator string,
	timestamp entity.NanoTime) *entity.Transfer {
	return &entity.Transfer{
		TransactionID: transactionID,
		SourceChainID: sourceChainId,
		TargetChainID: targetChainId,
		NativeChainID: nativeChainId,
		SourceAsset:   sourceAsset,
		TargetAsset:   targetAsset,
		NativeAsset:   nativeAsset,
		Receiver:      receiver,
		Fee:           fee,
		Status:        status,
		SerialNumber:  serialNumber,
		Metadata:      metadata,
		IsNft:         true,
		Timestamp:     timestamp,
		Originator:    originator,
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

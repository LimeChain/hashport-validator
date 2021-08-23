/*
 * Copyright 2021 LimeChain Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package util

import (
	"database/sql"
	"github.com/hashgraph/hedera-sdk-go/v2"
	hederahelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	burn_event "github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/burn-event"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/fee"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/e2e/service/database"
	"github.com/limechain/hedera-eth-bridge-validator/e2e/setup"
	"strconv"
	"testing"
)

func PrepareExpectedBurnEventRecord(scheduleID string, amount int64, recipient hedera.AccountID, burnEventId string, transactionID string) *entity.BurnEvent {
	return &entity.BurnEvent{
		Id:         burnEventId,
		ScheduleID: scheduleID,
		Amount:     amount,
		Recipient:  recipient.String(),
		Status:     burn_event.StatusCompleted,
		TransactionId: sql.NullString{
			String: transactionID,
			Valid:  true,
		},
	}
}

func PrepareExpectedLockEventRecord(amount int64, recipient hedera.AccountID, burnEventId, transferTransactionID, mintTransactionID, scheduleTransferID, scheduleMintID, nativeAsset, wrappedAsset string, sourceChainID, targetChainID int64, status string) *entity.LockEvent {
	return &entity.LockEvent{
		Id:                 burnEventId,
		ScheduleTransferID: scheduleTransferID,
		ScheduleMintID:     scheduleMintID,
		NativeAsset:        nativeAsset,
		WrappedAsset:       wrappedAsset,
		Amount:             amount,
		SourceChainID:      sourceChainID,
		TargetChainID:      targetChainID,
		Recipient:          recipient.String(),
		Status:             status,
		ScheduleTransferTxId: sql.NullString{
			String: transferTransactionID,
			Valid:  true,
		},
		ScheduleMintTxId: sql.NullString{
			String: mintTransactionID,
			Valid:  true,
		},
	}
}

func PrepareExpectedFeeRecord(transactionID, scheduleID string, amount int64, transferID, burnEventID string) *entity.Fee {
	fee := &entity.Fee{
		TransactionID: transactionID,
		ScheduleID:    scheduleID,
		Amount:        strconv.FormatInt(amount, 10),
		Status:        fee.StatusCompleted,
	}

	if transferID != "" {
		fee.TransferID = sql.NullString{
			String: transferID,
			Valid:  true,
		}
	}

	if burnEventID != "" {
		fee.BurnEventID = sql.NullString{
			String: burnEventID,
			Valid:  true,
		}
	}

	return fee
}

func PrepareExpectedTransfer(assetMappings config.AssetMappings, sourceChainId, targetChainId int64, transactionID hedera.TransactionID, routerAddress, nativeAsset, amount, receiver string, statuses database.ExpectedStatuses, t *testing.T) *entity.Transfer {
	expectedTxId := hederahelper.FromHederaTransactionID(&transactionID)

	wrappedAsset, err := setup.NativeToWrappedAsset(assetMappings, sourceChainId, targetChainId, nativeAsset)
	if err != nil {
		t.Fatalf("Expecting Token [%s] is not supported. - Error: [%s]", nativeAsset, err)
	}
	return &entity.Transfer{
		TransactionID:      expectedTxId.String(),
		Receiver:           receiver,
		NativeAsset:        nativeAsset,
		RouterAddress:      routerAddress,
		WrappedAsset:       wrappedAsset.String(),
		Amount:             amount,
		Status:             statuses.Status,
		SignatureMsgStatus: statuses.StatusSignature,
	}
}

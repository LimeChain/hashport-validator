/*
 * Copyright 2022 LimeChain Ltd.
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
	"strconv"

	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
)

func PrepareExpectedFeeRecord(transactionID, scheduleID string, amount int64, transferID string) *entity.Fee {
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

func PrepareExpectedTransfer(
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

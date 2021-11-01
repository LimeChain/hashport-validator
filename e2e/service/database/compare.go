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

package database

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
)

func transfersFieldsMatch(comparing, comparable entity.Transfer) bool {
	return comparable.TransactionID == comparing.TransactionID &&
		comparable.SourceChainID == comparing.SourceChainID &&
		comparable.TargetChainID == comparing.TargetChainID &&
		comparable.NativeChainID == comparing.NativeChainID &&
		comparable.SourceAsset == comparing.SourceAsset &&
		comparable.TargetAsset == comparing.TargetAsset &&
		comparable.NativeAsset == comparing.NativeAsset &&
		comparable.Receiver == comparing.Receiver &&
		comparable.Amount == comparing.Amount &&
		comparable.Status == comparing.Status
}

func messagesFieldsMatch(comparing, comparable entity.Message) bool {
	return comparable.TransferID == comparing.TransferID &&
		comparable.Signature == comparing.Signature &&
		comparable.Hash == comparing.Hash &&
		transfersFieldsMatch(comparable.Transfer, comparing.Transfer) &&
		comparable.Signer == comparing.Signer
}

func feeFieldsMatch(comparing, comparable entity.Fee) bool {
	return comparing.Status == comparable.Status &&
		comparing.TransactionID == comparable.TransactionID &&
		comparing.Amount == comparable.Amount &&
		comparing.ScheduleID == comparable.ScheduleID &&
		comparing.TransferID == comparable.TransferID
}

func scheduleFieldsMatch(comparing, comparable entity.Schedule) bool {
	return comparing.TransactionID == comparable.TransactionID &&
		comparing.ScheduleID == comparable.ScheduleID &&
		comparing.Operation == comparable.Operation &&
		comparing.TransferID.String == comparable.TransferID.String &&
		comparing.TransferID.Valid == comparable.TransferID.Valid &&
		comparing.HasReceiver == comparable.HasReceiver &&
		comparing.Status == comparable.Status
}

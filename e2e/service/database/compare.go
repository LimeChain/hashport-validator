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
		comparable.RouterAddress == comparing.RouterAddress &&
		comparable.Receiver == comparing.Receiver &&
		comparable.NativeAsset == comparing.NativeAsset &&
		comparable.WrappedAsset == comparing.WrappedAsset &&
		comparable.Amount == comparing.Amount &&
		comparable.Status == comparing.Status &&
		comparable.SignatureMsgStatus == comparing.SignatureMsgStatus
}

func messagesFieldsMatch(comparing, comparable entity.Message) bool {
	return comparable.TransferID == comparing.TransferID &&
		comparable.Signature == comparing.Signature &&
		comparable.Hash == comparing.Hash &&
		transfersFieldsMatch(comparable.Transfer, comparing.Transfer) &&
		comparable.Signer == comparing.Signer
}

func feeFieldsMatch(comparing, comparable *entity.Fee) bool {
	return comparing.Status == comparable.Status &&
		comparing.TransactionID == comparable.TransactionID &&
		comparing.Amount == comparable.Amount &&
		comparable.ScheduleID == comparable.ScheduleID &&
		comparing.TransferID == comparable.TransferID &&
		comparing.BurnEventID == comparable.BurnEventID
}

func burnEventsFieldsMatch(comparing, comparable *entity.BurnEvent) bool {
	return comparing.Status == comparable.Status &&
		comparing.ScheduleID == comparable.ScheduleID &&
		comparing.Recipient == comparable.Recipient &&
		comparing.Amount == comparable.Amount &&
		comparing.Id == comparable.Id &&
		comparing.TransactionId == comparable.TransactionId
}

func lockEventsFieldsMatch(comparing, comparable *entity.LockEvent) bool {
	return comparing.Recipient == comparable.Recipient &&
		comparing.Amount == comparable.Amount &&
		comparing.Id == comparable.Id
	// TODO: Come up with a way to track ALL statuses asynchronously
	//comparing.ScheduleMintID == comparable.ScheduleMintID &&
	//comparing.ScheduleMintTxId == comparable.ScheduleMintTxId &&
	//comparing.ScheduleTransferID == comparable.ScheduleTransferID &&
	//comparing.Status == comparable.Status &&
	//comparing.ScheduleTransferTxId == comparable.ScheduleTransferTxId
}

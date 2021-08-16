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

package entity

import "database/sql"

type LockEvent struct {
	Id                   string `gorm:"primaryKey"` // represents {ethTxHash}-{logIndex}
	ScheduleTransferID   string
	ScheduleMintID       string
	NativeAsset          string
	WrappedAsset         string
	Amount               int64
	SourceChainID        int64
	TargetChainID        int64
	Recipient            string
	Status               string
	ScheduleTransferTxId sql.NullString `gorm:"unique"` // id of the original scheduled transfer transaction
	ScheduleMintTxId     sql.NullString `gorm:"unique"` // id of the original scheduled token mint transaction
}

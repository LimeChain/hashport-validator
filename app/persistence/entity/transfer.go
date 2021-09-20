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

type Transfer struct {
	TransactionID string `gorm:"primaryKey"`
	SourceChainID int64
	TargetChainID int64
	NativeChainID int64
	SourceAsset   string
	TargetAsset   string
	NativeAsset   string
	Receiver      string
	Amount        string
	Status        string
	Messages      []Message  `gorm:"foreignKey:TransferID"`
	Fee           Fee        `gorm:"foreignKey:TransferID"`
	Schedules     []Schedule `gorm:"foreignKey:TransferID"`
}

// Message is a db model used to track the messages signed by validators for a given transfer
type Message struct {
	TransferID           string
	Transfer             Transfer `gorm:"foreignKey:TransferID;references:TransactionID;"`
	Hash                 string
	Signature            string `gorm:"unique"`
	Signer               string
	TransactionTimestamp int64
}

// Fee is a db model used only to mark native Hedera transfer fees to validators
type Fee struct {
	TransactionID string `gorm:"primaryKey"`
	ScheduleID    string `gorm:"unique"`
	Amount        string
	Status        string
	TransferID    sql.NullString
}

// Schedule is a db model used to track scheduled transactions for a given transfer
type Schedule struct {
	TransactionID string `gorm:"primaryKey"` // TransactionID  of the original scheduled transaction
	ScheduleID    string `gorm:"unique"`     // schedule ID
	Operation     string // type of scheduled transaction (TokenMint, TokenBurn, CryptoTransfer)
	Status        string
	TransferID    sql.NullString // foreign key to the transfer ID
}

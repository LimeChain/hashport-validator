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

package entity

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"time"

	transferModel "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
)

type Transfer struct {
	TransactionID string `gorm:"primaryKey" gorm:"index:,type:btree"`
	SourceChainID uint64
	TargetChainID uint64
	NativeChainID uint64
	SourceAsset   string
	TargetAsset   string
	NativeAsset   string
	Receiver      string
	Amount        string
	Fee           string
	Status        string
	SerialNumber  int64
	Metadata      string
	IsNft         bool     `gorm:"default:false"`
	Timestamp     NanoTime `sql:"type:bigint" gorm:"index:,sort:desc"`
	Originator    string
	Messages      []Message  `gorm:"foreignKey:TransferID"`
	Fees          []Fee      `gorm:"foreignKey:TransferID"`
	Schedules     []Schedule `gorm:"foreignKey:TransferID"`
}

func (t *Transfer) ToDto() *transferModel.Transfer {
	return &transferModel.Transfer{
		TransactionId: t.TransactionID,
		SourceChainId: t.SourceChainID,
		TargetChainId: t.TargetChainID,
		NativeChainId: t.NativeChainID,
		SourceAsset:   t.SourceAsset,
		TargetAsset:   t.TargetAsset,
		NativeAsset:   t.NativeAsset,
		Receiver:      t.Receiver,
		Amount:        t.Amount,
		SerialNum:     t.SerialNumber,
		Metadata:      t.Metadata,
		IsNft:         t.IsNft,
		Originator:    t.Originator,
		Timestamp:     t.Timestamp.Time,
	}
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
	ScheduleID    string // ScheduleID of the transaction. Can be empty if execution failed
	Amount        string
	Status        string
	TransferID    sql.NullString
}

// Schedule is a db model used to track scheduled transactions for a given transfer
type Schedule struct {
	TransactionID string `gorm:"primaryKey"` // TransactionID  of the original scheduled transaction
	ScheduleID    string // ScheduleID of the transaction. Can be empty if execution failed
	HasReceiver   bool   // True if the scheduled transaction includes the receiver of the TransferID in itself
	Operation     string // type of scheduled transaction (TokenMint, TokenBurn, CryptoTransfer)
	Status        string
	TransferID    sql.NullString // foreign key to the transfer ID
}

type NanoTime struct {
	time.Time
}

func (n NanoTime) Value() (driver.Value, error) {
	return n.UTC().UnixNano(), nil
}

func (n *NanoTime) Scan(value interface{}) error {
	if value == nil {
		*n = NanoTime{}
		return nil
	}
	t, ok := value.(int64)
	if !ok {
		return errors.New("value was not int64")
	}
	n.Time = time.Unix(0, t).UTC()
	return nil
}

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

package message

import (
	"gorm.io/gorm"
)

type TransactionMessage struct {
	gorm.Model
	TransactionId        string
	EthAddress           string
	Amount               string
	Fee                  string
	Signature            string
	Hash                 string
	SignerAddress        string
	TransactionTimestamp int64
}

type MessageRepository struct {
	dbClient *gorm.DB
}

func NewMessageRepository(dbClient *gorm.DB) *MessageRepository {
	return &MessageRepository{
		dbClient: dbClient,
	}
}

func (m MessageRepository) GetTransaction(txId, signature, hash string) (*TransactionMessage, error) {
	var message TransactionMessage
	err := m.dbClient.Model(&TransactionMessage{}).
		Where("transaction_id = ? and signature = ? and hash = ?", txId, signature, hash).
		First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (m MessageRepository) Create(message *TransactionMessage) error {
	return m.dbClient.Create(message).Error
}

func (m MessageRepository) GetTransactions(txId string, hash string) ([]TransactionMessage, error) {
	var messages []TransactionMessage
	err := m.dbClient.Where("transaction_id = ? and hash = ?", txId, hash).Order("transaction_timestamp").Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

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
	"errors"
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

type Repository struct {
	dbClient *gorm.DB
}

func NewRepository(dbClient *gorm.DB) *Repository {
	return &Repository{
		dbClient: dbClient,
	}
}

func (m Repository) GetMessageWith(txId, signature, hash string) (*TransactionMessage, error) {
	var message TransactionMessage
	err := m.dbClient.Model(&TransactionMessage{}).
		Where("transaction_id = ? and signature = ? and hash = ?", txId, signature, hash).
		First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (m Repository) Exist(txId, signature, hash string) (bool, error) {
	_, err := m.GetMessageWith(txId, signature, hash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func (m Repository) Create(message *TransactionMessage) error {
	return m.dbClient.Create(message).Error
}

func (m Repository) GetMessagesFor(txId string) ([]TransactionMessage, error) {
	var messages []TransactionMessage
	err := m.dbClient.Where("transaction_id = ?", txId).Order("transaction_timestamp").Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

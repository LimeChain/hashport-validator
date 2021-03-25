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
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"gorm.io/gorm"
)

type Repository struct {
	dbClient *gorm.DB
}

func NewRepository(dbClient *gorm.DB) *Repository {
	return &Repository{
		dbClient: dbClient,
	}
}

func (m Repository) GetMessageWith(txId, signature, hash string) (*entity.Message, error) {
	var message entity.Message
	err := m.dbClient.Model(&entity.Message{}).
		Where("transfer_id = ? and signature = ? and hash = ?", txId, signature, hash).
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

func (m Repository) Create(message *entity.Message) error {
	return m.dbClient.Create(message).Error
}

func (m Repository) Get(txId string) ([]entity.Message, error) {
	var messages []entity.Message
	err := m.dbClient.
		Preload("Transfer").
		Where("transfer_id = ?", txId).
		Order("transaction_timestamp").
		Find(&messages).
		Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

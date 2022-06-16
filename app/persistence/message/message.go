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

package message

import (
	"errors"

	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(dbClient *gorm.DB) *Repository {
	return &Repository{
		db: dbClient,
	}
}

func (r *Repository) GetMessageWith(transferID, signature, hash string) (*entity.Message, error) {
	var message entity.Message
	err := r.db.Model(&entity.Message{}).
		Where("transfer_id = ? and signature = ? and hash = ?", transferID, signature, hash).
		First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *Repository) Exist(transferID, signature, hash string) (bool, error) {
	_, err := r.GetMessageWith(transferID, signature, hash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func (r *Repository) Create(message *entity.Message) error {
	return r.db.Create(message).Error
}

func (r *Repository) Get(transferID string) ([]entity.Message, error) {
	var messages []entity.Message
	err := r.db.
		Preload("Transfer").
		Where("transfer_id = ?", transferID).
		Order("transaction_timestamp").
		Find(&messages).
		Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

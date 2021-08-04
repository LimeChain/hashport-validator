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

package lock_event

import (
	"errors"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	burn_event "github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/burn-event"
	"gorm.io/gorm"
)

// TODO: Put proper repository model with correct parameters here

type Repository struct {
	dbClient *gorm.DB
}

func NewRepository(dbClient *gorm.DB) *Repository {
	return &Repository{
		dbClient: dbClient,
	}
}

func (sr Repository) Create(id string, amount int64, recipient string) error {
	return sr.dbClient.Create(&entity.LockEvent{
		Id: id,
	}).Error
}

func (sr Repository) UpdateStatusSubmitted(ethTxHash, scheduleID, transactionId string) error {
	return sr.dbClient.
		Model(entity.LockEvent{}).
		Where("id = ?", ethTxHash).
		Updates(entity.LockEvent{}).
		Error
}

func (sr Repository) UpdateStatusCompleted(id string) error {
	return sr.updateStatus(id, burn_event.StatusCompleted)
}

func (sr Repository) UpdateStatusFailed(id string) error {
	return sr.updateStatus(id, burn_event.StatusFailed)
}

func (sr Repository) updateStatus(id string, status string) error {
	return sr.dbClient.
		Model(entity.LockEvent{}).
		Where("id = ?", id).
		UpdateColumn("status", status).
		Error
}

func (sr Repository) Get(id string) (*entity.LockEvent, error) {
	LockEvent := &entity.LockEvent{}
	result := sr.dbClient.
		Model(entity.LockEvent{}).
		Where("id = ?", id).
		First(LockEvent)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	return LockEvent, nil
}

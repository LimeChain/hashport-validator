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
	"database/sql"
	"errors"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	lock_event_status "github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/lock-event"
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

func (sr Repository) Create(id string, amount int64, recipient, nativeAsset, wrappedAsset string, sourceChainId, targetChainId int64) error {
	return sr.dbClient.Create(&entity.LockEvent{
		Id:            id,
		Amount:        amount,
		Recipient:     recipient,
		NativeAsset:   nativeAsset,
		WrappedAsset:  wrappedAsset,
		SourceChainID: sourceChainId,
		TargetChainID: targetChainId,
		Status:        lock_event_status.StatusInitial,
	}).Error
}

func (sr Repository) UpdateStatusScheduledTokenMintSubmitted(ethTxHash, scheduledTokenMintID, originalScheduledTokenMintID string) error {
	return sr.dbClient.
		Model(entity.LockEvent{}).
		Where("id = ?", ethTxHash).
		Updates(entity.LockEvent{Status: lock_event_status.StatusMintSubmitted, ScheduleMintID: scheduledTokenMintID, ScheduleMintTxId: sql.NullString{
			String: originalScheduledTokenMintID,
			Valid:  true,
		}}).
		Error
}

func (sr Repository) UpdateStatusScheduledTokenTransferSubmitted(ethTxHash, scheduledTokenTransferID, originalScheduledTokenTransferID string) error {
	return sr.dbClient.
		Model(entity.LockEvent{}).
		Where("id = ?", ethTxHash).
		Updates(entity.LockEvent{Status: lock_event_status.StatusTransferSubmitted, ScheduleTransferID: scheduledTokenTransferID, ScheduleTransferTxId: sql.NullString{
			String: originalScheduledTokenTransferID,
			Valid:  true,
		}}).
		Error
}

func (sr Repository) UpdateStatusScheduledTokenMintCompleted(id string) error {
	return sr.updateStatus(id, lock_event_status.StatusMintCompleted)
}

func (sr Repository) UpdateStatusCompleted(id string) error {
	return sr.updateStatus(id, lock_event_status.StatusCompleted)
}

func (sr Repository) UpdateStatusFailed(id string) error {
	return sr.updateStatus(id, lock_event_status.StatusFailed)
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

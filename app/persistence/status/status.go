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

package status

import (
	"fmt"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/process"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Repository struct {
	dbClient                 *gorm.DB
	lastFetchedTimestampCode string //"LAST_FETCHED_TIMESTAMP"
}

func NewRepositoryForStatus(dbClient *gorm.DB, statusType string) *Repository {
	typeCheck(statusType)
	return &Repository{
		dbClient:                 dbClient,
		lastFetchedTimestampCode: fmt.Sprintf("LAST_%s_TIMESTAMP", statusType),
	}
}

func typeCheck(statusType string) {
	switch statusType {
	case process.HCSMessageType:
	case process.CryptoTransferMessageType:
		return
	default:
		log.Fatal("Invalid status type.")
	}
}

func (s Repository) GetLastFetchedTimestamp(entityID string) (int64, error) {
	lastFetchedStatus := &entity.Status{}
	err := s.dbClient.
		Where("code = ? and entity_id = ?", s.lastFetchedTimestampCode, entityID).
		First(&lastFetchedStatus).Error
	if err != nil {
		return 0, err
	}
	return lastFetchedStatus.Timestamp, nil
}

func (s Repository) CreateTimestamp(entityID string, timestamp int64) error {
	return s.dbClient.Create(entity.Status{
		Name:      "Last fetched timestamp",
		EntityID:  entityID,
		Code:      s.lastFetchedTimestampCode,
		Timestamp: timestamp,
	}).Error
}

func (s Repository) UpdateLastFetchedTimestamp(entityID string, timestamp int64) error {
	return s.dbClient.
		Where("code = ? and entity_id = ?", s.lastFetchedTimestampCode, entityID).
		Save(entity.Status{
			Name:      "Last fetched timestamp",
			EntityID:  entityID,
			Code:      s.lastFetchedTimestampCode,
			Timestamp: timestamp,
		}).
		Error
}

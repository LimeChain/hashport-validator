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

package status

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Hedera Watcher SDK message types
const (
	Transfer = "TRANSFER"
	Message  = "TOPIC_MESSAGE"
)

type Repository struct {
	dbClient *gorm.DB
}

func NewRepositoryForStatus(dbClient *gorm.DB, statusType string) *Repository {
	typeCheck(statusType)
	return &Repository{
		dbClient: dbClient,
	}
}

func typeCheck(statusType string) {
	switch statusType {
	case Message:
	case Transfer:
		return
	default:
		log.Fatal("Invalid status type.")
	}
}

func (s Repository) Get(entityID string) (int64, error) {
	lastFetchedStatus := &entity.Status{}
	err := s.dbClient.
		Where("entity_id = ?", entityID).
		First(&lastFetchedStatus).Error
	if err != nil {
		return 0, err
	}
	return lastFetchedStatus.Last, nil
}

func (s Repository) Create(entityID string, timestampOrBlockNumber int64) error {
	return s.dbClient.Create(entity.Status{
		EntityID: entityID,
		Last:     timestampOrBlockNumber,
	}).Error
}

func (s Repository) Update(entityID string, timestampOrBlockNumber int64) error {
	return s.dbClient.
		Where("entity_id = ?", entityID).
		Save(entity.Status{
			EntityID: entityID,
			Last:     timestampOrBlockNumber,
		}).
		Error
}

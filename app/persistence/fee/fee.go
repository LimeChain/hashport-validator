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

package fee

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/fee"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Repository struct {
	dbClient *gorm.DB
	logger   *log.Entry
}

func NewRepository(dbClient *gorm.DB) *Repository {
	return &Repository{
		dbClient: dbClient,
		logger:   config.GetLoggerFor("Fee Repository"),
	}
}

func (r Repository) Create(entity *entity.Fee) error {
	return r.dbClient.Create(entity).Error
}

func (r Repository) UpdateStatusCompleted(txId string) error {
	return r.updateStatus(txId, fee.StatusCompleted)
}

func (r Repository) UpdateStatusFailed(txId string) error {
	return r.updateStatus(txId, fee.StatusFailed)
}

func (r Repository) updateStatus(txId string, status string) error {
	err := r.dbClient.
		Model(entity.Fee{}).
		Where("transaction_id = ?", txId).
		UpdateColumn("status", status).
		Error
	if err == nil {
		r.logger.Debugf("[%s] - Updated Status to [%s]", txId, status)
	}
	return err
}

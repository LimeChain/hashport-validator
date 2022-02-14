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

package transfer

import (
	"errors"
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
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
		logger:   config.GetLoggerFor("Transfer Repository"),
	}
}

// Returns Transfer. Returns nil if not found
func (tr Repository) GetByTransactionId(txId string) (*entity.Transfer, error) {
	tx := &entity.Transfer{}
	result := tr.dbClient.
		Model(entity.Transfer{}).
		Where("transaction_id = ?", txId).
		First(tx)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return tx, nil
}

func (tr Repository) GetWithPreloads(txId string) (*entity.Transfer, error) {
	tx := &entity.Transfer{}
	result := tr.dbClient.
		Preload("Fees").
		Preload("Messages").
		Model(entity.Transfer{}).
		Where("transaction_id = ?", txId).
		Find(tx)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	return tx, nil
}

// Returns Transfer with preloaded Fee table. Returns nil if not found
func (tr Repository) GetWithFee(txId string) (*entity.Transfer, error) {
	tx := &entity.Transfer{}
	result := tr.dbClient.
		Preload("Fees").
		Model(entity.Transfer{}).
		Where("transaction_id = ?", txId).
		First(tx)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return tx, nil
}

// Create creates new record of Transfer
func (tr Repository) Create(ct *model.Transfer) (*entity.Transfer, error) {
	return tr.create(ct, status.Initial)
}

// Save updates the provided Transfer instance
func (tr Repository) Save(tx *entity.Transfer) error {
	return tr.dbClient.Save(tx).Error
}

func (tr Repository) UpdateFee(txId string, fee string) error {
	err := tr.dbClient.
		Model(entity.Transfer{}).
		Where("transaction_id = ?", txId).
		UpdateColumn("fee", fee).
		Error
	if err == nil {
		tr.logger.Debugf("Updated Fee of TX [%s] to [%s]", txId, fee)
	}
	return err
}

func (tr Repository) UpdateStatusCompleted(txId string) error {
	return tr.updateStatus(txId, status.Completed)
}

func (tr Repository) UpdateStatusFailed(txId string) error {
	return tr.updateStatus(txId, status.Failed)
}

func (tr Repository) create(ct *model.Transfer, status string) (*entity.Transfer, error) {
	tx := &entity.Transfer{
		TransactionID: ct.TransactionId,
		SourceChainID: ct.SourceChainId,
		TargetChainID: ct.TargetChainId,
		NativeChainID: ct.NativeChainId,
		SourceAsset:   ct.SourceAsset,
		TargetAsset:   ct.TargetAsset,
		NativeAsset:   ct.NativeAsset,
		Receiver:      ct.Receiver,
		Amount:        ct.Amount,
		SerialNumber:  ct.SerialNum,
		Metadata:      ct.Metadata,
		IsNft:         ct.IsNft,
		Status:        status,
	}
	err := tr.dbClient.Create(tx).Error

	return tx, err
}

func (tr Repository) updateStatus(txId string, s string) error {
	// Sanity check
	if s != status.Initial &&
		s != status.Completed &&
		s != status.Failed {
		return errors.New("invalid status")
	}

	err := tr.dbClient.
		Model(entity.Transfer{}).
		Where("transaction_id = ?", txId).
		UpdateColumn("status", s).
		Error
	if err == nil {
		tr.logger.Debugf("Updated Status of TX [%s] to [%s]", txId, s)
	}
	return err
}

func (tr Repository) baseUpdateStatus(statusColumn, txId, status string, possibleStatuses []string) error {
	if !isValidStatus(status, possibleStatuses) {
		return errors.New("invalid status")
	}

	err := tr.dbClient.
		Model(entity.Transfer{}).
		Where("transaction_id = ?", txId).
		UpdateColumn(statusColumn, status).
		Error
	if err == nil {
		tr.logger.Debugf("[%s] - Column [%s] status to [%s]", txId, statusColumn, status)
	}
	return err
}

func isValidStatus(status string, possibleStatuses []string) bool {
	for _, option := range possibleStatuses {
		if status == option {
			return true
		}
	}
	return false
}

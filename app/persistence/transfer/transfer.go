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

package transfer

import (
	"errors"

	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/payload"

	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	countQuery = `SELECT COUNT(*) FROM (SELECT DISTINCT transaction_id FROM transfers) AS t`
)

type Repository struct {
	db     *gorm.DB
	logger *log.Entry
}

func NewRepository(dbClient *gorm.DB) *Repository {
	return &Repository{
		db:     dbClient,
		logger: config.GetLoggerFor("Transfer Repository"),
	}
}

// Returns Transfer. Returns nil if not found
func (r Repository) GetByTransactionId(txId string) (*entity.Transfer, error) {
	tx := &entity.Transfer{}
	result := r.db.
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

func (r Repository) GetWithPreloads(txId string) (*entity.Transfer, error) {
	tx := &entity.Transfer{}
	result := r.db.
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
func (r Repository) GetWithFee(txId string) (*entity.Transfer, error) {
	tx := &entity.Transfer{}
	result := r.db.
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
func (r Repository) Create(ct *payload.Transfer) (*entity.Transfer, error) {
	return r.create(ct, status.Initial)
}

// Save updates the provided Transfer instance
func (r Repository) Save(tx *entity.Transfer) error {
	return r.db.Save(tx).Error
}

func (r Repository) UpdateFee(txId string, fee string) error {
	err := r.db.
		Model(entity.Transfer{}).
		Where("transaction_id = ?", txId).
		UpdateColumn("fee", fee).
		Error
	if err == nil {
		r.logger.Debugf("Updated Fee of TX [%s] to [%s]", txId, fee)
	}
	return err
}

func (r Repository) UpdateStatusCompleted(txId string) error {
	return r.updateStatus(txId, status.Completed)
}

func (r Repository) UpdateStatusFailed(txId string) error {
	return r.updateStatus(txId, status.Failed)
}

func (r Repository) Paged(req *transfer.PagedRequest) ([]*entity.Transfer, error) {
	offset := (req.Page - 1) * req.PerPage
	res := make([]*entity.Transfer, 0, req.PerPage)
	q := r.db.
		Model(entity.Transfer{}).
		Order("timestamp desc, status asc").
		Offset(int(offset)).
		Limit(int(req.PerPage))

	f := req.Filter
	if f.Originator != "" {
		q.Where("originator = ?", f.Originator)
	}
	if f.TokenId != "" {
		q.Where("source_asset = ?", f.TokenId).
			Or("target_asset = ?", f.TokenId)
	}
	if f.TransactionId != "" {
		q.Where("transaction_id LIKE ?%", f.TransactionId).
			Or("transaction_id = ?", f.TransactionId)
	}
	if !f.Timestamp.IsZero() {
		q.Where("timestamp = ?", f.Timestamp.UnixNano())
	}

	err := q.Find(&res).Error
	if err != nil {
		r.logger.Errorf("Failed to get paged transfers: [%s]", err)
		return nil, err
	}

	return res, nil
}

func (r Repository) Count() (int64, error) {
	db, err := r.db.DB()
	if err != nil {
		return 0, err
	}

	cur, err := db.Query(countQuery)
	if err != nil {
		return 0, err
	}
	defer cur.Close()

	var res int64
	if cur.Next() {
		if err := cur.Scan(&res); err != nil {
			return 0, err
		}
	}

	return res, nil
}

func (r Repository) create(ct *payload.Transfer, status string) (*entity.Transfer, error) {
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
		Status:        status,
		SerialNumber:  ct.SerialNum,
		Metadata:      ct.Metadata,
		IsNft:         ct.IsNft,
		Timestamp:     entity.NanoTime{Time: ct.Timestamp},
		Originator:    ct.Originator,
	}
	err := r.db.Create(tx).Error

	return tx, err
}

func (r Repository) updateStatus(txId string, s string) error {
	// Sanity check
	if s != status.Initial &&
		s != status.Completed &&
		s != status.Failed {
		return errors.New("invalid status")
	}

	err := r.db.
		Model(entity.Transfer{}).
		Where("transaction_id = ?", txId).
		UpdateColumn("status", s).
		Error
	if err == nil {
		r.logger.Debugf("Updated Status of TX [%s] to [%s]", txId, s)
	}
	return err
}

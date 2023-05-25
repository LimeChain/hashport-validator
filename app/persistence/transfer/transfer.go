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
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/limechain/hedera-eth-bridge-validator/constants"

	"github.com/ethereum/go-ethereum/common"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/payload"

	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
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
func (r *Repository) GetByTransactionId(txId string) (*entity.Transfer, error) {
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
	r.updateHederaChainId(tx)

	return tx, nil
}

func (r *Repository) GetWithPreloads(txId string) (*entity.Transfer, error) {
	tx := &entity.Transfer{}
	result := r.db.
		Preload("Fees").
		Preload("Messages").
		Model(entity.Transfer{}).
		Where("transaction_id = ?", txId).
		First(tx)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	r.updateHederaChainId(tx)

	return tx, nil
}

// Returns Transfer with preloaded Fee table. Returns nil if not found
func (r *Repository) GetWithFee(txId string) (*entity.Transfer, error) {
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
	r.updateHederaChainId(tx)

	return tx, nil
}

// Create creates new record of Transfer
func (r *Repository) Create(ct *payload.Transfer) (*entity.Transfer, error) {
	return r.create(ct, status.Initial)
}

// Save updates the provided Transfer instance
func (r *Repository) Save(tx *entity.Transfer) error {
	return r.db.Save(tx).Error
}

func (r *Repository) UpdateFee(txId string, fee string) error {
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

func (r *Repository) UpdateStatusCompleted(txId string) error {
	return r.updateStatus(txId, status.Completed)
}

func (r *Repository) UpdateStatusFailed(txId string) error {
	return r.updateStatus(txId, status.Failed)
}

func formatTimestampFilter(q *gorm.DB, ts_query string) (*gorm.DB, error) {
	qParams := strings.Split(ts_query, "&")
	operators := map[string]string{
		"eq":  "=",
		"gt":  ">",
		"lt":  "<",
		"gte": ">=",
		"lte": "<=",
	}

	// This if statement handles legacy timestamp filter like:
	// "timestamp": "2023-04-19T04:41:47.104114905Z"
	if len(qParams) == 1 && !strings.Contains(ts_query, "=") {
		timestamp, err := time.Parse(time.RFC3339Nano, qParams[0])
		if err != nil {
			return q, err
		}

		q = q.Where("timestamp = ?", timestamp.UnixNano())
		return q, nil
	}

	// This for loop handles new timestamp filter like:
	// "timestamp": "lte=2023-05-19T04:41:47.104114905Z&gte=2023-04-19T04:41:47.104114905Z"
	// "timestamp": "eq=2023-04-19T04:41:47.104114905Z"
	for _, param := range qParams {
		parts := strings.Split(param, "=")

		operator := parts[0]
		datetime, err := time.Parse(time.RFC3339Nano, parts[1])
		if err != nil {
			return q, err
		}

		timestamp := datetime.UnixNano()

		op, exists := operators[operator]
		if !exists {
			return q, service.ErrWrongQuery
		}

		q = q.Where("timestamp "+op+" ?", timestamp)

	}

	return q, nil
}

func (r *Repository) Paged(req *transfer.PagedRequest) ([]*entity.Transfer, int64, error) {
	var (
		err   error
		count int64
	)

	offset := (req.Page - 1) * req.PageSize
	res := make([]*entity.Transfer, 0, req.PageSize)
	f := req.Filter
	q := r.db.
		Model(entity.Transfer{}).
		Order("timestamp desc, status asc")

	if f.Originator != "" {
		if strings.Contains(f.Originator, "0x") {
			a := common.HexToAddress(f.Originator).String()
			q = q.Where("originator = ?", a)
		} else {
			q = q.Where("originator = ?", f.Originator)
		}
	}

	if f.TimestampQuery != "" {
		q, err = formatTimestampFilter(q, f.TimestampQuery)
		if err != nil {
			r.logger.Errorf("Failed to get paged transfers: [%s]", err)
			return nil, 0, err
		}

	}
	if f.TokenId != "" {
		if strings.Contains(f.TokenId, "0x") {
			a := common.HexToAddress(f.TokenId).String()
			q = q.Where("(source_asset = @address OR target_asset = @address)", sql.Named("address", a))
		} else {
			q = q.Where("(source_asset = @tokenId OR target_asset = @tokenId)", sql.Named("tokenId", f.TokenId))
		}
	}
	if f.TransactionId != "" {
		q = q.Where("transaction_id LIKE ?", fmt.Sprintf(`%s%%`, f.TransactionId))
	}

	q = q.Count(&count).
		Offset(int(offset)).
		Limit(int(req.PageSize))

	err = q.Find(&res).Error
	if err != nil {
		r.logger.Errorf("Failed to get paged transfers: [%s]", err)
		return nil, 0, err
	}

	return res, count, nil
}

func (r *Repository) create(ct *payload.Transfer, status string) (*entity.Transfer, error) {
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

func (r *Repository) updateStatus(txId string, s string) error {
	// Sanity check
	if s != status.Initial &&
		s != status.Completed &&
		s != status.Failed {
		return errors.New("invalid status")
	}

	result := r.db.
		Model(entity.Transfer{}).
		Where("transaction_id = ?", txId).
		UpdateColumn("status", s)
	if result.Error == nil {
		if s == status.Failed {
			r.logger.Errorf("Updated Status of TX [%s] to [%s]", txId, s)
			return nil
		}
		r.logger.Infof("Updated Status of TX [%s] to [%s]", txId, s)
	}

	if result.RowsAffected != 1 {
		return fmt.Errorf("updated %d rows, expected 1", result.RowsAffected)
	}

	return result.Error
}

func (r *Repository) updateHederaChainId(tx *entity.Transfer) {
	// SourceChainID
	if tx.SourceChainID == constants.OldHederaNetworkId {
		tx.SourceChainID = constants.HederaNetworkId
	}
	// NativeChainID
	if tx.NativeChainID == constants.OldHederaNetworkId {
		tx.NativeChainID = constants.HederaNetworkId
	}
	// TargetChainID
	if tx.TargetChainID == constants.OldHederaNetworkId {
		tx.TargetChainID = constants.HederaNetworkId
	}
}

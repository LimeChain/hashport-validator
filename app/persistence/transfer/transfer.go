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
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/transfer"
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
		Preload("Fee").
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
		Preload("Fee").
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

func (tr Repository) GetInitialAndSignatureSubmittedTx() ([]*entity.Transfer, error) {
	var transfers []*entity.Transfer

	err := tr.dbClient.
		Model(entity.Transfer{}).
		Where("status = ? OR status = ?", transfer.StatusInitial, transfer.StatusSignatureSubmitted).
		Find(&transfers).Error
	if err != nil {
		return nil, err
	}

	return transfers, nil
}

// Create creates new record of Transfer
func (tr Repository) Create(ct *model.Transfer) (*entity.Transfer, error) {
	return tr.create(ct, transfer.StatusInitial)
}

// Save updates the provided Transfer instance
func (tr Repository) Save(tx *entity.Transfer) error {
	return tr.dbClient.Save(tx).Error
}

func (tr *Repository) SaveRecoveredTxn(ct *model.Transfer) error {
	_, err := tr.create(ct, transfer.StatusRecovered)
	return err
}

func (tr Repository) UpdateStatusCompleted(txId string) error {
	return tr.updateStatus(txId, transfer.StatusCompleted)
}

func (tr Repository) UpdateStatusSignatureSubmitted(txId string) error {
	err := tr.dbClient.
		Model(entity.Transfer{}).
		Where("transaction_id = ?", txId).
		Updates(entity.Transfer{SignatureMsgStatus: transfer.StatusSignatureSubmitted, Status: transfer.StatusInProgress}).
		Error
	if err == nil {
		tr.logger.Debugf("[%s] - Updated Status to [%s] and SignatureMsgStatus to [%s]", txId, transfer.StatusInProgress, transfer.StatusSignatureSubmitted)
	}
	return err
}

func (tr Repository) UpdateStatusSignatureMined(txId string) error {
	return tr.updateSignatureStatus(txId, transfer.StatusSignatureMined)
}

func (tr Repository) UpdateStatusSignatureFailed(txId string) error {
	return tr.updateSignatureStatus(txId, transfer.StatusSignatureFailed)
}

func (tr Repository) UpdateStatusScheduledTokenBurnSubmitted(txId string) error {
	return tr.updateSignatureStatus(txId, transfer.StatusScheduledTokenBurnSubmitted)
}

func (tr Repository) UpdateStatusScheduledTokenBurnFailed(txId string) error {
	return tr.updateSignatureStatus(txId, transfer.StatusScheduledTokenBurnFailed)
}

func (tr Repository) UpdateStatusScheduledTokenBurnCompleted(txId string) error {
	return tr.updateSignatureStatus(txId, transfer.StatusScheduledTokenBurnCompleted)
}

func (tr Repository) create(ct *model.Transfer, status string) (*entity.Transfer, error) {
	tx := &entity.Transfer{
		TransactionID: ct.TransactionId,
		Receiver:      ct.Receiver,
		Amount:        ct.Amount,
		Status:        status,
		NativeAsset:   ct.NativeAsset,
		WrappedAsset:  ct.WrappedAsset,
		RouterAddress: ct.RouterAddress,
	}
	err := tr.dbClient.Create(tx).Error

	return tx, err
}

func (tr Repository) updateStatus(txId string, status string) error {
	// Sanity check
	if status != transfer.StatusInitial &&
		status != transfer.StatusInProgress &&
		status != transfer.StatusCompleted {
		return errors.New("invalid signature status")
	}

	err := tr.dbClient.
		Model(entity.Transfer{}).
		Where("transaction_id = ?", txId).
		UpdateColumn("status", status).
		Error
	if err == nil {
		tr.logger.Debugf("Updated Status of TX [%s] to [%s]", txId, status)
	}
	return err
}

func (tr Repository) updateSignatureStatus(txId string, status string) error {
	return tr.baseUpdateStatus("signature_msg_status", txId, status, []string{transfer.StatusSignatureSubmitted, transfer.StatusSignatureMined, transfer.StatusSignatureFailed})
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

func (tr *Repository) GetUnprocessedTransfers() ([]*entity.Transfer, error) {
	var transfers []*entity.Transfer

	err := tr.dbClient.
		Where("status IN ?", []string{transfer.StatusInitial, transfer.StatusRecovered}).
		Find(&transfers).Error
	if err != nil {
		return nil, err
	}

	return transfers, nil
}

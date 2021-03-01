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

package scheduled

import (
	"gorm.io/gorm"
)

const (
	StatusCompleted = "COMPLETED"
	StatusFailed    = "FAILED"
	StatusInitial   = "INITIAL"
	StatusSubmitted = "SUBMITTED"
)

// contains information about scheduled (unwrapping) transactions
type Scheduled struct {
	gorm.Model
	Amount                   int64
	Nonce                    string `gorm:"unique"`
	BridgeThresholdAccountID string
	PayerAccountID           string
	Recipient                string
	Status                   string
	ScheduleID               string
	SubmissionTxId           string `gorm:"unique"`
}

type ScheduledRepository struct {
	dbClient *gorm.DB
}

func NewScheduledRepository(dbClient *gorm.DB) *ScheduledRepository {
	return &ScheduledRepository{
		dbClient: dbClient,
	}
}

func (sr *ScheduledRepository) Create(amount int64, nonce, recipient, bridgeThresholdAccountID, payerAccountID string) error {
	return sr.dbClient.Create(&Scheduled{
		Model:                    gorm.Model{},
		Amount:                   amount,
		Nonce:                    nonce,
		BridgeThresholdAccountID: bridgeThresholdAccountID,
		PayerAccountID:           payerAccountID,
		Recipient:                recipient,
		Status:                   StatusInitial,
	}).Error
}

func (sr *ScheduledRepository) UpdateStatusSubmitted(nonce, scheduleID, submissionTxId string) error {
	return sr.dbClient.
		Model(Scheduled{}).
		Where("nonce = ?", nonce).
		Updates(Scheduled{Status: StatusSubmitted, ScheduleID: scheduleID, SubmissionTxId: submissionTxId}).
		Error
}

func (sr *ScheduledRepository) UpdateStatusCompleted(txId string) error {
	return sr.updateStatus(txId, StatusCompleted)
}

func (sr *ScheduledRepository) UpdateStatusFailed(txId string) error {
	return sr.updateStatus(txId, StatusFailed)
}

func (sr *ScheduledRepository) updateStatus(txId string, status string) error {
	return sr.dbClient.
		Model(Scheduled{}).
		Where("submission_tx_id = ?", txId).
		UpdateColumn("status", status).
		Error
}

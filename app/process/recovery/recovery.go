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

package recovery

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

type Recovery struct {
	feeRepository      repository.Fee
	scheduleRepository repository.Schedule
	mirrorClient       hedera.MirrorNode
	logger             *log.Entry
}

func New(
	feeRepository repository.Fee,
	scheduleRepository repository.Schedule,
	mirrorClient hedera.MirrorNode) *Recovery {
	return &Recovery{
		feeRepository:      feeRepository,
		scheduleRepository: scheduleRepository,
		mirrorClient:       mirrorClient,
		logger:             config.GetLoggerFor("Recovery"),
	}
}

func (r Recovery) Execute() {
	go r.checkSubmittedFees()
	go r.checkSubmittedSchedules()
}

func (r Recovery) checkSubmittedFees() {
	fees, err := r.feeRepository.GetAllSubmittedIds()
	if err != nil {
		r.logger.Errorf("Failed to get all submitted fees. Error: [%s].", err)
		return
	}

	for _, fee := range fees {
		onSuccess, onRevert := r.callbacks(fee.TransactionID, true)
		r.mirrorClient.WaitForScheduledTransaction(fee.TransactionID, onSuccess, onRevert)
	}
}

func (r Recovery) checkSubmittedSchedules() {
	schedules, err := r.scheduleRepository.GetAllSubmittedIds()
	if err != nil {
		r.logger.Errorf("Failed to get all submitted fees. Error: [%s].", err)
		return
	}

	for _, schedule := range schedules {
		onSuccess, onRevert := r.callbacks(schedule.TransactionID, false)
		r.mirrorClient.WaitForScheduledTransaction(schedule.TransactionID, onSuccess, onRevert)
	}
}

func (r Recovery) callbacks(transactionID string, isFee bool) (onSuccess, onRevert func()) {
	if isFee {
		onSuccess = func() {
			err := r.feeRepository.UpdateStatusCompleted(transactionID)
			if err != nil {
				r.logger.Errorf("[%s] - Failed to update fee status completed. Error [%s].", transactionID, err)
				return
			}
		}

		onRevert = func() {
			err := r.feeRepository.UpdateStatusFailed(transactionID)
			if err != nil {
				r.logger.Errorf("[%s] - Failed to update fee status failed. Error [%s].", transactionID, err)
				return
			}
		}
	} else {
		onSuccess = func() {
			err := r.scheduleRepository.UpdateStatusCompleted(transactionID)
			if err != nil {
				r.logger.Errorf("[%s] - Failed to update schedule status completed. Error [%s].", transactionID, err)
				return
			}
		}

		onRevert = func() {
			err := r.scheduleRepository.UpdateStatusFailed(transactionID)
			if err != nil {
				r.logger.Errorf("[%s] - Failed to update schedule status failed. Error [%s].", transactionID, err)
				return
			}
		}
	}

	return onSuccess, onRevert
}

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
	hederaHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/schedule"
	"sync"

	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/payload"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

// Handler is transfers event handler
type Handler struct {
	bridgeAccount      hedera.AccountID
	repository         repository.Transfer
	scheduleRepository repository.Schedule
	scheduledService   service.Scheduled
	transfersService   service.Transfers
	logger             *log.Entry
}

func NewHandler(
	bridgeAccount string,
	repository repository.Transfer,
	scheduleRepository repository.Schedule,
	transfersService service.Transfers,
	scheduledService service.Scheduled,
) *Handler {
	bridgeAcc, err := hedera.AccountIDFromString(bridgeAccount)
	if err != nil {
		log.Fatalf("Invalid account id [%s]. Error: [%s]", bridgeAccount, err)
	}

	return &Handler{
		bridgeAccount:      bridgeAcc,
		repository:         repository,
		scheduleRepository: scheduleRepository,
		scheduledService:   scheduledService,
		transfersService:   transfersService,
		logger:             config.GetLoggerFor("Hedera Native Scheduled Nft Transfer Handler"),
	}
}

func (nth Handler) Handle(p interface{}) {
	transfer, ok := p.(*payload.Transfer)
	if !ok {
		nth.logger.Errorf("Could not cast payload [%s]", p)
		return
	}

	receiver, err := hedera.AccountIDFromString(transfer.Receiver)
	if err != nil {
		nth.logger.Errorf("[%s] - Failed to parse event account [%s]. Error [%s].", transfer.TransactionId, transfer.Receiver, err)
		return
	}

	token, err := hedera.TokenIDFromString(transfer.TargetAsset)
	if err != nil {
		nth.logger.Errorf("[%s] - Failed to parse token [%s]. Error [%s].", transfer.TransactionId, transfer.TargetAsset, err)
		return
	}
	nftID := hedera.NftID{
		TokenID:      token,
		SerialNumber: transfer.SerialNum,
	}

	transactionRecord, err := nth.transfersService.InitiateNewTransfer(*transfer)
	if err != nil {
		nth.logger.Errorf("[%s] - Error occurred while initiating processing. Error: [%s]", transfer.TransactionId, err)
		return
	}

	if transactionRecord.Status != status.Initial {
		nth.logger.Debugf("[%s] - Previously added with status [%s]. Skipping further execution.", transactionRecord.TransactionID, transactionRecord.Status)
		return
	}

	var statusResult string
	wg := new(sync.WaitGroup)
	wg.Add(1)
	onExecutionSuccess, onExecutionFail := hederaHelper.ScheduledNftTxExecutionCallbacks(nth.repository, nth.scheduleRepository, nth.logger, transfer.TransactionId, true, &statusResult, schedule.APPROVE, wg)
	onSuccess, onFail := hederaHelper.ScheduledNftTxMinedCallbacks(nth.repository, nth.scheduleRepository, nth.logger, transfer.TransactionId, &statusResult, wg)

	nth.scheduledService.ExecuteScheduledNftAllowTransaction(transfer.TransactionId, nftID, nth.bridgeAccount, receiver, onExecutionSuccess, onExecutionFail, onSuccess, onFail)
}

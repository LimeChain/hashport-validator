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

package mint_hts

import (
	"database/sql"

	"github.com/limechain/hedera-eth-bridge-validator/app/process/payload"

	"github.com/hashgraph/hedera-sdk-go/v2"
	mirrorNode "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/metrics"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/schedule"
	entityStatus "github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/status"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

// Handler is transfers event handler
type Handler struct {
	scheduleRepository repository.Schedule
	mirrorNode         client.MirrorNode
	bridgeAccount      hedera.AccountID
	transfersService   service.Transfers
	readOnlyService    service.ReadOnly
	prometheusService  service.Prometheus
	logger             *log.Entry
}

func NewHandler(
	scheduleRepository repository.Schedule,
	bridgeAccount string,
	mirrorNode client.MirrorNode,
	transfersService service.Transfers,
	readOnlyService service.ReadOnly,
	prometheusService service.Prometheus) *Handler {

	bridgeAcc, err := hedera.AccountIDFromString(bridgeAccount)
	if err != nil {
		log.Fatalf("Invalid account id [%s]. Error: [%s]", bridgeAccount, err)
	}

	return &Handler{
		scheduleRepository: scheduleRepository,
		bridgeAccount:      bridgeAcc,
		mirrorNode:         mirrorNode,
		logger:             config.GetLoggerFor("Hedera Mint and Transfer Handler"),
		transfersService:   transfersService,
		readOnlyService:    readOnlyService,
		prometheusService:  prometheusService,
	}
}

func (fmh *Handler) Handle(p interface{}) {
	transferMsg, ok := p.(*payload.Transfer)
	if !ok {
		fmh.logger.Errorf("Could not cast payload [%s]", p)
		return
	}

	transactionRecord, err := fmh.transfersService.InitiateNewTransfer(*transferMsg)
	if err != nil {
		fmh.logger.Errorf("[%s] - Error occurred while initiating processing. Error: [%s]", transferMsg.TransactionId, err)
		return
	}

	if transactionRecord.Status != entityStatus.Initial {
		fmh.logger.Debugf("[%s] - Previously added with status [%s]. Skipping further execution.", transactionRecord.TransactionID, transactionRecord.Status)
		return
	}

	fmh.readOnlyService.FindTransfer(transferMsg.TransactionId,
		func() (*mirrorNode.Response, error) {
			return fmh.mirrorNode.GetAccountTokenMintTransactionsAfterTimestampString(fmh.bridgeAccount, transferMsg.NetworkTimestamp)
		},
		func(transactionID, scheduleID, status string) error {
			return fmh.scheduleRepository.Create(&entity.Schedule{
				TransactionID: transactionID,
				ScheduleID:    scheduleID,
				Operation:     schedule.MINT,
				Status:        status,
				TransferID: sql.NullString{
					String: transferMsg.TransactionId,
					Valid:  true,
				},
			})
		})

	fmh.readOnlyService.FindTransfer(
		transferMsg.TransactionId,
		func() (*mirrorNode.Response, error) {
			return fmh.mirrorNode.GetAccountDebitTransactionsAfterTimestampString(fmh.bridgeAccount, transferMsg.NetworkTimestamp)
		},
		func(transactionID, scheduleID, status string) error {

			if status == entityStatus.Completed {
				metrics.SetUserGetHisTokens(
					transferMsg.SourceChainId,
					transferMsg.TargetChainId,
					transferMsg.SourceAsset,
					transferMsg.TransactionId,
					fmh.prometheusService,
					fmh.logger,
				)
			}

			return fmh.scheduleRepository.Create(&entity.Schedule{
				TransactionID: transactionID,
				ScheduleID:    scheduleID,
				Operation:     schedule.TRANSFER,
				HasReceiver:   true,
				Status:        status,
				TransferID: sql.NullString{
					String: transferMsg.TransactionId,
					Valid:  true,
				},
			})
		})
}

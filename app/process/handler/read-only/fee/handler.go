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
	"database/sql"
	"github.com/hashgraph/hedera-sdk-go/v2"
	mirror_node "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/fee"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/schedule"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	log "github.com/sirupsen/logrus"
	"strconv"
)

// Handler is transfers event handler
type Handler struct {
	feeRepository      repository.Fee
	scheduleRepository repository.Schedule
	mirrorNode         client.MirrorNode
	bridgeAccount      hedera.AccountID
	distributor        service.Distributor
	feeService         service.Fee
	transfersService   service.Transfers
	transferRepository repository.Transfer
	logger             *log.Entry
}

func NewHandler(
	feeRepository repository.Fee,
	scheduleRepository repository.Schedule,
	mirrorNode client.MirrorNode,
	bridgeAccount string,
	distributor service.Distributor,
	feeService service.Fee,
	transferRepository repository.Transfer,
	transfersService service.Transfers) *Handler {
	bridgeAcc, err := hedera.AccountIDFromString(bridgeAccount)
	if err != nil {
		log.Fatalf("Invalid account id [%s]. Error: [%s]", bridgeAccount, err)
	}
	return &Handler{
		feeRepository:      feeRepository,
		scheduleRepository: scheduleRepository,
		mirrorNode:         mirrorNode,
		bridgeAccount:      bridgeAcc,
		logger:             config.GetLoggerFor("Hedera Transfer and Topic Submission Read-only Handler"),
		transfersService:   transfersService,
		distributor:        distributor,
		feeService:         feeService,
		transferRepository: transferRepository,
	}
}

func (fmh Handler) Handle(payload interface{}) {
	transferMsg, ok := payload.(*model.Transfer)
	if !ok {
		fmh.logger.Errorf("Could not cast payload [%s]", payload)
		return
	}

	transactionRecord, err := fmh.transfersService.InitiateNewTransfer(*transferMsg)
	if err != nil {
		fmh.logger.Errorf("[%s] - Error occurred while initiating processing. Error: [%s]", transferMsg.TransactionId, err)
		return
	}

	if transactionRecord.Status != transfer.StatusInitial {
		fmh.logger.Debugf("[%s] - Previously added with status [%s]. Skipping further execution.", transactionRecord.TransactionID, transactionRecord.Status)
		return
	}

	// TODO: Find Transfer TX with fee, messages will come by themselves (NH -> WEVM)

	intAmount, err := strconv.ParseInt(transferMsg.Amount, 10, 64)
	if err != nil {
		fmh.logger.Errorf("[%s] - Failed to parse amount. Error: [%s]", transferMsg.TransactionId, err)
		return
	}

	calculatedFee, _ := fmh.feeService.CalculateFee(transferMsg.SourceAsset, intAmount)
	validFee := fmh.distributor.ValidAmount(calculatedFee)

	expectedTransfers, err := fmh.distributor.PrepareTransfers(validFee, transferMsg.SourceAsset)
	if err != nil {
		fmh.logger.Errorf("[%s] Fee - Failed to Distribute to Members. Error: [%s].", transferMsg.TransactionId, err)
		return
	}

	if transferMsg.SourceAsset == constants.Hbar {
		expectedTransfers = append(expectedTransfers,
			mirror_node.Transfer{
				Account: fmh.bridgeAccount.String(),
				Amount:  -validFee,
			})
	} else {
		expectedTransfers = append(expectedTransfers,
			mirror_node.Transfer{
				Account: fmh.bridgeAccount.String(),
				Amount:  -validFee,
				Token:   transferMsg.SourceAsset,
			})
	}

	for {
		response, err := fmh.mirrorNode.GetAccountDebitTransactionsAfterTimestampString(fmh.bridgeAccount, transferMsg.Timestamp)
		if err != nil {
			fmh.logger.Errorf("[%s] - Failed to get token burn transactions after timestamp. Error: [%s]", transactionRecord.TransactionID, err)
		}

		finished := false
		for _, transaction := range response.Transactions {
			found := false
			for _, expectedTransfer := range expectedTransfers {
				contained := false
				if transferMsg.SourceAsset == constants.Hbar {
					for _, t := range transaction.Transfers {
						if expectedTransfer == t {
							contained = true
							break
						}
					}
				} else {
					for _, t := range transaction.TokenTransfers {
						if expectedTransfer == t {
							contained = true
							break
						}
					}
				}

				if !contained {
					found = false
					break
				} else {
					found = true
				}
			}
			if found {
				isFound := false
				scheduledTx, err := fmh.mirrorNode.GetScheduledTransaction(transaction.TransactionID)
				if err != nil {
					fmh.logger.Errorf("[%s] - Failed to retrieve scheduled transaction [%s]. Error: [%s]", transferMsg.TransactionId, transaction.TransactionID, err)
					continue
				}
				for _, tx := range scheduledTx.Transactions {
					if tx.Result == hedera.StatusSuccess.String() {
						scheduleID, err := fmh.mirrorNode.GetSchedule(tx.EntityId)
						if err != nil {
							fmh.logger.Errorf("[%s] - Failed to get scheduled entity [%s]. Error: [%s]", transferMsg.TransactionId, scheduleID, err)
							break
						}
						if scheduleID.Memo == transferMsg.TransactionId {
							err := fmh.scheduleRepository.Create(&entity.Schedule{
								TransactionID: transaction.TransactionID,
								ScheduleID:    tx.EntityId,
								Operation:     schedule.TRANSFER,
								Status:        schedule.StatusCompleted,
								TransferID: sql.NullString{
									String: transferMsg.TransactionId,
									Valid:  true,
								},
							})
							if err != nil {
								fmh.logger.Errorf("[%s] - Failed to create scheduled entity [%s]. Error: [%s]", transferMsg.TransactionId, tx.EntityId, err)
								break
							}
							err = fmh.feeRepository.Create(&entity.Fee{
								TransactionID: transaction.TransactionID,
								ScheduleID:    tx.EntityId,
								Amount:        strconv.FormatInt(validFee, 10),
								Status:        fee.StatusCompleted,
								TransferID: sql.NullString{
									String: transferMsg.TransactionId,
									Valid:  true,
								},
							})
							if err != nil {
								fmh.logger.Errorf("[%s] - Failed to create fee  entity [%s]. Error: [%s]", transferMsg.TransactionId, tx.EntityId, err)
								break
							}

							err = fmh.transferRepository.UpdateStatusCompleted(transferMsg.TransactionId)
							if err != nil {
								fmh.logger.Errorf("[%s] - Failed to update status completed [%s]. Error: [%s]", transferMsg.TransactionId, tx.EntityId, err)
								break
							}
							isFound = true
						}
					}
				}
				if isFound {
					finished = true
					break
				}
			}
		}
		if finished {
			break
		}
	}
}

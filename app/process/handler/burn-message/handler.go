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

package burn_message

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	transfersService service.Transfers
	logger           *log.Entry
}

func NewHandler(transferService service.Transfers) *Handler {
	return &Handler{
		transfersService: transferService,
		logger:           config.GetLoggerFor("Hedera Burn and Topic Message Handler"),
	}
}

func (mhh Handler) Handle(payload interface{}) {
	transferMsg, ok := payload.(*model.Transfer)
	if !ok {
		mhh.logger.Errorf("Could not cast payload [%s]", payload)
		return
	}

	transactionRecord, err := mhh.transfersService.InitiateNewTransfer(*transferMsg)
	if err != nil {
		mhh.logger.Errorf("[%s] - Error occurred while initiating processing. Error: [%s]", transferMsg.TransactionId, err)
		return
	}

	if transactionRecord.Status != transfer.StatusInitial {
		mhh.logger.Debugf("[%s] - Previously added with status [%s]. Skipping further execution.", transactionRecord.TransactionID, transactionRecord.Status)
		return
	}

	err = mhh.transfersService.ProcessWrappedTransfer(*transferMsg)
	if err != nil {
		mhh.logger.Errorf("[%s] - Processing failed. Error: [%s]", transferMsg.TransactionId, err)
		return
	}
}

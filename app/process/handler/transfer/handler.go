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
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/encoding"
	txRepo "github.com/limechain/hedera-eth-bridge-validator/app/persistence/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

// Handler is transfers event handler
type Handler struct {
	transfersService service.Transfers
	logger           *log.Entry
}

func NewHandler(transfersService service.Transfers) *Handler {

	return &Handler{
		logger:           config.GetLoggerFor("Account Transfer Handler"),
		transfersService: transfersService,
	}
}

func (th Handler) Handle(payload []byte) {
	transferMsg, err := encoding.NewTransferMessageFromBytes(payload)
	if err != nil {
		th.logger.Errorf("Failed to parse incoming payload. Error [%s].", err)
		return
	}

	transactionRecord, err := th.transfersService.InitiateNewTransfer(*transferMsg)
	if err != nil {
		th.logger.Errorf("Error occurred while initiating TX ID [%s] processing", transferMsg.TransactionId)
		return
	}

	if transactionRecord.Status != txRepo.StatusInitial {
		th.logger.Debugf("Previously added Transaction with TransactionID [%s] has status [%s]. Skipping further execution.", transactionRecord.TransactionId, transactionRecord.Status)
		return
	}

	if transferMsg.ExecuteEthTransaction {
		err = th.transfersService.VerifyFee(*transferMsg)
		if err != nil {
			th.logger.Errorf("Fee validation failed for TX [%s]. Skipping further execution", transferMsg.TransactionId)
			return
		}
	}

	err = th.transfersService.ProcessTransfer(*transferMsg)
	if err != nil {
		th.logger.Errorf("Processing of TX [%s] failed", transferMsg.TransactionId)
		return
	}
}

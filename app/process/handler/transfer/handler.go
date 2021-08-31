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
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	model "github.com/limechain/hedera-eth-bridge-validator/app/model/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/transfer"
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

func (th Handler) Handle(payload interface{}) {
	var err error
	switch t := payload.(type) {
	case *model.NativeTransfer:
		err = th.initiateTransferAndCheckStatus(t.Transfer)
		if err != nil {
			return
		}

		err = th.transfersService.ProcessNativeTransfer(t.Transfer)
		if err != nil {
			th.logger.Errorf("[%s] - Processing failed. Error: [%s]", t.TransactionId, err)
			return
		}
	case *model.WrappedTransfer:
		err = th.initiateTransferAndCheckStatus(t.Transfer)
		if err != nil {
			return
		}

		err = th.transfersService.ProcessWrappedTransfer(t.Transfer)
		if err != nil {
			th.logger.Errorf("[%s] - Processing failed. Error: [%s]", t.TransactionId, err)
			return
		}
	}
}

func (th Handler) initiateTransferAndCheckStatus(transferMsg model.Transfer) error {
	transactionRecord, err := th.transfersService.InitiateNewTransfer(transferMsg)
	if err != nil {
		th.logger.Errorf("[%s] - Error occurred while initiating processing. Error: [%s]", transferMsg.TransactionId, err)
		return err
	}

	if transactionRecord.Status != transfer.StatusInitial {
		th.logger.Debugf("[%s] - Previously added with status [%s]. Skipping further execution.", transactionRecord.TransactionID, transactionRecord.Status)
		return errors.New("previously-added-record")
	}
	return nil
}

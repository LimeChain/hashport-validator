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

package fee_transfer

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/payload"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	burnService service.BurnEvent
	logger      *log.Entry
}

func NewHandler(burnService service.BurnEvent) *Handler {
	return &Handler{
		burnService: burnService,
		logger:      config.GetLoggerFor("Hedera Fee and Schedule Transfer Handler"),
	}
}

func (fth Handler) Handle(p interface{}) {
	event, ok := p.(*payload.Transfer)
	if !ok {
		fth.logger.Errorf("Could not cast payload [%s]", p)
		return
	}
	fth.burnService.ProcessEvent(*event)
}

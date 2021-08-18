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

package evm

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	burn_event "github.com/limechain/hedera-eth-bridge-validator/app/model/burn-event"
	lock_event "github.com/limechain/hedera-eth-bridge-validator/app/model/lock-event"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	lockService service.LockEvent
	burnService service.BurnEvent
	logger      *log.Entry
}

func NewHandler(burnService service.BurnEvent, lockService service.LockEvent) *Handler {
	return &Handler{
		lockService: lockService,
		burnService: burnService,
		logger:      config.GetLoggerFor("EVM Event Handler"),
	}
}

func (sth Handler) Handle(payload interface{}) {
	switch event := payload.(type) {
	case *burn_event.BurnEvent:
		sth.burnService.ProcessEvent(*event)
	case *lock_event.LockEvent:
		sth.lockService.ProcessEvent(*event)
	default:
		sth.logger.Errorf("Could not cast payload [%s] to any of the events: [burnEvent, lockEvent]", payload)
	}
}

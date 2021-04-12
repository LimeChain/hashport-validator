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

package burn_event

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	burn_event "github.com/limechain/hedera-eth-bridge-validator/app/model/burn-event"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	service service.BurnEvent
	logger  *log.Entry
}

func NewHandler(service service.BurnEvent) *Handler {
	return &Handler{
		service: service,
		logger:  config.GetLoggerFor("Scheduled Transaction Handler"),
	}
}

func (sth Handler) Handle(payload interface{}) {
	burnEvent, ok := payload.(*burn_event.BurnEvent)
	if !ok {
		sth.logger.Errorf("Error could not cast BurnEvent payload [%s]", payload)
		return
	}

	sth.service.ProcessEvent(*burnEvent)
}

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

package lock_event

import (
	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

// TODO: Create LockEvent Service

type Service struct {
	bridgeAccount      hedera.AccountID
	feeRepository      repository.Fee
	repository         repository.LockEvent
	distributorService service.Distributor
	feeService         service.Fee
	scheduledService   service.Scheduled
	logger             *log.Entry
}

func NewService(
	bridgeAccount string,
	repository repository.LockEvent,
	feeRepository repository.Fee,
	distributor service.Distributor,
	scheduled service.Scheduled,
	feeService service.Fee) *Service {

	bridgeAcc, err := hedera.AccountIDFromString(bridgeAccount)
	if err != nil {
		log.Fatalf("Invalid bridge account: [%s].", bridgeAccount)
	}

	return &Service{
		bridgeAccount:      bridgeAcc,
		feeRepository:      feeRepository,
		repository:         repository,
		distributorService: distributor,
		feeService:         feeService,
		scheduledService:   scheduled,
		logger:             config.GetLoggerFor("Lock Event Service"),
	}
}

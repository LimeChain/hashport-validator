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

package price

import (
	qi "github.com/limechain/hedera-eth-bridge-validator/app/domain/queue"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
	"time"
)

type Watcher struct {
	pricingService service.Pricing
	logger         *log.Entry
}

func NewWatcher(pricingService service.Pricing) *Watcher {
	return &Watcher{
		pricingService: pricingService,
		logger:         config.GetLoggerFor("Price Watcher"),
	}
}

func (pw *Watcher) Watch(q qi.Queue) {
	// there will be no handler, so the q is to implement the interface
	go pw.beginWatching()
}

func (pw *Watcher) beginWatching() {
	err := pw.pricingService.FetchAndUpdateUsdPrices(false)
	if err != nil {
		pw.logger.Errorf(err.Error())
	}

	time.Sleep(10 * time.Minute)
}

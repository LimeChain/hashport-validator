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

package scheduler

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

// Scheduler implements the required scheduling logic for submitting Ethereum transactions using a slot-based algorithm
type Scheduler struct {
	logger          *log.Entry
	tasks           *sync.Map
	executionWindow int64
}

// Schedule - Schedules new Transaction for execution at the right leader elected slot
func (s *Scheduler) Schedule(id string, firstTimestamp, slot int64, task func()) error {
	// TODO in the future we should move the calculation of the slot in the scheduler for better cohesion
	et := s.computeExecutionTime(firstTimestamp, slot)

	executeIn := time.Until(et)
	timer := time.NewTimer(executeIn)

	storedValue, alreadyExisted := s.tasks.LoadOrStore(id, &Storage{
		Executed: false,
		Timer:    timer,
	})

	if alreadyExisted {
		s.logger.Infof("Job for TX [%s] already scheduled for execution/executed.", id)
		return nil
	}

	go func() {
		<-timer.C
		storedValue.(*Storage).Executed = true
		task()
		s.logger.Infof("Execution for TX [%s] completed", id)
	}()

	s.logger.Infof("Scheduled new Job for TX [%s] for execution in [%s]", id, executeIn)

	return nil
}

// Cancel - Removes and cancels an already scheduled Transaction
func (s *Scheduler) Cancel(id string) {
	t, exists := s.tasks.Load(id)
	if !exists {
		s.logger.Warnf("Scheduled transaction execution for [%s] not found.", id)
	}

	storage := t.(*Storage)

	if !storage.Executed {
		storage.Timer.Stop()
		s.logger.Infof("Cancelled scheduled execution for TX [%s].", id)
	} else {
		s.logger.Infof("TX [%s] was already broadcast/executed.", id)
	}
}

// NewScheduler - Creates new instance of Scheduler
func NewScheduler(executionWindow int64) *Scheduler {
	return &Scheduler{
		logger:          config.GetLoggerFor("Scheduler"),
		tasks:           new(sync.Map),
		executionWindow: executionWindow,
	}
}

// computeExecutionTime - computes the time at which the TX must be executed based on first message timestamp and slot provided
func (s *Scheduler) computeExecutionTime(firstTimestamp, slot int64) time.Time {
	executionTimeNanos := firstTimestamp + timestamp.ToNanos(slot*s.executionWindow)
	return time.Unix(0, executionTimeNanos)
}

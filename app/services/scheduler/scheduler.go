package scheduler

import (
	"errors"
	"fmt"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Scheduler struct {
	logger          *log.Entry
	tasks           *sync.Map
	operator        string
	executionWindow int64
}

// Schedule - Schedules new Transaction for execution at the right leader elected slot
func (s *Scheduler) Schedule(id string, messages []message.TransactionMessage) error {
	_, exists := s.tasks.Load(id)
	if exists {
		return errors.New(fmt.Sprintf("Transaction with ID [%s] already scheduled for execution", id))
	}

	et, err := s.computeExecutionTime(messages)
	if err != nil {
		return err
	}

	executeIn := time.Until(et)
	timer := time.NewTimer(executeIn)
	s.tasks.Store(id, timer)
	go func() {
		<-timer.C

		// TODO Submit ETH TX

		s.tasks.Delete(id)
		s.logger.Infof("Executed Scheduled TX [%s]", id)
	}()

	s.logger.Infof("Scheduled new TX with ID [%s] for execution in [%s]", id, executeIn)

	return nil
}

// Cancel - Removes and cancels an already scheduled Transaction
func (s *Scheduler) Cancel(id string) error {
	t, exists := s.tasks.Load(id)
	if !exists {
		return errors.New("transaction not found")
	}
	s.tasks.Delete(id)

	timer := t.(*time.Timer)
	timer.Stop()

	s.logger.Infof("Cancelled scheduled execution for TX [%s]", id)
	return nil
}

// NewScheduler - Creates new instance of Scheduler
func NewScheduler(operator string, executionWindow int64) *Scheduler {
	return &Scheduler{
		logger:          config.GetLoggerFor("Scheduler"),
		tasks:           new(sync.Map),
		operator:        operator,
		executionWindow: executionWindow,
	}
}

// computeExecutionTime - computes the time at which the TX must be executed based on the first signature and the current validator
// Important! Transaction messages ARE expected to be sorted by ascending Timestamp
func (s *Scheduler) computeExecutionTime(messages []message.TransactionMessage) (time.Time, error) {
	slot, err := s.computeExecutionSlot(messages)
	if err != nil {
		return time.Unix(0, 0), err
	}

	firstSignatureTimestamp := messages[0].TransactionTimestamp
	executionTimeNanos := firstSignatureTimestamp + timestamp.ToNanos(int64(slot)*s.executionWindow)

	return time.Unix(0, executionTimeNanos), nil
}

func (s *Scheduler) computeExecutionSlot(messages []message.TransactionMessage) (int, error) {
	for i := 0; i < len(messages); i++ {
		if messages[i].SignerAddress == s.operator {
			return i, nil
		}
	}

	return -1, errors.New(fmt.Sprintf("Operator is not amongst the potential leaders - [%v]", s.operator))
}

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

package service

import (
	"errors"
	"fmt"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/retry"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	sleepPeriod = 5 * time.Second
)

var (
	timeoutError = errors.New(fmt.Sprintf("Timeout after [%d]", sleepPeriod))
)

// timeout is a function that returns an error after sleepPeriod.
func timeout() <-chan retry.Result {
	r := make(chan retry.Result)

	go func() {
		defer close(r)

		time.Sleep(sleepPeriod)
		r <- retry.Result{
			Value: nil,
			Error: timeoutError,
		}
	}()

	return r
}

// Retry executes two functions in race condition ({@param executionFunction} and timeout function).
// It takes the first result from both functions.
// If timeout function finishes first, it will retry the same mechanism {@param retries} times.
// If {@param executionFunction} finishes first, it will directly resolve its result.
// This function finds usability in the execution of EVM queries, which from time to time do not return response -
// the query is stuck forever and breaks the business logic. This way, if the query takes more than sleepPeriod, it will
// retry the query {@param retries} times.
// If {@param retries} is reached, it will return an error.
func Retry(executionFunction func() <-chan retry.Result, retries int) (interface{}, error) {
	times := 0
	var retryFunction func() (interface{}, error)

	retryFunction = func() (interface{}, error) {
		var executionResult retry.Result
		select {
		case executionResult = <-timeout():
		case executionResult = <-executionFunction():
		}

		if executionResult.Error != nil {
			if errors.Is(executionResult.Error, timeoutError) {
				times++
				if times >= retries {
					log.Warnf("Function execution timeouted. [%d/%d] tries.", times, retries)
					return 0, errors.New("too many retries")
				}

				log.Warnf("Function execution timeout. [%d/%d] tries.", times, retries)
				return retryFunction()
			}
			return nil, executionResult.Error
		}

		return executionResult.Value, executionResult.Error
	}

	return retryFunction()
}

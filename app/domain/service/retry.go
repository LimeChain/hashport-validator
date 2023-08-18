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
	"context"
	"errors"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/retry"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	sleepPeriod = 5 * time.Second
)

// Retry executes the given function with a timeout of {@param sleepPeriod}.
// If the function timeouts, it will retry the execution until the given {@param retries} is reached.
// If the function returns an error, this will return the error.
// If the function is executed successfully, this will return the result.
// This function finds usability in the execution of EVM queries, which from time to time do not return response -
// the query is stuck forever and breaks the business logic. This way, if the query takes more than sleepPeriod, it will
// retry the query {@param retries} times.
// If {@param retries} is reached, it will return an error.
func Retry(executionFunction func(context.Context) retry.Result, retries int) (interface{}, error) {
	times := 0

	for {
		ctx, cancel := context.WithTimeout(context.Background(), sleepPeriod)
		executionResult := executionFunction(ctx)
		cancel()

		if executionResult.Error != nil {
			if errors.Is(executionResult.Error, context.DeadlineExceeded) {
				times++
				if times >= retries {
					log.Warnf("Function execution timeouted. [%d/%d] tries.", times, retries)
					return nil, ErrTooManyRetires
				}

				log.Warnf("Function execution timeout. [%d/%d] tries.", times, retries)
				continue
			}
			return nil, executionResult.Error
		}

		return executionResult.Value, executionResult.Error
	}
}

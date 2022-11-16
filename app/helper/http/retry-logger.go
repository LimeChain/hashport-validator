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

package http

import (
	"fmt"
	log "github.com/sirupsen/logrus"
)

// RetryLogger This is a wrapper of logrus for `go-retryablehttp`
type RetryLogger struct {
	logger *log.Entry
}

func NewRetryLogger(logger *log.Entry) *RetryLogger {
	return &RetryLogger{logger}
}

func argsToFields(keysAndValues ...interface{}) log.Fields {
	fields := make(map[string]interface{})
	for i := 0; i < len(keysAndValues); i += 2 {
		fields[fmt.Sprint(keysAndValues[i])] = keysAndValues[i+1]
	}
	return fields
}

func (l *RetryLogger) Error(msg string, keysAndValues ...interface{}) {
	l.logger.WithFields(argsToFields(keysAndValues...)).Error(msg)
}

func (l *RetryLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.WithFields(argsToFields(keysAndValues...)).Info(msg)
}

// Debug This redirects the Debug messages to Trace level as the library is displaying way too many messages
func (l *RetryLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.logger.WithFields(argsToFields(keysAndValues...)).Trace(msg)
}

func (l *RetryLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.logger.WithFields(argsToFields(keysAndValues...)).Warn(msg)
}

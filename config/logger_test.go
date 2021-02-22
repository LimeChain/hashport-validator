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

package config

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func Test_GetLoggerFor(t *testing.T) {
	ctx := "testContext"
	logEntry := GetLoggerFor(ctx)

	if logEntry.Data["context"] != ctx {
		t.Fatalf(`Expected to return logger with context: [%s]`, ctx)
	}
}

func Test_InitLogger(t *testing.T) {
	debugMode := false
	debugLevel := InitLogger(&debugMode)
	if *debugLevel != log.InfoLevel {
		t.Fatalf(`Expected to return log level with context: [%s]`, log.InfoLevel)
	}

	debugMode = true
	debugLevel = InitLogger(&debugMode)
	if *debugLevel != log.DebugLevel {
		t.Fatalf(`Expected to return log level with context: [%s]`, log.DebugLevel)
	}
}

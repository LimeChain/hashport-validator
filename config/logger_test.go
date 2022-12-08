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

package config

import (
	log "github.com/sirupsen/logrus"
	"testing"
)

func Test_GetLoggerFor(t *testing.T) {
	ctx := "testContext"
	logEntry := GetLoggerFor(ctx)

	if logEntry.Data["context"] != ctx {
		t.Fatalf(`Expected to return logger with context: [%s]`, ctx)
	}
}

func Test_LevelsWorkCorrectly(t *testing.T) {
	ctx := "testContext"
	logEntry := GetLoggerFor(ctx)

	InitLogger("trace", "gcp")
	if logEntry.Logger.Level != log.TraceLevel {
		t.Fatalf(`Expected to logger level to be [%s], but got [%s]`, log.TraceLevel, logEntry.Level)
	}

	InitLogger("debug", "gcp")
	if logEntry.Logger.Level != log.DebugLevel {
		t.Fatalf(`Expected to logger level to be [%s], but got [%s]`, log.DebugLevel, logEntry.Level)
	}

	InitLogger("info", "gcp")
	if logEntry.Logger.Level != log.InfoLevel {
		t.Fatalf(`Expected to logger level to be [%s], but got [%s]`, log.InfoLevel, logEntry.Level)
	}

	InitLogger("trace", "")
	if logEntry.Logger.Level != log.TraceLevel {
		t.Fatalf(`Expected to logger level to be [%s], but got [%s]`, log.TraceLevel, logEntry.Level)
	}

	InitLogger("debug", "")
	if logEntry.Logger.Level != log.DebugLevel {
		t.Fatalf(`Expected to logger level to be [%s], but got [%s]`, log.DebugLevel, logEntry.Level)
	}

	InitLogger("info", "")
	if logEntry.Logger.Level != log.InfoLevel {
		t.Fatalf(`Expected to logger level to be [%s], but got [%s]`, log.InfoLevel, logEntry.Level)
	}
}

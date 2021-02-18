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
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

// GetLoggerFor returns a logger defined with a context
func GetLoggerFor(ctx string) *log.Entry {
	return log.WithField("context", ctx)
}

// InitLogger sets the initial configuration of the used logger
func InitLogger(debugMode *bool) *log.Level {
	log.SetOutput(os.Stdout)

	if *debugMode == true {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339Nano,
	})

	debugLevel := log.GetLevel()

	return &debugLevel
}

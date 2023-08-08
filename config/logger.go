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
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// GetLoggerFor returns a logger defined with a context
func GetLoggerFor(ctx string) *log.Entry {
	return log.WithField("context", ctx)
}

// InitLogger sets the initial configuration of the used logger
func InitLogger(level string, format string) {

	switch strings.ToLower(format) {
	case "gcp":
		log.SetFormatter(NewGCEFormatter(true))
	case "default", "":
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339Nano,
		})
	default:
		log.Fatalf("Unsupported log format: %s", format)
	}

	log.SetOutput(os.Stdout)
	log.SetReportCaller(true)

	switch strings.ToLower(level) {
	case "trace":
		log.SetLevel(log.TraceLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info", "":
		log.SetLevel(log.InfoLevel)
	default:
		log.Fatalf("Unsupported log level: %s", level)
	}

	log.Infof("Configured Log Level [%s]", log.GetLevel())
}

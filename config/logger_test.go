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

package config

import (
	"testing"
)

func Test_GetLoggerFor(t *testing.T) {
	ctx := "testContext"
	logEntry := GetLoggerFor(ctx)

	if logEntry.Data["context"] != ctx {
		t.Fatalf(`Expected to return logger with context: [%s]`, ctx)
	}
}

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

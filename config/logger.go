package config

import log "github.com/sirupsen/logrus"

// GetLoggerFor returns a logger defined with a context
func GetLoggerFor(ctx string) *log.Entry {
	return log.WithField("context", ctx)
}

package config

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// GetLoggerFor returns a logger defined with a context
func GetLoggerFor(ctx string) *log.Entry {
	return log.WithField("context", ctx)
}

// InitLogger sets the initial configuration of the used logger
func InitLogger(debugMode *bool) {
	log.SetOutput(os.Stdout)

	if *debugMode == true {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2000-01-02T16:20:00.999999999Z",
	})
}

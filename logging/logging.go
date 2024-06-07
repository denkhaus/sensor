package logging

import (
	"os"

	"github.com/sirupsen/logrus"
)

var (
	logger *logrus.Logger
)

func init() {
	logger = &logrus.Logger{
		Out: os.Stdout,
		Formatter: &logrus.TextFormatter{
			DisableTimestamp: false,
			ForceColors:      true,
			FullTimestamp:    true,
		},
		Level: logrus.InfoLevel,
	}
}

func SwitchLogLevel(logLevel string) {
	switch logLevel {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.Fatalf("unknown log level: %s", logLevel)
	}
}

func Logger() *logrus.Logger {
	return logger
}

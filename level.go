package glogp

import (
	"strings"

	logrus "github.com/sirupsen/logrus"
)

type LogLevel uint8

const (
	LogLevelNone LogLevel = iota
	LogLevelError
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
)

func LogLevelValue(LevelName string) LogLevel {
	switch strings.ToLower(LevelName) {
	case "error":
		return LogLevelError
	case "warn":
		return LogLevelWarn
	case "warning":
		return LogLevelWarn
	case "info":
		return LogLevelInfo
	case "debug":
		return LogLevelDebug
	}
	return LogLevelNone
}

func SetLevel(level string) {
	Print.level = LogLevelValue(level)
}

func setLevel(logger *logrus.Logger, level LogLevel) {
	logger.SetReportCaller(false)
	switch level {
	case LogLevelDebug:
		logger.SetLevel(logrus.DebugLevel)
	case LogLevelInfo:
		logger.SetLevel(logrus.InfoLevel)
	case LogLevelWarn:
		logger.SetLevel(logrus.WarnLevel)
	case LogLevelError:
		logger.SetLevel(logrus.ErrorLevel)
	}
}

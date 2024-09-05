package logger

import (
	"github.com/charmbracelet/log"
	"os"
	"time"
)

var logger *log.Logger

func InitLog(appName string, level log.Level) {
	logger = log.New(os.Stderr)
	logger.SetLevel(level)
	logger.SetPrefix(appName)
	logger.SetReportTimestamp(true)
	logger.SetTimeFormat(time.DateTime)
}

func Fatal(format string, args ...interface{}) {
	if len(args) == 0 {
		logger.Fatal(format)
	} else {
		logger.Fatalf(format, args...)
	}
}

func Info(format string, args ...interface{}) {
	if len(args) == 0 {
		logger.Info(format)
	} else {
		logger.Infof(format, args...)
	}
}

func Debug(format string, args ...interface{}) {
	if len(args) == 0 {
		logger.Debug(format)
	} else {
		logger.Debugf(format, args...)
	}
}

func Error(format string, args ...interface{}) {
	if len(args) == 0 {
		logger.Error(format)
	} else {
		logger.Errorf(format, args...)
	}
}

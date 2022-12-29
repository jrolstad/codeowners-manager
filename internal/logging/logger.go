package logging

import (
	"go.uber.org/zap"
	"sync"
)

func LogInfo(message string, keysAndValues ...interface{}) {
	logger := getLogger()
	logger.Infow(message, keysAndValues...)
}

func LogError(err error, keysAndValues ...interface{}) {
	logger := getLogger()
	logger.Errorw(err.Error(), keysAndValues...)
}

func LogPanic(err error, keysAndValues ...interface{}) {
	logger := getLogger()
	logger.Panicw(err.Error(), keysAndValues...)
}

var logger *zap.SugaredLogger
var locker = &sync.Mutex{}

func getLogger() *zap.SugaredLogger {
	if logger == nil {
		locker.Lock()

		prodLogger, _ := zap.NewProduction()
		logger = prodLogger.Sugar()
		
		locker.Unlock()
	}

	return logger
}

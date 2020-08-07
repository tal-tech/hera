package bootstrap

import (
	logger "github.com/tal-tech/loggerX"
)

type AfterServerStopFunc func()

func CloseLogger() AfterServerStopFunc {
	return func() {
		logger.Close()
	}
}

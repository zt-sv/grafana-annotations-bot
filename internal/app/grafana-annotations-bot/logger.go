package app

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"os"
)

// GetLogger : create application logger
func GetLogger(config Configuration) log.Logger {
	levelFilter := map[string]level.Option{
		levelError: level.AllowError(),
		levelWarn:  level.AllowWarn(),
		levelInfo:  level.AllowInfo(),
		levelDebug: level.AllowDebug(),
	}

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))

	if config.LogJSON {
		logger = log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	}

	logger = level.NewFilter(logger, levelFilter[config.LogLevel])
	logger = log.With(logger,
		"ts", log.DefaultTimestampUTC,
		"caller", log.DefaultCaller,
	)

	return logger
}

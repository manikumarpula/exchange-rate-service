package utils

import (
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

// NewLogger creates a new Go-Kit logger
func NewLogger() log.Logger {
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = level.NewFilter(logger, level.AllowInfo())
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}
	return logger
}
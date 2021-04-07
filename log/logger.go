package log

import (
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

// New returns the same logger all the time
func New() *zap.SugaredLogger {
	if logger != nil {
		return logger
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.Encoding = "json"
	cfg.OutputPaths = []string{"stdout", "log/upchain.log"}
	base, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	logger = base.Sugar()
	return logger
}

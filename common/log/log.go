package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	messageKey = "log"
)

// Init initialize the logger, given production flag and set global logger
func Init(prod bool) {
	zap.ReplaceGlobals(newLogger(prod))
}

func newLogger(prod bool) *zap.Logger {
	var logger *zap.Logger
	if prod {
		cfg := zap.NewProductionConfig()
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		cfg.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
		cfg.EncoderConfig.MessageKey = messageKey
		cfg.DisableCaller = true
		cfg.DisableStacktrace = true

		logger, _ = cfg.Build()
		logger.Info("Running in the production mode")
	} else {
		cfg := zap.NewDevelopmentConfig()
		cfg.DisableCaller = true
		cfg.DisableStacktrace = true
		cfg.EncoderConfig.MessageKey = messageKey

		logger, _ = cfg.Build()
		logger.Info("Running in the development mode")
	}
	return logger
}

func L() *zap.Logger {
	return zap.L()
}

func S() *zap.SugaredLogger {
	return zap.S()
}

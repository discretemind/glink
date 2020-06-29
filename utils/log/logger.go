package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var instance *zap.Logger

func Init(isProductionMode bool, fields ...zap.Field) *zap.Logger {
	instance = newLogger(isProductionMode).With(fields...)
	return instance
}

func Get() (res *zap.Logger) {
	if instance != nil {
		return instance
	}
	return newLogger(true)
}

func Debug(msg string, fields ...zap.Field) {
	Get().Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	Get().Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Get().Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Get().Error(msg, fields...)
}

// Info lo
func newLogger(isProductionMode bool) (res *zap.Logger) {
	cfg := zap.Config{}
	if isProductionMode {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	res, _ = cfg.Build()
	return res
}

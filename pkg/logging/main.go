package logging

import (
	"log"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func initLogger() {
	mode := viper.GetString("general_config.mode")
	developmentConfig := zap.NewDevelopmentConfig()
	developmentConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	developmentConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	developmentConfig.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	developmentConfig.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	developmentConfig.EncoderConfig.EncodeName = zapcore.FullNameEncoder
	developmentConfig.EncoderConfig.LineEnding = zapcore.DefaultLineEnding

	productionConfig := zap.NewProductionConfig()
	productionConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	productionConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	productionConfig.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	productionConfig.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	productionConfig.EncoderConfig.EncodeName = zapcore.FullNameEncoder
	productionConfig.EncoderConfig.LineEnding = zapcore.DefaultLineEnding
	productionConfig.EncoderConfig.StacktraceKey = "stacktrace"
	productionConfig.EncoderConfig.MessageKey = "msg"
	productionConfig.EncoderConfig.NameKey = "logger"
	productionConfig.EncoderConfig.CallerKey = "caller"
	productionConfig.EncoderConfig.TimeKey = "ts"
	productionConfig.EncoderConfig.LevelKey = "level"
	var cfg zap.Config
	if mode == "development" {
		cfg = developmentConfig
	} else {
		cfg = productionConfig
	}
	var err error
	logger, err = cfg.Build()
	if err != nil {
		log.Fatal("failed to build logger: ", err)
	}
	defer func() { _ = logger.Sync() }()
	logger.Info("logger initialized")
}

func GetSugaredLogger() *zap.SugaredLogger {
	if logger == nil {
		initLogger()
	}
	return logger.Sugar()
}

package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewZapConf(env string) zap.Config {
	var conf zap.Config
	switch env {
	case "development", "dev":
		conf = zap.NewDevelopmentConfig()
	case "production", "prod":
		fallthrough
	default:
		conf = zap.NewProductionConfig()
	}

	conf.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	return conf
}

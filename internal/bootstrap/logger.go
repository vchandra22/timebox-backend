package bootstrap

import "go.uber.org/zap"

func LoggerInit() *zap.Logger {
	zap, _ := zap.NewProduction()
	return zap
}

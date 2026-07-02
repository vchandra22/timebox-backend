package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ZapLogger(log *zap.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		path := ctx.Request.URL.Path
		method := ctx.Request.Method
		query := ctx.Request.URL.RawQuery
		clientIP := ctx.ClientIP()

		ctx.Next()

		latency := time.Since(start)
		statusCode := ctx.Writer.Status()

		log.Info("HTTP Request",
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("client_ip", clientIP),
			zap.Int("status_code", statusCode),
			zap.Duration("latency", latency),
		)
	}
}

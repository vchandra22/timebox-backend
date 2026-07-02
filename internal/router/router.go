package router

import (
	"boilerplate-golang/internal/handler"
	"boilerplate-golang/internal/middleware"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func NewRouter(handlers *handler.Handler, log *zap.Logger, allowedOrigins []string, ginMode string) *gin.Engine {
	gin.SetMode(ginMode)

	r := gin.New()

	r.Use(
		middleware.ZapLogger(log),
		cors.New(cors.Config{
			AllowOrigins:     allowedOrigins,
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}),
		gin.Recovery(),
	)

	registerRoutes(r, handlers)

	return r
}

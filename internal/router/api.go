package router

import (
	"boilerplate-golang/internal/handler"

	"github.com/gin-gonic/gin"
)

func registerRoutes(r *gin.Engine, handlers *handler.Handler) {
	api := r.Group("/api/v1")

	handlers.Health.RegisterRoutes(api)
	handlers.User.RegisterRoutes(api)
}

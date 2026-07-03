package router

import (
	"timebox-backend/internal/handler"

	"github.com/gin-gonic/gin"
)

func registerRoutes(r *gin.Engine, handlers *handler.Handler) {
	api := r.Group("/api/v1")

	handlers.Auth.RegisterRoutes(api)
	handlers.Collaboration.RegisterRoutes(api)
	handlers.Execution.RegisterRoutes(api)
	handlers.Health.RegisterRoutes(api)
	handlers.Planning.RegisterRoutes(api)
	handlers.Workspace.RegisterRoutes(api)
}

package handler

import (
	"net/http"
	"timebox-backend/internal/response"
	"timebox-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	healthService *service.HealthService
}

func newHealthHandler(services *service.Service) *HealthHandler {
	return &HealthHandler{
		healthService: services.Health,
	}
}

func (h *HealthHandler) RegisterRoutes(routeGroup *gin.RouterGroup) {
	r := routeGroup.Group("/health")

	r.GET("/", h.Ready)
	r.GET("/ready", h.Ready)
	r.GET("/live", h.Live)
}

func (h *HealthHandler) Live(ctx *gin.Context) {
	response.WithoutData(ctx, "OK", http.StatusOK)
}

func (h *HealthHandler) Ready(ctx *gin.Context) {
	res, err := h.healthService.Get(ctx)
	if err != nil {
		response.Error(ctx, "internal server error", "health check failed", http.StatusInternalServerError)
		return
	}

	response.WithoutData(ctx, res, http.StatusOK)
}

package config

import (
	"general-service/internal/dto"
	"general-service/internal/handlers"

	"github.com/gin-gonic/gin"
)

// CheckHealth godoc
// @Summary Check service health
// @Description Returns pong and service status
// @Tags health
// @Produce json
// @Success 200 {object} dto.HealthResponse
// @Router /api/v1/ping [get]
func CheckHealth(c *gin.Context) {
	c.JSON(200, dto.HealthResponse{
		Message: "pong",
		Status:  "healthy",
	})
}

func SetupHealthRoutes(router *gin.RouterGroup) {
	router.GET("/ping", CheckHealth)
}

func SetupAPIRoutes(router *gin.Engine, h *handlers.Handlers) {
	v1 := router.Group("/api/v1")
	{
		SetupHealthRoutes(v1)
		// Add other routes here later
	}
}

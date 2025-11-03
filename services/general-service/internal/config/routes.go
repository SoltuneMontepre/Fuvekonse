package config

import (
	"general-service/internal/handlers"

	"github.com/gin-gonic/gin"
)

// CheckHealth godoc
// @Summary Check service health
// @Description Returns healthy if service is running
// @Tags Health
// @Produce json
// @Success 200 {object}
// @Router /ping [get]
func CheckHealth(c *gin.Context) {
	c.JSON(200, gin.H{"message": "pong", "status": "healthy"})
}

func SetupHealthRoutes(router *gin.RouterGroup) {
	router.GET("/ping", CheckHealth)
}

func SetupAPIRoutes(router *gin.Engine, h *handlers.Handlers) {
	v1 := router.Group("/api/v1")
	{
		SetupHealthRoutes(v1)
		// Thêm các routes khác vào đây sau
	}
}

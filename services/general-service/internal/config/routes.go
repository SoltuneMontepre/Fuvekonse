package config

import (
	"general-service/internal/handlers"

	"github.com/gin-gonic/gin"
)

func SetupHealthRoutes(router *gin.Engine) {
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong", "status": "healthy"})
	})
}

func SetupAPIRoutes(router *gin.Engine, h *handlers.Handlers) {
	SetupHealthRoutes(router)
}

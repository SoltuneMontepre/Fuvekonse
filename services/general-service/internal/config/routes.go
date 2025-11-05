package config

import (
	"general-service/internal/dto/common"
	"general-service/internal/handlers"

	"github.com/gin-gonic/gin"
)

// CheckHealth godoc
//
//	@Summary		Check service health
//	@Description	Returns pong and service status
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	common.HealthResponse
//	@Router			/ping [get]
func CheckHealth(c *gin.Context) {
	healthData := common.HealthResponse{
		Message: "pong",
		Status:  "healthy",
	}
	c.JSON(200, common.SuccessResponse(&healthData, "Service is healthy", 200))
}

func SetupHealthRoutes(router *gin.RouterGroup) {
	router.GET("/ping", CheckHealth)
}

func SetupAuthRoutes(router *gin.RouterGroup, h *handlers.Handlers) {
	auth := router.Group("/auth")
	{
		auth.POST("/login", h.Auth.Login)
		auth.POST("/reset-password", h.Auth.ResetPassword)
	}
}

func SetupAPIRoutes(router *gin.Engine, h *handlers.Handlers) {
	v1 := router.Group("/api/v1")
	{
		SetupHealthRoutes(v1)
		SetupAuthRoutes(v1, h)
	}
}

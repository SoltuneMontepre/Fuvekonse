package config

import (
	"general-service/internal/dto/common"
	"general-service/internal/handlers"
	"general-service/internal/middlewares"

	"github.com/gin-gonic/gin"
)

// CheckHealth godoc
// @Summary Check service health
// @Description Returns pong and service status
// @Tags health
// @Produce json
// @Success 200 {object} common.HealthResponse
// @Router /ping [get]
func CheckHealth(c *gin.Context) {
	healthData := common.HealthResponse{
		Message: "pong",
		Status:  "healthy",
	}
	c.JSON(200, common.SuccessResponse(&healthData, "Service is healthy", 200))
}

func SetupAuthRoutes(router *gin.RouterGroup, h *handlers.Handlers) {
	auth := router.Group("/auth")
	{
		auth.POST("/login", h.Auth.Login)
	}
}

func SetupAPIRoutes(router *gin.Engine, h *handlers.Handlers) {
	v1 := router.Group("/api/v1")
	{
		v1.GET("/ping", CheckHealth)
		SetupAuthRoutes(v1, h)

		// Protected routes - require JWT authentication
		protected := v1.Group("")
		protected.Use(middlewares.JWTAuthMiddleware())
		{
			// User routes
			users := protected.Group("/users")
			{
				users.GET("/me", h.User.GetMe)
			}
		}

		// Admin only routes - require JWT authentication and admin role
		// admin := v1.Group("")
		// admin.Use(middlewares.JWTAuthMiddleware())
		// admin.Use(middlewares.RequireRole("admin", "superadmin"))
		// {
		// 	// Example: Add your admin-only routes here
		// 	// admin.GET("/users", h.User.GetAllUsers)
		// 	// admin.DELETE("/users/:id", h.User.DeleteUser)
		// }
	}
}

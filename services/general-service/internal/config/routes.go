package config

import (
	"context"
	"general-service/internal/dto/common"
	"general-service/internal/handlers"
	"general-service/internal/middlewares"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

func SetupAuthRoutes(router *gin.RouterGroup, h *handlers.Handlers) {
	auth := router.Group("/auth")
	{
		auth.POST("/login", h.Auth.Login)

		//add jwt auth
		auth.POST("/reset-password", middlewares.JWTAuthMiddleware(), h.Auth.ResetPassword)
	}
}


func SetupAPIRoutes(router *gin.Engine, h *handlers.Handlers, db *gorm.DB, redisSetFunc func(ctx context.Context, key string, value interface{}, expiration time.Duration) error) {
	router.GET("/health/db", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(500, gin.H{"error": "Database connection error"})
			return
		}
		if err := sqlDB.Ping(); err != nil {
			c.JSON(500, gin.H{"error": "Database ping failed"})
			return
		}
		c.JSON(200, gin.H{"status": "database healthy"})
	})

	router.GET("/health/redis", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := redisSetFunc(ctx, "health_check", "ok", time.Minute); err != nil {
			c.JSON(500, gin.H{"error": "Redis connection failed"})
			return
		}
		c.JSON(200, gin.H{"status": "redis healthy"})
	})

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

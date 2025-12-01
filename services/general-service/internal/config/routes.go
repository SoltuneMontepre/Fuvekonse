package config

import (
	"context"
	role "general-service/internal/common/constants"
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
		auth.POST("/register", h.Auth.Register)
		auth.POST("/login", h.Auth.Login)
		auth.POST("/logout", middlewares.JWTAuthMiddleware(), h.Auth.Logout)

		//add jwt auth
		auth.POST("/reset-password", middlewares.JWTAuthMiddleware(), h.Auth.ResetPassword)
		auth.POST("/verify-otp", h.Auth.VerifyOtp)

		auth.POST("/forgot-password", h.Auth.ForgotPassword)
		auth.POST("/reset-password/confirm", h.Auth.ResetPasswordConfirm)
	}
}

func SetupAPIRoutes(router gin.IRouter, h *handlers.Handlers, db *gorm.DB, redisSetFunc func(ctx context.Context, key string, value interface{}, expiration time.Duration) error) {
	// Root endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "General Service API",
			"version": "1.0",
			"status":  "running",
		})
	})

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

	v1 := router.Group("/v1")
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
				users.PUT("/me", h.User.UpdateProfile)
				users.PATCH("/me/avatar", h.User.UpdateAvatar)
			}
		}

		// Admin only routes - require JWT authentication and admin role
		admin := v1.Group("/admin")
		admin.Use(middlewares.JWTAuthMiddleware())
		admin.Use(middlewares.RequireRole(role.RoleAdmin))
		{
			// Admin user management routes
			adminUsers := admin.Group("/users")
			{
				adminUsers.GET("", h.User.GetAllUsers)
				adminUsers.GET("/:id", h.User.GetUserByIDForAdmin)
				adminUsers.PUT("/:id", h.User.UpdateUserByAdmin)
				adminUsers.DELETE("/:id", h.User.DeleteUser)
				adminUsers.PATCH("/:id/verify", h.User.VerifyUser)
			}
		}
	}
}

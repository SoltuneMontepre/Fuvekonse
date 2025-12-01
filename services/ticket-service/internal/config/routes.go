package config

import (
	"context"
	"ticket-service/internal/handlers"
	"ticket-service/internal/middlewares"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoleRoutes(rg *gin.RouterGroup, h *handlers.RoleHandler) {
	roles := rg.Group("/roles")
	{
		roles.POST("", h.CreateRole)
		roles.GET("", h.GetRoles)
		roles.GET("/:id", h.GetRole)
		roles.PUT("/:id", h.UpdateRole)
		roles.DELETE("/:id", h.DeleteRole)
		roles.GET("/:id/permissions", h.GetRoleWithPermissions)
		roles.POST("/:id/permissions", h.AddPermissionToRole)
		roles.DELETE("/:id/permissions/:permission_id", h.RemovePermissionFromRole)
	}
}

func SetupPermissionRoutes(rg *gin.RouterGroup, h *handlers.PermissionHandler) {
	permissions := rg.Group("/permissions")
	{
		permissions.POST("", h.CreatePermission)
		permissions.GET("", h.GetPermissions)
		permissions.GET("/:id", h.GetPermission)
		permissions.PUT("/:id", h.UpdatePermission)
		permissions.DELETE("/:id", h.DeletePermission)
		permissions.GET("/:id/roles", h.GetPermissionWithRoles)
	}
}

func SetupUserBanRoutes(rg *gin.RouterGroup, h *handlers.UserBanHandler) {
	users := rg.Group("/users")
	{
		users.POST("/:user_id/bans", h.BanUser)
		users.GET("/:user_id/bans", h.GetUserBans)
		users.DELETE("/:user_id/bans/:permission_id", h.UnbanUser)
		users.GET("/:user_id/bans/check", h.CheckUserBan)
	}

	bans := rg.Group("/bans")
	{
		bans.GET("", h.GetAllUserBans)
		bans.GET("/:id", h.GetUserBan)
		bans.PUT("/:id", h.UpdateBanReason)
	}
}

func SetupAPIRoutes(router gin.IRouter, h *handlers.Handlers, db *gorm.DB, redisSetFunc func(ctx context.Context, key string, value interface{}, expiration time.Duration) error) {
	// Health endpoints
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong", "status": "healthy"})
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

	api := router.Group("/v1")
	{
		SetupRoleRoutes(api, h.Role)
		SetupPermissionRoutes(api, h.Permission)
		SetupUserBanRoutes(api, h.UserBan)
		SetupTicketRoutes(api, h.Ticket)
		SetupPaymentRoutes(api, h.Payment)
	}
}

func SetupTicketRoutes(rg *gin.RouterGroup, h *handlers.TicketHandler) {
	tickets := rg.Group("/tickets")
	{
		tickets.GET("/tiers", h.GetTicketTiers)
		tickets.GET("/tiers/active", h.GetActiveTicketTiers)
		tickets.GET("/tiers/:id", h.GetTicketTier)
		tickets.GET("/:id", h.GetTicket)
		tickets.GET("/user/:user_id", h.GetUserTickets)
	}
}

func SetupPaymentRoutes(rg *gin.RouterGroup, h *handlers.PaymentHandler) {
	payments := rg.Group("/payments")
	{
		payments.POST("/payment-link", middlewares.JWTAuthMiddleware(), h.CreatePaymentLink)
		payments.POST("/webhook", h.HandleWebhook)
		payments.POST("/cleanup-stuck", h.CleanupStuckPayments)
		payments.GET("/status/:orderCode", h.GetPaymentStatus)
		payments.POST("/cancel-by-order", h.CancelPaymentByOrderCode)
	}
}

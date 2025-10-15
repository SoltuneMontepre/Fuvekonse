package config

import (
	"gin/internal/handlers"

	"github.com/gin-gonic/gin"
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

func SetupHealthRoutes(router *gin.Engine) {
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong", "status": "healthy"})
	})
}

func SetupAPIRoutes(router *gin.Engine, h *handlers.Handlers) {
	SetupHealthRoutes(router)

	api := router.Group("/api/v1")
	{
		SetupRoleRoutes(api, h.Role)
		SetupPermissionRoutes(api, h.Permission)
		SetupUserBanRoutes(api, h.UserBan)
	}
}
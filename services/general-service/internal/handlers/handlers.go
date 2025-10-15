package handlers

import (
	"gin/internal/services"
)

type Handlers struct {
	Role       *RoleHandler
	Permission *PermissionHandler
	UserBan    *UserBanHandler
}

func NewHandlers(services *services.Services) *Handlers {
	return &Handlers{
		Role:       NewRoleHandler(services.Role),
		Permission: NewPermissionHandler(services.Permission),
		UserBan:    NewUserBanHandler(services.UserBan),
	}
}
package services

import (
	"gin/internal/repositories"
)

type Services struct {
	Role       RoleServiceInterface
	Permission PermissionServiceInterface
	UserBan    UserBanServiceInterface
}

func NewServices(repos *repositories.Repositories) *Services {
	return &Services{
		Role:       NewRoleService(repos.Role, repos.Permission),
		Permission: NewPermissionService(repos.Permission),
		UserBan:    NewUserBanService(repos.UserBan, repos.Permission),
	}
}
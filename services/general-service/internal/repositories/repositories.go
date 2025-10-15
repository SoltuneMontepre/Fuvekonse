package repositories

import "gorm.io/gorm"

type Repositories struct {
	Role       RoleRepositoryInterface
	Permission PermissionRepositoryInterface
	UserBan    UserBanRepositoryInterface
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Role:       NewRoleRepository(db),
		Permission: NewPermissionRepository(db),
		UserBan:    NewUserBanRepository(db),
	}
}
package repositories

import (
	"gin/internal/models"

	"gorm.io/gorm"
)

type RoleRepositoryInterface interface {
	Create(role *models.Role) error
	GetByID(id uint) (*models.Role, error)
	GetByName(name string) (*models.Role, error)
	GetAll() ([]models.Role, error)
	Update(role *models.Role) error
	Delete(id uint) error
	GetWithPermissions(id uint) (*models.Role, error)
	AddPermission(roleID, permissionID uint) error
	RemovePermission(roleID, permissionID uint) error
}

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepositoryInterface {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) Create(role *models.Role) error {
	return r.db.Create(role).Error
}

func (r *RoleRepository) GetByID(id uint) (*models.Role, error) {
	var role models.Role
	err := r.db.First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) GetByName(name string) (*models.Role, error) {
	var role models.Role
	err := r.db.Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) GetAll() ([]models.Role, error) {
	var roles []models.Role
	err := r.db.Find(&roles).Error
	return roles, err
}

func (r *RoleRepository) Update(role *models.Role) error {
	return r.db.Save(role).Error
}

func (r *RoleRepository) Delete(id uint) error {
	return r.db.Delete(&models.Role{}, id).Error
}

func (r *RoleRepository) GetWithPermissions(id uint) (*models.Role, error) {
	var role models.Role
	err := r.db.Preload("Permissions").First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) AddPermission(roleID, permissionID uint) error {
	var role models.Role
	var permission models.Permission
	
	if err := r.db.First(&role, roleID).Error; err != nil {
		return err
	}
	
	if err := r.db.First(&permission, permissionID).Error; err != nil {
		return err
	}
	
	return r.db.Model(&role).Association("Permissions").Append(&permission)
}

func (r *RoleRepository) RemovePermission(roleID, permissionID uint) error {
	var role models.Role
	var permission models.Permission
	
	if err := r.db.First(&role, roleID).Error; err != nil {
		return err
	}
	
	if err := r.db.First(&permission, permissionID).Error; err != nil {
		return err
	}
	
	return r.db.Model(&role).Association("Permissions").Delete(&permission)
}
package repositories

import (
	"gin/internal/models"

	"gorm.io/gorm"
)

type PermissionRepositoryInterface interface {
	Create(permission *models.Permission) error
	GetByID(id uint) (*models.Permission, error)
	GetByName(name string) (*models.Permission, error)
	GetAll() ([]models.Permission, error)
	Update(permission *models.Permission) error
	Delete(id uint) error
	GetWithRoles(id uint) (*models.Permission, error)
}

type PermissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepositoryInterface {
	return &PermissionRepository{db: db}
}

func (p *PermissionRepository) Create(permission *models.Permission) error {
	return p.db.Create(permission).Error
}

func (p *PermissionRepository) GetByID(id uint) (*models.Permission, error) {
	var permission models.Permission
	err := p.db.First(&permission, id).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (p *PermissionRepository) GetByName(name string) (*models.Permission, error) {
	var permission models.Permission
	err := p.db.Where("name = ?", name).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (p *PermissionRepository) GetAll() ([]models.Permission, error) {
	var permissions []models.Permission
	err := p.db.Find(&permissions).Error
	return permissions, err
}

func (p *PermissionRepository) Update(permission *models.Permission) error {
	return p.db.Save(permission).Error
}

func (p *PermissionRepository) Delete(id uint) error {
	return p.db.Delete(&models.Permission{}, id).Error
}

func (p *PermissionRepository) GetWithRoles(id uint) (*models.Permission, error) {
	var permission models.Permission
	err := p.db.Preload("Roles").First(&permission, id).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}
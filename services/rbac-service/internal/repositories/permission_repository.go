package repositories

import (
	"context"
	"rbac-service/internal/models"

	"gorm.io/gorm"
)

type PermissionRepositoryInterface interface {
	Create(ctx context.Context, permission *models.Permission) error
	GetByID(ctx context.Context, id uint) (*models.Permission, error)
	GetByName(ctx context.Context, name string) (*models.Permission, error)
	GetAll(ctx context.Context) ([]models.Permission, error)
	Update(ctx context.Context, permission *models.Permission) error
	Delete(ctx context.Context, id uint) error
	GetWithRoles(ctx context.Context, id uint) (*models.Permission, error)
}

type PermissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepositoryInterface {
	return &PermissionRepository{db: db}
}

func (p *PermissionRepository) Create(ctx context.Context, permission *models.Permission) error {
	return p.db.WithContext(ctx).Create(permission).Error
}

func (p *PermissionRepository) GetByID(ctx context.Context, id uint) (*models.Permission, error) {
	var permission models.Permission
	err := p.db.WithContext(ctx).First(&permission, id).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (p *PermissionRepository) GetByName(ctx context.Context, name string) (*models.Permission, error) {
	var permission models.Permission
	err := p.db.WithContext(ctx).Where("name = ?", name).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (p *PermissionRepository) GetAll(ctx context.Context) ([]models.Permission, error) {
	var permissions []models.Permission
	err := p.db.WithContext(ctx).Find(&permissions).Error
	return permissions, err
}

func (p *PermissionRepository) Update(ctx context.Context, permission *models.Permission) error {
	return p.db.WithContext(ctx).Save(permission).Error
}

func (p *PermissionRepository) Delete(ctx context.Context, id uint) error {
	return p.db.WithContext(ctx).Delete(&models.Permission{}, id).Error
}

func (p *PermissionRepository) GetWithRoles(ctx context.Context, id uint) (*models.Permission, error) {
	var permission models.Permission
	err := p.db.WithContext(ctx).Preload("Roles").First(&permission, id).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

package services

import (
	"context"
	"fmt"
	"rbac-service/internal/models"
	"rbac-service/internal/repositories"
)

type PermissionServiceInterface interface {
	CreatePermission(ctx context.Context, name string) (*models.Permission, error)
	GetPermissionByID(ctx context.Context, id uint) (*models.Permission, error)
	GetPermissionByName(ctx context.Context, name string) (*models.Permission, error)
	GetAllPermissions(ctx context.Context) ([]models.Permission, error)
	UpdatePermission(ctx context.Context, permission *models.Permission) error
	DeletePermission(ctx context.Context, id uint) error
	GetPermissionWithRoles(ctx context.Context, id uint) (*models.Permission, error)
}

type PermissionService struct {
	permissionRepo repositories.PermissionRepositoryInterface
}

func NewPermissionService(permissionRepo repositories.PermissionRepositoryInterface) PermissionServiceInterface {
	return &PermissionService{
		permissionRepo: permissionRepo,
	}
}

func (s *PermissionService) CreatePermission(ctx context.Context, name string) (*models.Permission, error) {
	if name == "" {
		return nil, fmt.Errorf("permission name cannot be empty")
	}

	existingPermission, err := s.permissionRepo.GetByName(ctx, name)
	if err == nil && existingPermission != nil {
		return nil, fmt.Errorf("permission with name '%s' already exists", name)
	}

	permission := &models.Permission{
		Name: name,
	}

	if err := s.permissionRepo.Create(ctx, permission); err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	return permission, nil
}

func (s *PermissionService) GetPermissionByID(ctx context.Context, id uint) (*models.Permission, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid permission ID")
	}

	permission, err := s.permissionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("permission not found: %w", err)
	}

	return permission, nil
}

func (s *PermissionService) GetPermissionByName(ctx context.Context, name string) (*models.Permission, error) {
	if name == "" {
		return nil, fmt.Errorf("permission name cannot be empty")
	}

	permission, err := s.permissionRepo.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("permission not found: %w", err)
	}

	return permission, nil
}

func (s *PermissionService) GetAllPermissions(ctx context.Context) ([]models.Permission, error) {
	return s.permissionRepo.GetAll(ctx)
}

func (s *PermissionService) UpdatePermission(ctx context.Context, permission *models.Permission) error {
	if permission == nil {
		return fmt.Errorf("permission cannot be nil")
	}

	if permission.Name == "" {
		return fmt.Errorf("permission name cannot be empty")
	}

	existingPermission, err := s.permissionRepo.GetByID(ctx, permission.PermID)
	if err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	if permissionWithSameName, err := s.permissionRepo.GetByName(ctx, permission.Name); err == nil && permissionWithSameName.PermID != permission.PermID {
		return fmt.Errorf("permission with name '%s' already exists", permission.Name)
	}

	existingPermission.Name = permission.Name
	return s.permissionRepo.Update(ctx, existingPermission)
}

func (s *PermissionService) DeletePermission(ctx context.Context, id uint) error {
	if id == 0 {
		return fmt.Errorf("invalid permission ID")
	}

	_, err := s.permissionRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	return s.permissionRepo.Delete(ctx, id)
}

func (s *PermissionService) GetPermissionWithRoles(ctx context.Context, id uint) (*models.Permission, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid permission ID")
	}

	permission, err := s.permissionRepo.GetWithRoles(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("permission not found: %w", err)
	}

	return permission, nil
}

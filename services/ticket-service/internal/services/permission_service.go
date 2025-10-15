package services

import (
	"fmt"
	"gin/internal/models"
	"gin/internal/repositories"
)

type PermissionServiceInterface interface {
	CreatePermission(name string) (*models.Permission, error)
	GetPermissionByID(id uint) (*models.Permission, error)
	GetPermissionByName(name string) (*models.Permission, error)
	GetAllPermissions() ([]models.Permission, error)
	UpdatePermission(permission *models.Permission) error
	DeletePermission(id uint) error
	GetPermissionWithRoles(id uint) (*models.Permission, error)
}

type PermissionService struct {
	permissionRepo repositories.PermissionRepositoryInterface
}

func NewPermissionService(permissionRepo repositories.PermissionRepositoryInterface) PermissionServiceInterface {
	return &PermissionService{
		permissionRepo: permissionRepo,
	}
}

func (s *PermissionService) CreatePermission(name string) (*models.Permission, error) {
	if name == "" {
		return nil, fmt.Errorf("permission name cannot be empty")
	}

	existingPermission, err := s.permissionRepo.GetByName(name)
	if err == nil && existingPermission != nil {
		return nil, fmt.Errorf("permission with name '%s' already exists", name)
	}

	permission := &models.Permission{
		Name: name,
	}

	if err := s.permissionRepo.Create(permission); err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	return permission, nil
}

func (s *PermissionService) GetPermissionByID(id uint) (*models.Permission, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid permission ID")
	}

	permission, err := s.permissionRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("permission not found: %w", err)
	}

	return permission, nil
}

func (s *PermissionService) GetPermissionByName(name string) (*models.Permission, error) {
	if name == "" {
		return nil, fmt.Errorf("permission name cannot be empty")
	}

	permission, err := s.permissionRepo.GetByName(name)
	if err != nil {
		return nil, fmt.Errorf("permission not found: %w", err)
	}

	return permission, nil
}

func (s *PermissionService) GetAllPermissions() ([]models.Permission, error) {
	return s.permissionRepo.GetAll()
}

func (s *PermissionService) UpdatePermission(permission *models.Permission) error {
	if permission == nil {
		return fmt.Errorf("permission cannot be nil")
	}

	if permission.Name == "" {
		return fmt.Errorf("permission name cannot be empty")
	}

	existingPermission, err := s.permissionRepo.GetByID(permission.PermID)
	if err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	if permissionWithSameName, err := s.permissionRepo.GetByName(permission.Name); err == nil && permissionWithSameName.PermID != permission.PermID {
		return fmt.Errorf("permission with name '%s' already exists", permission.Name)
	}

	existingPermission.Name = permission.Name
	return s.permissionRepo.Update(existingPermission)
}

func (s *PermissionService) DeletePermission(id uint) error {
	if id == 0 {
		return fmt.Errorf("invalid permission ID")
	}

	_, err := s.permissionRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	return s.permissionRepo.Delete(id)
}

func (s *PermissionService) GetPermissionWithRoles(id uint) (*models.Permission, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid permission ID")
	}

	permission, err := s.permissionRepo.GetWithRoles(id)
	if err != nil {
		return nil, fmt.Errorf("permission not found: %w", err)
	}

	return permission, nil
}
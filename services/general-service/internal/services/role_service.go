package services

import (
	"fmt"
	"gin/internal/models"
	"gin/internal/repositories"
)

type RoleServiceInterface interface {
	CreateRole(name string) (*models.Role, error)
	GetRoleByID(id uint) (*models.Role, error)
	GetRoleByName(name string) (*models.Role, error)
	GetAllRoles() ([]models.Role, error)
	UpdateRole(role *models.Role) error
	DeleteRole(id uint) error
	GetRoleWithPermissions(id uint) (*models.Role, error)
	AddPermissionToRole(roleID, permissionID uint) error
	RemovePermissionFromRole(roleID, permissionID uint) error
}

type RoleService struct {
	roleRepo       repositories.RoleRepositoryInterface
	permissionRepo repositories.PermissionRepositoryInterface
}

func NewRoleService(roleRepo repositories.RoleRepositoryInterface, permissionRepo repositories.PermissionRepositoryInterface) RoleServiceInterface {
	return &RoleService{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
	}
}

func (s *RoleService) CreateRole(name string) (*models.Role, error) {
	if name == "" {
		return nil, fmt.Errorf("role name cannot be empty")
	}

	// Check if role already exists
	existingRole, err := s.roleRepo.GetByName(name)
	if err == nil && existingRole != nil {
		return nil, fmt.Errorf("role with name '%s' already exists", name)
	}

	role := &models.Role{
		Name: name,
	}

	if err := s.roleRepo.Create(role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return role, nil
}

func (s *RoleService) GetRoleByID(id uint) (*models.Role, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid role ID")
	}

	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	return role, nil
}

func (s *RoleService) GetRoleByName(name string) (*models.Role, error) {
	if name == "" {
		return nil, fmt.Errorf("role name cannot be empty")
	}

	role, err := s.roleRepo.GetByName(name)
	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	return role, nil
}

func (s *RoleService) GetAllRoles() ([]models.Role, error) {
	return s.roleRepo.GetAll()
}

func (s *RoleService) UpdateRole(role *models.Role) error {
	if role == nil {
		return fmt.Errorf("role cannot be nil")
	}

	if role.Name == "" {
		return fmt.Errorf("role name cannot be empty")
	}

	// Check if role exists
	existingRole, err := s.roleRepo.GetByID(role.RoleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Check if another role with the same name exists (excluding current role)
	if roleWithSameName, err := s.roleRepo.GetByName(role.Name); err == nil && roleWithSameName.RoleID != role.RoleID {
		return fmt.Errorf("role with name '%s' already exists", role.Name)
	}

	existingRole.Name = role.Name
	return s.roleRepo.Update(existingRole)
}

func (s *RoleService) DeleteRole(id uint) error {
	if id == 0 {
		return fmt.Errorf("invalid role ID")
	}

	_, err := s.roleRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	return s.roleRepo.Delete(id)
}

func (s *RoleService) GetRoleWithPermissions(id uint) (*models.Role, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid role ID")
	}

	role, err := s.roleRepo.GetWithPermissions(id)
	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	return role, nil
}

func (s *RoleService) AddPermissionToRole(roleID, permissionID uint) error {
	if roleID == 0 || permissionID == 0 {
		return fmt.Errorf("invalid role ID or permission ID")
	}

	_, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Verify permission exists
	_, err = s.permissionRepo.GetByID(permissionID)
	if err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	// Check if permission is already assigned to role
	roleWithPermissions, err := s.roleRepo.GetWithPermissions(roleID)
	if err != nil {
		return fmt.Errorf("failed to get role permissions: %w", err)
	}

	for _, permission := range roleWithPermissions.Permissions {
		if permission.PermID == permissionID {
			return fmt.Errorf("permission already assigned to role")
		}
	}

	// Add permission using GORM association
	// Since we removed AddPermission from repository, we'll use a different approach
	permission, _ := s.permissionRepo.GetByID(permissionID)
	roleWithPermissions.Permissions = append(roleWithPermissions.Permissions, *permission)
	
	return s.roleRepo.Update(roleWithPermissions)
}

func (s *RoleService) RemovePermissionFromRole(roleID, permissionID uint) error {
	if roleID == 0 || permissionID == 0 {
		return fmt.Errorf("invalid role ID or permission ID")
	}

	roleWithPermissions, err := s.roleRepo.GetWithPermissions(roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	var updatedPermissions []models.Permission
	found := false
	for _, permission := range roleWithPermissions.Permissions {
		if permission.PermID != permissionID {
			updatedPermissions = append(updatedPermissions, permission)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("permission not assigned to role")
	}

	roleWithPermissions.Permissions = updatedPermissions
	return s.roleRepo.Update(roleWithPermissions)
}
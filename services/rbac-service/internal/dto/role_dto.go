package dto

type CreateRoleRequest struct {
	Name string `json:"name" binding:"required" validate:"min=1,max=100"`
}

type UpdateRoleRequest struct {
	Name string `json:"name" binding:"required" validate:"min=1,max=100"`
}

type AddPermissionToRoleRequest struct {
	PermissionID uint `json:"permission_id" binding:"required"`
}

type RoleResponse struct {
	RoleID uint   `json:"role_id"`
	Name   string `json:"name"`
}

type RoleWithPermissionsResponse struct {
	RoleID      uint                 `json:"role_id"`
	Name        string               `json:"name"`
	Permissions []PermissionResponse `json:"permissions"`
}

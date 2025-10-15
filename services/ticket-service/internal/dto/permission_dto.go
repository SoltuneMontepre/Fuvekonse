package dto

type CreatePermissionRequest struct {
	Name string `json:"name" binding:"required" validate:"min=1,max=100"`
}

type UpdatePermissionRequest struct {
	Name string `json:"name" binding:"required" validate:"min=1,max=100"`
}

type PermissionResponse struct {
	PermID uint   `json:"perm_id"`
	Name   string `json:"name"`
}

type PermissionWithRolesResponse struct {
	PermID uint           `json:"perm_id"`
	Name   string         `json:"name"`
	Roles  []RoleResponse `json:"roles"`
}
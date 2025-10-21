package handlers

import (
	"net/http"
	"strconv"

	"github.com/SoltuneMontepre/Fuvekonse/services/general-service/internal/dto"
	"github.com/SoltuneMontepre/Fuvekonse/services/general-service/internal/models"
	"github.com/SoltuneMontepre/Fuvekonse/services/general-service/internal/services"
	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	roleService services.RoleServiceInterface
}

func NewRoleHandler(roleService services.RoleServiceInterface) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
	}
}

// CreateRole godoc
// @Summary Create a new role
// @Description Create a new role with the given name
// @Tags roles
// @Accept json
// @Produce json
// @Param role body dto.CreateRoleRequest true "Role data"
// @Success 201 {object} map[string]dto.RoleResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /roles [post]
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req dto.CreateRoleRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	role, err := h.roleService.CreateRole(req.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	response := dto.RoleResponse{
		RoleID: role.RoleID,
		Name:   role.Name,
	}

	c.JSON(http.StatusCreated, gin.H{"role": response})
}

// GetRole godoc
// @Summary Get a role by ID
// @Description Get a specific role by its ID
// @Tags roles
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Success 200 {object} map[string]dto.RoleResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /roles/{id} [get]
func (h *RoleHandler) GetRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid role ID"})
		return
	}

	role, err := h.roleService.GetRoleByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		return
	}

	response := dto.RoleResponse{
		RoleID: role.RoleID,
		Name:   role.Name,
	}

	c.JSON(http.StatusOK, gin.H{"role": response})
}

// GetRoles godoc
// @Summary Get all roles
// @Description Get a list of all roles
// @Tags roles
// @Accept json
// @Produce json
// @Success 200 {object} map[string][]dto.RoleResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /roles [get]
func (h *RoleHandler) GetRoles(c *gin.Context) {
	roles, err := h.roleService.GetAllRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	var response []dto.RoleResponse
	for _, role := range roles {
		response = append(response, dto.RoleResponse{
			RoleID: role.RoleID,
			Name:   role.Name,
		})
	}

	c.JSON(http.StatusOK, gin.H{"roles": response})
}

// UpdateRole godoc
// @Summary Update a role
// @Description Update an existing role
// @Tags roles
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Param role body dto.UpdateRoleRequest true "Role data"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /roles/{id} [put]
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid role ID"})
		return
	}

	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	role := &models.Role{
		RoleID: uint(id),
		Name:   req.Name,
	}

	if err := h.roleService.UpdateRole(role); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{Message: "Role updated successfully"})
}

// DeleteRole godoc
// @Summary Delete a role
// @Description Delete a role by its ID
// @Tags roles
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /roles/{id} [delete]
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid role ID"})
		return
	}

	if err := h.roleService.DeleteRole(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{Message: "Role deleted successfully"})
}

func (h *RoleHandler) GetRoleWithPermissions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid role ID"})
		return
	}

	role, err := h.roleService.GetRoleWithPermissions(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		return
	}

	var permissions []dto.PermissionResponse
	for _, perm := range role.Permissions {
		permissions = append(permissions, dto.PermissionResponse{
			PermID: perm.PermID,
			Name:   perm.Name,
		})
	}

	response := dto.RoleWithPermissionsResponse{
		RoleID:      role.RoleID,
		Name:        role.Name,
		Permissions: permissions,
	}

	c.JSON(http.StatusOK, gin.H{"role": response})
}

func (h *RoleHandler) AddPermissionToRole(c *gin.Context) {
	idStr := c.Param("id")
	roleID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid role ID"})
		return
	}

	var req dto.AddPermissionToRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.roleService.AddPermissionToRole(uint(roleID), req.PermissionID); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{Message: "Permission added to role successfully"})
}

func (h *RoleHandler) RemovePermissionFromRole(c *gin.Context) {
	idStr := c.Param("id")
	roleID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid role ID"})
		return
	}

	permissionIDStr := c.Param("permission_id")
	permissionID, err := strconv.ParseUint(permissionIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid permission ID"})
		return
	}

	if err := h.roleService.RemovePermissionFromRole(uint(roleID), uint(permissionID)); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{Message: "Permission removed from role successfully"})
}

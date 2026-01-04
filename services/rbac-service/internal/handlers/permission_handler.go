package handlers

import (
	"net/http"
	"rbac-service/internal/models"
	"rbac-service/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PermissionHandler struct {
	permissionService services.PermissionServiceInterface
}

func NewPermissionHandler(permissionService services.PermissionServiceInterface) *PermissionHandler {
	return &PermissionHandler{
		permissionService: permissionService,
	}
}

// CreatePermission godoc
// @Summary Create a new permission
// @Description Create a new permission with the given name
// @Tags permissions
// @Accept json
// @Produce json
// @Param permission body object{name=string} true "Permission data"
// @Success 201 {object} map[string]models.Permission
// @Failure 400 {object} map[string]string
// @Router /permissions [post]
func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	permission, err := h.permissionService.CreatePermission(ctx, req.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"permission": permission})
}

// GetPermission godoc
// @Summary Get a permission by ID
// @Description Get a specific permission by its ID
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path int true "Permission ID"
// @Success 200 {object} map[string]models.Permission
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /permissions/{id} [get]
func (h *PermissionHandler) GetPermission(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission ID"})
		return
	}

	permission, err := h.permissionService.GetPermissionByID(ctx, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"permission": permission})
}

// GetPermissions godoc
// @Summary Get all permissions
// @Description Get a list of all permissions
// @Tags permissions
// @Accept json
// @Produce json
// @Success 200 {object} map[string][]models.Permission
// @Failure 500 {object} map[string]string
// @Router /permissions [get]
func (h *PermissionHandler) GetPermissions(c *gin.Context) {
	ctx := c.Request.Context()

	permissions, err := h.permissionService.GetAllPermissions(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"permissions": permissions})
}

// UpdatePermission godoc
// @Summary Update a permission
// @Description Update an existing permission
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path int true "Permission ID"
// @Param permission body object{name=string} true "Permission data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /permissions/{id} [put]
func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission ID"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	permission := &models.Permission{
		PermID: uint(id),
		Name:   req.Name,
	}

	if err := h.permissionService.UpdatePermission(ctx, permission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permission updated successfully"})
}

// DeletePermission godoc
// @Summary Delete a permission
// @Description Delete a permission by its ID
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path int true "Permission ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /permissions/{id} [delete]
func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission ID"})
		return
	}

	if err := h.permissionService.DeletePermission(ctx, uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permission deleted successfully"})
}

func (h *PermissionHandler) GetPermissionWithRoles(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission ID"})
		return
	}

	permission, err := h.permissionService.GetPermissionWithRoles(ctx, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"permission": permission})
}

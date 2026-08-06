package rbac

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler provides HTTP handlers for RBAC management endpoints.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// mustParseCompanyID parses the companyId set in context by the auth middleware.
// On failure it writes a 400 response and returns false.
func mustParseCompanyID(c *gin.Context) (uuid.UUID, bool) {
	cid, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return uuid.Nil, false
	}
	return cid, true
}

// RoleResponse is the API shape for a single role.
type RoleResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	IsSystem    bool    `json:"isSystem"`
} //@name RoleResponse

// PermissionResponse is the API shape for a single permission.
type PermissionResponse struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Module      string `json:"module"`
	Description string `json:"description"`
} //@name PermissionResponse

// RolePermissionsResponse wraps the permission IDs for a role.
type RolePermissionsResponse struct {
	PermissionIDs []string `json:"permissionIds,omitempty"`
} //@name RolePermissionsResponse

// MyPermissionsResponse is the shape of GET /hris/me/permissions.
type MyPermissionsResponse struct {
	Permissions []string       `json:"permissions,omitempty"`
	Roles       []RoleResponse `json:"roles,omitempty"`
} //@name MyPermissionsResponse

// SetRolePermissionsRequest is the body for PUT /hris/roles/:roleId/permissions.
type SetRolePermissionsRequest struct {
	PermissionIDs []string `json:"permissionIds,omitempty"`
} //@name SetRolePermissionsRequest

// CreateRoleRequest is the body for POST /hris/roles.
type CreateRoleRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description,omitempty"`
} //@name CreateRoleRequest

// UpdateRoleRequest is the body for PUT /hris/roles/:roleId.
type UpdateRoleRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description,omitempty"`
} //@name UpdateRoleRequest

// AssignRoleRequest is the body for POST /hris/employees/:id/roles.
type AssignRoleRequest struct {
	RoleID uuid.UUID `json:"roleId" binding:"required"`
} //@name AssignRoleRequest

func roleToResponse(r Role) RoleResponse {
	return RoleResponse{
		ID:          r.ID.String(),
		Name:        r.Name,
		Description: r.Description,
		IsSystem:    r.IsSystem,
	}
}

func permToResponse(p Permission) PermissionResponse {
	return PermissionResponse{
		ID:          p.ID,
		Code:        p.Code,
		Module:      p.Module,
		Description: p.Description,
	}
}

// ListRoles godoc
// @Summary      List company roles
// @Description  Returns all HRIS roles for the authenticated company
// @Tags         rbac
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} RoleResponse
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /hris/roles [get]
func (h *Handler) ListRoles(c *gin.Context) {
	cid, ok := mustParseCompanyID(c)
	if !ok {
		return
	}
	roles, err := h.svc.ListRoles(c.Request.Context(), cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list roles"})
		return
	}
	resp := make([]RoleResponse, len(roles))
	for i, r := range roles {
		resp[i] = roleToResponse(r)
	}
	c.JSON(http.StatusOK, resp)
}

// ListPermissions godoc
// @Summary      List all permissions
// @Description  Returns all permissions grouped by module
// @Tags         rbac
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} PermissionResponse
// @Failure      500 {object} map[string]string
// @Router       /hris/permissions [get]
func (h *Handler) ListPermissions(c *gin.Context) {
	perms, err := h.svc.ListPermissions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list permissions"})
		return
	}
	resp := make([]PermissionResponse, len(perms))
	for i, p := range perms {
		resp[i] = permToResponse(p)
	}
	c.JSON(http.StatusOK, resp)
}

// GetRolePermissions godoc
// @Summary      Get role permission IDs
// @Description  Returns the permission IDs currently granted to a specific role
// @Tags         rbac
// @Produce      json
// @Security     BearerAuth
// @Param        roleId path string true "Role UUID"
// @Success      200 {object} RolePermissionsResponse
// @Failure      400 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /hris/roles/{roleId}/permissions [get]
func (h *Handler) GetRolePermissions(c *gin.Context) {
	cid, ok := mustParseCompanyID(c)
	if !ok {
		return
	}
	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}
	ids, err := h.svc.GetRolePermissionIDs(c.Request.Context(), cid, roleID)
	if err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load role permissions"})
		return
	}
	c.JSON(http.StatusOK, RolePermissionsResponse{PermissionIDs: ids})
}

// SetRolePermissions godoc
// @Summary      Replace role permissions
// @Description  Replaces the entire permission set for a role. System roles are protected.
// @Tags         rbac
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        roleId path string true "Role UUID"
// @Param        body body SetRolePermissionsRequest true "Permission IDs"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /hris/roles/{roleId}/permissions [put]
func (h *Handler) SetRolePermissions(c *gin.Context) {
	cid, ok := mustParseCompanyID(c)
	if !ok {
		return
	}
	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}
	var req SetRolePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if err := h.svc.SetRolePermissions(c.Request.Context(), cid, roleID, req.PermissionIDs); err != nil {
		switch {
		case errors.Is(err, ErrRoleNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		case errors.Is(err, ErrSystemRole):
			c.JSON(http.StatusBadRequest, gin.H{"error": "System roles cannot be modified"})
		case errors.Is(err, ErrInvalidPermission):
			slog.Error("set role permissions: invalid permission IDs", "roleID", roleID, "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "One or more permission IDs are invalid"})
		default:
			slog.Error("set role permissions failed", "roleID", roleID, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role permissions"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Permissions updated"})
}

// CreateRole godoc
// @Summary      Create a custom role
// @Description  Creates a new non-system role for the company
// @Tags         rbac
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body CreateRoleRequest true "Role details"
// @Success      201 {object} RoleResponse
// @Failure      400 {object} map[string]string
// @Router       /hris/roles [post]
func (h *Handler) CreateRole(c *gin.Context) {
	cid, ok := mustParseCompanyID(c)
	if !ok {
		return
	}
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	role, err := h.svc.CreateRole(c.Request.Context(), cid, req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, roleToResponse(*role))
}

// UpdateRole godoc
// @Summary      Update a custom role
// @Description  Updates name/description of a non-system role
// @Tags         rbac
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        roleId path string true "Role UUID"
// @Param        body body UpdateRoleRequest true "Role details"
// @Success      200 {object} RoleResponse
// @Failure      400 {object} map[string]string
// @Router       /hris/roles/{roleId} [put]
func (h *Handler) UpdateRole(c *gin.Context) {
	cid, ok := mustParseCompanyID(c)
	if !ok {
		return
	}
	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}
	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	role, err := h.svc.UpdateRole(c.Request.Context(), cid, roleID, req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, roleToResponse(*role))
}

// DeleteRole godoc
// @Summary      Delete a custom role
// @Description  Deletes a non-system role from the company
// @Tags         rbac
// @Produce      json
// @Security     BearerAuth
// @Param        roleId path string true "Role UUID"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Router       /hris/roles/{roleId} [delete]
func (h *Handler) DeleteRole(c *gin.Context) {
	cid, ok := mustParseCompanyID(c)
	if !ok {
		return
	}
	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}
	if err := h.svc.DeleteRole(c.Request.Context(), cid, roleID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Role deleted"})
}

// GetEmployeeRoles godoc
// @Summary      List employee roles
// @Description  Returns all roles assigned to a specific employee
// @Tags         rbac
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Employee UUID"
// @Success      200 {array} RoleResponse
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /hris/employees/{id}/roles [get]
func (h *Handler) GetEmployeeRoles(c *gin.Context) {
	cid, ok := mustParseCompanyID(c)
	if !ok {
		return
	}
	employeeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	roles, err := h.svc.GetEmployeeRoles(c.Request.Context(), cid, employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list employee roles"})
		return
	}
	resp := make([]RoleResponse, len(roles))
	for i, r := range roles {
		resp[i] = roleToResponse(r)
	}
	c.JSON(http.StatusOK, resp)
}

// AssignRole godoc
// @Summary      Assign role to employee
// @Description  Assigns a role to an employee
// @Tags         rbac
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Employee UUID"
// @Param        body body AssignRoleRequest true "Role to assign"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /hris/employees/{id}/roles [post]
func (h *Handler) AssignRole(c *gin.Context) {
	cid, ok := mustParseCompanyID(c)
	if !ok {
		return
	}
	requesterID, err := uuid.Parse(c.GetString("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	employeeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if err := h.svc.AssignRole(c.Request.Context(), cid, requesterID, employeeID, req.RoleID); err != nil {
		switch {
		case errors.Is(err, ErrCannotSelfAssign):
			c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot assign a role to yourself"})
		case errors.Is(err, ErrRoleNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		case errors.Is(err, ErrEmployeeNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		default:
			slog.Error("assign role failed", "employeeID", employeeID, "roleID", req.RoleID, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign role"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Role assigned"})
}

// RemoveRole godoc
// @Summary      Remove role from employee
// @Description  Removes a role from an employee
// @Tags         rbac
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Employee UUID"
// @Param        roleId path string true "Role UUID"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /hris/employees/{id}/roles/{roleId} [delete]
func (h *Handler) RemoveRole(c *gin.Context) {
	cid, ok := mustParseCompanyID(c)
	if !ok {
		return
	}
	employeeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	roleID, err := uuid.Parse(c.Param("roleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}
	if err := h.svc.RemoveRole(c.Request.Context(), cid, employeeID, roleID); err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found for this employee"})
			return
		}
		slog.Error("remove role failed", "employeeID", employeeID, "roleID", roleID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Role removed"})
}

// GetMyPermissions godoc
// @Summary      Get my permissions and roles
// @Description  Returns the current employee's permissions and roles
// @Tags         rbac
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} MyPermissionsResponse
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /hris/me/permissions [get]
func (h *Handler) GetMyPermissions(c *gin.Context) {
	cid, ok := mustParseCompanyID(c)
	if !ok {
		return
	}
	employeeID, err := uuid.Parse(c.GetString("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	perms, err := h.svc.GetEmployeePermissions(c.Request.Context(), cid, employeeID)
	if err != nil {
		slog.Error("get my permissions failed", "employeeID", employeeID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get permissions"})
		return
	}
	roles, err := h.svc.GetEmployeeRoles(c.Request.Context(), cid, employeeID)
	if err != nil {
		slog.Error("get my roles failed", "employeeID", employeeID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get roles"})
		return
	}
	roleResp := make([]RoleResponse, len(roles))
	for i, r := range roles {
		roleResp[i] = roleToResponse(r)
	}
	if perms == nil {
		perms = []string{}
	}
	c.JSON(http.StatusOK, MyPermissionsResponse{Permissions: perms, Roles: roleResp})
}

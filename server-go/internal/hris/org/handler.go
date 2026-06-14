package org

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc *Service
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{svc: NewService(db)}
}

func companyID(c *gin.Context) uuid.UUID {
	return uuid.MustParse(c.GetString("companyId"))
}

func parseID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return uuid.Nil, false
	}
	return id, true
}

// ============================================================
// Branches
// ============================================================

func (h *Handler) CreateBranch(c *gin.Context) {
	var req CreateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	b, err := h.svc.CreateBranch(c.Request.Context(), companyID(c), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create branch"})
		return
	}
	c.JSON(http.StatusCreated, b)
}

func (h *Handler) ListBranches(c *gin.Context) {
	branches, err := h.svc.ListBranches(c.Request.Context(), companyID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list branches"})
		return
	}
	c.JSON(http.StatusOK, branches)
}

func (h *Handler) GetBranch(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	b, err := h.svc.GetBranch(c.Request.Context(), companyID(c), id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Branch not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get branch"})
		return
	}
	c.JSON(http.StatusOK, b)
}

func (h *Handler) UpdateBranch(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req UpdateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	b, err := h.svc.UpdateBranch(c.Request.Context(), companyID(c), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update branch"})
		return
	}
	c.JSON(http.StatusOK, b)
}

func (h *Handler) DeleteBranch(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteBranch(c.Request.Context(), companyID(c), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deactivate branch"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Branch deactivated"})
}

// ============================================================
// Departments
// ============================================================

func (h *Handler) CreateDepartment(c *gin.Context) {
	var req CreateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	d, err := h.svc.CreateDepartment(c.Request.Context(), companyID(c), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create department"})
		return
	}
	c.JSON(http.StatusCreated, d)
}

func (h *Handler) ListDepartments(c *gin.Context) {
	depts, err := h.svc.ListDepartments(c.Request.Context(), companyID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list departments"})
		return
	}
	c.JSON(http.StatusOK, depts)
}

func (h *Handler) UpdateDepartment(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req UpdateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	d, err := h.svc.UpdateDepartment(c.Request.Context(), companyID(c), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update department"})
		return
	}
	c.JSON(http.StatusOK, d)
}

func (h *Handler) DeleteDepartment(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteDepartment(c.Request.Context(), companyID(c), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete department"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Department deleted"})
}

// ============================================================
// Positions
// ============================================================

func (h *Handler) CreatePosition(c *gin.Context) {
	var req CreatePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.svc.CreatePosition(c.Request.Context(), companyID(c), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create position"})
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *Handler) ListPositions(c *gin.Context) {
	positions, err := h.svc.ListPositions(c.Request.Context(), companyID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list positions"})
		return
	}
	c.JSON(http.StatusOK, positions)
}

func (h *Handler) UpdatePosition(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req UpdatePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.svc.UpdatePosition(c.Request.Context(), companyID(c), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update position"})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *Handler) DeletePosition(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.DeletePosition(c.Request.Context(), companyID(c), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete position"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Position deleted"})
}

// ============================================================
// Org Tree
// ============================================================

func (h *Handler) GetOrgTree(c *gin.Context) {
	tree, err := h.svc.GetOrgTree(c.Request.Context(), companyID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get org tree"})
		return
	}
	c.JSON(http.StatusOK, tree)
}

// ============================================================
// Enhanced Org Chart (employee-level)
// ============================================================

func (h *Handler) GetOrgChart(c *gin.Context) {
	chart, err := h.svc.GetOrgChart(c.Request.Context(), companyID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get org chart"})
		return
	}
	c.JSON(http.StatusOK, chart)
}

// ============================================================
// Working Calendars
// ============================================================

func (h *Handler) CreateCalendar(c *gin.Context) {
	var req CreateCalendarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	wc, err := h.svc.CreateCalendar(c.Request.Context(), companyID(c), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create calendar"})
		return
	}
	c.JSON(http.StatusCreated, wc)
}

func (h *Handler) ListCalendars(c *gin.Context) {
	var yearPtr *int
	if y := c.Query("year"); y != "" {
		n, err := strconv.Atoi(y)
		if err == nil {
			yearPtr = &n
		}
	}
	calendars, err := h.svc.ListCalendars(c.Request.Context(), companyID(c), yearPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list calendars"})
		return
	}
	c.JSON(http.StatusOK, calendars)
}

func (h *Handler) UpdateCalendar(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req UpdateCalendarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	wc, err := h.svc.UpdateCalendar(c.Request.Context(), companyID(c), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update calendar"})
		return
	}
	c.JSON(http.StatusOK, wc)
}

func (h *Handler) DeleteCalendar(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteCalendar(c.Request.Context(), companyID(c), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete calendar"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Calendar deleted"})
}

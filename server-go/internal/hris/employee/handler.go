package employee

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

func (h *Handler) Create(c *gin.Context) {
	companyID := mustParseCompanyID(c)

	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	emp, err := h.svc.Create(c.Request.Context(), companyID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create employee"})
		return
	}

	c.JSON(http.StatusCreated, emp)
}

func (h *Handler) Get(c *gin.Context) {
	companyID := mustParseCompanyID(c)
	employeeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}

	emp, err := h.svc.Get(c.Request.Context(), companyID, employeeID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get employee"})
		return
	}

	c.JSON(http.StatusOK, emp)
}

func (h *Handler) List(c *gin.Context) {
	companyID := mustParseCompanyID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	params := ListParams{
		CompanyID: companyID,
		Status:    c.Query("status"),
		Search:    c.Query("search"),
		Page:      page,
		PageSize:  pageSize,
	}

	if deptID := c.Query("departmentId"); deptID != "" {
		id, err := uuid.Parse(deptID)
		if err == nil {
			params.DepartmentID = &id
		}
	}
	if branchID := c.Query("branchId"); branchID != "" {
		id, err := uuid.Parse(branchID)
		if err == nil {
			params.BranchID = &id
		}
	}

	result, err := h.svc.List(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list employees"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) Update(c *gin.Context) {
	companyID := mustParseCompanyID(c)
	employeeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	emp, err := h.svc.Update(c.Request.Context(), companyID, employeeID, req)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update employee"})
		return
	}

	c.JSON(http.StatusOK, emp)
}

func mustParseCompanyID(c *gin.Context) uuid.UUID {
	return uuid.MustParse(c.GetString("companyId"))
}

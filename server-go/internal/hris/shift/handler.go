package shift

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc *Service
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{svc: NewService(db)}
}

func (h *Handler) ListTemplates(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	list, err := h.svc.ListTemplates(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list shift templates"})
		return
	}
	if list == nil {
		list = []ShiftTemplate{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) CreateTemplate(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	var t ShiftTemplate
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	t.CompanyID = companyID
	if err := h.svc.CreateTemplate(c.Request.Context(), companyID, &t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create shift template"})
		return
	}
	c.JSON(http.StatusCreated, t)
}

func (h *Handler) UpdateTemplate(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}
	var t ShiftTemplate
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if err := h.svc.UpdateTemplate(c.Request.Context(), companyID, id, &t); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Shift template not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update shift template"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Updated"})
}

func (h *Handler) DeleteTemplate(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}
	if err := h.svc.DeleteTemplate(c.Request.Context(), companyID, id); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Shift template not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete shift template"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

func (h *Handler) AssignShift(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	employeeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	_ = companyID

	var body struct {
		ShiftID       string  `json:"shiftId"`
		EffectiveDate string  `json:"effectiveDate"`
		EndDate       *string `json:"endDate"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	shiftID, err := uuid.Parse(body.ShiftID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid shift ID"})
		return
	}
	_ = shiftID

	es, err := h.svc.AssignShift(c.Request.Context(), employeeID, shiftID, body.EffectiveDate, body.EndDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign shift"})
		return
	}
	c.JSON(http.StatusCreated, es)
}

func (h *Handler) ListEmployeeShifts(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	employeeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	list, err := h.svc.ListEmployeeShifts(c.Request.Context(), companyID, employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list employee shifts"})
		return
	}
	if list == nil {
		list = []EmployeeShift{}
	}
	c.JSON(http.StatusOK, list)
}

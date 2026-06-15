package payroll

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

// ── Salary Components ──

func (h *Handler) ListComponents(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	list, err := h.svc.ListComponents(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list salary components"})
		return
	}
	if list == nil {
		list = []SalaryComponent{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) CreateComponent(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	var comp SalaryComponent
	if err := c.ShouldBindJSON(&comp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if comp.Type != "earning" && comp.Type != "deduction" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Type must be 'earning' or 'deduction'"})
		return
	}
	comp.CompanyID = companyID
	if err := h.svc.CreateComponent(c.Request.Context(), companyID, &comp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create salary component"})
		return
	}
	c.JSON(http.StatusCreated, comp)
}

func (h *Handler) UpdateComponent(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid component ID"})
		return
	}
	var comp SalaryComponent
	if err := c.ShouldBindJSON(&comp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if comp.Type != "earning" && comp.Type != "deduction" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Type must be 'earning' or 'deduction'"})
		return
	}
	if err := h.svc.UpdateComponent(c.Request.Context(), companyID, id, &comp); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Component not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update component"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Updated"})
}

func (h *Handler) DeleteComponent(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid component ID"})
		return
	}
	if err := h.svc.DeleteComponent(c.Request.Context(), companyID, id); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Component not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete component"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

// ── Employee Salary ──

func (h *Handler) GetEmployeeSalary(c *gin.Context) {
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
	list, err := h.svc.GetEmployeeSalary(c.Request.Context(), companyID, employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get employee salary"})
		return
	}
	if list == nil {
		list = []EmployeeSalary{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) SetEmployeeSalary(c *gin.Context) {
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
	var items []EmployeeSalary
	if err := c.ShouldBindJSON(&items); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if err := h.svc.SetEmployeeSalary(c.Request.Context(), companyID, employeeID, items); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Salary updated"})
}

// ── Payroll Runs ──

func (h *Handler) ListRuns(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	list, err := h.svc.ListRuns(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list payroll runs"})
		return
	}
	if list == nil {
		list = []PayrollRun{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) CreateRun(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	runBy, err := uuid.Parse(c.GetString("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	var body struct {
		PeriodStart string  `json:"periodStart" binding:"required"`
		PeriodEnd   string  `json:"periodEnd" binding:"required"`
		Notes       *string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	run, err := h.svc.CreateRun(c.Request.Context(), companyID, runBy, body.PeriodStart, body.PeriodEnd, body.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payroll run"})
		return
	}
	c.JSON(http.StatusCreated, run)
}

func (h *Handler) CalculateRun(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	runID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid run ID"})
		return
	}
	if err := h.svc.CalculateRun(c.Request.Context(), companyID, runID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Payroll calculated"})
}

func (h *Handler) ApproveRun(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	approverID, err := uuid.Parse(c.GetString("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	runID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid run ID"})
		return
	}
	if err := h.svc.ApproveRun(c.Request.Context(), companyID, runID, approverID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Payroll approved"})
}

func (h *Handler) MarkPaid(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	runID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid run ID"})
		return
	}
	if err := h.svc.MarkPaid(c.Request.Context(), companyID, runID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Payroll marked as paid"})
}

// ── Payslips ──

func (h *Handler) ListPayslips(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	runID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid run ID"})
		return
	}
	list, err := h.svc.ListPayslips(c.Request.Context(), companyID, runID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list payslips"})
		return
	}
	if list == nil {
		list = []Payslip{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) MyPayslips(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	employeeID, err := uuid.Parse(c.GetString("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	list, err := h.svc.GetMyPayslips(c.Request.Context(), companyID, employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get payslips"})
		return
	}
	if list == nil {
		list = []Payslip{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) GetPayslip(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	payslipID, err := uuid.Parse(c.Param("payslipId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payslip ID"})
		return
	}
	slip, err := h.svc.GetPayslip(c.Request.Context(), companyID, payslipID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Payslip not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get payslip"})
		return
	}
	c.JSON(http.StatusOK, slip)
}

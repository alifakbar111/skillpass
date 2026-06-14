package leave

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc *Service
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{svc: NewService(db)}
}

// ── Leave Types ──

func (h *Handler) ListTypes(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	list, err := h.svc.ListTypes(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list leave types"})
		return
	}
	if list == nil {
		list = []LeaveType{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) CreateType(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	var t LeaveType
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	t.CompanyID = companyID
	if err := h.svc.CreateType(c.Request.Context(), companyID, &t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create leave type"})
		return
	}
	c.JSON(http.StatusCreated, t)
}

func (h *Handler) UpdateType(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid leave type ID"})
		return
	}
	var t LeaveType
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if err := h.svc.UpdateType(c.Request.Context(), companyID, id, &t); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Leave type not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update leave type"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Updated"})
}

func (h *Handler) DeleteType(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid leave type ID"})
		return
	}
	if err := h.svc.DeleteType(c.Request.Context(), companyID, id); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Leave type not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete leave type"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

// ── Balances ──

func (h *Handler) GetBalances(c *gin.Context) {
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
	year := time.Now().Year()
	if y := c.Query("year"); y != "" {
		if parsed, e := strconv.Atoi(y); e == nil {
			year = parsed
		}
	}
	list, err := h.svc.GetBalances(c.Request.Context(), companyID, employeeID, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get balances"})
		return
	}
	if list == nil {
		list = []LeaveBalance{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) InitBalances(c *gin.Context) {
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
	year := time.Now().Year()
	if err := h.svc.InitBalances(c.Request.Context(), companyID, employeeID, year); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize balances"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Balances initialized"})
}

// ── Leave Requests ──

func (h *Handler) CreateRequest(c *gin.Context) {
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
	var req LeaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if err := h.svc.CreateRequest(c.Request.Context(), companyID, employeeID, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, req)
}

func (h *Handler) ListRequests(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	status := c.Query("status")
	list, err := h.svc.ListRequests(c.Request.Context(), companyID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list leave requests"})
		return
	}
	if list == nil {
		list = []LeaveRequest{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) MyRequests(c *gin.Context) {
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
	list, err := h.svc.MyRequests(c.Request.Context(), companyID, employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list requests"})
		return
	}
	if list == nil {
		list = []LeaveRequest{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) ReviewRequest(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	reviewerID, err := uuid.Parse(c.GetString("employeeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reviewer ID"})
		return
	}
	reqID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}
	var body struct {
		Status  string `json:"status"`
		Comment string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if body.Status != "approved" && body.Status != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status must be 'approved' or 'rejected'"})
		return
	}
	if err := h.svc.ReviewRequest(c.Request.Context(), companyID, reqID, reviewerID, body.Status, body.Comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Request reviewed"})
}

func (h *Handler) CancelRequest(c *gin.Context) {
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
	reqID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}
	if err := h.svc.CancelRequest(c.Request.Context(), companyID, employeeID, reqID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Request cancelled"})
}

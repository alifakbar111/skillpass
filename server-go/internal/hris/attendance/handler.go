package attendance

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc *Service
	hub *Hub
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{
		svc: NewService(db),
		hub: NewHub(),
	}
}

func (h *Handler) Hub() *Hub {
	return h.hub
}

func (h *Handler) ClockIn(c *gin.Context) {
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

	var req ClockInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	log, err := h.svc.ClockIn(c.Request.Context(), companyID, employeeID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.hub.Broadcast(companyID.String(), map[string]any{
		"type": "clock_in",
		"data": log,
	})

	c.JSON(http.StatusCreated, log)
}

func (h *Handler) ClockOut(c *gin.Context) {
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

	var req ClockOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	log, err := h.svc.ClockOut(c.Request.Context(), companyID, employeeID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.hub.Broadcast(companyID.String(), map[string]any{
		"type": "clock_out",
		"data": log,
	})

	c.JSON(http.StatusOK, log)
}

func (h *Handler) Dashboard(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	date := c.DefaultQuery("date", time.Now().Format("2006-01-02"))

	stats, logs, err := h.svc.GetDashboard(c.Request.Context(), companyID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load dashboard"})
		return
	}
	if logs == nil {
		logs = []AttendanceLog{}
	}
	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
		"logs":  logs,
	})
}

func (h *Handler) MyAttendance(c *gin.Context) {
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

	month := c.DefaultQuery("month", time.Now().Format("2006-01"))

	logs, err := h.svc.GetMyAttendance(c.Request.Context(), companyID, employeeID, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load attendance"})
		return
	}
	if logs == nil {
		logs = []AttendanceLog{}
	}
	c.JSON(http.StatusOK, logs)
}

func (h *Handler) CreateException(c *gin.Context) {
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

	var ex AttendanceException
	if err := c.ShouldBindJSON(&ex); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.svc.CreateException(c.Request.Context(), companyID, employeeID, &ex); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create exception"})
		return
	}
	c.JSON(http.StatusCreated, ex)
}

func (h *Handler) ListExceptions(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	status := c.Query("status")
	list, err := h.svc.ListExceptions(c.Request.Context(), companyID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list exceptions"})
		return
	}
	if list == nil {
		list = []AttendanceException{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) ReviewException(c *gin.Context) {
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
	exID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exception ID"})
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

	if err := h.svc.ReviewException(c.Request.Context(), companyID, exID, reviewerID, body.Status, body.Comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Exception reviewed"})
}

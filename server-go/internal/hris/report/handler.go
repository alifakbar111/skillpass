package report

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

func (h *Handler) ExportAttendance(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	from := c.Query("from")
	to := c.Query("to")
	if from == "" || to == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from and to query params required (YYYY-MM-DD)"})
		return
	}

	format := c.DefaultQuery("format", "json")

	rows, err := h.svc.ExportAttendance(c.Request.Context(), companyID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export attendance"})
		return
	}

	if format == "csv" {
		csv := h.svc.ToCSV(rows)
		c.Header("Content-Disposition", "attachment; filename=attendance_"+from+"_"+to+".csv")
		c.Data(http.StatusOK, "text/csv", []byte(csv))
		return
	}

	c.JSON(http.StatusOK, rows)
}

func (h *Handler) GetHeadcountStats(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	stats, err := h.svc.GetHeadcountStats(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get headcount stats"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *Handler) GenerateSnapshot(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	var req struct {
		Month string `json:"month" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "month required (YYYY-MM)"})
		return
	}

	snap, err := h.svc.GenerateSnapshot(c.Request.Context(), companyID, req.Month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate snapshot"})
		return
	}
	c.JSON(http.StatusOK, snap)
}

func (h *Handler) ListSnapshots(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	snapshots, err := h.svc.ListSnapshots(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list snapshots"})
		return
	}
	c.JSON(http.StatusOK, snapshots)
}

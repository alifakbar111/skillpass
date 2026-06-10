package analytics

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	db      *sql.DB
	service *Service
}

func NewHandler(db *sql.DB, service *Service) *Handler {
	return &Handler{db: db, service: service}
}

// CompanyAnalytics	godoc
// @Summary		Company hiring analytics
// @Description	Aggregate hiring metrics for the authenticated verified company: jobs, applications by status, time-to-decision, per-job funnels
// @Tags		analytics
// @Produce		json
// @Security	BearerAuth
// @Success		200 {object} analytics.CompanyAnalytics
// @Failure		401 {object} map[string]string
// @Failure		403 {object} map[string]string
// @Router		/company/analytics [get]
func (h *Handler) CompanyAnalytics(c *gin.Context) {
	companyIDVal, ok := c.Get("companyId")
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	companyID, ok := companyIDVal.(string)
	if !ok || companyID == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}

	result, err := h.service.ForCompany(c.Request.Context(), companyID)
	if err != nil {
		slog.Error("company analytics failed", "companyID", companyID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load analytics"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// JobseekerAnalytics	godoc
// @Summary		Jobseeker stats
// @Description	Application counts by status, response rate, and passport views for the authenticated jobseeker
// @Tags		analytics
// @Produce		json
// @Security	BearerAuth
// @Success		200 {object} analytics.JobseekerAnalytics
// @Failure		401 {object} map[string]string
// @Failure		404 {object} map[string]string
// @Router		/profiles/me/analytics [get]
func (h *Handler) JobseekerAnalytics(c *gin.Context) {
	userIDVal, ok := c.Get("userId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, ok := userIDVal.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var profileID uuid.UUID
	err := h.db.QueryRowContext(c.Request.Context(),
		`SELECT id FROM jobseeker_profiles WHERE user_id = $1`, userID,
	).Scan(&profileID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Jobseeker profile not found"})
			return
		}
		slog.Error("profile lookup failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load analytics"})
		return
	}

	result, err := h.service.ForJobseeker(c.Request.Context(), profileID.String())
	if err != nil {
		slog.Error("jobseeker analytics failed", "profileID", profileID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load analytics"})
		return
	}

	c.JSON(http.StatusOK, result)
}

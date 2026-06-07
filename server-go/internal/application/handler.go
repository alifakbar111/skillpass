package application

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler uses gin context to handle application requests.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func getUserID(c *gin.Context) (string, bool) {
	val, ok := c.Get("userId")
	if !ok {
		return "", false
	}
	s, ok := val.(string)
	return s, ok && s != ""
}

func getCompanyID(c *gin.Context) (string, bool) {
	val, ok := c.Get("companyId")
	if !ok {
		return "", false
	}
	s, ok := val.(string)
	return s, ok && s != ""
}

// Apply applies the authenticated jobseeker to a job posting.
// POST /api/v1/jobs/:id/apply
func (h *Handler) Apply(c *gin.Context) {
	userIDStr, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	jobPostingID := c.Param("id")
	if _, err := uuid.Parse(jobPostingID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job posting ID"})
		return
	}

	// Look up the jobseeker profile ID from user ID
	profileID, err := h.service.LookupJobseekerProfileID(c.Request.Context(), userIDStr)
	if err != nil {
		slog.Error("failed to lookup jobseeker profile", "userID", userIDStr, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Jobseeker profile not found"})
		return
	}

	result, err := h.service.Apply(c.Request.Context(), profileID, jobPostingID)
	if err != nil {
		slog.Error("application failed", "error", err)
		switch {
		case errors.Is(err, ErrJobNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, ErrJobClosed):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, ErrDuplicate):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to apply"})
		}
		return
	}

	c.JSON(http.StatusCreated, result)
}

// ListMyApplications returns the jobseeker's applications (kanban data).
// GET /api/v1/applications/me
func (h *Handler) ListMyApplications(c *gin.Context) {
	userIDStr, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	profileID, err := h.service.LookupJobseekerProfileID(c.Request.Context(), userIDStr)
	if err != nil {
		slog.Error("failed to lookup jobseeker profile", "userID", userIDStr, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Jobseeker profile not found"})
		return
	}

	applications, err := h.service.ListForJobseeker(c.Request.Context(), profileID)
	if err != nil {
		slog.Error("failed to list applications", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list applications"})
		return
	}

	if applications == nil {
		applications = []ApplicationResult{}
	}

	c.JSON(http.StatusOK, applications)
}

// UpdateStatus updates an application's status (company action).
// PUT /api/v1/applications/:id/status
func (h *Handler) UpdateStatus(c *gin.Context) {
	companyID, ok := getCompanyID(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}

	applicationID := c.Param("id")
	if _, err := uuid.Parse(applicationID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: status is required"})
		return
	}

	result, err := h.service.UpdateStatus(c.Request.Context(), applicationID, companyID, req.Status)
	if err != nil {
		slog.Error("failed to update application status", "applicationID", applicationID, "error", err)
		switch {
		case errors.Is(err, ErrAppNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, ErrInvalidStatus):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

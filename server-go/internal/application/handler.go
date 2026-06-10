package application

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Notifier creates notifications in response to application events.
// Implemented by the notification.Service; kept as an interface here to avoid
// a hard dependency and to allow nil (no-op) in tests.
type Notifier interface {
	NotifyCompanyOfApplication(ctx context.Context, jobPostingID, jobseekerProfileID string) error
	NotifyJobseekerOfStatus(ctx context.Context, applicationID, status string) error
}

// Handler uses gin context to handle application requests.
type Handler struct {
	service  *Service
	notifier Notifier
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// SetNotifier attaches a notifier used to emit notifications on application events.
// Optional — when nil, no notifications are emitted.
func (h *Handler) SetNotifier(n Notifier) {
	h.notifier = n
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

// Apply		godoc
// @Summary		Apply to a job
// @Description	Apply the authenticated jobseeker to a job posting
// @Tags		applications
// @Produce		json
// @Security	BearerAuth
// @Param		id path string true "Job posting UUID"
// @Success		201 {object} application.ApplicationResult
// @Failure		400 {object} map[string]string
// @Failure		401 {object} map[string]string
// @Failure		404 {object} map[string]string
// @Failure		409 {object} map[string]string
// @Router		/jobs/{id}/apply [post]
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

	// Best-effort: notify the company of the new application.
	if h.notifier != nil {
		if err := h.notifier.NotifyCompanyOfApplication(c.Request.Context(), jobPostingID, profileID); err != nil {
			slog.Warn("failed to notify company of application", "error", err)
		}
	}

	c.JSON(http.StatusCreated, result)
}

// ListMyApplications	godoc
// @Summary		List my applications
// @Description	Get all applications for the authenticated jobseeker (kanban-style data with job title and company name)
// @Tags		applications
// @Produce		json
// @Security	BearerAuth
// @Success		200 {array} application.ApplicationResult
// @Failure		401 {object} map[string]string
// @Failure		404 {object} map[string]string
// @Router		/applications/me [get]
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

// ListCompanyApplications	godoc
// @Summary		List applications for company's jobs
// @Description	Get all applications for jobs owned by the authenticated verified company
// @Tags		applications
// @Produce		json
// @Security	BearerAuth
// @Success		200 {array} application.CompanyApplicationResult
// @Failure		401 {object} map[string]string
// @Failure		403 {object} map[string]string
// @Router		/company/applications [get]
func (h *Handler) ListCompanyApplications(c *gin.Context) {
	companyID, ok := getCompanyID(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}

	applications, err := h.service.ListForCompany(c.Request.Context(), companyID)
	if err != nil {
		slog.Error("failed to list company applications", "companyID", companyID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list applications"})
		return
	}

	if applications == nil {
		applications = []CompanyApplicationResult{}
	}

	c.JSON(http.StatusOK, applications)
}

// AddMessage	godoc
// @Summary		Add a note to an application
// @Description	Company adds a note/message to an application it owns (visible to the candidate)
// @Tags		applications
// @Accept		json
// @Produce		json
// @Security	BearerAuth
// @Param		id path string true "Application UUID"
// @Param		body body object{body=string} true "Message body"
// @Success		201 {object} application.Message
// @Failure		400 {object} map[string]string
// @Failure		403 {object} map[string]string
// @Failure		404 {object} map[string]string
// @Router		/applications/{id}/messages [post]
func (h *Handler) AddMessage(c *gin.Context) {
	companyID, ok := getCompanyID(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	senderUserID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	applicationID := c.Param("id")
	if _, err := uuid.Parse(applicationID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID"})
		return
	}

	var req struct {
		Body string `json:"body" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: body is required"})
		return
	}

	msg, err := h.service.AddMessage(c.Request.Context(), applicationID, companyID, senderUserID, req.Body)
	if err != nil {
		switch {
		case errors.Is(err, ErrAppNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			slog.Error("failed to add message", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add message"})
		}
		return
	}

	// Best-effort: notify the jobseeker of the new note.
	if h.notifier != nil {
		if err := h.notifier.NotifyJobseekerOfStatus(c.Request.Context(), applicationID, "updated with a note"); err != nil {
			slog.Warn("failed to notify jobseeker of note", "error", err)
		}
	}

	c.JSON(http.StatusCreated, msg)
}

// ListMessages	godoc
// @Summary		List application messages
// @Description	Company lists the message thread for an application it owns
// @Tags		applications
// @Produce		json
// @Security	BearerAuth
// @Param		id path string true "Application UUID"
// @Success		200 {array} application.Message
// @Failure		403 {object} map[string]string
// @Failure		404 {object} map[string]string
// @Router		/applications/{id}/messages [get]
func (h *Handler) ListMessages(c *gin.Context) {
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

	messages, err := h.service.ListMessages(c.Request.Context(), applicationID, companyID)
	if err != nil {
		switch {
		case errors.Is(err, ErrAppNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			slog.Error("failed to list messages", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list messages"})
		}
		return
	}

	c.JSON(http.StatusOK, messages)
}

// UpdateStatus		godoc
// @Summary		Update application status
// @Description	Update an application's status (reviewed, interviewed, offered, rejected). Company action for verified companies.
// @Tags		applications
// @Accept		json
// @Produce		json
// @Security	BearerAuth
// @Param		id path string true "Application UUID"
// @Param		body body object{status=string} true "New status value"
// @Success		200 {object} application.ApplicationResult
// @Failure		400 {object} map[string]string
// @Failure		403 {object} map[string]string
// @Failure		404 {object} map[string]string
// @Router		/applications/{id}/status [put]
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

	// Best-effort: notify the jobseeker of the status change.
	if h.notifier != nil {
		if err := h.notifier.NotifyJobseekerOfStatus(c.Request.Context(), applicationID, req.Status); err != nil {
			slog.Warn("failed to notify jobseeker of status change", "error", err)
		}
	}

	c.JSON(http.StatusOK, result)
}

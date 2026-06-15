package feedback

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

// PostFeedback godoc
// @Summary      Submit feedback for a candidate
// @Description  Company submits feedback for a jobseeker profile
// @Tags         feedback
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        profile_id path string true "Jobseeker profile ID"
// @Param        body body CreateFeedbackRequest true "Feedback payload"
// @Success      201 {object} Feedback
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /feedback/{profile_id} [post]
func (h *Handler) PostFeedback(c *gin.Context) {
	companyID, err := getCompanyID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	profileID := c.Param("profile_id")
	if _, err := uuid.Parse(profileID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid profile ID"})
		return
	}

	if err := h.lookupProfileID(c, profileID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Jobseeker profile not found"})
			return
		}
		slog.Error("failed to lookup profile", "profileID", profileID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to lookup profile"})
		return
	}

	var req CreateFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Content is required"})
		return
	}

	fb, err := h.service.Create(c.Request.Context(), profileID, companyID, &req)
	if err != nil {
		slog.Error("failed to create feedback", "profileID", profileID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create feedback"})
		return
	}

	c.JSON(http.StatusCreated, fb)
}

// GetMyFeedback godoc
// @Summary      Get feedback for current user's profile
// @Description  Jobseeker retrieves all feedback submitted for their profile
// @Tags         feedback
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} Feedback
// @Failure      401 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /feedback/me [get]
func (h *Handler) GetMyFeedback(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	profileID, err := h.lookupProfileIDByUserID(c, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Jobseeker profile not found"})
			return
		}
		slog.Error("failed to lookup profile", "userID", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to lookup profile"})
		return
	}

	feedbacks, err := h.service.GetByProfileID(c.Request.Context(), profileID)
	if err != nil {
		slog.Error("failed to get feedback", "profileID", profileID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get feedback"})
		return
	}

	if feedbacks == nil {
		feedbacks = []Feedback{}
	}

	c.JSON(http.StatusOK, feedbacks)
}

// GetCompanyFeedback godoc
// @Summary      Get feedback for current company
// @Description  Company retrieves all feedback they have submitted
// @Tags         feedback
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} Feedback
// @Failure      401 {object} map[string]string
// @Router       /feedback/company [get]
func (h *Handler) GetCompanyFeedback(c *gin.Context) {
	companyID, err := getCompanyID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	feedbacks, err := h.service.GetByCompanyID(c.Request.Context(), companyID)
	if err != nil {
		slog.Error("failed to get feedback", "companyID", companyID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get feedback"})
		return
	}

	if feedbacks == nil {
		feedbacks = []Feedback{}
	}

	c.JSON(http.StatusOK, feedbacks)
}

// GetMySuggestions godoc
// @Summary      Get AI suggestions for current user
// @Description  Jobseeker retrieves aggregated AI suggestions from all feedback
// @Tags         feedback
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} AISuggestion
// @Failure      401 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /feedback/suggestions/me [get]
func (h *Handler) GetMySuggestions(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	profileID, err := h.lookupProfileIDByUserID(c, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Jobseeker profile not found"})
			return
		}
		slog.Error("failed to lookup profile", "userID", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to lookup profile"})
		return
	}

	suggestions, err := h.service.GetAISuggestionsByProfileID(c.Request.Context(), profileID)
	if err != nil {
		slog.Error("failed to get AI suggestions", "profileID", profileID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get AI suggestions"})
		return
	}

	if suggestions == nil {
		suggestions = []AISuggestion{}
	}

	c.JSON(http.StatusOK, suggestions)
}

func (h *Handler) lookupProfileID(c *gin.Context, profileID string) error {
	var id uuid.UUID
	return h.db.QueryRowContext(c.Request.Context(),
		`SELECT id FROM jobseeker_profiles WHERE id = $1`, profileID,
	).Scan(&id)
}

func (h *Handler) lookupProfileIDByUserID(c *gin.Context, userID string) (string, error) {
	var profileID uuid.UUID
	err := h.db.QueryRowContext(c.Request.Context(),
		`SELECT id FROM jobseeker_profiles WHERE user_id = $1`, userID,
	).Scan(&profileID)
	if err != nil {
		return "", err
	}
	return profileID.String(), nil
}

func getUserID(c *gin.Context) (string, error) {
	userIDVal, ok := c.Get("userId")
	if !ok {
		return "", errors.New("unauthorized")
	}
	userIDStr, ok := userIDVal.(string)
	if !ok || userIDStr == "" {
		return "", errors.New("unauthorized")
	}
	return userIDStr, nil
}

func getCompanyID(c *gin.Context) (string, error) {
	companyIDVal, ok := c.Get("companyId")
	if !ok {
		return "", errors.New("unauthorized: company ID not found")
	}
	companyIDStr, ok := companyIDVal.(string)
	if !ok || companyIDStr == "" {
		return "", errors.New("unauthorized: invalid company ID")
	}
	return companyIDStr, nil
}

package career

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

// ListCareerPaths godoc
// @Summary      List career paths
// @Description  Get all career paths, optionally filtered by industry
// @Tags         career
// @Produce      json
// @Param        industry query string false "Filter by industry"
// @Success      200 {array} CareerPath
// @Router       /career/paths [get]
func (h *Handler) ListCareerPaths(c *gin.Context) {
	industry := c.Query("industry")

	paths, err := h.service.ListPaths(c.Request.Context(), industry)
	if err != nil {
		slog.Error("failed to list career paths", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list career paths"})
		return
	}

	c.JSON(http.StatusOK, paths)
}

// GetSkillGap godoc
// @Summary      Get skill gap analysis
// @Description  Compare authenticated jobseeker's skills against career path requirements
// @Tags         career
// @Produce      json
// @Security     BearerAuth
// @Param        industry query string true "Industry to compare against"
// @Success      200 {object} SkillGapResult
// @Failure      401 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /career/skill-gap/me [get]
func (h *Handler) GetSkillGap(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	profileID, err := h.lookupProfileID(c, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Jobseeker profile not found"})
			return
		}
		slog.Error("failed to lookup profile", "userID", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to lookup profile"})
		return
	}

	industry := c.Query("industry")
	if industry == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "industry query parameter is required"})
		return
	}

	result, err := h.service.GetSkillGap(c.Request.Context(), profileID, industry)
	if err != nil {
		slog.Error("failed to get skill gap", "profileID", profileID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate skill gap"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetCareerPath godoc
// @Summary      Get AI career path prediction
// @Description  Generate AI-powered career path prediction for the authenticated jobseeker
// @Tags         career
// @Produce      json
// @Security     BearerAuth
// @Param        industry query string true "Industry context for prediction"
// @Success      200 {object} CareerPrediction
// @Failure      401 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /career/path/me [get]
func (h *Handler) GetCareerPath(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	profileID, err := h.lookupProfileID(c, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Jobseeker profile not found"})
			return
		}
		slog.Error("failed to lookup profile", "userID", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to lookup profile"})
		return
	}

	industry := c.Query("industry")
	if industry == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "industry query parameter is required"})
		return
	}

	result, err := h.service.PredictPath(c.Request.Context(), profileID, industry)
	if err != nil {
		slog.Error("career path prediction failed", "profileID", profileID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate career prediction"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) lookupProfileID(c *gin.Context, userID string) (string, error) {
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

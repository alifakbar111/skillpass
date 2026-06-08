package evaluation

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

// EvaluationResponse is the public shape returned to the frontend.
type EvaluationResponse struct {
	ID           string           `json:"id"`
	OverallScore int              `json:"overallScore"`
	Strengths    []SkillNote      `json:"strengths"`
	Weaknesses   []SkillNote      `json:"weaknesses"`
	Suggestions  []Suggestion     `json:"suggestions"`
	SkillScores  []SkillScoreItem `json:"skillScores"`
	CreatedAt    string           `json:"createdAt"`
}

// PostEvaluate triggers a fresh AI evaluation for the authenticated jobseeker.
func (h *Handler) PostEvaluate(c *gin.Context) {
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

	result, err := h.service.Evaluate(c.Request.Context(), profileID)
	if err != nil {
		slog.Error("evaluation failed", "profileID", profileID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Evaluation failed"})
		return
	}

	c.JSON(http.StatusOK, toResponse(result))
}

// GetLatestEvaluation returns the most recent evaluation.
func (h *Handler) GetLatestEvaluation(c *gin.Context) {
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

	result, err := h.service.GetLatest(c.Request.Context(), profileID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "No evaluation found. Trigger one first."})
			return
		}
		slog.Error("failed to get latest evaluation", "profileID", profileID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get evaluation"})
		return
	}

	c.JSON(http.StatusOK, toResponse(result))
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

func toResponse(eval *EvaluationResult) EvaluationResponse {
	return EvaluationResponse{
		ID:           eval.ID,
		OverallScore: eval.OverallScore,
		Strengths:    eval.Strengths,
		Weaknesses:   eval.Weaknesses,
		Suggestions:  eval.Suggestion,
		SkillScores:  eval.SkillScores,
		CreatedAt:    eval.CreatedAt,
	}
}

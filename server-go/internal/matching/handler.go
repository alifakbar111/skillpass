package matching

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// lookupProfileID resolves the jobseeker profile id for a user via raw SQL
// (ad-hoc go-jet destinations without alias tags scan zero values silently).
func (h *Handler) lookupProfileID(c *gin.Context, userID string) (string, error) {
	var profileID uuid.UUID
	err := h.service.db.QueryRowContext(c.Request.Context(),
		`SELECT id FROM jobseeker_profiles WHERE user_id = $1`, userID,
	).Scan(&profileID)
	if err != nil {
		return "", err
	}
	return profileID.String(), nil
}

// MatchJobs		godoc
// @Summary		Match jobs for jobseeker
// @Description	Get job recommendations based on the authenticated jobseeker's skills and AI evaluation
// @Tags		matching
// @Produce		json
// @Security	BearerAuth
// @Success		200 {array} matching.JobMatch
// @Failure		401 {object} map[string]string
// @Router		/jobs/matches [get]
func (h *Handler) MatchJobs(c *gin.Context) {
	userID, ok := c.Get("userId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userIDStr := userID.(string)

	profileID, err := h.lookupProfileID(c, userIDStr)
	if err != nil {
		slog.Error("profile lookup failed", "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	matches, err := h.service.MatchJobs(c.Request.Context(), profileID)
	if err != nil {
		slog.Error("job matching failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Matching failed"})
		return
	}

	if matches == nil {
		matches = []JobMatch{}
	}

	c.JSON(http.StatusOK, matches)
}

// SkillsGap		godoc
// @Summary		Skills gap for a job
// @Description	Compare the authenticated jobseeker's evaluated skills against a job's required skills
// @Tags		matching
// @Produce		json
// @Security	BearerAuth
// @Param		id path string true "Job posting UUID"
// @Success		200 {object} matching.SkillsGap
// @Failure		401 {object} map[string]string
// @Failure		404 {object} map[string]string
// @Router		/jobs/{id}/skills-gap [get]
func (h *Handler) SkillsGap(c *gin.Context) {
	userID, ok := c.Get("userId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userIDStr := userID.(string)

	jobPostingID := c.Param("id")
	if _, err := uuid.Parse(jobPostingID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job posting ID"})
		return
	}

	profileID, err := h.lookupProfileID(c, userIDStr)
	if err != nil {
		slog.Error("profile lookup failed", "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	gap, err := h.service.ComputeSkillsGap(c.Request.Context(), profileID, jobPostingID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
			return
		}
		slog.Error("skills gap failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to compute skills gap"})
		return
	}

	c.JSON(http.StatusOK, gap)
}

// MatchCandidates		godoc
// @Summary		Match candidates for a job
// @Description	Get candidate recommendations for a specific job posting based on skill matching
// @Tags		matching
// @Produce		json
// @Security	BearerAuth
// @Param		jobId query string true "Job posting UUID to find candidates for"
// @Success		200 {array} matching.CandidateMatch
// @Failure		400 {object} map[string]string
// @Failure		401 {object} map[string]string
// @Router		/candidates/matches [get]
func (h *Handler) MatchCandidates(c *gin.Context) {
	jobPostingID := c.Query("jobId")
	if jobPostingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "jobId query parameter is required"})
		return
	}

	matches, err := h.service.MatchCandidates(c.Request.Context(), jobPostingID)
	if err != nil {
		slog.Error("candidate matching failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Matching failed"})
		return
	}

	if matches == nil {
		matches = []CandidateMatch{}
	}

	// Blind hiring mode: mask candidate identities for companies that opted in.
	if companyIDVal, ok := c.Get("companyId"); ok {
		if companyID, ok := companyIDVal.(string); ok && companyID != "" {
			if h.service.IsBlindMode(c.Request.Context(), companyID) {
				for i := range matches {
					short := matches[i].ProfileID
					if len(short) > 8 {
						short = short[:8]
					}
					matches[i].Name = "Candidate " + short
				}
			}
		}
	}

	c.JSON(http.StatusOK, matches)
}

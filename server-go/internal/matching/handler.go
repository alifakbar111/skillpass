package matching

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"

	"skillpass-server-go/internal/gen"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
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

	profileStmt := SELECT(
		gen.JobseekerProfiles.ID,
	).FROM(
		gen.JobseekerProfiles,
	).WHERE(
		gen.JobseekerProfiles.UserID.EQ(UUID(uuid.MustParse(userIDStr))),
	)

	var profile struct{ ID uuid.UUID }
	err := profileStmt.QueryContext(c.Request.Context(), h.service.db, &profile)
	if err != nil {
		slog.Error("profile lookup failed", "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	matches, err := h.service.MatchJobs(c.Request.Context(), profile.ID.String())
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

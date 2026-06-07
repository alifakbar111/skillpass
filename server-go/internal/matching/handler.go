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
		gen.JobseekerProfiles.UserID.EQ(String(userIDStr)),
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

	c.JSON(http.StatusOK, matches)
}

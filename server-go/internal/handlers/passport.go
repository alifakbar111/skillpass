package handlers

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"

	"skillpass-server-go/internal/models"
)

type PublicProfileResponse struct {
	Name        string       `json:"name"`
	AvatarURL   *string      `json:"avatarUrl,omitempty"`
	Headline    *string      `json:"headline,omitempty"`
	About       *string      `json:"about,omitempty"`
	YearsOfExp  *int         `json:"yearsOfExperience,omitempty"`
	ViewCount   int          `json:"viewCount"`
	Experiences []Experience `json:"experiences,omitempty"`
} //@name PublicProfileResponse

type PassportHandler struct {
	db    *sql.DB
	bunDB *bun.DB
}

func NewPassportHandler(db *sql.DB, bunDB *bun.DB) *PassportHandler {
	return &PassportHandler{db: db, bunDB: bunDB}
}

// GetProfile		godoc
// @Summary		Get public profile
// @Description	Get a jobseeker's public profile by username/slug (no auth required)
// @Tags		passport
// @Produce		json
// @Param		username path string true "Profile username/slug"
// @Success		200 {object} PublicProfileResponse
// @Failure		404 {object} map[string]string
// @Router		/profiles/{username} [get]
func (h *PassportHandler) GetProfile(c *gin.Context) {
	username := c.Param("username")

	var profile models.JobseekerProfile
	err := h.bunDB.NewSelect().Model(&profile).
		Where("slug = ?", username).
		Scan(c.Request.Context())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
			return
		}
		slog.Error("failed to load profile", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load profile"})
		return
	}

	var user models.User
	err = h.bunDB.NewSelect().Model(&user).Column("name", "avatar_url").
		Where("id = ?", profile.UserID).
		Scan(c.Request.Context())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
			return
		}
		slog.Error("failed to load user", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load user"})
		return
	}

	var exps []models.JobExperience
	err = h.bunDB.NewSelect().Model(&exps).
		Where("profile_id = ?", profile.ID).
		Order("start_date ASC").
		Scan(c.Request.Context())
	if err != nil {
		slog.Error("failed to query experiences", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query experiences"})
		return
	}

	experiences := make([]Experience, len(exps))
	for i, exp := range exps {
		experiences[i] = mapExperience(exp)
	}

	// view_count is intentionally not written on public GETs. The previous
	// behavior issued an UPDATE on every unauthenticated request — a DoS
	// amplifier and trivially gameable. A dedicated authenticated stats
	// endpoint backed by the profile_views table is the path forward.
	c.JSON(http.StatusOK, PublicProfileResponse{
		Name:        user.Name,
		AvatarURL:   user.AvatarURL,
		Headline:    profile.Headline,
		About:       profile.About,
		YearsOfExp:  int32ToIntPtr(profile.YearsOfExperience),
		ViewCount:   0,
		Experiences: experiences,
	})
}

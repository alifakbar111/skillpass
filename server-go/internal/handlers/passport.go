package handlers

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	. "github.com/go-jet/jet/v2/postgres"

	"skillpass-server-go/.gen/skillpass/public/model"
	"skillpass-server-go/internal/gen"
)

type PublicProfileResponse struct {
	Name        string       `json:"name"`
	AvatarURL   *string      `json:"avatarUrl"`
	Headline    *string      `json:"headline"`
	About       *string      `json:"about"`
	YearsOfExp  *int         `json:"yearsOfExperience"`
	ViewCount   int          `json:"viewCount"`
	Experiences []Experience `json:"experiences"`
} //@name PublicProfileResponse

type PassportHandler struct {
	db *sql.DB
}

func NewPassportHandler(db *sql.DB) *PassportHandler {
	return &PassportHandler{db: db}
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

	profileStmt := SELECT(
		gen.JobseekerProfiles.ID, gen.JobseekerProfiles.UserID, gen.JobseekerProfiles.Headline,
		gen.JobseekerProfiles.About, gen.JobseekerProfiles.YearsOfExperience,
	).FROM(
		gen.JobseekerProfiles,
	).WHERE(
		gen.JobseekerProfiles.Slug.EQ(String(username)),
	)

	var profiles []model.JobseekerProfiles
	err := profileStmt.QueryContext(c.Request.Context(), h.db, &profiles)
	if err != nil {
		slog.Error("failed to load profile", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load profile"})
		return
	}
	if len(profiles) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}
	profile := profiles[0]

	userStmt := SELECT(
		gen.Users.Name, gen.Users.AvatarURL,
	).FROM(
		gen.Users,
	).WHERE(
		gen.Users.ID.EQ(UUID(profile.UserID)),
	)

	var users []model.Users
	err = userStmt.QueryContext(c.Request.Context(), h.db, &users)
	if err != nil {
		slog.Error("failed to load user", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load user"})
		return
	}
	if len(users) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}
	user := users[0]

	expStmt := SELECT(
		gen.JobExperiences.ID, gen.JobExperiences.ProfileID, gen.JobExperiences.Type,
		gen.JobExperiences.Title, gen.JobExperiences.Organization, gen.JobExperiences.StartDate,
		gen.JobExperiences.EndDate, gen.JobExperiences.IsCurrent, gen.JobExperiences.Description,
		gen.JobExperiences.Industry, gen.JobExperiences.SkillsUsed, gen.JobExperiences.URL,
	).FROM(
		gen.JobExperiences,
	).WHERE(
		gen.JobExperiences.ProfileID.EQ(UUID(profile.ID)),
	).ORDER_BY(
		gen.JobExperiences.StartDate.ASC(),
	)

	var exps []model.JobExperiences
	err = expStmt.QueryContext(c.Request.Context(), h.db, &exps)
	if err != nil {
		slog.Error("failed to query experiences", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query experiences"})
		return
	}

	experiences := make([]Experience, len(exps))
	for i, exp := range exps {
		experiences[i] = mapExperience(exp)
	}

	// Increment the view counter and read the updated value (raw SQL — column not in go-jet model).
	var viewCount int
	if err := h.db.QueryRowContext(c.Request.Context(),
		`UPDATE jobseeker_profiles SET view_count = view_count + 1 WHERE id = $1 RETURNING view_count`,
		profile.ID,
	).Scan(&viewCount); err != nil {
		slog.Warn("failed to increment view count", "error", err)
	}

	c.JSON(http.StatusOK, PublicProfileResponse{
		Name:        user.Name,
		AvatarURL:   user.AvatarURL,
		Headline:    profile.Headline,
		About:       profile.About,
		YearsOfExp:  int32ToIntPtr(profile.YearsOfExperience),
		ViewCount:   viewCount,
		Experiences: experiences,
	})
}

package handlers

import (
	"database/sql"
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
	Experiences []Experience `json:"experiences"`
}

type PassportHandler struct {
	db *sql.DB
}

func NewPassportHandler(db *sql.DB) *PassportHandler {
	return &PassportHandler{db: db}
}

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

	var profile model.JobseekerProfiles
	err := profileStmt.QueryContext(c.Request.Context(), h.db, &profile)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	userStmt := SELECT(
		gen.Users.Name, gen.Users.AvatarURL,
	).FROM(
		gen.Users,
	).WHERE(
		gen.Users.ID.EQ(String(profile.UserID.String())),
	)

	var user model.Users
	err = userStmt.QueryContext(c.Request.Context(), h.db, &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
		return
	}

	expStmt := SELECT(
		gen.JobExperiences.ID, gen.JobExperiences.ProfileID, gen.JobExperiences.Type,
		gen.JobExperiences.Title, gen.JobExperiences.Organization, gen.JobExperiences.StartDate,
		gen.JobExperiences.EndDate, gen.JobExperiences.IsCurrent, gen.JobExperiences.Description,
		gen.JobExperiences.Industry, gen.JobExperiences.SkillsUsed, gen.JobExperiences.URL,
	).FROM(
		gen.JobExperiences,
	).WHERE(
		gen.JobExperiences.ProfileID.EQ(String(profile.ID.String())),
	).ORDER_BY(
		gen.JobExperiences.StartDate.ASC(),
	)

	var exps []model.JobExperiences
	err = expStmt.QueryContext(c.Request.Context(), h.db, &exps)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query experiences"})
		return
	}

	experiences := make([]Experience, len(exps))
	for i, exp := range exps {
		experiences[i] = mapExperience(exp)
	}

	c.JSON(http.StatusOK, PublicProfileResponse{
		Name:        user.Name,
		AvatarURL:   user.AvatarURL,
		Headline:    profile.Headline,
		About:       profile.About,
		YearsOfExp:  int32ToIntPtr(profile.YearsOfExperience),
		Experiences: experiences,
	})
}

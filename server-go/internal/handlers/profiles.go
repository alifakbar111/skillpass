package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"net/url"
	"regexp"

	"github.com/gin-gonic/gin"
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"skillpass-server-go/.gen/skillpass/public/model"
	"skillpass-server-go/internal/gen"
)

type UpdateProfileRequest struct {
	Headline          *string `json:"headline"`
	About             *string `json:"about"`
	YearsOfExperience *int    `json:"yearsOfExperience"`
	Slug              *string `json:"slug" binding:"omitempty,min=3,max=64"`
} //@name UpdateProfileRequest

type CreateExperienceRequest struct {
	Type         string   `json:"type" binding:"required,oneof=employment gig education certification project volunteering"`
	Title        string   `json:"title" binding:"required"`
	Organization string   `json:"organization" binding:"required"`
	StartDate    string   `json:"startDate" binding:"required"`
	EndDate      *string  `json:"endDate"`
	IsCurrent    *bool    `json:"isCurrent"`
	Description  *string  `json:"description"`
	Industry     *string  `json:"industry"`
	SkillsUsed   []string `json:"skillsUsed"`
	URL          *string  `json:"url"`
} //@name CreateExperienceRequest

type UpdateExperienceRequest struct {
	Type         *string  `json:"type" binding:"omitempty,oneof=employment gig education certification project volunteering"`
	Title        *string  `json:"title"`
	Organization *string  `json:"organization"`
	StartDate    *string  `json:"startDate"`
	EndDate      *string  `json:"endDate"`
	IsCurrent    *bool    `json:"isCurrent"`
	Description  *string  `json:"description"`
	Industry     *string  `json:"industry"`
	SkillsUsed   []string `json:"skillsUsed"`
	URL          *string  `json:"url"`
} //@name UpdateExperienceRequest

type Experience struct {
	ID           string   `json:"id"`
	ProfileID    string   `json:"profileId"`
	Type         string   `json:"type"`
	Title        string   `json:"title"`
	Organization string   `json:"organization"`
	StartDate    string   `json:"startDate"`
	EndDate      *string  `json:"endDate"`
	IsCurrent    bool     `json:"isCurrent"`
	Description  *string  `json:"description"`
	Industry     *string  `json:"industry"`
	SkillsUsed   []string `json:"skillsUsed"`
	URL          *string  `json:"url"`
} //@name Experience

type ProfileResponse struct {
	ID          string       `json:"id"`
	UserID      string       `json:"userId"`
	Headline    *string      `json:"headline"`
	About       *string      `json:"about"`
	YearsOfExp  *int         `json:"yearsOfExperience"`
	Slug        string       `json:"slug"`
	Name        string       `json:"name"`
	Email       string       `json:"email"`
	Username    string       `json:"username"`
	Role        string       `json:"role"`
	AvatarURL   *string      `json:"avatarUrl"`
	Experiences []Experience `json:"experiences"`
} //@name ProfileResponse

type ProfileHandler struct {
	db *sql.DB
}

func NewProfileHandler(db *sql.DB) *ProfileHandler {
	return &ProfileHandler{db: db}
}

func mapExperience(exp model.JobExperiences) Experience {
	skillsUsed := []string{}
	if exp.SkillsUsed != nil {
		skillsUsed = []string(*exp.SkillsUsed)
	}
	return Experience{
		ID:           exp.ID.String(),
		ProfileID:    exp.ProfileID.String(),
		Type:         string(exp.Type),
		Title:        exp.Title,
		Organization: exp.Organization,
		StartDate:    exp.StartDate,
		EndDate:      exp.EndDate,
		IsCurrent:    exp.IsCurrent,
		Description:  exp.Description,
		Industry:     exp.Industry,
		SkillsUsed:   skillsUsed,
		URL:          exp.URL,
	}
}

var slugPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{1,62}[a-z0-9])?$`)

var reservedSlugs = map[string]struct{}{
	"admin":       {},
	"administrator": {},
	"api":         {},
	"auth":        {},
	"company":     {},
	"companies":   {},
	"jobseeker":   {},
	"jobseekers":  {},
	"profile":     {},
	"profiles":    {},
	"public":      {},
	"login":       {},
	"register":    {},
	"logout":      {},
	"search":      {},
	"settings":    {},
	"help":        {},
	"about":       {},
	"terms":       {},
	"privacy":     {},
}

var experienceTypeMap = map[string]StringExpression{
	"employment":    gen.ExperienceTypeEmployment,
	"gig":           gen.ExperienceTypeGig,
	"education":     gen.ExperienceTypeEducation,
	"certification": gen.ExperienceTypeCertification,
	"project":       gen.ExperienceTypeProject,
	"volunteering":  gen.ExperienceTypeVolunteering,
}

func isValidSlug(s string) bool {
	if !slugPattern.MatchString(s) {
		return false
	}
	if _, ok := reservedSlugs[s]; ok {
		return false
	}
	return true
}

func int32ToIntPtr(v *int32) *int {
	if v == nil {
		return nil
	}
	val := int(*v)
	return &val
}

// GetMyProfile	godoc
// @Summary		Get own profile
// @Description	Get the authenticated jobseeker's full profile (including experiences)
// @Tags		profiles
// @Produce		json
// @Security	BearerAuth
// @Success		200 {object} ProfileResponse
// @Failure		401 {object} map[string]string
// @Failure		404 {object} map[string]string
// @Router		/profiles/me [get]
func (h *ProfileHandler) GetMyProfile(c *gin.Context) {
	userIDVal, ok := c.Get("userId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	profileStmt := SELECT(
		gen.JobseekerProfiles.ID, gen.JobseekerProfiles.UserID, gen.JobseekerProfiles.Headline,
		gen.JobseekerProfiles.About, gen.JobseekerProfiles.YearsOfExperience, gen.JobseekerProfiles.Slug,
	).FROM(
		gen.JobseekerProfiles,
	).WHERE(
		gen.JobseekerProfiles.UserID.EQ(UUID(uuid.MustParse(userIDStr))),
	)

	var profiles []model.JobseekerProfiles
	err := profileStmt.QueryContext(c.Request.Context(), h.db, &profiles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load profile"})
		return
	}
	if len(profiles) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}
	profile := profiles[0]

	userStmt := SELECT(
		gen.Users.Name, gen.Users.Email, gen.Users.Username, gen.Users.Role, gen.Users.AvatarURL,
	).FROM(
		gen.Users,
	).WHERE(
		gen.Users.ID.EQ(UUID(uuid.MustParse(userIDStr))),
	)

	var users []model.Users
	err = userStmt.QueryContext(c.Request.Context(), h.db, &users)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query experiences"})
		return
	}

	experiences := make([]Experience, len(exps))
	for i, exp := range exps {
		experiences[i] = mapExperience(exp)
	}

	c.JSON(http.StatusOK, ProfileResponse{
		ID:          profile.ID.String(),
		UserID:      profile.UserID.String(),
		Headline:    profile.Headline,
		About:       profile.About,
		YearsOfExp:  int32ToIntPtr(profile.YearsOfExperience),
		Slug:        profile.Slug,
		Name:        user.Name,
		Email:       user.Email,
		Username:    user.Username,
		Role:        string(user.Role),
		AvatarURL:   user.AvatarURL,
		Experiences: experiences,
	})
}

// UpdateMyProfile	godoc
// @Summary		Update own profile
// @Description	Update the authenticated jobseeker's profile fields (headline, about, years of experience, slug)
// @Tags		profiles
// @Accept		json
// @Produce		json
// @Security	BearerAuth
// @Param		body body UpdateProfileRequest true "Profile fields to update"
// @Success		200 {object} UpdateProfileResponse
// @Failure		400 {object} map[string]string
// @Failure		401 {object} map[string]string
// @Failure		404 {object} map[string]string
// @Failure		409 {object} map[string]string
// @Router		/profiles/me [put]
func (h *ProfileHandler) UpdateMyProfile(c *gin.Context) {
	userIDVal, ok := c.Get("userId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var setVals []interface{}
	if req.Headline != nil {
		setVals = append(setVals, gen.JobseekerProfiles.Headline.SET(String(*req.Headline)))
	}
	if req.About != nil {
		setVals = append(setVals, gen.JobseekerProfiles.About.SET(String(*req.About)))
	}
	if req.YearsOfExperience != nil {
		setVals = append(setVals, gen.JobseekerProfiles.YearsOfExperience.SET(Int64(int64(*req.YearsOfExperience))))
	}
	if req.Slug != nil {
		if !isValidSlug(*req.Slug) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid slug format or reserved word"})
			return
		}
		setVals = append(setVals, gen.JobseekerProfiles.Slug.SET(String(*req.Slug)))
	}

	if len(setVals) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	stmt := gen.JobseekerProfiles.UPDATE().SET(setVals[0], setVals[1:]...).WHERE(
		gen.JobseekerProfiles.UserID.EQ(UUID(uuid.MustParse(userIDStr))),
	).RETURNING(
		gen.JobseekerProfiles.ID, gen.JobseekerProfiles.UserID, gen.JobseekerProfiles.Headline,
		gen.JobseekerProfiles.About, gen.JobseekerProfiles.YearsOfExperience, gen.JobseekerProfiles.Slug,
	)

	var profiles []model.JobseekerProfiles
	err := stmt.QueryContext(c.Request.Context(), h.db, &profiles)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
			return
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "Slug already taken"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}
	if len(profiles) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}
	profile := profiles[0]

	c.JSON(http.StatusOK, UpdateProfileResponse{
		ID:                profile.ID.String(),
		UserID:            profile.UserID.String(),
		Headline:          profile.Headline,
		About:             profile.About,
		YearsOfExperience: int32ToIntPtr(profile.YearsOfExperience),
		Slug:              profile.Slug,
	})
}

// isValidExperienceURL validates that the URL uses http or https scheme only.
// This prevents stored XSS via javascript: or data: URLs in experience links.
func isValidExperienceURL(raw *string) bool {
	if raw == nil || *raw == "" {
		return true
	}
	u, err := url.Parse(*raw)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

// CreateExperience	godoc
// @Summary		Add experience entry
// @Description	Add a new employment, gig, education, certification, project, or volunteering entry to the profile
// @Tags		profiles
// @Accept		json
// @Produce		json
// @Security	BearerAuth
// @Param		body body CreateExperienceRequest true "Experience details"
// @Success		201 {object} Experience
// @Failure		400 {object} map[string]string
// @Failure		401 {object} map[string]string
// @Router		/profiles/me/experience [post]
func (h *ProfileHandler) CreateExperience(c *gin.Context) {
	userIDVal, ok := c.Get("userId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req CreateExperienceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if !isValidDate(req.StartDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "startDate must be YYYY-MM"})
		return
	}
	if req.EndDate != nil && *req.EndDate != "" && !isValidDate(*req.EndDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endDate must be YYYY-MM"})
		return
	}

	if !isValidExperienceURL(req.URL) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL must use http or https scheme"})
		return
	}

	profileStmt := SELECT(
		gen.JobseekerProfiles.ID,
	).FROM(
		gen.JobseekerProfiles,
	).WHERE(
		gen.JobseekerProfiles.UserID.EQ(UUID(uuid.MustParse(userIDStr))),
	)

	var profiles []model.JobseekerProfiles
	err := profileStmt.QueryContext(c.Request.Context(), h.db, &profiles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load profile"})
		return
	}
	if len(profiles) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}
	profile := profiles[0]

	isCurrent := false
	if req.IsCurrent != nil {
		isCurrent = *req.IsCurrent
	}

	insertStmt := gen.JobExperiences.INSERT(
		gen.JobExperiences.ProfileID, gen.JobExperiences.Type, gen.JobExperiences.Title,
		gen.JobExperiences.Organization, gen.JobExperiences.StartDate, gen.JobExperiences.EndDate,
		gen.JobExperiences.IsCurrent, gen.JobExperiences.Description, gen.JobExperiences.Industry,
		gen.JobExperiences.SkillsUsed, gen.JobExperiences.URL,
	).VALUES(
		profile.ID, req.Type, req.Title, req.Organization, req.StartDate, req.EndDate,
		isCurrent, req.Description, req.Industry, StringArray(req.SkillsUsed...), req.URL,
	).RETURNING(
		gen.JobExperiences.ID, gen.JobExperiences.ProfileID, gen.JobExperiences.Type,
		gen.JobExperiences.Title, gen.JobExperiences.Organization, gen.JobExperiences.StartDate,
		gen.JobExperiences.EndDate, gen.JobExperiences.IsCurrent, gen.JobExperiences.Description,
		gen.JobExperiences.Industry, gen.JobExperiences.SkillsUsed, gen.JobExperiences.URL,
	)

	var exp model.JobExperiences
	err = insertStmt.QueryContext(c.Request.Context(), h.db, &exp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create experience"})
		return
	}

	c.JSON(http.StatusCreated, mapExperience(exp))
}

// UpdateExperience	godoc
// @Summary		Update an experience entry
// @Description	Update specific fields of an experience entry owned by the authenticated jobseeker
// @Tags		profiles
// @Accept		json
// @Produce		json
// @Security	BearerAuth
// @Param		id path string true "Experience entry UUID"
// @Param		body body UpdateExperienceRequest true "Fields to update"
// @Success		200 {object} Experience
// @Failure		400 {object} map[string]string
// @Failure		401 {object} map[string]string
// @Failure		404 {object} map[string]string
// @Router		/profiles/me/experience/{id} [put]
func (h *ProfileHandler) UpdateExperience(c *gin.Context) {
	userIDVal, ok := c.Get("userId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	expID := c.Param("id")
	if _, err := uuid.Parse(expID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid experience id"})
		return
	}

	profileStmt := SELECT(
		gen.JobseekerProfiles.ID,
	).FROM(
		gen.JobseekerProfiles,
	).WHERE(
		gen.JobseekerProfiles.UserID.EQ(UUID(uuid.MustParse(userIDStr))),
	)

	var profiles []model.JobseekerProfiles
	err := profileStmt.QueryContext(c.Request.Context(), h.db, &profiles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load profile"})
		return
	}
	if len(profiles) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}
	profile := profiles[0]

	var req UpdateExperienceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.StartDate != nil && *req.StartDate != "" && !isValidDate(*req.StartDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "startDate must be YYYY-MM"})
		return
	}
	if req.EndDate != nil && *req.EndDate != "" && !isValidDate(*req.EndDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endDate must be YYYY-MM"})
		return
	}

	if !isValidExperienceURL(req.URL) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL must use http or https scheme"})
		return
	}

	var setVals []interface{}
	if req.Type != nil {
		if expr, ok := experienceTypeMap[*req.Type]; ok {
			setVals = append(setVals, gen.JobExperiences.Type.SET(expr))
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid experience type"})
			return
		}
	}
	if req.Title != nil {
		setVals = append(setVals, gen.JobExperiences.Title.SET(String(*req.Title)))
	}
	if req.Organization != nil {
		setVals = append(setVals, gen.JobExperiences.Organization.SET(String(*req.Organization)))
	}
	if req.StartDate != nil {
		setVals = append(setVals, gen.JobExperiences.StartDate.SET(String(*req.StartDate)))
	}
	if req.EndDate != nil {
		setVals = append(setVals, gen.JobExperiences.EndDate.SET(String(*req.EndDate)))
	}
	if req.IsCurrent != nil {
		setVals = append(setVals, gen.JobExperiences.IsCurrent.SET(Bool(*req.IsCurrent)))
	}
	if req.Description != nil {
		setVals = append(setVals, gen.JobExperiences.Description.SET(String(*req.Description)))
	}
	if req.Industry != nil {
		setVals = append(setVals, gen.JobExperiences.Industry.SET(String(*req.Industry)))
	}
	if req.SkillsUsed != nil {
		setVals = append(setVals, gen.JobExperiences.SkillsUsed.SET(StringArray(req.SkillsUsed...)))
	}
	if req.URL != nil {
		setVals = append(setVals, gen.JobExperiences.URL.SET(String(*req.URL)))
	}

	if len(setVals) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	stmt := gen.JobExperiences.UPDATE().SET(setVals[0], setVals[1:]...).WHERE(
		gen.JobExperiences.ID.EQ(UUID(uuid.MustParse(expID))).AND(
			gen.JobExperiences.ProfileID.EQ(UUID(profile.ID)),
		),
	).RETURNING(
		gen.JobExperiences.ID, gen.JobExperiences.ProfileID, gen.JobExperiences.Type,
		gen.JobExperiences.Title, gen.JobExperiences.Organization, gen.JobExperiences.StartDate,
		gen.JobExperiences.EndDate, gen.JobExperiences.IsCurrent, gen.JobExperiences.Description,
		gen.JobExperiences.Industry, gen.JobExperiences.SkillsUsed, gen.JobExperiences.URL,
	)

	var exps []model.JobExperiences
	err = stmt.QueryContext(c.Request.Context(), h.db, &exps)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update experience"})
		return
	}
	if len(exps) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Experience not found"})
		return
	}

	c.JSON(http.StatusOK, mapExperience(exps[0]))
}

// DeleteExperience	godoc
// @Summary		Delete an experience entry
// @Description	Delete an experience entry owned by the authenticated jobseeker
// @Tags		profiles
// @Produce		json
// @Security	BearerAuth
// @Param		id path string true "Experience entry UUID"
// @Success		200 {object} MessageResponse
// @Failure		400 {object} map[string]string
// @Failure		401 {object} map[string]string
// @Failure		404 {object} map[string]string
// @Router		/profiles/me/experience/{id} [delete]
func (h *ProfileHandler) DeleteExperience(c *gin.Context) {
	userIDVal, ok := c.Get("userId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	expID := c.Param("id")
	if _, err := uuid.Parse(expID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid experience id"})
		return
	}

	profileStmt := SELECT(
		gen.JobseekerProfiles.ID,
	).FROM(
		gen.JobseekerProfiles,
	).WHERE(
		gen.JobseekerProfiles.UserID.EQ(UUID(uuid.MustParse(userIDStr))),
	)

	var profiles []model.JobseekerProfiles
	err := profileStmt.QueryContext(c.Request.Context(), h.db, &profiles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load profile"})
		return
	}
	if len(profiles) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}
	profile := profiles[0]

	deleteStmt := gen.JobExperiences.DELETE().WHERE(
		gen.JobExperiences.ID.EQ(UUID(uuid.MustParse(expID))).AND(
			gen.JobExperiences.ProfileID.EQ(UUID(profile.ID)),
		),
	)

	result, err := deleteStmt.ExecContext(c.Request.Context(), h.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete experience"})
		return
	}
	ra, _ := result.RowsAffected()
	if ra == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Experience not found"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "Deleted"})
}

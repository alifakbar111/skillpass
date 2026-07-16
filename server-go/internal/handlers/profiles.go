package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
	"github.com/uptrace/bun"

	"skillpass-server-go/internal/lib"
	"skillpass-server-go/internal/models"
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
	Organization string   `json:"organization,omitempty"` // not required — project/volunteering may have no org; fallback to title
	StartDate    string   `json:"startDate" binding:"required"`
	EndDate      *string  `json:"endDate,omitempty"`
	IsCurrent    *bool    `json:"isCurrent,omitempty"`
	Description  *string  `json:"description,omitempty"`
	Industry     *string  `json:"industry,omitempty"`
	SkillsUsed   []string `json:"skillsUsed,omitempty"`
	URL          *string  `json:"url,omitempty"`
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

type ReorderExperienceRequest struct {
	Experiences []ReorderItem `json:"experiences,omitempty" binding:"required,min=1,dive"`
} //@name ReorderExperienceRequest

type ReorderItem struct {
	ID        string `json:"id" binding:"required"`
	SortOrder int    `json:"sortOrder" binding:"required"`
} //@name ReorderItem

type Experience struct {
	ID           string   `json:"id"`
	ProfileID    string   `json:"profileId"`
	Type         string   `json:"type"`
	Title        string   `json:"title"`
	Organization string   `json:"organization"`
	StartDate    string   `json:"startDate"`
	EndDate      *string  `json:"endDate,omitempty"`
	IsCurrent    bool     `json:"isCurrent,omitempty"`
	Description  *string  `json:"description,omitempty"`
	Industry     *string  `json:"industry,omitempty"`
	SkillsUsed   []string `json:"skillsUsed,omitempty"`
	URL          *string  `json:"url,omitempty"`
} //@name Experience

type ProfileResponse struct {
	ID          string       `json:"id"`
	UserID      string       `json:"userId"`
	Headline    *string      `json:"headline,omitempty"`
	About       *string      `json:"about,omitempty"`
	YearsOfExp  *int         `json:"yearsOfExperience,omitempty"`
	Slug        string       `json:"slug"`
	Name        string       `json:"name"`
	Email       string       `json:"email"`
	Username    string       `json:"username"`
	Role        string       `json:"role"`
	AvatarURL   *string      `json:"avatarUrl,omitempty"`
	Experiences []Experience `json:"experiences,omitempty"`
} //@name ProfileResponse

type ProfileHandler struct {
	db    *sql.DB
	bunDB *bun.DB
}

func NewProfileHandler(db *sql.DB, bunDB *bun.DB) *ProfileHandler {
	return &ProfileHandler{db: db, bunDB: bunDB}
}

func mapExperience(exp models.JobExperience) Experience {
	skillsUsed := []string{}
	if exp.SkillsUsed != nil {
		skillsUsed = []string(*exp.SkillsUsed)
	}
	return Experience{
		ID:           exp.ID.String(),
		ProfileID:    exp.ProfileID.String(),
		Type:         exp.Type,
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

	userUUID, err := lib.ParseUUID(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid user ID: %v", err)})
		return
	}

	var profile models.JobseekerProfile
	err = h.bunDB.NewSelect().Model(&profile).Where("user_id = ?", userUUID).Scan(c.Request.Context())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load profile"})
		return
	}

	var user models.User
	err = h.bunDB.NewSelect().Model(&user).Where("id = ?", userUUID).Scan(c.Request.Context())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
		return
	}

	var exps []models.JobExperience
	err = h.bunDB.NewSelect().Model(&exps).
		Where("profile_id = ?", profile.ID).
		Order("sort_order ASC", "start_date DESC").
		Scan(c.Request.Context())
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
		Role:        user.Role,
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

	userUUID, err := lib.ParseUUID(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid user ID: %v", err)})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	query := h.bunDB.NewUpdate().Model((*models.JobseekerProfile)(nil))
	hasFields := false
	if req.Headline != nil {
		query = query.Set("headline = ?", *req.Headline)
		hasFields = true
	}
	if req.About != nil {
		query = query.Set("about = ?", *req.About)
		hasFields = true
	}
	if req.YearsOfExperience != nil {
		query = query.Set("years_of_experience = ?", int32(*req.YearsOfExperience))
		hasFields = true
	}
	if req.Slug != nil {
		if !isValidSlug(*req.Slug) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid slug format or reserved word"})
			return
		}
		query = query.Set("slug = ?", *req.Slug)
		hasFields = true
	}

	if !hasFields {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	var profile models.JobseekerProfile
	err = query.Where("user_id = ?", userUUID).Returning("*").Scan(c.Request.Context(), &profile)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
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

// ReorderExperience	godoc
// @Summary		Reorder experiences
// @Description	Update sort_order for multiple experiences at once
// @Tags		profiles
// @Accept		json
// @Produce		json
// @Security	BearerAuth
// @Param		body body ReorderExperienceRequest true "Experience IDs with new sort orders"
// @Success		200 {object} MessageResponse
// @Failure		400 {object} map[string]string
// @Failure		401 {object} map[string]string
// @Router		/profiles/me/experience/reorder [put]
func (h *ProfileHandler) ReorderExperience(c *gin.Context) {
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

	userUUID, err := lib.ParseUUID(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req ReorderExperienceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var profile models.JobseekerProfile
	err = h.bunDB.NewSelect().Model(&profile).Column("id").Where("user_id = ?", userUUID).Scan(c.Request.Context())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load profile"})
		return
	}

	tx, err := h.bunDB.BeginTx(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	for _, item := range req.Experiences {
		expUUID, err := lib.ParseUUID(item.ID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid experience ID: %s", item.ID)})
			return
		}
		result, err := tx.NewUpdate().Model((*models.JobExperience)(nil)).
			Set("sort_order = ?", int32(item.SortOrder)).
			Where("id = ? AND profile_id = ?", expUUID, profile.ID).
			Exec(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update experience order"})
			return
		}
		ra, _ := result.RowsAffected()
		if ra == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Experience %s not found", item.ID)})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit reorder"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "Reordered"})
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

	userUUID, err := lib.ParseUUID(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid user ID: %v", err)})
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

	var profile models.JobseekerProfile
	err = h.bunDB.NewSelect().Model(&profile).Column("id").Where("user_id = ?", userUUID).Scan(c.Request.Context())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load profile"})
		return
	}

	isCurrent := false
	if req.IsCurrent != nil {
		isCurrent = *req.IsCurrent
	}

	org := req.Organization
	if org == "" {
		org = req.Title // DB column is NOT NULL; fallback for project/volunteering with no org
	}
	exp := &models.JobExperience{
		ProfileID:    profile.ID,
		Type:         req.Type,
		Title:        req.Title,
		Organization: org,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		IsCurrent:    isCurrent,
		Description:  req.Description,
		Industry:     req.Industry,
		URL:          req.URL,
	}
	if len(req.SkillsUsed) > 0 {
		arr := pq.StringArray(req.SkillsUsed)
		exp.SkillsUsed = &arr
	}

	err = h.bunDB.NewInsert().Model(exp).Returning("*").Scan(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create experience"})
		return
	}

	// Upsert skills into skills table for autocomplete (best-effort, non-blocking)
	for _, skill := range req.SkillsUsed {
		if skill == "" {
			continue
		}
		_, _ = h.bunDB.ExecContext(c.Request.Context(),
			`INSERT INTO skills (name) VALUES (?) ON CONFLICT (name) DO NOTHING`, skill)
	}

	c.JSON(http.StatusCreated, mapExperience(*exp))
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

	userUUID, err := lib.ParseUUID(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid user ID: %v", err)})
		return
	}

	expID := c.Param("id")
	expUUID, err := lib.ParseUUID(expID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid experience ID: %v", err)})
		return
	}

	var profile models.JobseekerProfile
	err = h.bunDB.NewSelect().Model(&profile).Column("id").Where("user_id = ?", userUUID).Scan(c.Request.Context())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load profile"})
		return
	}

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

	query := h.bunDB.NewUpdate().Model((*models.JobExperience)(nil))
	hasFields := false
	if req.Type != nil {
		query = query.Set("type = ?", *req.Type)
		hasFields = true
	}
	if req.Title != nil {
		query = query.Set("title = ?", *req.Title)
		hasFields = true
	}
	if req.Organization != nil {
		query = query.Set("organization = ?", *req.Organization)
		hasFields = true
	}
	if req.StartDate != nil {
		query = query.Set("start_date = ?", *req.StartDate)
		hasFields = true
	}
	if req.EndDate != nil {
		query = query.Set("end_date = ?", *req.EndDate)
		hasFields = true
	}
	if req.IsCurrent != nil {
		query = query.Set("is_current = ?", *req.IsCurrent)
		hasFields = true
	}
	if req.Description != nil {
		query = query.Set("description = ?", *req.Description)
		hasFields = true
	}
	if req.Industry != nil {
		query = query.Set("industry = ?", *req.Industry)
		hasFields = true
	}
	if req.SkillsUsed != nil {
		query = query.Set("skills_used = ?", pq.StringArray(req.SkillsUsed))
		hasFields = true
	}
	if req.URL != nil {
		query = query.Set("url = ?", *req.URL)
		hasFields = true
	}

	if !hasFields {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	var exp models.JobExperience
	err = query.Where("id = ? AND profile_id = ?", expUUID, profile.ID).Returning("*").Scan(c.Request.Context(), &exp)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Experience not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update experience"})
		return
	}

	// Upsert skills into skills table for autocomplete
	if req.SkillsUsed != nil {
		for _, skill := range req.SkillsUsed {
			if skill == "" {
				continue
			}
			_, _ = h.bunDB.ExecContext(c.Request.Context(),
				`INSERT INTO skills (name) VALUES (?) ON CONFLICT (name) DO NOTHING`, skill)
		}
	}

	c.JSON(http.StatusOK, mapExperience(exp))
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

	userUUID, err := lib.ParseUUID(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid user ID: %v", err)})
		return
	}

	expID := c.Param("id")
	expUUID, err := lib.ParseUUID(expID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid experience ID: %v", err)})
		return
	}

	var profile models.JobseekerProfile
	err = h.bunDB.NewSelect().Model(&profile).Column("id").Where("user_id = ?", userUUID).Scan(c.Request.Context())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load profile"})
		return
	}

	result, err := h.bunDB.NewDelete().Model((*models.JobExperience)(nil)).
		Where("id = ? AND profile_id = ?", expUUID, profile.ID).
		Exec(c.Request.Context())
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

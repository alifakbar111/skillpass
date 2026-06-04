package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	. "github.com/go-jet/jet/v2/postgres"

	"skillpass-server-go/.gen/skillpass/public/model"
	"skillpass-server-go/internal/gen"
)

type UpdateProfileRequest struct {
	Headline          *string `json:"headline"`
	About             *string `json:"about"`
	YearsOfExperience *int    `json:"yearsOfExperience"`
	Slug              *string `json:"slug"`
}

type CreateExperienceRequest struct {
	Type         string   `json:"type" binding:"required"`
	Title        string   `json:"title" binding:"required"`
	Organization string   `json:"organization" binding:"required"`
	StartDate    string   `json:"startDate" binding:"required"`
	EndDate      *string  `json:"endDate"`
	IsCurrent    *bool    `json:"isCurrent"`
	Description  *string  `json:"description"`
	Industry     *string  `json:"industry"`
	SkillsUsed   []string `json:"skillsUsed"`
	URL          *string  `json:"url"`
}

type UpdateExperienceRequest struct {
	Type         *string  `json:"type"`
	Title        *string  `json:"title"`
	Organization *string  `json:"organization"`
	StartDate    *string  `json:"startDate"`
	EndDate      *string  `json:"endDate"`
	IsCurrent    *bool    `json:"isCurrent"`
	Description  *string  `json:"description"`
	Industry     *string  `json:"industry"`
	SkillsUsed   []string `json:"skillsUsed"`
	URL          *string  `json:"url"`
}

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
}

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
}

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

func int32ToIntPtr(v *int32) *int {
	if v == nil {
		return nil
	}
	val := int(*v)
	return &val
}

func (h *ProfileHandler) GetMyProfile(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr := userID.(string)

	profileStmt := SELECT(
		gen.JobseekerProfiles.ID, gen.JobseekerProfiles.UserID, gen.JobseekerProfiles.Headline,
		gen.JobseekerProfiles.About, gen.JobseekerProfiles.YearsOfExperience, gen.JobseekerProfiles.Slug,
	).FROM(
		gen.JobseekerProfiles,
	).WHERE(
		gen.JobseekerProfiles.UserID.EQ(String(userIDStr)),
	)

	var profile model.JobseekerProfiles
	err := profileStmt.QueryContext(c.Request.Context(), h.db, &profile)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	userStmt := SELECT(
		gen.Users.Name, gen.Users.Email, gen.Users.Username, gen.Users.Role, gen.Users.AvatarURL,
	).FROM(
		gen.Users,
	).WHERE(
		gen.Users.ID.EQ(String(userIDStr)),
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

func (h *ProfileHandler) UpdateMyProfile(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr := userID.(string)

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
		setVals = append(setVals, gen.JobseekerProfiles.Slug.SET(String(*req.Slug)))
	}

	if len(setVals) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	stmt := gen.JobseekerProfiles.UPDATE().SET(setVals[0], setVals[1:]...).WHERE(
		gen.JobseekerProfiles.UserID.EQ(String(userIDStr)),
	).RETURNING(
		gen.JobseekerProfiles.ID, gen.JobseekerProfiles.UserID, gen.JobseekerProfiles.Headline,
		gen.JobseekerProfiles.About, gen.JobseekerProfiles.YearsOfExperience, gen.JobseekerProfiles.Slug,
	)

	var profile model.JobseekerProfiles
	err := stmt.QueryContext(c.Request.Context(), h.db, &profile)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":                profile.ID.String(),
		"userId":            profile.UserID.String(),
		"headline":          profile.Headline,
		"about":             profile.About,
		"yearsOfExperience": int32ToIntPtr(profile.YearsOfExperience),
		"slug":              profile.Slug,
	})
}

func (h *ProfileHandler) CreateExperience(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr := userID.(string)

	var req CreateExperienceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	profileStmt := SELECT(
		gen.JobseekerProfiles.ID,
	).FROM(
		gen.JobseekerProfiles,
	).WHERE(
		gen.JobseekerProfiles.UserID.EQ(String(userIDStr)),
	)

	var profile model.JobseekerProfiles
	err := profileStmt.QueryContext(c.Request.Context(), h.db, &profile)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

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

func (h *ProfileHandler) UpdateExperience(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr := userID.(string)
	expID := c.Param("id")

	profileStmt := SELECT(
		gen.JobseekerProfiles.ID,
	).FROM(
		gen.JobseekerProfiles,
	).WHERE(
		gen.JobseekerProfiles.UserID.EQ(String(userIDStr)),
	)

	var profile model.JobseekerProfiles
	err := profileStmt.QueryContext(c.Request.Context(), h.db, &profile)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	var req UpdateExperienceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var setVals []interface{}
	if req.Type != nil {
		setVals = append(setVals, gen.JobExperiences.Type.SET(String(*req.Type)))
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
		gen.JobExperiences.ID.EQ(String(expID)).AND(
			gen.JobExperiences.ProfileID.EQ(String(profile.ID.String())),
		),
	).RETURNING(
		gen.JobExperiences.ID, gen.JobExperiences.ProfileID, gen.JobExperiences.Type,
		gen.JobExperiences.Title, gen.JobExperiences.Organization, gen.JobExperiences.StartDate,
		gen.JobExperiences.EndDate, gen.JobExperiences.IsCurrent, gen.JobExperiences.Description,
		gen.JobExperiences.Industry, gen.JobExperiences.SkillsUsed, gen.JobExperiences.URL,
	)

	var exp model.JobExperiences
	err = stmt.QueryContext(c.Request.Context(), h.db, &exp)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Experience not found"})
		return
	}

	c.JSON(http.StatusOK, mapExperience(exp))
}

func (h *ProfileHandler) DeleteExperience(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr := userID.(string)
	expID := c.Param("id")

	profileStmt := SELECT(
		gen.JobseekerProfiles.ID,
	).FROM(
		gen.JobseekerProfiles,
	).WHERE(
		gen.JobseekerProfiles.UserID.EQ(String(userIDStr)),
	)

	var profile model.JobseekerProfiles
	err := profileStmt.QueryContext(c.Request.Context(), h.db, &profile)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	deleteStmt := gen.JobExperiences.DELETE().WHERE(
		gen.JobExperiences.ID.EQ(String(expID)).AND(
			gen.JobExperiences.ProfileID.EQ(String(profile.ID.String())),
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

	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

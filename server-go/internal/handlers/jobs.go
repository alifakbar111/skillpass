package handlers

import (
	"database/sql"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/uptrace/bun"

	"skillpass-server-go/internal/lib"
	"skillpass-server-go/internal/models"
)

const dateFormat = "2006-01"

var datePattern = regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])$`)

func isValidDate(s string) bool {
	if !datePattern.MatchString(s) {
		return false
	}
	_, err := time.Parse(dateFormat, s)
	return err == nil
}

type JobResponse struct {
	ID                 string    `json:"id"`
	CompanyID          string    `json:"companyId"`
	Title              string    `json:"title"`
	Description        string    `json:"description"`
	Industry           string    `json:"industry"`
	Tags               []string  `json:"tags,omitempty"`
	RequiredSkills     []string  `json:"requiredSkills,omitempty"`
	ExperienceLevel    *string   `json:"experienceLevel,omitempty"`
	Location           *string   `json:"location,omitempty"`
	SalaryRange        *string   `json:"salaryRange,omitempty"`
	Requirements       *string   `json:"requirements,omitempty"`
	Benefits           *string   `json:"benefits,omitempty"`
	YearsExperienceMin *int      `json:"yearsExperienceMin,omitempty"`
	YearsExperienceMax *int      `json:"yearsExperienceMax,omitempty"`
	Status             string    `json:"status"`
	IsFreshGradFriendly bool     `json:"isFreshGradFriendly"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
} //@name JobResponse

type CreateJobRequest struct {
	Title              string   `json:"title" binding:"required"`
	Description        string   `json:"description" binding:"required"`
	Industry           string   `json:"industry" binding:"required"`
	Tags               []string `json:"tags,omitempty"`
	RequiredSkills     []string `json:"requiredSkills,omitempty"`
	ExperienceLevel    *string  `json:"experienceLevel,omitempty" binding:"omitempty,oneof=entry mid senior lead"`
	Location           *string  `json:"location,omitempty"`
	SalaryRange        *string  `json:"salaryRange,omitempty"`
	Requirements       *string  `json:"requirements,omitempty"`
	Benefits           *string  `json:"benefits,omitempty"`
	YearsExperienceMin *int     `json:"yearsExperienceMin,omitempty"`
	YearsExperienceMax *int     `json:"yearsExperienceMax,omitempty"`
	IsFreshGradFriendly bool    `json:"isFreshGradFriendly"`
} //@name CreateJobRequest

type UpdateJobRequest struct {
	Title              *string  `json:"title" binding:"omitempty,min=1"`
	Description        *string  `json:"description" binding:"omitempty,min=1"`
	Industry           *string  `json:"industry" binding:"omitempty,min=1"`
	Tags               []string `json:"tags"`
	RequiredSkills     []string `json:"requiredSkills"`
	ExperienceLevel    *string  `json:"experienceLevel" binding:"omitempty,oneof=entry mid senior lead"`
	Location           *string  `json:"location"`
	SalaryRange        *string  `json:"salaryRange"`
	Requirements       *string  `json:"requirements"`
	Benefits           *string  `json:"benefits"`
	YearsExperienceMin *int     `json:"yearsExperienceMin"`
	YearsExperienceMax *int     `json:"yearsExperienceMax"`
	IsFreshGradFriendly *bool   `json:"isFreshGradFriendly"`
	Status             *string  `json:"status" binding:"omitempty,oneof=open closed"`
} //@name UpdateJobRequest

type JobHandler struct {
	bunDB *bun.DB
}

func NewJobHandler(bunDB *bun.DB) *JobHandler {
	return &JobHandler{bunDB: bunDB}
}

func jobFromModel(j models.JobPosting) JobResponse {
	var tags []string
	if j.Tags != nil {
		tags = []string(*j.Tags)
	}
	var requiredSkills []string
	if j.RequiredSkills != nil {
		requiredSkills = []string(*j.RequiredSkills)
	}
	var yearsExpMin *int
	if j.YearsExperienceMin != nil {
		v := int(*j.YearsExperienceMin)
		yearsExpMin = &v
	}
	var yearsExpMax *int
	if j.YearsExperienceMax != nil {
		v := int(*j.YearsExperienceMax)
		yearsExpMax = &v
	}
	return JobResponse{
		ID:                  j.ID.String(),
		CompanyID:           j.CompanyID.String(),
		Title:               j.Title,
		Description:         j.Description,
		Industry:            j.Industry,
		Tags:                tags,
		RequiredSkills:      requiredSkills,
		ExperienceLevel:     j.ExperienceLevel,
		Location:            j.Location,
		SalaryRange:         j.SalaryRange,
		Requirements:        j.Requirements,
		Benefits:            j.Benefits,
		YearsExperienceMin:  yearsExpMin,
		YearsExperienceMax:  yearsExpMax,
		IsFreshGradFriendly: j.IsFreshGradFriendly,
		Status:              j.Status,
		CreatedAt:           j.CreatedAt,
		UpdatedAt:           j.UpdatedAt,
	}
}

func parseJobLimit(c *gin.Context) int64 {
	limit := int64(defaultListLimit)
	if v, err := strconv.ParseInt(c.Query("limit"), 10, 64); err == nil && v > 0 {
		limit = v
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}
	return limit
}

func parseJobOffset(c *gin.Context) int64 {
	offset := int64(0)
	if v, err := strconv.ParseInt(c.Query("offset"), 10, 64); err == nil && v >= 0 {
		offset = v
	}
	return offset
}

func int32Ptr(v *int) *int32 {
	if v == nil {
		return nil
	}
	val := int32(*v)
	return &val
}

func pqToStringArray(s []string) *pq.StringArray {
	if s == nil {
		return nil
	}
	v := pq.StringArray(s)
	return &v
}

// ListJobs		godoc
// @Summary		List open job postings
// @Description	Get all open job postings with optional filters
// @Tags		jobs
// @Produce		json
// @Param		industry query string false "Filter by industry name"
// @Param		experience_level query string false "Filter by experience level (entry, mid, senior, lead)"
// @Param		limit query int false "Max results (default 50, max 200)"
// @Param		offset query int false "Result offset for pagination"
// @Success		200 {array} JobResponse
// @Router		/jobs [get]
func (h *JobHandler) ListJobs(c *gin.Context) {
	var jobs []models.JobPosting
	query := h.bunDB.NewSelect().Model(&jobs).Where("status = ?", "open")

	if industry := c.Query("industry"); industry != "" {
		query = query.Where("industry = ?", industry)
	}
	if expLevel := c.Query("experience_level"); expLevel != "" {
		query = query.Where("experience_level = ?", expLevel)
	}

	query = query.Order("created_at DESC").Limit(int(parseJobLimit(c))).Offset(int(parseJobOffset(c)))

	if err := query.Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query jobs"})
		return
	}

	result := make([]JobResponse, len(jobs))
	for i, j := range jobs {
		result[i] = jobFromModel(j)
	}

	c.JSON(http.StatusOK, result)
}

// GetJob		godoc
// @Summary		Get a job posting
// @Description	Get a single job posting by ID
// @Tags		jobs
// @Produce		json
// @Param		id path string true "Job posting UUID"
// @Success		200 {object} JobResponse
// @Failure		404 {object} map[string]string
// @Router		/jobs/{id} [get]
func (h *JobHandler) GetJob(c *gin.Context) {
	id := c.Param("id")
	jobUUID, err := lib.ParseUUID(id)
	if err != nil {
		slog.Warn("invalid job ID", "raw", c.Param("id"), "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	var job models.JobPosting
	err = h.bunDB.NewSelect().Model(&job).Where("id = ?", jobUUID).Scan(c.Request.Context())
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query job"})
		return
	}

	c.JSON(http.StatusOK, jobFromModel(job))
}

// ListMyJobs	godoc
// @Summary		List company's job postings
// @Description	Get all job postings for the authenticated company
// @Tags		jobs
// @Produce		json
// @Security	BearerAuth
// @Success		200 {array} JobResponse
// @Router		/jobs/me [get]
func (h *JobHandler) ListMyJobs(c *gin.Context) {
	companyIDVal, ok := c.Get("companyId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	companyIDStr, ok := companyIDVal.(string)
	if !ok || companyIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	companyUUID, err := lib.ParseUUID(companyIDStr)
	if err != nil {
		slog.Warn("invalid company ID", "raw", c.Param("id"), "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	var jobs []models.JobPosting
	err = h.bunDB.NewSelect().Model(&jobs).
		Where("company_id = ?", companyUUID).
		Order("created_at DESC").
		Scan(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query jobs"})
		return
	}

	result := make([]JobResponse, len(jobs))
	for i, j := range jobs {
		result[i] = jobFromModel(j)
	}

	c.JSON(http.StatusOK, result)
}

// CreateJob	godoc
// @Summary		Create a job posting
// @Description	Create a new job posting for the authenticated company
// @Tags		jobs
// @Accept		json
// @Produce		json
// @Security	BearerAuth
// @Param		body body CreateJobRequest true "Job posting details"
// @Success		201 {object} JobResponse
// @Failure		400 {object} map[string]string
// @Router		/jobs [post]
func (h *JobHandler) CreateJob(c *gin.Context) {
	companyIDVal, ok := c.Get("companyId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	companyIDStr, ok := companyIDVal.(string)
	if !ok || companyIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	companyUUID, err := lib.ParseUUID(companyIDStr)
	if err != nil {
		slog.Warn("invalid company ID", "raw", c.Param("id"), "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	var req CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	job := &models.JobPosting{
		CompanyID:           companyUUID,
		Title:               req.Title,
		Description:         req.Description,
		Industry:            req.Industry,
		Tags:                pqToStringArray(req.Tags),
		RequiredSkills:      pqToStringArray(req.RequiredSkills),
		ExperienceLevel:     req.ExperienceLevel,
		Location:            req.Location,
		SalaryRange:         req.SalaryRange,
		Requirements:        req.Requirements,
		Benefits:            req.Benefits,
		YearsExperienceMin:  int32Ptr(req.YearsExperienceMin),
		YearsExperienceMax:  int32Ptr(req.YearsExperienceMax),
		IsFreshGradFriendly: req.IsFreshGradFriendly,
	}

	err = h.bunDB.NewInsert().Model(job).Returning("*").Scan(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create job"})
		return
	}

	c.JSON(http.StatusCreated, jobFromModel(*job))
}

// UpdateJob	godoc
// @Summary		Update a job posting
// @Description	Update specific fields of a job posting owned by the authenticated company
// @Tags		jobs
// @Accept		json
// @Produce		json
// @Security	BearerAuth
// @Param		id path string true "Job posting UUID"
// @Param		body body UpdateJobRequest true "Fields to update"
// @Success		200 {object} JobResponse
// @Failure		400 {object} map[string]string
// @Failure		404 {object} map[string]string
// @Router		/jobs/{id} [put]
func (h *JobHandler) UpdateJob(c *gin.Context) {
	id := c.Param("id")
	jobUUID, err := lib.ParseUUID(id)
	if err != nil {
		slog.Warn("invalid job ID", "raw", c.Param("id"), "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}
	companyIDVal, ok := c.Get("companyId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	companyIDStr, ok := companyIDVal.(string)
	if !ok || companyIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	companyUUID, err := lib.ParseUUID(companyIDStr)
	if err != nil {
		slog.Warn("invalid company ID", "raw", c.Param("id"), "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	var req UpdateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	query := h.bunDB.NewUpdate().Model((*models.JobPosting)(nil))
	hasUpdates := false

	if req.Title != nil {
		query = query.Set("title = ?", *req.Title)
		hasUpdates = true
	}
	if req.Description != nil {
		query = query.Set("description = ?", *req.Description)
		hasUpdates = true
	}
	if req.Industry != nil {
		query = query.Set("industry = ?", *req.Industry)
		hasUpdates = true
	}
	if req.Tags != nil {
		query = query.Set("tags = ?", pq.StringArray(req.Tags))
		hasUpdates = true
	}
	if req.RequiredSkills != nil {
		query = query.Set("required_skills = ?", pq.StringArray(req.RequiredSkills))
		hasUpdates = true
	}
	if req.ExperienceLevel != nil {
		query = query.Set("experience_level = ?", *req.ExperienceLevel)
		hasUpdates = true
	}
	if req.Location != nil {
		query = query.Set("location = ?", *req.Location)
		hasUpdates = true
	}
	if req.SalaryRange != nil {
		query = query.Set("salary_range = ?", *req.SalaryRange)
		hasUpdates = true
	}
	if req.Requirements != nil {
		query = query.Set("requirements = ?", *req.Requirements)
		hasUpdates = true
	}
	if req.Benefits != nil {
		query = query.Set("benefits = ?", *req.Benefits)
		hasUpdates = true
	}
	if req.YearsExperienceMin != nil {
		query = query.Set("years_experience_min = ?", *req.YearsExperienceMin)
		hasUpdates = true
	}
	if req.YearsExperienceMax != nil {
		query = query.Set("years_experience_max = ?", *req.YearsExperienceMax)
		hasUpdates = true
	}
	if req.IsFreshGradFriendly != nil {
		query = query.Set("is_fresh_grad_friendly = ?", *req.IsFreshGradFriendly)
		hasUpdates = true
	}
	if req.Status != nil {
		query = query.Set("status = ?", *req.Status)
		hasUpdates = true
	}

	if !hasUpdates {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	// Always update updated_at on any actual update
	query = query.Set("updated_at = NOW()")

	var jobs []models.JobPosting
	err = query.Where("id = ? AND company_id = ?", jobUUID, companyUUID).
		Returning("*").
		Scan(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update job"})
		return
	}
	if len(jobs) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	c.JSON(http.StatusOK, jobFromModel(jobs[0]))
}

// DeleteJob	godoc
// @Summary		Delete a job posting
// @Description	Delete a job posting owned by the authenticated company
// @Tags		jobs
// @Produce		json
// @Security	BearerAuth
// @Param		id path string true "Job posting UUID"
// @Success		200 {object} MessageResponse
// @Failure		404 {object} map[string]string
// @Router		/jobs/{id} [delete]
func (h *JobHandler) DeleteJob(c *gin.Context) {
	id := c.Param("id")
	jobUUID, err := lib.ParseUUID(id)
	if err != nil {
		slog.Warn("invalid job ID", "raw", c.Param("id"), "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}
	companyIDVal, ok := c.Get("companyId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	companyIDStr, ok := companyIDVal.(string)
	if !ok || companyIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	companyUUID, err := lib.ParseUUID(companyIDStr)
	if err != nil {
		slog.Warn("invalid company ID", "raw", c.Param("id"), "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	result, err := h.bunDB.NewDelete().Model((*models.JobPosting)(nil)).
		Where("id = ? AND company_id = ?", jobUUID, companyUUID).
		Exec(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete job"})
		return
	}
	ra, _ := result.RowsAffected()
	if ra == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	c.JSON(http.StatusOK, MessageResponse{Message: "Deleted"})
}

package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"

	"skillpass-server-go/.gen/skillpass/public/model"
	"skillpass-server-go/internal/gen"
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

var experienceLevelMap = map[string]StringExpression{
	"entry":  gen.ExperienceLevelEntry,
	"mid":    gen.ExperienceLevelMid,
	"senior": gen.ExperienceLevelSenior,
	"lead":   gen.ExperienceLevelLead,
}

var jobStatusMap = map[string]StringExpression{
	"open":   gen.JobStatusOpen,
	"closed": gen.JobStatusClosed,
}

type JobResponse struct {
	ID              string    `json:"id"`
	CompanyID       string    `json:"companyId"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Industry        string    `json:"industry"`
	Tags            []string  `json:"tags"`
	RequiredSkills  []string  `json:"requiredSkills"`
	ExperienceLevel *string   `json:"experienceLevel"`
	Location        *string   `json:"location"`
	SalaryRange     *string   `json:"salaryRange"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"createdAt"`
} //@name JobResponse

type CreateJobRequest struct {
	Title           string   `json:"title" binding:"required"`
	Description     string   `json:"description" binding:"required"`
	Industry        string   `json:"industry" binding:"required"`
	Tags            []string `json:"tags"`
	RequiredSkills  []string `json:"requiredSkills"`
	ExperienceLevel *string  `json:"experienceLevel" binding:"omitempty,oneof=entry mid senior lead"`
	Location        *string  `json:"location"`
	SalaryRange     *string  `json:"salaryRange"`
} //@name CreateJobRequest

type UpdateJobRequest struct {
	Title           *string  `json:"title" binding:"omitempty,min=1"`
	Description     *string  `json:"description" binding:"omitempty,min=1"`
	Industry        *string  `json:"industry" binding:"omitempty,min=1"`
	Tags            []string `json:"tags"`
	RequiredSkills  []string `json:"requiredSkills"`
	ExperienceLevel *string  `json:"experienceLevel" binding:"omitempty,oneof=entry mid senior lead"`
	Location        *string  `json:"location"`
	SalaryRange     *string  `json:"salaryRange"`
	Status          *string  `json:"status" binding:"omitempty,oneof=open closed"`
} //@name UpdateJobRequest

type JobHandler struct {
	db *sql.DB
}

func NewJobHandler(db *sql.DB) *JobHandler {
	return &JobHandler{db: db}
}

func jobFromModel(j model.JobPostings) JobResponse {
	var tags []string
	if j.Tags != nil {
		tags = []string(*j.Tags)
	}
	var requiredSkills []string
	if j.RequiredSkills != nil {
		requiredSkills = []string(*j.RequiredSkills)
	}
	var expLevel *string
	if j.ExperienceLevel != nil {
		v := string(*j.ExperienceLevel)
		expLevel = &v
	}
	return JobResponse{
		ID:              j.ID.String(),
		CompanyID:       j.CompanyID.String(),
		Title:           j.Title,
		Description:     j.Description,
		Industry:        j.Industry,
		Tags:            tags,
		RequiredSkills:  requiredSkills,
		ExperienceLevel: expLevel,
		Location:        j.Location,
		SalaryRange:     j.SalaryRange,
		Status:          string(j.Status),
		CreatedAt:       j.CreatedAt,
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
	var whereCond BoolExpression = gen.JobPostings.Status.EQ(gen.JobStatusOpen)

	if industry := c.Query("industry"); industry != "" {
		whereCond = whereCond.AND(gen.JobPostings.Industry.EQ(String(industry)))
	}
	if expLevel := c.Query("experience_level"); expLevel != "" {
		if expr, ok := experienceLevelMap[expLevel]; ok {
			whereCond = whereCond.AND(gen.JobPostings.ExperienceLevel.EQ(expr))
		}
	}

	stmt := SELECT(
		gen.JobPostings.AllColumns,
	).FROM(
		gen.JobPostings,
	).WHERE(
		whereCond,
	).ORDER_BY(
		gen.JobPostings.CreatedAt.DESC(),
	).LIMIT(parseJobLimit(c)).OFFSET(parseJobOffset(c))

	var jobs []model.JobPostings
	if err := stmt.QueryContext(c.Request.Context(), h.db, &jobs); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusOK, []JobResponse{})
			return
		}
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
	jobUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job id"})
		return
	}

	stmt := SELECT(
		gen.JobPostings.AllColumns,
	).FROM(
		gen.JobPostings,
	).WHERE(
		gen.JobPostings.ID.EQ(UUID(jobUUID)),
	)

	var jobs []model.JobPostings
	if err := stmt.QueryContext(c.Request.Context(), h.db, &jobs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query job"})
		return
	}
	if len(jobs) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	c.JSON(http.StatusOK, jobFromModel(jobs[0]))
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

	companyUUID, err := uuid.Parse(companyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	stmt := SELECT(
		gen.JobPostings.AllColumns,
	).FROM(
		gen.JobPostings,
	).WHERE(
		gen.JobPostings.CompanyID.EQ(UUID(companyUUID)),
	).ORDER_BY(
		gen.JobPostings.CreatedAt.DESC(),
	)

	var jobs []model.JobPostings
	if err := stmt.QueryContext(c.Request.Context(), h.db, &jobs); err != nil {
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

	var req CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	stmt := gen.JobPostings.INSERT(
		gen.JobPostings.CompanyID, gen.JobPostings.Title, gen.JobPostings.Description,
		gen.JobPostings.Industry, gen.JobPostings.Tags, gen.JobPostings.RequiredSkills,
		gen.JobPostings.ExperienceLevel, gen.JobPostings.Location, gen.JobPostings.SalaryRange,
	).VALUES(
		companyIDStr, req.Title, req.Description, req.Industry,
		StringArray(req.Tags...), StringArray(req.RequiredSkills...),
		req.ExperienceLevel, req.Location, req.SalaryRange,
	).RETURNING(
		gen.JobPostings.AllColumns,
	)

	var job model.JobPostings
	if err := stmt.QueryContext(c.Request.Context(), h.db, &job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create job"})
		return
	}

	c.JSON(http.StatusCreated, jobFromModel(job))
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
	jobUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job id"})
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

	companyUUID, err := uuid.Parse(companyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	var req UpdateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var setVals []interface{}
	if req.Title != nil {
		setVals = append(setVals, gen.JobPostings.Title.SET(String(*req.Title)))
	}
	if req.Description != nil {
		setVals = append(setVals, gen.JobPostings.Description.SET(String(*req.Description)))
	}
	if req.Industry != nil {
		setVals = append(setVals, gen.JobPostings.Industry.SET(String(*req.Industry)))
	}
	if req.Tags != nil {
		setVals = append(setVals, gen.JobPostings.Tags.SET(StringArray(req.Tags...)))
	}
	if req.RequiredSkills != nil {
		setVals = append(setVals, gen.JobPostings.RequiredSkills.SET(StringArray(req.RequiredSkills...)))
	}
	if req.ExperienceLevel != nil {
		if expr, ok := experienceLevelMap[*req.ExperienceLevel]; ok {
			setVals = append(setVals, gen.JobPostings.ExperienceLevel.SET(expr))
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid experience level"})
			return
		}
	}
	if req.Location != nil {
		setVals = append(setVals, gen.JobPostings.Location.SET(String(*req.Location)))
	}
	if req.SalaryRange != nil {
		setVals = append(setVals, gen.JobPostings.SalaryRange.SET(String(*req.SalaryRange)))
	}
	if req.Status != nil {
		if expr, ok := jobStatusMap[*req.Status]; ok {
			setVals = append(setVals, gen.JobPostings.Status.SET(expr))
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
			return
		}
	}

	if len(setVals) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	stmt := gen.JobPostings.UPDATE().SET(setVals[0], setVals[1:]...	).WHERE(
		gen.JobPostings.ID.EQ(UUID(jobUUID)).AND(
			gen.JobPostings.CompanyID.EQ(UUID(companyUUID)),
		),
	).RETURNING(
		gen.JobPostings.AllColumns,
	)

	var jobs []model.JobPostings
	if err := stmt.QueryContext(c.Request.Context(), h.db, &jobs); err != nil {
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
	jobUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job id"})
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

	companyUUID, err := uuid.Parse(companyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	stmt := gen.JobPostings.DELETE().WHERE(
		gen.JobPostings.ID.EQ(UUID(jobUUID)).AND(
			gen.JobPostings.CompanyID.EQ(UUID(companyUUID)),
		),
	)

	result, err := stmt.ExecContext(c.Request.Context(), h.db)
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

package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	. "github.com/go-jet/jet/v2/postgres"
	"database/sql"

	"skillpass-server-go/.gen/skillpass/public/model"
	"skillpass-server-go/internal/gen"
)

type JobResponse struct {
	ID              string   `json:"id"`
	CompanyID       string   `json:"companyId"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	Industry        string   `json:"industry"`
	Tags            []string `json:"tags"`
	RequiredSkills  []string `json:"requiredSkills"`
	ExperienceLevel *string  `json:"experienceLevel"`
	Location        *string  `json:"location"`
	SalaryRange     *string  `json:"salaryRange"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"createdAt"`
}

type CreateJobRequest struct {
	Title           string   `json:"title" binding:"required"`
	Description     string   `json:"description" binding:"required"`
	Industry        string   `json:"industry" binding:"required"`
	Tags            []string `json:"tags"`
	RequiredSkills  []string `json:"requiredSkills"`
	ExperienceLevel *string  `json:"experienceLevel"`
	Location        *string  `json:"location"`
	SalaryRange     *string  `json:"salaryRange"`
}

type UpdateJobRequest struct {
	Title           *string   `json:"title"`
	Description     *string   `json:"description"`
	Industry        *string   `json:"industry"`
	Tags            []string  `json:"tags"`
	RequiredSkills  []string  `json:"requiredSkills"`
	ExperienceLevel *string   `json:"experienceLevel"`
	Location        *string   `json:"location"`
	SalaryRange     *string   `json:"salaryRange"`
	Status          *string   `json:"status"`
}

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

func (h *JobHandler) ListJobs(c *gin.Context) {
	var whereCond BoolExpression = gen.JobPostings.Status.EQ(String("open"))

	if industry := c.Query("industry"); industry != "" {
		whereCond = whereCond.AND(gen.JobPostings.Industry.EQ(String(industry)))
	}
	if expLevel := c.Query("experience_level"); expLevel != "" {
		whereCond = whereCond.AND(gen.JobPostings.ExperienceLevel.EQ(String(expLevel)))
	}

	stmt := SELECT(
		gen.JobPostings.AllColumns,
	).FROM(
		gen.JobPostings,
	).WHERE(
		whereCond,
	).ORDER_BY(
		gen.JobPostings.CreatedAt.ASC(),
	)

	var jobs []model.JobPostings
	err := stmt.QueryContext(c.Request.Context(), h.db, &jobs)
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

func (h *JobHandler) GetJob(c *gin.Context) {
	id := c.Param("id")

	stmt := SELECT(
		gen.JobPostings.AllColumns,
	).FROM(
		gen.JobPostings,
	).WHERE(
		gen.JobPostings.ID.EQ(String(id)),
	)

	var job model.JobPostings
	err := stmt.QueryContext(c.Request.Context(), h.db, &job)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	c.JSON(http.StatusOK, jobFromModel(job))
}

func (h *JobHandler) ListMyJobs(c *gin.Context) {
	companyID, _ := c.Get("companyId")
	companyIDStr := companyID.(string)

	stmt := SELECT(
		gen.JobPostings.AllColumns,
	).FROM(
		gen.JobPostings,
	).WHERE(
		gen.JobPostings.CompanyID.EQ(String(companyIDStr)),
	).ORDER_BY(
		gen.JobPostings.CreatedAt.ASC(),
	)

	var jobs []model.JobPostings
	err := stmt.QueryContext(c.Request.Context(), h.db, &jobs)
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

func (h *JobHandler) CreateJob(c *gin.Context) {
	companyID, _ := c.Get("companyId")
	companyIDStr := companyID.(string)

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
	err := stmt.QueryContext(c.Request.Context(), h.db, &job)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create job"})
		return
	}

	c.JSON(http.StatusCreated, jobFromModel(job))
}

func (h *JobHandler) UpdateJob(c *gin.Context) {
	id := c.Param("id")
	companyID, _ := c.Get("companyId")
	companyIDStr := companyID.(string)

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
		setVals = append(setVals, gen.JobPostings.ExperienceLevel.SET(String(*req.ExperienceLevel)))
	}
	if req.Location != nil {
		setVals = append(setVals, gen.JobPostings.Location.SET(String(*req.Location)))
	}
	if req.SalaryRange != nil {
		setVals = append(setVals, gen.JobPostings.SalaryRange.SET(String(*req.SalaryRange)))
	}
	if req.Status != nil {
		setVals = append(setVals, gen.JobPostings.Status.SET(String(*req.Status)))
	}

	if len(setVals) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	stmt := gen.JobPostings.UPDATE().SET(setVals[0], setVals[1:]...).WHERE(
		gen.JobPostings.ID.EQ(String(id)).AND(
			gen.JobPostings.CompanyID.EQ(String(companyIDStr)),
		),
	).RETURNING(
		gen.JobPostings.AllColumns,
	)

	var job model.JobPostings
	err := stmt.QueryContext(c.Request.Context(), h.db, &job)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	c.JSON(http.StatusOK, jobFromModel(job))
}

func (h *JobHandler) DeleteJob(c *gin.Context) {
	id := c.Param("id")
	companyID, _ := c.Get("companyId")
	companyIDStr := companyID.(string)

	stmt := gen.JobPostings.DELETE().WHERE(
		gen.JobPostings.ID.EQ(String(id)).AND(
			gen.JobPostings.CompanyID.EQ(String(companyIDStr)),
		),
	)

	result, err := stmt.ExecContext(c.Request.Context(), h.db)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}
	ra, _ := result.RowsAffected()
	if ra == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

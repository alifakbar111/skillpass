package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"

	"skillpass-server-go/internal/gen"
)

type ReferenceHandler struct {
	db *sql.DB
}

func NewReferenceHandler(db *sql.DB) *ReferenceHandler {
	return &ReferenceHandler{db: db}
}

// IndustryResponse is the camelCase JSON shape for an industry category.
type IndustryResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
}

// GetIndustries	godoc
// @Summary		List industries
// @Description	Get all industry categories for filtering
// @Tags		reference
// @Produce		json
// @Success		200 {array} handlers.IndustryResponse
// @Router		/industries [get]
func (h *ReferenceHandler) GetIndustries(c *gin.Context) {
	stmt := SELECT(
		gen.IndustryCategories.AllColumns,
	).FROM(
		gen.IndustryCategories,
	).ORDER_BY(
		gen.IndustryCategories.Name,
	)

	var dest []gen.IndustryCategory
	err := stmt.QueryContext(c.Request.Context(), h.db, &dest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query industries"})
		return
	}

	resp := make([]IndustryResponse, 0, len(dest))
	for _, ind := range dest {
		resp = append(resp, IndustryResponse{
			ID:          ind.ID,
			Name:        ind.Name,
			Description: ind.Description,
		})
	}

	c.JSON(http.StatusOK, resp)
}

// TagResponse is the camelCase JSON shape for a skill tag.
type TagResponse struct {
	ID                 uuid.UUID  `json:"id"`
	Name               string     `json:"name"`
	IndustryCategoryID *uuid.UUID `json:"industryCategoryId"`
}

// GetTags		godoc
// @Summary		List tags
// @Description	Get skill tags, optionally filtered by industry category
// @Tags		reference
// @Produce		json
// @Param		industry query string false "Industry category ID to filter tags by"
// @Success		200 {array} handlers.TagResponse
// @Failure		400 {object} map[string]string
// @Router		/tags [get]
func (h *ReferenceHandler) GetTags(c *gin.Context) {
	industryID := c.Query("industry")

	var stmt SelectStatement
	if industryID != "" {
		parsed, err := uuid.Parse(industryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid industry ID"})
			return
		}

		stmt = SELECT(
			gen.Tags.AllColumns,
		).FROM(
			gen.Tags,
		).WHERE(
			gen.Tags.IndustryCategoryID.EQ(UUID(parsed)),
		).ORDER_BY(
			gen.Tags.Name,
		)
	} else {
		stmt = SELECT(
			gen.Tags.AllColumns,
		).FROM(
			gen.Tags,
		).ORDER_BY(
			gen.Tags.Name,
		)
	}

	var dest []gen.Tag
	err := stmt.QueryContext(c.Request.Context(), h.db, &dest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query tags"})
		return
	}

	resp := make([]TagResponse, 0, len(dest))
	for _, t := range dest {
		resp = append(resp, TagResponse{
			ID:                 t.ID,
			Name:               t.Name,
			IndustryCategoryID: t.IndustryCategoryID,
		})
	}

	c.JSON(http.StatusOK, resp)
}

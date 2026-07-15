package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"skillpass-server-go/internal/models"
)

type ReferenceHandler struct {
	bunDB *bun.DB
}

func NewReferenceHandler(bunDB *bun.DB) *ReferenceHandler {
	return &ReferenceHandler{bunDB: bunDB}
}

// IndustryResponse is the camelCase JSON shape for an industry category.
type IndustryResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
} //@name IndustryResponse

// GetIndustries	godoc
// @Summary		List industries
// @Description	Get all industry categories for filtering
// @Tags		reference
// @Produce		json
// @Success		200 {array} handlers.IndustryResponse
// @Router		/industries [get]
func (h *ReferenceHandler) GetIndustries(c *gin.Context) {
	var dest []models.IndustryCategory
	err := h.bunDB.NewSelect().
		Model(&dest).
		Order("name ASC").
		Scan(c.Request.Context())
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
	IndustryCategoryID *uuid.UUID `json:"industryCategoryId,omitempty"`
} //@name TagResponse

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

	var dest []models.Tag
	query := h.bunDB.NewSelect().Model(&dest).Order("name ASC")
	if industryID != "" {
		parsed, err := uuid.Parse(industryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid industry ID"})
			return
		}
		query = query.Where("industry_category_id = ?", parsed)
	}

	err := query.Scan(c.Request.Context())
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

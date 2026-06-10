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

// GetIndustries	godoc
// @Summary		List industries
// @Description	Get all industry categories for filtering
// @Tags		reference
// @Produce		json
// @Success		200 {array} gen.IndustryCategory
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

	c.JSON(http.StatusOK, dest)
}

// GetTags		godoc
// @Summary		List tags
// @Description	Get skill tags, optionally filtered by industry category
// @Tags		reference
// @Produce		json
// @Param		industry query string false "Industry category ID to filter tags by"
// @Success		200 {array} gen.Tag
// @Router		/tags [get]
func (h *ReferenceHandler) GetTags(c *gin.Context) {
	industryID := c.Query("industry")

	var stmt SelectStatement
	if industryID != "" {
		stmt = SELECT(
			gen.Tags.AllColumns,
		).FROM(
			gen.Tags,
		).WHERE(
			gen.Tags.IndustryCategoryID.EQ(UUID(uuid.MustParse(industryID))),
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

	c.JSON(http.StatusOK, dest)
}

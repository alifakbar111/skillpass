package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	. "github.com/go-jet/jet/v2/postgres"

	"skillpass-server-go/internal/gen"
)

type ReferenceHandler struct {
	db *sql.DB
}

func NewReferenceHandler(db *sql.DB) *ReferenceHandler {
	return &ReferenceHandler{db: db}
}

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

func (h *ReferenceHandler) GetTags(c *gin.Context) {
	industryID := c.Query("industry")

	var stmt SelectStatement
	if industryID != "" {
		stmt = SELECT(
			gen.Tags.AllColumns,
		).FROM(
			gen.Tags,
		).WHERE(
			gen.Tags.IndustryCategoryID.EQ(String(industryID)),
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

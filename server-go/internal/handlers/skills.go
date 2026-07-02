package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"

	"skillpass-server-go/internal/gen"
)

type SkillsHandler struct {
	db *sql.DB
}

func NewSkillsHandler(db *sql.DB) *SkillsHandler {
	return &SkillsHandler{db: db}
}

// SkillResponse is the camelCase JSON shape for a skill.
type SkillResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
} //@name SkillResponse

// SearchSkills		godoc
// @Summary		Search skills
// @Description	Autocomplete search for skills by name prefix
// @Tags		skills
// @Produce		json
// @Param		q query string false "Search query (prefix match)"
// @Success		200 {array} handlers.SkillResponse
// @Router		/skills [get]
func (h *SkillsHandler) SearchSkills(c *gin.Context) {
	query := c.Query("q")

	var stmt SelectStatement
	if query == "" {
		stmt = SELECT(
			gen.Skills.ID, gen.Skills.Name,
		).FROM(
			gen.Skills,
		).ORDER_BY(
			gen.Skills.Name,
		).LIMIT(20)
	} else {
		stmt = SELECT(
			gen.Skills.ID, gen.Skills.Name,
		).FROM(
			gen.Skills,
		).WHERE(
			LOWER(gen.Skills.Name).LIKE(LOWER(String("%"+query+"%"))),
		).ORDER_BY(
			gen.Skills.Name,
		).LIMIT(20)
	}

	var skills []gen.Skill
	err := stmt.QueryContext(c.Request.Context(), h.db, &skills)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query skills"})
		return
	}

	resp := make([]SkillResponse, len(skills))
	for i, s := range skills {
		resp[i] = SkillResponse{ID: s.ID, Name: s.Name}
	}
	c.JSON(http.StatusOK, resp)
}

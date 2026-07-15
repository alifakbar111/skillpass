package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"skillpass-server-go/internal/models"
)

type SkillsHandler struct {
	bunDB *bun.DB
}

func NewSkillsHandler(bunDB *bun.DB) *SkillsHandler {
	return &SkillsHandler{bunDB: bunDB}
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
	var skills []models.Skill
	q := h.bunDB.NewSelect().Model(&skills).Column("id", "name").Order("name ASC").Limit(20)
	if query := c.Query("q"); query != "" {
		q = q.Where("LOWER(name) LIKE LOWER(?)", "%"+query+"%")
	}
	if err := q.Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query skills"})
		return
	}

	resp := make([]SkillResponse, len(skills))
	for i, s := range skills {
		resp[i] = SkillResponse{ID: s.ID, Name: s.Name}
	}
	c.JSON(http.StatusOK, resp)
}

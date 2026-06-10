package resume

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ParseResume	godoc
// @Summary		Parse resume text
// @Description	Extract structured profile data (headline, about, experiences) from pasted resume text using AI. Does not modify the profile — the client reviews and saves entries.
// @Tags		profiles
// @Accept		json
// @Produce		json
// @Security	BearerAuth
// @Param		body body object{text=string} true "Raw resume text"
// @Success		200 {object} resume.ParsedResume
// @Failure		400 {object} map[string]string
// @Failure		401 {object} map[string]string
// @Failure		500 {object} map[string]string
// @Router		/profiles/me/resume-parse [post]
func (h *Handler) ParseResume(c *gin.Context) {
	var req struct {
		Text string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: text is required"})
		return
	}

	text := strings.TrimSpace(req.Text)
	if len(text) < 30 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Resume text is too short to parse"})
		return
	}
	// Cap input to keep LLM calls bounded.
	const maxLen = 20000
	if len(text) > maxLen {
		text = text[:maxLen]
	}

	result, err := h.service.Parse(c.Request.Context(), text)
	if err != nil {
		slog.Error("resume parse failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse resume"})
		return
	}

	c.JSON(http.StatusOK, result)
}

package resume

import (
	"io"
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

	h.parseAndRespond(c, req.Text)
}

const maxResumePDFBytes = 8 << 20 // 8 MB

// UploadResume	godoc
// @Summary		Upload resume PDF
// @Description	Upload a PDF resume as multipart field "file"; its text is extracted and parsed by AI into structured experiences. Text-based PDFs only (scanned PDFs need OCR and are rejected as unreadable).
// @Tags		profiles
// @Accept		mpfd
// @Produce		json
// @Security	BearerAuth
// @Param		file formData file true "PDF resume"
// @Success		200 {object} resume.ParsedResume
// @Failure		400 {object} map[string]string
// @Failure		401 {object} map[string]string
// @Failure		500 {object} map[string]string
// @Router		/profiles/me/resume-upload [post]
func (h *Handler) UploadResume(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A 'file' upload field is required"})
		return
	}
	if fileHeader.Size > maxResumePDFBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "PDF too large (max 8MB)"})
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not read upload"})
		return
	}
	defer f.Close()

	// Sniff for the PDF magic header instead of trusting the client.
	head := make([]byte, 5)
	if _, err := io.ReadFull(f, head); err != nil || string(head) != "%PDF-" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is not a PDF"})
		return
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not process upload"})
		return
	}

	text, err := extractPDFText(f, fileHeader.Size)
	if err != nil {
		slog.Warn("pdf text extraction failed", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not read text from this PDF"})
		return
	}
	if len(strings.TrimSpace(text)) < 30 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No readable text found in this PDF — if it's a scanned document, paste the text instead"})
		return
	}

	h.parseAndRespond(c, text)
}

// parseAndRespond trims and bounds the resume text, runs the LLM parse, and
// writes the response. Shared by the paste and PDF-upload endpoints.
func (h *Handler) parseAndRespond(c *gin.Context, raw string) {
	text := strings.TrimSpace(raw)
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

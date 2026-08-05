package documents

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func ctxUUID(c *gin.Context, key string) (uuid.UUID, bool) {
	v := c.GetString(key)
	id, err := uuid.Parse(v)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

// Upload	godoc
// @Summary		Upload a document
// @Tags		documents
// @Accept		mpfd
// @Produce		json
// @Security	BearerAuth
// @Param		file formData file true "File (max 25MB)"
// @Param		category formData string false "identity|contract|certificate|payslip|tax|other"
// @Param		employeeId formData string false "Employee to attach the document to"
// @Success		201 {object} DocumentResponse
// @Router		/hris/documents [post]
func (h *Handler) Upload(c *gin.Context) {
	companyID, ok := ctxUUID(c, "companyId")
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	userID, _ := ctxUUID(c, "userId")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file field is required"})
		return
	}
	defer file.Close()

	var empID *uuid.UUID
	if raw := c.PostForm("employeeId"); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			empID = &id
		}
	}
	mime := header.Header.Get("Content-Type")
	if mime == "" {
		mime = "application/octet-stream"
	}

	res, err := h.svc.Upload(c.Request.Context(), UploadParams{
		CompanyID:  companyID,
		UploadedBy: userID,
		EmployeeID: empID,
		Category:   c.PostForm("category"),
		Filename:   header.Filename,
		MimeType:   mime,
		IPAddress:  c.ClientIP(),
	}, file)
	if err != nil {
		switch {
		case errors.Is(err, ErrTooLarge):
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error()})
		case errors.Is(err, ErrBadRequest):
			c.JSON(http.StatusBadRequest, gin.H{"error": "empty or invalid file"})
		default:
			slog.Error("document upload failed", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload document"})
		}
		return
	}
	c.JSON(http.StatusCreated, res)
}

// List	godoc
// @Summary		List documents
// @Tags		documents
// @Produce		json
// @Security	BearerAuth
// @Success		200 {array} DocumentResponse
// @Router		/hris/documents [get]
func (h *Handler) List(c *gin.Context) {
	companyID, ok := ctxUUID(c, "companyId")
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	res, err := h.svc.List(c.Request.Context(), companyID, c.Query("employeeId"), c.Query("category"))
	if err != nil {
		slog.Error("list documents", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list documents"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// Download	godoc
// @Summary		Download a document
// @Tags		documents
// @Produce		octet-stream
// @Security	BearerAuth
// @Param		id path string true "Document ID"
// @Router		/hris/documents/{id}/download [get]
func (h *Handler) Download(c *gin.Context) {
	companyID, ok := ctxUUID(c, "companyId")
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}
	var userID *uuid.UUID
	if id, ok := ctxUUID(c, "userId"); ok {
		userID = &id
	}

	target, err := h.svc.OpenForDownload(c.Request.Context(), companyID, docID, userID, c.ClientIP())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open document"})
		return
	}
	defer target.Body.Close()

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", target.Filename))
	c.Header("Content-Type", target.MimeType)
	c.Header("X-Content-Type-Options", "nosniff")
	if _, err := io.Copy(c.Writer, target.Body); err != nil {
		slog.Warn("stream document", "documentID", docID, "error", err)
	}
}

// Delete	godoc
// @Summary		Delete a document
// @Tags		documents
// @Produce		json
// @Security	BearerAuth
// @Param		id path string true "Document ID"
// @Router		/hris/documents/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	companyID, ok := ctxUUID(c, "companyId")
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}
	if err := h.svc.Delete(c.Request.Context(), companyID, docID); err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete document"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Document deleted"})
}

// AuditLog	godoc
// @Summary		Document access audit log
// @Tags		documents
// @Produce		json
// @Security	BearerAuth
// @Success		200 {array} DocumentAccessLog
// @Router		/hris/documents/audit-log [get]
func (h *Handler) AuditLog(c *gin.Context) {
	companyID, ok := ctxUUID(c, "companyId")
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	res, err := h.svc.AuditLog(c.Request.Context(), companyID)
	if err != nil {
		slog.Error("document audit log", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load audit log"})
		return
	}
	c.JSON(http.StatusOK, res)
}

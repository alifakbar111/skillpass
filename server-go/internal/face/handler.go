package face

import (
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"strings"

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
	id, err := uuid.Parse(c.GetString(key))
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

// decodeImage accepts a raw or data-URL base64 image string.
func decodeImage(s string) ([]byte, error) {
	if i := strings.Index(s, ","); strings.HasPrefix(s, "data:") && i >= 0 {
		s = s[i+1:]
	}
	return base64.StdEncoding.DecodeString(strings.TrimSpace(s))
}

type enrollRequest struct {
	Image string `json:"image" binding:"required"`
}

// Enroll	godoc
// @Summary		Enrol the current employee's face
// @Tags		face
// @Accept		json
// @Produce		json
// @Security	BearerAuth
// @Param		body body object true "{ image: base64 }"
// @Success		200 {object} FaceEnrollResponse
// @Router		/hris/face/enroll [post]
func (h *Handler) Enroll(c *gin.Context) {
	companyID, ok := ctxUUID(c, "companyId")
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	employeeID, ok := ctxUUID(c, "employeeId")
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "No employee record"})
		return
	}
	var enrolledBy *uuid.UUID
	if id, ok := ctxUUID(c, "userId"); ok {
		enrolledBy = &id
	}

	var req enrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image is required"})
		return
	}
	img, err := decodeImage(req.Image)
	if err != nil || len(img) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image data"})
		return
	}

	res, err := h.svc.Enroll(c.Request.Context(), companyID, employeeID, enrolledBy, img)
	if err != nil {
		switch {
		case errors.Is(err, ErrDisabled):
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Face recognition is not configured"})
		case errors.Is(err, ErrEmployeeNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		default:
			slog.Error("face enroll failed", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enrol face"})
		}
		return
	}
	c.JSON(http.StatusOK, res)
}

// Status	godoc
// @Summary		Current employee's face-enrolment status
// @Tags		face
// @Produce		json
// @Security	BearerAuth
// @Success		200 {object} FaceStatusResponse
// @Router		/hris/face/status [get]
func (h *Handler) Status(c *gin.Context) {
	companyID, ok := ctxUUID(c, "companyId")
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	employeeID, ok := ctxUUID(c, "employeeId")
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "No employee record"})
		return
	}
	h.respondStatus(c, companyID, employeeID)
}

// EmployeeStatus	godoc
// @Summary		A specific employee's face-enrolment status (admin)
// @Tags		face
// @Produce		json
// @Security	BearerAuth
// @Param		id path string true "Employee ID"
// @Success		200 {object} FaceStatusResponse
// @Router		/hris/face/employees/{id} [get]
func (h *Handler) EmployeeStatus(c *gin.Context) {
	companyID, ok := ctxUUID(c, "companyId")
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	employeeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	h.respondStatus(c, companyID, employeeID)
}

func (h *Handler) respondStatus(c *gin.Context, companyID, employeeID uuid.UUID) {
	res, err := h.svc.Status(c.Request.Context(), companyID, employeeID)
	if err != nil {
		if errors.Is(err, ErrEmployeeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
			return
		}
		slog.Error("face status failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get status"})
		return
	}
	c.JSON(http.StatusOK, res)
}

package webhook

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func getCompanyID(c *gin.Context) (string, bool) {
	val, ok := c.Get("companyId")
	if !ok {
		return "", false
	}
	s, ok := val.(string)
	return s, ok && s != ""
}

// List	godoc
// @Summary		List webhooks
// @Description	List the authenticated company's registered webhooks (secrets omitted)
// @Tags		webhooks
// @Produce		json
// @Security	BearerAuth
// @Success		200 {array} webhook.Webhook
// @Failure		403 {object} map[string]string
// @Router		/company/webhooks [get]
func (h *Handler) List(c *gin.Context) {
	companyID, ok := getCompanyID(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}

	webhooks, err := h.service.List(c.Request.Context(), companyID)
	if err != nil {
		slog.Error("failed to list webhooks", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list webhooks"})
		return
	}

	c.JSON(http.StatusOK, webhooks)
}

// Create	godoc
// @Summary		Register a webhook
// @Description	Register a webhook URL. The signing secret is returned once — store it; events are signed with HMAC-SHA256 in the X-SkillPass-Signature header.
// @Tags		webhooks
// @Accept		json
// @Produce		json
// @Security	BearerAuth
// @Param		body body object{url=string} true "Webhook URL"
// @Success		201 {object} webhook.Webhook
// @Failure		400 {object} map[string]string
// @Failure		403 {object} map[string]string
// @Router		/company/webhooks [post]
func (h *Handler) Create(c *gin.Context) {
	companyID, ok := getCompanyID(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}

	var req struct {
		URL string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: url is required"})
		return
	}

	created, err := h.service.Create(c.Request.Context(), companyID, req.URL)
	if err != nil {
		// URL validation errors are user errors.
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// Delete	godoc
// @Summary		Delete a webhook
// @Description	Delete a webhook owned by the authenticated company
// @Tags		webhooks
// @Produce		json
// @Security	BearerAuth
// @Param		id path string true "Webhook UUID"
// @Success		200 {object} map[string]string
// @Failure		403 {object} map[string]string
// @Failure		404 {object} map[string]string
// @Router		/company/webhooks/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	companyID, ok := getCompanyID(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}

	webhookID := c.Param("id")
	if _, err := uuid.Parse(webhookID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
		return
	}

	found, err := h.service.Delete(c.Request.Context(), webhookID, companyID)
	if err != nil {
		slog.Error("failed to delete webhook", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete webhook"})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

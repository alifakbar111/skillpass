package notification

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

func getUserID(c *gin.Context) (string, bool) {
	val, ok := c.Get("userId")
	if !ok {
		return "", false
	}
	s, ok := val.(string)
	return s, ok && s != ""
}

// ListMine	godoc
// @Summary		List my notifications
// @Description	Get recent notifications and unread count for the authenticated user
// @Tags		notifications
// @Produce		json
// @Security	BearerAuth
// @Success		200 {object} notification.ListResult
// @Failure		401 {object} map[string]string
// @Router		/notifications/me [get]
func (h *Handler) ListMine(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	result, err := h.service.ListForUser(c.Request.Context(), userID, 50)
	if err != nil {
		slog.Error("failed to list notifications", "userID", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list notifications"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// MarkRead	godoc
// @Summary		Mark notification read
// @Description	Mark a single notification as read
// @Tags		notifications
// @Produce		json
// @Security	BearerAuth
// @Param		id path string true "Notification UUID"
// @Success		200 {object} map[string]string
// @Failure		401 {object} map[string]string
// @Failure		404 {object} map[string]string
// @Router		/notifications/{id}/read [put]
func (h *Handler) MarkRead(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	notificationID := c.Param("id")
	if _, err := uuid.Parse(notificationID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	found, err := h.service.MarkRead(c.Request.Context(), notificationID, userID)
	if err != nil {
		slog.Error("failed to mark notification read", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark read"})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Marked read"})
}

// MarkAllRead	godoc
// @Summary		Mark all notifications read
// @Description	Mark all unread notifications as read for the authenticated user
// @Tags		notifications
// @Produce		json
// @Security	BearerAuth
// @Success		200 {object} map[string]string
// @Failure		401 {object} map[string]string
// @Router		/notifications/read-all [put]
func (h *Handler) MarkAllRead(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := h.service.MarkAllRead(c.Request.Context(), userID); err != nil {
		slog.Error("failed to mark all read", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark all read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All marked read"})
}

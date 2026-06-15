package profileviews

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	db      *sql.DB
	service *Service
}

func NewHandler(db *sql.DB, service *Service) *Handler {
	return &Handler{db: db, service: service}
}

func (h *Handler) GetMyProfileViews(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	profileID, err := lookupProfileID(h.db, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		return
	}

	views, err := h.service.GetViewsByProfile(c.Request.Context(), profileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch views"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"views": views})
}

func (h *Handler) RecordView(c *gin.Context) {
	profileIDStr := c.Param("profile_id")
	profileID, err := uuid.Parse(profileIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid profile id"})
		return
	}

	viewerID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	companyIDStr := c.GetString("companyId")
	var companyID *uuid.UUID
	if companyIDStr != "" {
		cid, err := uuid.Parse(companyIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid company ID"})
			return
		}
		companyID = &cid
	}

	if err := h.service.RecordView(c.Request.Context(), profileID, viewerID, companyID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record view"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "view recorded"})
}

func lookupProfileID(db *sql.DB, userID uuid.UUID) (uuid.UUID, error) {
	var profileID uuid.UUID
	err := db.QueryRow(`SELECT id FROM profiles WHERE user_id = $1`, userID).Scan(&profileID)
	return profileID, err
}

func getUserID(c *gin.Context) (uuid.UUID, error) {
	val, exists := c.Get("userId")
	if !exists {
		return uuid.Nil, sql.ErrNoRows
	}
	return uuid.Parse(val.(string))
}

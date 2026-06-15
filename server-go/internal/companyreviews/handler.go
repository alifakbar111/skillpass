package companyreviews

import (
	"errors"
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

// PostReview		godoc
// @Summary		Review a company
// @Description	Candidate rates and reviews a company after applying or interviewing
// @Tags		company-reviews
// @Accept		json
// @Produce		json
// @Security	BearerAuth
// @Param		id path string true "Company UUID"
// @Param		body body companyreviews.CreateReviewRequest true "Review payload"
// @Success		201 {object} companyreviews.CompanyReview
// @Failure		400 {object} map[string]string
// @Failure		401 {object} map[string]string
// @Failure		404 {object} map[string]string
// @Router		/companies/{id}/reviews [post]
func (h *Handler) PostReview(c *gin.Context) {
	userID, ok := c.Get("userId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	companyID := c.Param("id")
	if _, err := uuid.Parse(companyID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	profileID, err := h.lookupCandidateProfile(c, userIDStr)
	if err != nil {
		slog.Error("failed to lookup candidate profile", "userID", userIDStr, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Candidate profile not found"})
		return
	}

	var req CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	review, err := h.service.Create(c.Request.Context(), companyID, profileID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidRating):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, ErrInvalidInteraction):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, ErrCompanyNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
		default:
			slog.Error("failed to create review", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create review"})
		}
		return
	}

	c.JSON(http.StatusCreated, review)
}

// ListReviews			godoc
// @Summary		List reviews for a company
// @Description	Get all reviews for a company
// @Tags		company-reviews
// @Produce		json
// @Security	BearerAuth
// @Param		id path string true "Company UUID"
// @Success		200 {array} companyreviews.CompanyReview
// @Failure		400 {object} map[string]string
// @Failure		401 {object} map[string]string
// @Router		/companies/{id}/reviews [get]
func (h *Handler) ListReviews(c *gin.Context) {
	companyID := c.Param("id")
	if _, err := uuid.Parse(companyID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	reviews, err := h.service.ListByCompanyID(c.Request.Context(), companyID)
	if err != nil {
		slog.Error("failed to list reviews", "companyID", companyID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list reviews"})
		return
	}

	c.JSON(http.StatusOK, reviews)
}

// GetReputation		godoc
// @Summary		Get company reputation
// @Description	Get aggregated review score for a company
// @Tags		company-reviews
// @Produce		json
// @Param		id path string true "Company UUID"
// @Success		200 {object} companyreviews.Reputation
// @Failure		400 {object} map[string]string
// @Router		/companies/{id}/reputation [get]
func (h *Handler) GetReputation(c *gin.Context) {
	companyID := c.Param("id")
	if _, err := uuid.Parse(companyID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	reputation, err := h.service.GetReputation(c.Request.Context(), companyID)
	if err != nil {
		slog.Error("failed to get reputation", "companyID", companyID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get reputation"})
		return
	}

	c.JSON(http.StatusOK, reputation)
}

func (h *Handler) lookupCandidateProfile(c *gin.Context, userID string) (string, error) {
	return h.service.LookupCandidateProfile(c.Request.Context(), userID)
}

package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"skillpass-server-go/internal/models"
)

const defaultListLimit = 50
const maxListLimit = 200

type PendingCompany struct {
	ID                 string          `json:"id"`
	UserID             string          `json:"userId"`
	CompanyName        string          `json:"companyName"`
	Website            *string         `json:"website,omitempty"`
	Industry           string          `json:"industry"`
	Description        *string         `json:"description,omitempty"`
	VerificationStatus string          `json:"verificationStatus"`
	VerificationDocs   json.RawMessage `json:"verificationDocs,omitempty" swaggertype:"array,object"`
	VerifiedAt         *time.Time      `json:"verifiedAt,omitempty"`
	CreatedAt          time.Time       `json:"createdAt"`
} //@name PendingCompany

type VerificationActionRequest struct {
	Action string  `json:"action" binding:"required,oneof=approve reject"`
	Reason *string `json:"reason,omitempty"`
} //@name VerificationActionRequest

type AdminHandler struct {
	bunDB *bun.DB
}

func NewAdminHandler(bunDB *bun.DB) *AdminHandler {
	return &AdminHandler{bunDB: bunDB}
}

func pendingCompanyFromModel(c models.Company) PendingCompany {
	var docs json.RawMessage
	if c.VerificationDocs != nil {
		docs = json.RawMessage(*c.VerificationDocs)
	}
	return PendingCompany{
		ID:                 c.ID.String(),
		UserID:             c.UserID.String(),
		CompanyName:        c.CompanyName,
		Website:            c.Website,
		Industry:           c.Industry,
		Description:        c.Description,
		VerificationStatus: c.VerificationStatus,
		VerificationDocs:   docs,
		VerifiedAt:         c.VerifiedAt,
		CreatedAt:          c.CreatedAt,
	}
}

func parseLimit(c *gin.Context) int64 {
	limit := int64(defaultListLimit)
	if s := c.Query("limit"); s != "" {
		if v, err := strconv.ParseInt(s, 10, 64); err == nil && v > 0 {
			limit = v
		}
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}
	return limit
}

func parseOffset(c *gin.Context) int64 {
	offset := int64(0)
	if s := c.Query("offset"); s != "" {
		if v, err := strconv.ParseInt(s, 10, 64); err == nil && v >= 0 {
			offset = v
		}
	}
	return offset
}

// ListPendingVerifications	godoc
// @Summary		List pending company verifications
// @Description	Get all companies with pending verification status. Requires admin role.
// @Tags		admin
// @Produce		json
// @Security	BearerAuth
// @Param		limit query int false "Max results (default 50, max 200)"
// @Param		offset query int false "Result offset for pagination"
// @Success		200 {array} PendingCompany
// @Router		/admin/verifications/pending [get]
func (h *AdminHandler) ListPendingVerifications(c *gin.Context) {
	limit := parseLimit(c)
	offset := parseOffset(c)

	var companies []models.Company
	err := h.bunDB.NewSelect().
		Model(&companies).
		Where("verification_status = 'pending'").
		Order("created_at ASC").
		Limit(int(limit)).
		Offset(int(offset)).
		Scan(c.Request.Context())
	if err != nil {
		slog.Error("failed to query pending verifications", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query pending verifications"})
		return
	}

	result := make([]PendingCompany, len(companies))
	for i, co := range companies {
		result[i] = pendingCompanyFromModel(co)
	}

	c.JSON(http.StatusOK, result)
}

// HandleVerification	godoc
// @Summary		Approve or reject company verification
// @Description	Process a company verification request (approve or reject). Requires admin role.
// @Tags		admin
// @Accept		json
// @Produce		json
// @Security	BearerAuth
// @Param		id path string true "Company UUID"
// @Param		body body VerificationActionRequest true "Verification action"
// @Success		200 {object} map[string]interface{}
// @Failure		400 {object} map[string]string
// @Failure		404 {object} map[string]string
// @Router		/admin/verifications/{id} [post]
func (h *AdminHandler) HandleVerification(c *gin.Context) {
	id := c.Param("id")
	companyUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company id"})
		return
	}

	adminIDVal, ok := c.Get("userId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	adminIDStr, _ := adminIDVal.(string)
	adminUUID, err := uuid.Parse(adminIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req VerificationActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	tx, err := h.bunDB.BeginTx(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction"})
		return
	}
	defer tx.Rollback()

	var existing models.Company
	err = tx.NewSelect().Model(&existing).
		Column("id", "user_id").
		Where("id = ?", companyUUID).
		For("UPDATE").
		Scan(c.Request.Context())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
			return
		}
		slog.Error("failed to find company", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process verification"})
		return
	}

	var updated models.Company
	switch req.Action {
	case "approve":
		err = tx.NewUpdate().
			Model((*models.Company)(nil)).
			Set("verification_status = ?", "verified").
			Set("verified_at = ?", time.Now().UTC()).
			Where("id = ?", companyUUID).
			Returning("*").
			Scan(c.Request.Context(), &updated)
		if err != nil {
			slog.Error("failed to approve company", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to approve"})
			return
		}
		_, err = tx.NewUpdate().
			Model((*models.User)(nil)).
			Set("is_verified = ?", true).
			Where("id = ?", existing.UserID).
			Exec(c.Request.Context())
		if err != nil {
			slog.Error("failed to mark user verified", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to approve"})
			return
		}
	case "reject":
		err = tx.NewUpdate().
			Model((*models.Company)(nil)).
			Set("verification_status = ?", "rejected").
			Where("id = ?", companyUUID).
			Returning("*").
			Scan(c.Request.Context(), &updated)
		if err != nil {
			slog.Error("failed to reject company", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reject"})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action"})
		return
	}

	_, err = tx.NewInsert().Model(&models.AdminAudit{
		AdminID:   adminUUID,
		CompanyID: companyUUID,
		Action:    req.Action,
		Reason:    req.Reason,
	}).Exec(c.Request.Context())
	if err != nil {
		slog.Error("failed to write audit log", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record audit"})
		return
	}

	if err = tx.Commit(); err != nil {
		slog.Error("failed to commit verification transaction", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit"})
		return
	}

	c.JSON(http.StatusOK, pendingCompanyFromModel(updated))
}

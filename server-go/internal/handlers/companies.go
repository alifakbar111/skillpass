package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"log/slog"

	"skillpass-server-go/internal/lib"
	"skillpass-server-go/internal/models"
)

type CompanyResponse struct {
	ID                 string          `json:"id"`
	UserID             string          `json:"userId"`
	CompanyName        string          `json:"companyName"`
	Website            *string         `json:"website,omitempty"`
	Industry           string          `json:"industry"`
	Description        *string         `json:"description,omitempty"`
	VerificationStatus string          `json:"verificationStatus"`
	VerificationDocs   json.RawMessage `json:"verificationDocs,omitempty" swaggertype:"array,object"`
	VerifiedAt         *time.Time      `json:"verifiedAt,omitempty"`
	BlindMode          bool            `json:"blindMode"`
	CreatedAt          time.Time       `json:"createdAt"`
}

type UpdateCompanyRequest struct {
	CompanyName *string `json:"companyName" binding:"omitempty,min=1"`
	Website     *string `json:"website" binding:"omitempty,url"`
	Industry    *string `json:"industry" binding:"omitempty,min=1"`
	Description *string `json:"description"`
	BlindMode   *bool   `json:"blindMode"`
} //@name UpdateCompanyRequest

type VerificationRequest struct {
	BusinessRegistration string `json:"businessRegistration" binding:"required"`
	Website              string `json:"website" binding:"required,url"`
	Address              string `json:"address" binding:"required"`
	Contact              string `json:"contact" binding:"required"`
} //@name VerificationRequest

type CompanyHandler struct {
	db    *sql.DB
	bunDB *bun.DB
}

func NewCompanyHandler(db *sql.DB, bunDB *bun.DB) *CompanyHandler {
	return &CompanyHandler{db: db, bunDB: bunDB}
}

func companyFromModel(c models.Company) CompanyResponse {
	var docs json.RawMessage
	if c.VerificationDocs != nil {
		docs = json.RawMessage(*c.VerificationDocs)
	}
	return CompanyResponse{
		ID:                 c.ID.String(),
		UserID:             c.UserID.String(),
		CompanyName:        c.CompanyName,
		Website:            c.Website,
		Industry:           c.Industry,
		Description:        c.Description,
		VerificationStatus: c.VerificationStatus,
		VerificationDocs:   docs,
		VerifiedAt:         c.VerifiedAt,
		BlindMode:          c.BlindMode,
		CreatedAt:          c.CreatedAt,
	}
}

// GetProfile		godoc
// @Summary		Get company profile
// @Description	Get the authenticated company's profile
// @Tags		companies
// @Produce		json
// @Security	BearerAuth
// @Success		200 {object} map[string]interface{}
// @Failure		401 {object} map[string]string
// @Failure		404 {object} map[string]string
// @Router		/company/profile [get]
func (h *CompanyHandler) GetProfile(c *gin.Context) {
	userIDVal, ok := c.Get("userId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userUUID, err := lib.ParseUUID(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid user ID: %v", err)})
		return
	}

	var company models.Company
	err = h.bunDB.NewSelect().
		Model(&company).
		Where("user_id = ?", userUUID).
		Scan(c.Request.Context())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
			return
		}
		slog.Error("failed to load company", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load company"})
		return
	}

	resp := companyFromModel(company)
	resp.BlindMode = CompanyBlindMode(c.Request.Context(), h.db, company.ID.String())
	c.JSON(http.StatusOK, resp)
}

// UpdateProfile	godoc
// @Summary		Update company profile
// @Description	Update the authenticated company's profile fields (company name, website, industry, description)
// @Tags		companies
// @Accept		json
// @Produce		json
// @Security	BearerAuth
// @Param		body body UpdateCompanyRequest true "Company profile fields to update"
// @Success		200 {object} map[string]interface{}
// @Failure		400 {object} map[string]string
// @Failure		401 {object} map[string]string
// @Failure		404 {object} map[string]string
// @Router		/company/profile [put]
func (h *CompanyHandler) UpdateProfile(c *gin.Context) {
	userIDVal, ok := c.Get("userId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userUUID, err := lib.ParseUUID(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid user ID: %v", err)})
		return
	}

	var req UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.CompanyName == nil && req.Website == nil && req.Industry == nil && req.Description == nil && req.BlindMode == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	var company models.Company
	query := h.bunDB.NewUpdate().Model((*models.Company)(nil))
	if req.CompanyName != nil {
		query = query.Set("company_name = ?", *req.CompanyName)
	}
	if req.Website != nil {
		query = query.Set("website = ?", *req.Website)
	}
	if req.Industry != nil {
		query = query.Set("industry = ?", *req.Industry)
	}
	if req.Description != nil {
		query = query.Set("description = ?", *req.Description)
	}
	if req.BlindMode != nil {
		query = query.Set("blind_mode = ?", *req.BlindMode)
	}
	err = query.Where("user_id = ?", userUUID).Returning("*").Scan(c.Request.Context(), &company)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
			return
		}
		slog.Error("failed to update company", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update company"})
		return
	}

	resp := companyFromModel(company)
	c.JSON(http.StatusOK, resp)
}

// SubmitVerification	godoc
// @Summary		Submit verification documents
// @Description	Submit business verification documents for company approval
// @Tags		companies
// @Accept		json
// @Produce		json
// @Security	BearerAuth
// @Param		body body VerificationRequest true "Verification documents"
// @Success		200 {object} VerificationSubmittedResponse
// @Failure		400 {object} map[string]string
// @Failure		401 {object} map[string]string
// @Failure		404 {object} map[string]string
// @Router		/company/verification [post]
func (h *CompanyHandler) SubmitVerification(c *gin.Context) {
	userIDVal, ok := c.Get("userId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userUUID, err := lib.ParseUUID(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid user ID: %v", err)})
		return
	}

	var req VerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	docs, _ := json.Marshal(req)

	result, err := h.bunDB.NewUpdate().
		Model((*models.Company)(nil)).
		Set("verification_docs = ?", string(docs)).
		Set("verification_status = ?", "pending").
		Where("user_id = ?", userUUID).
		Exec(c.Request.Context())
	if err != nil {
		slog.Error("failed to submit verification", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to submit verification"})
		return
	}
	ra, _ := result.RowsAffected()
	if ra == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
		return
	}

	c.JSON(http.StatusOK, VerificationSubmittedResponse{Message: "Verification submitted", Status: "pending"})
}

// GetVerificationStatus	godoc
// @Summary		Get verification status
// @Description	Get the current verification status for the authenticated company
// @Tags		companies
// @Produce		json
// @Security	BearerAuth
// @Success		200 {object} VerificationStatusResponse
// @Failure		401 {object} map[string]string
// @Router		/company/verification-status [get]
func (h *CompanyHandler) GetVerificationStatus(c *gin.Context) {
	userIDVal, ok := c.Get("userId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userUUID, err := lib.ParseUUID(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid user ID: %v", err)})
		return
	}

	var company models.Company
	err = h.bunDB.NewSelect().
		Model(&company).
		Column("verification_status").
		Where("user_id = ?", userUUID).
		Scan(c.Request.Context())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusOK, VerificationStatusResponse{VerificationStatus: "none"})
			return
		}
		slog.Error("failed to load verification status", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load verification status"})
		return
	}

	c.JSON(http.StatusOK, VerificationStatusResponse{VerificationStatus: company.VerificationStatus})
}

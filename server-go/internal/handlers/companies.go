package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	"log/slog"

	"skillpass-server-go/.gen/skillpass/public/model"
	"skillpass-server-go/internal/gen"
)

type CompanyResponse struct {
	ID                 string          `json:"id"`
	UserID             string          `json:"userId"`
	CompanyName        string          `json:"companyName"`
	Website            *string         `json:"website"`
	Industry           string          `json:"industry"`
	Description        *string         `json:"description"`
	VerificationStatus string          `json:"verificationStatus"`
	VerificationDocs   json.RawMessage `json:"verificationDocs"`
	VerifiedAt         *time.Time      `json:"verifiedAt"`
	CreatedAt          time.Time       `json:"createdAt"`
}

type UpdateCompanyRequest struct {
	CompanyName *string `json:"companyName" binding:"omitempty,min=1"`
	Website     *string `json:"website" binding:"omitempty,url"`
	Industry    *string `json:"industry" binding:"omitempty,min=1"`
	Description *string `json:"description"`
}

type VerificationRequest struct {
	BusinessRegistration string `json:"businessRegistration" binding:"required"`
	Website              string `json:"website" binding:"required,url"`
	Address              string `json:"address" binding:"required"`
	Contact              string `json:"contact" binding:"required"`
}

type CompanyHandler struct {
	db *sql.DB
}

func NewCompanyHandler(db *sql.DB) *CompanyHandler {
	return &CompanyHandler{db: db}
}

func companyFromModel(c model.Companies) CompanyResponse {
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
		VerificationStatus: string(c.VerificationStatus),
		VerificationDocs:   docs,
		VerifiedAt:         c.VerifiedAt,
		CreatedAt:          c.CreatedAt,
	}
}

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

	stmt := SELECT(
		gen.Companies.AllColumns,
	).FROM(
		gen.Companies,
	).WHERE(
		gen.Companies.UserID.EQ(UUID(uuid.MustParse(userIDStr))),
	)

	var company model.Companies
	if err := stmt.QueryContext(c.Request.Context(), h.db, &company); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
			return
		}
		slog.Error("failed to load company", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load company"})
		return
	}

	c.JSON(http.StatusOK, companyFromModel(company))
}

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

	var req UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var setVals []interface{}
	if req.CompanyName != nil {
		setVals = append(setVals, gen.Companies.CompanyName.SET(String(*req.CompanyName)))
	}
	if req.Website != nil {
		setVals = append(setVals, gen.Companies.Website.SET(String(*req.Website)))
	}
	if req.Industry != nil {
		setVals = append(setVals, gen.Companies.Industry.SET(String(*req.Industry)))
	}
	if req.Description != nil {
		setVals = append(setVals, gen.Companies.Description.SET(String(*req.Description)))
	}

	if len(setVals) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	stmt := gen.Companies.UPDATE().SET(setVals[0], setVals[1:]...).WHERE(
		gen.Companies.UserID.EQ(UUID(uuid.MustParse(userIDStr))),
	).RETURNING(
		gen.Companies.AllColumns,
	)

	var company model.Companies
	if err := stmt.QueryContext(c.Request.Context(), h.db, &company); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
			return
		}
		slog.Error("failed to update company", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update company"})
		return
	}

	c.JSON(http.StatusOK, companyFromModel(company))
}

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

	var req VerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	docs, _ := json.Marshal(req)

	stmt := gen.Companies.UPDATE().SET(
		gen.Companies.VerificationDocs.SET(StringExp(CAST(String(string(docs))).AS("jsonb"))),
		gen.Companies.VerificationStatus.SET(gen.VerificationStatusPending),
	).WHERE(
		gen.Companies.UserID.EQ(UUID(uuid.MustParse(userIDStr))),
	)

	result, err := stmt.ExecContext(c.Request.Context(), h.db)
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

	c.JSON(http.StatusOK, gin.H{"message": "Verification submitted", "status": "pending"})
}

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

	stmt := SELECT(
		gen.Companies.VerificationStatus,
	).FROM(
		gen.Companies,
	).WHERE(
		gen.Companies.UserID.EQ(UUID(uuid.MustParse(userIDStr))),
	)

	var company model.Companies
	err := stmt.QueryContext(c.Request.Context(), h.db, &company)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusOK, gin.H{"verificationStatus": "none"})
			return
		}
		slog.Error("failed to load verification status", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load verification status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"verificationStatus": string(company.VerificationStatus)})
}

package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	. "github.com/go-jet/jet/v2/postgres"
	"database/sql"

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
	CompanyName *string `json:"companyName"`
	Website     *string `json:"website"`
	Industry    *string `json:"industry"`
	Description *string `json:"description"`
}

type VerificationRequest struct {
	BusinessRegistration string `json:"businessRegistration" binding:"required"`
	Website              string `json:"website" binding:"required"`
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
	userID, _ := c.Get("userId")
	userIDStr := userID.(string)

	stmt := SELECT(
		gen.Companies.AllColumns,
	).FROM(
		gen.Companies,
	).WHERE(
		gen.Companies.UserID.EQ(String(userIDStr)),
	)

	var company model.Companies
	err := stmt.QueryContext(c.Request.Context(), h.db, &company)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
		return
	}

	c.JSON(http.StatusOK, companyFromModel(company))
}

func (h *CompanyHandler) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr := userID.(string)

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
		gen.Companies.UserID.EQ(String(userIDStr)),
	).RETURNING(
		gen.Companies.AllColumns,
	)

	var company model.Companies
	err := stmt.QueryContext(c.Request.Context(), h.db, &company)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
		return
	}

	c.JSON(http.StatusOK, companyFromModel(company))
}

func (h *CompanyHandler) SubmitVerification(c *gin.Context) {
	userID, _ := c.Get("userId")
	userIDStr := userID.(string)

	var req VerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	docs, _ := json.Marshal(req)

	stmt := gen.Companies.UPDATE().SET(
		gen.Companies.VerificationDocs.SET(String(string(docs))),
		gen.Companies.VerificationStatus.SET(String("pending")),
	).WHERE(
		gen.Companies.UserID.EQ(String(userIDStr)),
	)

	result, err := stmt.ExecContext(c.Request.Context(), h.db)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
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
	userID, _ := c.Get("userId")
	userIDStr := userID.(string)

	stmt := SELECT(
		gen.Companies.VerificationStatus,
	).FROM(
		gen.Companies,
	).WHERE(
		gen.Companies.UserID.EQ(String(userIDStr)),
	)

	var company model.Companies
	err := stmt.QueryContext(c.Request.Context(), h.db, &company)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"verificationStatus": "none"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"verificationStatus": string(company.VerificationStatus)})
}

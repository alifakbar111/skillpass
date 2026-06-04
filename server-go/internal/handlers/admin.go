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

type PendingCompany struct {
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

type VerificationActionRequest struct {
	Action string  `json:"action" binding:"required,oneof=approve reject"`
	Reason *string `json:"reason"`
}

type AdminHandler struct {
	db *sql.DB
}

func NewAdminHandler(db *sql.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

func pendingCompanyFromModel(c model.Companies) PendingCompany {
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
		VerificationStatus: string(c.VerificationStatus),
		VerificationDocs:   docs,
		VerifiedAt:         c.VerifiedAt,
		CreatedAt:          c.CreatedAt,
	}
}

func (h *AdminHandler) ListPendingVerifications(c *gin.Context) {
	stmt := SELECT(
		gen.Companies.AllColumns,
	).FROM(
		gen.Companies,
	).WHERE(
		gen.Companies.VerificationStatus.EQ(String("pending")),
	)

	var companies []model.Companies
	err := stmt.QueryContext(c.Request.Context(), h.db, &companies)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query pending verifications"})
		return
	}

	result := make([]PendingCompany, len(companies))
	for i, co := range companies {
		result[i] = pendingCompanyFromModel(co)
	}

	c.JSON(http.StatusOK, result)
}

func (h *AdminHandler) HandleVerification(c *gin.Context) {
	id := c.Param("id")

	var req VerificationActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	checkStmt := SELECT(
		gen.Companies.ID, gen.Companies.UserID,
	).FROM(
		gen.Companies,
	).WHERE(
		gen.Companies.ID.EQ(String(id)),
	)

	var company model.Companies
	err := checkStmt.QueryContext(c.Request.Context(), h.db, &company)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
		return
	}

	switch req.Action {
	case "approve":
		updateStmt := gen.Companies.UPDATE().SET(
			gen.Companies.VerificationStatus.SET(String("verified")),
			gen.Companies.VerifiedAt.SET(TimestampT(time.Now().UTC())),
		).WHERE(
			gen.Companies.ID.EQ(String(id)),
		).RETURNING(
			gen.Companies.AllColumns,
		)

		var companyModel model.Companies
		err = updateStmt.QueryContext(c.Request.Context(), h.db, &companyModel)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to approve"})
			return
		}

		gen.Users.UPDATE().SET(
			gen.Users.IsVerified.SET(Bool(true)),
		).WHERE(
			gen.Users.ID.EQ(String(company.UserID.String())),
		).ExecContext(c.Request.Context(), h.db)

		c.JSON(http.StatusOK, pendingCompanyFromModel(companyModel))

	case "reject":
		updateStmt := gen.Companies.UPDATE().SET(
			gen.Companies.VerificationStatus.SET(String("rejected")),
		).WHERE(
			gen.Companies.ID.EQ(String(id)),
		).RETURNING(
			gen.Companies.AllColumns,
		)

		var companyModel model.Companies
		err = updateStmt.QueryContext(c.Request.Context(), h.db, &companyModel)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reject"})
			return
		}

		c.JSON(http.StatusOK, pendingCompanyFromModel(companyModel))

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action"})
	}
}

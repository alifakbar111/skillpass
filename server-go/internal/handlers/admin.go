package handlers

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"

	"skillpass-server-go/.gen/skillpass/public/model"
	"skillpass-server-go/internal/gen"
)

const defaultListLimit = 50
const maxListLimit = 200

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
} //@name PendingCompany

type VerificationActionRequest struct {
	Action string  `json:"action" binding:"required,oneof=approve reject"`
	Reason *string `json:"reason"`
} //@name VerificationActionRequest

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

	stmt := SELECT(
		gen.Companies.AllColumns,
	).FROM(
		gen.Companies,
	).WHERE(
		gen.Companies.VerificationStatus.EQ(gen.VerificationStatusPending),
	).ORDER_BY(
		gen.Companies.CreatedAt.ASC(),
	).LIMIT(limit).OFFSET(offset)

	var companies []model.Companies
	if err := stmt.QueryContext(c.Request.Context(), h.db, &companies); err != nil {
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

	tx, err := h.db.BeginTx(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction"})
		return
	}
	defer tx.Rollback()

	var existing []model.Companies
	checkStmt := SELECT(
		gen.Companies.ID, gen.Companies.UserID,
	).FROM(
		gen.Companies,
	).WHERE(
		gen.Companies.ID.EQ(UUID(companyUUID)),
	).FOR(UPDATE())

	if err = checkStmt.QueryContext(c.Request.Context(), tx, &existing); err != nil {
		slog.Error("failed to find company", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process verification"})
		return
	}
	if len(existing) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
		return
	}

	var updated model.Companies
	switch req.Action {
	case "approve":
		updateStmt := gen.Companies.UPDATE().SET(
			gen.Companies.VerificationStatus.SET(gen.VerificationStatusVerified),
			gen.Companies.VerifiedAt.SET(TimestampzT(time.Now().UTC())),
		).WHERE(
			gen.Companies.ID.EQ(UUID(companyUUID)),
		).RETURNING(
			gen.Companies.AllColumns,
		)
		if err = updateStmt.QueryContext(c.Request.Context(), tx, &updated); err != nil {
			slog.Error("failed to approve company", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to approve"})
			return
		}
		if _, err = gen.Users.UPDATE().SET(
			gen.Users.IsVerified.SET(Bool(true)),
		).WHERE(
			gen.Users.ID.EQ(UUID(existing[0].UserID)),
		).ExecContext(c.Request.Context(), tx); err != nil {
			slog.Error("failed to mark user verified", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to approve"})
			return
		}
	case "reject":
		updateStmt := gen.Companies.UPDATE().SET(
			gen.Companies.VerificationStatus.SET(gen.VerificationStatusRejected),
		).WHERE(
			gen.Companies.ID.EQ(UUID(companyUUID)),
		).RETURNING(
			gen.Companies.AllColumns,
		)
		if err = updateStmt.QueryContext(c.Request.Context(), tx, &updated); err != nil {
			slog.Error("failed to reject company", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reject"})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action"})
		return
	}

	if _, err = gen.AdminAuditLog.INSERT(
		gen.AdminAuditLog.AdminID, gen.AdminAuditLog.CompanyID,
		gen.AdminAuditLog.Action, gen.AdminAuditLog.Reason,
	).VALUES(
		adminUUID, companyUUID, req.Action, req.Reason,
	).ExecContext(c.Request.Context(), tx); err != nil {
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

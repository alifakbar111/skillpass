package identity

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func companyIDFrom(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

// IssueDID issues (or returns) an employee's DID.
func (h *Handler) IssueDID(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	employeeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	res, err := h.svc.IssueDID(c.Request.Context(), companyID, employeeID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
			return
		}
		slog.Error("issue DID", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to issue DID"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GetDID returns an employee's DID.
func (h *Handler) GetDID(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	employeeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	res, err := h.svc.GetDID(c.Request.Context(), companyID, employeeID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "No DID issued"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get DID"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// ListCredentials returns an employee's signed credentials (with verification).
func (h *Handler) ListCredentials(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	employeeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	res, err := h.svc.ListCredentials(c.Request.Context(), companyID, employeeID)
	if err != nil {
		slog.Error("list credentials", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list credentials"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// Resolve is the public DID-document endpoint (no auth).
func (h *Handler) Resolve(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid DID id"})
		return
	}
	res, err := h.svc.Resolve(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "DID not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve DID"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// ---------- Sprint 6: attestations, verification, passport, public verify ----------

// AttestSkills signs an employee's skill scores from an evaluation.
func (h *Handler) AttestSkills(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	employeeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	var body struct {
		EvaluationID string `json:"evaluationId"`
	}
	_ = c.ShouldBindJSON(&body)
	var evalID *uuid.UUID
	if body.EvaluationID != "" {
		id, perr := uuid.Parse(body.EvaluationID)
		if perr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid evaluation ID"})
			return
		}
		evalID = &id
	}
	count, err := h.svc.AttestEmployeeSkills(c.Request.Context(), companyID, employeeID, evalID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "No evaluation found for this employee"})
			return
		}
		slog.Error("attest skills", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to attest skills"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"attested": count})
}

// ListAttestations returns an employee's signed skill attestations.
func (h *Handler) ListAttestations(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	employeeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	res, err := h.svc.ListAttestations(c.Request.Context(), companyID, employeeID)
	if err != nil {
		slog.Error("list attestations", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list attestations"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// RunVerification triggers a Dukcapil/PDDikti/manual identity check.
func (h *Handler) RunVerification(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	employeeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	var body struct {
		Provider string `json:"provider"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	res, err := h.svc.RunVerification(c.Request.Context(), companyID, employeeID, body.Provider)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		case errors.Is(err, ErrBadProvider):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown provider"})
		default:
			slog.Error("run verification", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Verification failed"})
		}
		return
	}
	c.JSON(http.StatusOK, res)
}

// ListVerifications returns an employee's identity verification history.
func (h *Handler) ListVerifications(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	employeeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	res, err := h.svc.ListVerifications(c.Request.Context(), companyID, employeeID)
	if err != nil {
		slog.Error("list verifications", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list verifications"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GetPassport returns/creates an employee's public passport settings.
func (h *Handler) GetPassport(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	employeeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	res, err := h.svc.GetOrCreatePassport(c.Request.Context(), companyID, employeeID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load passport"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// SetPassportVisibility toggles public visibility of an employee's passport.
func (h *Handler) SetPassportVisibility(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	employeeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}
	var body struct {
		IsPublic bool `json:"isPublic"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	res, err := h.svc.SetPassportPublic(c.Request.Context(), companyID, employeeID, body.IsPublic)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update passport"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// ---------- Public (no auth) ----------

// JWKS godoc
// @Summary Issuer public keys (JWKS)
// @Tags Verify
// @Produce json
// @Success 200 {object} JWKS
// @Router /.well-known/jwks.json [get]
func (h *Handler) JWKS(c *gin.Context) {
	c.JSON(http.StatusOK, h.svc.JWKS())
}

// VerifyCredential godoc
// @Summary Verify a signed credential (public)
// @Tags Verify
// @Produce json
// @Param id query string true "Attestation ID"
// @Success 200 {object} VerifiedCredential
// @Router /verify/credential [get]
func (h *Handler) VerifyCredential(c *gin.Context) {
	id, err := uuid.Parse(c.Query("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid credential id"})
		return
	}
	res, err := h.svc.VerifyCredential(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Credential not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Verification failed"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// PublicPassport godoc
// @Summary Public Skill Passport (verified badges)
// @Tags Verify
// @Produce json
// @Param slug path string true "Passport slug"
// @Success 200 {object} PublicPassport
// @Router /verify/passport/{slug} [get]
func (h *Handler) PublicPassport(c *gin.Context) {
	res, err := h.svc.GetPublicPassport(c.Request.Context(), c.Param("slug"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Passport not found or not public"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load passport"})
		return
	}
	c.JSON(http.StatusOK, res)
}

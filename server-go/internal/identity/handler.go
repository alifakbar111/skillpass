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

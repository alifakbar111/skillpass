package holiday

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc *Service
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{svc: NewService(db)}
}

func (h *Handler) List(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	year := time.Now().Year()
	if y := c.Query("year"); y != "" {
		if parsed, e := strconv.Atoi(y); e == nil {
			year = parsed
		}
	}
	list, err := h.svc.List(c.Request.Context(), companyID, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list holidays"})
		return
	}
	if list == nil {
		list = []Holiday{}
	}
	c.JSON(http.StatusOK, list)
}

func (h *Handler) Create(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	var hol Holiday
	if err := c.ShouldBindJSON(&hol); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	hol.CompanyID = companyID
	if err := h.svc.Create(c.Request.Context(), companyID, &hol); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create holiday"})
		return
	}
	c.JSON(http.StatusCreated, hol)
}

func (h *Handler) Update(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid holiday ID"})
		return
	}
	var hol Holiday
	if err := c.ShouldBindJSON(&hol); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if err := h.svc.Update(c.Request.Context(), companyID, id, &hol); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Holiday not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update holiday"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Updated"})
}

func (h *Handler) Delete(c *gin.Context) {
	companyID, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid holiday ID"})
		return
	}
	if err := h.svc.Delete(c.Request.Context(), companyID, id); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Holiday not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete holiday"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

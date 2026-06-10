package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetHealth		godoc
// @Summary			Health check
// @Description		Returns server health status with current timestamp
// @Tags			health
// @Produce			json
// @Success			200 {object} map[string]string
// @Router			/health [get]
func GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

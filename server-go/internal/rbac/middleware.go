package rbac

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequireCompanyMember(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("userId")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		userIDStr, ok := userIDVal.(string)
		if !ok || userIDStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		info, err := svc.GetEmployeeByUserID(c.Request.Context(), userID)
		if err != nil {
			if err == sql.ErrNoRows {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "No active employee record"})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Employee lookup failed"})
			return
		}

		c.Set("employeeId", info.EmployeeID.String())
		c.Set("companyId", info.CompanyID.String())
		c.Next()
	}
}

func RequirePermission(svc *Service, codes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		employeeIDVal, exists := c.Get("employeeId")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		employeeID, err := uuid.Parse(employeeIDVal.(string))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		has, err := svc.HasAnyPermission(c.Request.Context(), employeeID, codes)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Permission check failed"})
			return
		}
		if !has {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			return
		}

		c.Next()
	}
}

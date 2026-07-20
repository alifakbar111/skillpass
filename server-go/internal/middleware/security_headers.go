package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders adds a small set of defensive HTTP response headers to
// every request. The set is deliberately conservative so it does not
// break the SPA or the dev environment.
//
// HSTS is only sent in release mode because it is meaningless on a
// localhost-only dev server. (CSP is intentionally NOT added here; the
// SPA sets its own via meta tags, and a server-side CSP would block
// Vite HMR in dev.)
//
// Usage in cmd/server/main.go:
//   r.Use(middleware.SecurityHeaders())
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		if gin.Mode() == gin.ReleaseMode {
			h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}

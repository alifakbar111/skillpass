package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSecurityHeaders_AlwaysSent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SecurityHeaders())
	r.GET("/test", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options: got %q, want %q", got, "nosniff")
	}
	if got := w.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Errorf("X-Frame-Options: got %q, want %q", got, "DENY")
	}
	if got := w.Header().Get("Referrer-Policy"); got != "strict-origin-when-cross-origin" {
		t.Errorf("Referrer-Policy: got %q, want %q", got, "strict-origin-when-cross-origin")
	}
	if got := w.Header().Get("Permissions-Policy"); got != "camera=(), microphone=(), geolocation=()" {
		t.Errorf("Permissions-Policy: got %q", got)
	}
}

func TestSecurityHeaders_HSTSOnlyInRelease(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SecurityHeaders())
	r.GET("/test", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	if got := w.Header().Get("Strict-Transport-Security"); got != "" {
		t.Errorf("HSTS should be empty in test mode, got %q", got)
	}

	gin.SetMode(gin.ReleaseMode)
	defer gin.SetMode(gin.TestMode)
	r2 := gin.New()
	r2.Use(SecurityHeaders())
	r2.GET("/test", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	w2 := httptest.NewRecorder()
	r2.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/test", nil))
	if got := w2.Header().Get("Strict-Transport-Security"); got == "" {
		t.Errorf("HSTS should be set in release mode")
	}
}

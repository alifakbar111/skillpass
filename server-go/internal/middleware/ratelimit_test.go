package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRateLimiterAllow(t *testing.T) {
	rl := NewRateLimiter(10, 5) // 10 rps, burst of 5

	// First 5 requests should be allowed
	for i := 0; i < 5; i++ {
		if !rl.Allow("192.168.1.1") {
			t.Fatalf("request %d should be allowed within burst", i+1)
		}
	}
}

func TestRateLimiterMiddleware(t *testing.T) {
	rl := NewRateLimiter(1000, 100) // very high limit so we don't hit it

	router := gin.New()
	router.GET("/test", rl.Middleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRateLimiterDifferentIPs(t *testing.T) {
	rl := NewRateLimiter(100, 5)

	// Different IPs should have independent buckets
	if !rl.Allow("10.0.0.1") {
		t.Fatal("10.0.0.1 should be allowed")
	}
	if !rl.Allow("10.0.0.2") {
		t.Fatal("10.0.0.2 should be allowed")
	}
}

func TestClientIP(t *testing.T) {
	t.Run("uses X-Forwarded-For", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("X-Forwarded-For", "203.0.113.1, 10.0.0.1")
		c.Request.RemoteAddr = "10.0.0.1:12345"

		ip := clientIP(c)
		if ip != "203.0.113.1" {
			t.Fatalf("expected 203.0.113.1, got %s", ip)
		}
	})

	t.Run("falls back to RemoteAddr", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.RemoteAddr = "192.168.1.1:8080"

		ip := clientIP(c)
		if ip != "192.168.1.1" {
			t.Fatalf("expected 192.168.1.1, got %s", ip)
		}
	})
}

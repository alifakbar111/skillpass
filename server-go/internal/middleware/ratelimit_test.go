package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
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

func TestRateLimiterExceeded(t *testing.T) {
	rl := NewRateLimiter(100, 2) // 100 rps, burst of 2

	if !rl.Allow("10.0.0.1") {
		t.Fatal("request 1 should be allowed")
	}
	if !rl.Allow("10.0.0.1") {
		t.Fatal("request 2 should be allowed (within burst)")
	}
	if rl.Allow("10.0.0.1") {
		t.Fatal("request 3 should be rate limited (burst exceeded)")
	}
}

func TestRateLimiter429(t *testing.T) {
	rl := NewRateLimiter(100, 1) // burst of 1

	router := gin.New()
	router.GET("/test", rl.Middleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// First request should succeed
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-Forwarded-For", "10.0.0.1")
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w1.Code)
	}

	// Second request should be rate limited (burst exhausted)
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Forwarded-For", "10.0.0.1")
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w2.Code)
	}
}

func TestClientIP(t *testing.T) {
	// Reset the sync.Once-driven trusted-proxy cache so the test can
	// control which CIDRs are trusted.
	trustedProxyCIDRsOnce = sync.Once{}
	trustedProxyCIDRs = nil

	t.Run("ignores XFF when proxy is untrusted", func(t *testing.T) {
		t.Setenv("TRUSTED_PROXY_CIDRS", "")
		trustedProxyCIDRsOnce = sync.Once{}
		trustedProxyCIDRs = nil

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("X-Forwarded-For", "203.0.113.1, 10.0.0.1")
		c.Request.RemoteAddr = "10.0.0.1:12345"

		ip := clientIP(c)
		if ip != "10.0.0.1" {
			t.Fatalf("expected 10.0.0.1 (RemoteAddr), got %s", ip)
		}
	})

	t.Run("honors XFF when proxy is trusted", func(t *testing.T) {
		t.Setenv("TRUSTED_PROXY_CIDRS", "10.0.0.0/8")
		trustedProxyCIDRsOnce = sync.Once{}
		trustedProxyCIDRs = nil

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("X-Forwarded-For", "203.0.113.1, 10.0.0.1")
		c.Request.RemoteAddr = "10.0.0.1:12345"

		ip := clientIP(c)
		if ip != "203.0.113.1" {
			t.Fatalf("expected 203.0.113.1 (from XFF), got %s", ip)
		}
	})

	t.Run("falls back to RemoteAddr when no XFF", func(t *testing.T) {
		t.Setenv("TRUSTED_PROXY_CIDRS", "")
		trustedProxyCIDRsOnce = sync.Once{}
		trustedProxyCIDRs = nil

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

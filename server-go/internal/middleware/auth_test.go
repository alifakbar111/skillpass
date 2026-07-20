package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestAuthRequired(t *testing.T) {
	t.Run("no auth header", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)

		AuthRequired("secret")(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("Authorization", "Bearer invalid-token")

		AuthRequired("secret")(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("missing Bearer prefix", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("Authorization", "Token abc123")

		AuthRequired("secret")(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)

		claims := jwt.MapClaims{
			"userId": "test-user",
			"role":   "admin",
			"exp":    time.Now().Add(-1 * time.Hour).Unix(),
		}
		token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("secret"))
		c.Request.Header.Set("Authorization", "Bearer "+token)

		AuthRequired("secret")(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("malformed token", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("Authorization", "Bearer this-is-not-a-valid-jwt-token-format")

		AuthRequired("secret")(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("empty Bearer token", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("Authorization", "Bearer ")

		AuthRequired("secret")(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("query token is rejected", func(t *testing.T) {
		// HIGH-001 / C4: tokens in URL query strings leak into server
		// access logs, browser history, and HTTP referer headers. AuthRequired
		// must no longer accept them — SSE clients must use the exchange-ticket
		// flow instead.
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/notifications/stream?token=eyJhbGciOiJIUzI1NiJ9.eyJ1c2VySWQiOiJ1Ii.wrong", nil)

		AuthRequired("secret")(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 when ?token= is supplied, got %d", w.Code)
		}
		if c.IsAborted() != true {
			t.Fatal("expected request to be aborted when ?token= is supplied")
		}
	})
}

func TestRequireRole(t *testing.T) {
	t.Run("correct role passes", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("role", "admin")

		RequireRole("admin")(c)
		if c.IsAborted() {
			t.Fatal("should not abort for correct role")
		}
	})

	t.Run("wrong role returns 403", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("role", "jobseeker")

		RequireRole("company")(c)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
		if !c.IsAborted() {
			t.Fatal("should abort for wrong role")
		}
	})

	t.Run("no role set returns 403", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		RequireRole("admin")(c)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})

	t.Run("multiple allowed roles", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("role", "admin")

		RequireRole("admin")(c)
		RequireRole("admin")(c)
		if c.IsAborted() {
			t.Fatal("should not abort when all middleware match")
		}
	})
}

func TestSSERedeemMiddleware(t *testing.T) {
	t.Run("missing exchange ticket returns 401", func(t *testing.T) {
		store := NewStreamExchangeStore()
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/notifications/stream", nil)

		SSERedeemMiddleware(store)(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for missing exchange ticket, got %d", w.Code)
		}
	})

	t.Run("invalid exchange ticket returns 401", func(t *testing.T) {
		store := NewStreamExchangeStore()
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/notifications/stream?exchange=not-a-real-ticket", nil)

		SSERedeemMiddleware(store)(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for invalid exchange ticket, got %d", w.Code)
		}
	})

	t.Run("valid exchange ticket sets userId and role, single-use", func(t *testing.T) {
		store := NewStreamExchangeStore()
		nonce, err := store.Issue("user-123", "jobseeker", time.Minute)
		if err != nil {
			t.Fatalf("issue ticket: %v", err)
		}

		// First request: ticket is valid.
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/notifications/stream?exchange="+nonce, nil)
		SSERedeemMiddleware(store)(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 for valid exchange ticket, got %d", w.Code)
		}
		if c.IsAborted() {
			t.Fatal("valid ticket must not abort")
		}
		uid, _ := c.Get("userId")
		if uid != "user-123" {
			t.Fatalf("expected userId=user-123, got %v", uid)
		}
		role, _ := c.Get("role")
		if role != "jobseeker" {
			t.Fatalf("expected role=jobseeker, got %v", role)
		}

		// Second request with the same nonce: must be rejected (single-use).
		w2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(w2)
		c2.Request = httptest.NewRequest("GET", "/notifications/stream?exchange="+nonce, nil)
		SSERedeemMiddleware(store)(c2)
		if w2.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for replayed exchange ticket, got %d", w2.Code)
		}
	})

	t.Run("expired exchange ticket returns 401", func(t *testing.T) {
		store := NewStreamExchangeStore()
		nonce, err := store.Issue("user-456", "company", time.Microsecond)
		if err != nil {
			t.Fatalf("issue ticket: %v", err)
		}
		// Wait past the ttl.
		time.Sleep(10 * time.Millisecond)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/notifications/stream?exchange="+nonce, nil)
		SSERedeemMiddleware(store)(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for expired exchange ticket, got %d", w.Code)
		}
	})
}

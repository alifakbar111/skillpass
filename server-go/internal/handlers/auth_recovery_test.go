package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"

	"skillpass-server-go/internal/authtoken"
	"skillpass-server-go/internal/middleware"
	"skillpass-server-go/internal/testutil"
)

// recordingSender captures sent emails so tests can pull tokens out of links.
type recordingSender struct {
	mu    sync.Mutex
	to    string
	text  string
	count int
}

func (r *recordingSender) Send(_ context.Context, to, _, _, text string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.to = to
	r.text = text
	r.count++
	return nil
}

var tokenPattern = regexp.MustCompile(`token=([0-9a-f]+)`)

func (r *recordingSender) lastToken(t *testing.T) string {
	t.Helper()
	r.mu.Lock()
	defer r.mu.Unlock()
	m := tokenPattern.FindStringSubmatch(r.text)
	if m == nil {
		t.Fatalf("no token found in email text: %q", r.text)
	}
	return m[1]
}

func setupAuthRecoveryRouter(t *testing.T) (*gin.Engine, *recordingSender, func() string) {
	t.Helper()
	db := testutil.SetupTestDB()

	sender := &recordingSender{}
	h := NewAuthHandler(db, testutil.TestJWTSecret)
	h.SetEmailer(sender)
	h.SetTokenService(authtoken.NewService(db))

	router := gin.New()
	router.POST("/api/v1/auth/register", h.Register)
	router.POST("/api/v1/auth/login", h.Login)
	router.GET("/api/v1/auth/me", middleware.AuthRequired(testutil.TestJWTSecret), h.Me)
	router.GET("/api/v1/auth/verify-email", h.VerifyEmail)
	router.POST("/api/v1/auth/forgot-password", h.ForgotPassword)
	router.POST("/api/v1/auth/reset-password", h.ResetPassword)

	register := func() string {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(
			`{"email":"flow@ex.com","username":"flowuser","password":"password123","name":"Flow User","role":"jobseeker"}`))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("register: expected 201, got %d: %s", w.Code, w.Body.String())
		}
		var resp LoginResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		return resp.AccessToken
	}

	return router, sender, register
}

func TestEmailVerificationFlow(t *testing.T) {
	router, sender, register := setupAuthRecoveryRouter(t)
	accessToken := register()

	if sender.count != 1 {
		t.Fatalf("expected 1 verification email after register, got %d", sender.count)
	}
	if sender.to != "flow@ex.com" {
		t.Fatalf("verification email sent to %q", sender.to)
	}

	t.Run("me reports unverified before", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("me: expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var me UserResponse
		json.Unmarshal(w.Body.Bytes(), &me)
		if me.IsVerified {
			t.Fatal("expected isVerified false before verification")
		}
	})

	t.Run("verify with emailed token", func(t *testing.T) {
		token := sender.lastToken(t)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/auth/verify-email?token="+token, nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("verify: expected 200, got %d: %s", w.Code, w.Body.String())
		}

		// Token is idempotent — reusing the same token returns success
		// to handle React Strict Mode double-invocation and concurrent retries.
		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest("GET", "/api/v1/auth/verify-email?token="+token, nil)
		router.ServeHTTP(w2, req2)
		if w2.Code != http.StatusOK {
			t.Fatalf("reuse: expected 200, got %d: %s", w2.Code, w2.Body.String())
		}
	})

	t.Run("me reports verified after", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		router.ServeHTTP(w, req)
		var me UserResponse
		json.Unmarshal(w.Body.Bytes(), &me)
		if !me.IsVerified {
			t.Fatal("expected isVerified true after verification")
		}
	})

	t.Run("bogus token rejected", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/auth/verify-email?token=deadbeef", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})
}

func TestPasswordResetFlow(t *testing.T) {
	router, sender, register := setupAuthRecoveryRouter(t)
	register()

	t.Run("forgot password sends email", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/forgot-password", bytes.NewBufferString(`{"email":"flow@ex.com"}`))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		if sender.count != 2 { // 1 verification + 1 reset
			t.Fatalf("expected reset email to be sent, count=%d", sender.count)
		}
	})

	t.Run("unknown email gets same response, no email", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/forgot-password", bytes.NewBufferString(`{"email":"nobody@ex.com"}`))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 (anti-enumeration), got %d", w.Code)
		}
		if sender.count != 2 {
			t.Fatalf("no email should be sent for unknown account, count=%d", sender.count)
		}
	})

	t.Run("reset and login with new password", func(t *testing.T) {
		token := sender.lastToken(t)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/reset-password", bytes.NewBufferString(
			`{"token":"`+token+`","newPassword":"newpassword456"}`))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("reset: expected 200, got %d: %s", w.Code, w.Body.String())
		}

		// Old password no longer works.
		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(
			`{"email":"flow@ex.com","password":"password123"}`))
		req2.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w2, req2)
		if w2.Code != http.StatusUnauthorized {
			t.Fatalf("old password: expected 401, got %d", w2.Code)
		}

		// New password works.
		w3 := httptest.NewRecorder()
		req3 := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(
			`{"email":"flow@ex.com","password":"newpassword456"}`))
		req3.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w3, req3)
		if w3.Code != http.StatusOK {
			t.Fatalf("new password: expected 200, got %d: %s", w3.Code, w3.Body.String())
		}

		// Token is single-use.
		w4 := httptest.NewRecorder()
		req4 := httptest.NewRequest("POST", "/api/v1/auth/reset-password", bytes.NewBufferString(
			`{"token":"`+token+`","newPassword":"anotherpass789"}`))
		req4.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w4, req4)
		if w4.Code != http.StatusBadRequest {
			t.Fatalf("token reuse: expected 400, got %d", w4.Code)
		}
	})
}

func TestOGPage(t *testing.T) {
	db := testutil.SetupTestDB()
	_, _, err := testutil.CreateJobseeker(db, "og@example.com", "oguser", "password123", "OG User")
	if err != nil {
		t.Fatalf("create jobseeker: %v", err)
	}
	db.Exec(`UPDATE jobseeker_profiles SET headline = 'Senior Gopher' WHERE slug = 'oguser'`)

	router := gin.New()
	h := NewPassportHandler(db)
	router.GET("/p/:username", h.GetOGPage)

	t.Run("renders og tags", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/p/oguser", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		body := w.Body.String()
		for _, want := range []string{`og:title`, `OG User`, `Senior Gopher`, `/profiles/oguser`} {
			if !bytes.Contains([]byte(body), []byte(want)) {
				t.Fatalf("og page missing %q", want)
			}
		}
	})

	t.Run("404 for unknown slug", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/p/ghost", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})
}

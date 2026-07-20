package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"skillpass-server-go/internal/db"
	"skillpass-server-go/internal/middleware"
	"skillpass-server-go/internal/testutil"
)

func TestRegister(t *testing.T) {
	sqlDB := testutil.SetupTestDB()
	bunDB := db.NewBunDB(sqlDB)

	router := gin.New()
	h := NewAuthHandler(sqlDB, testutil.TestJWTSecret, bunDB)
	rl := middleware.NewRateLimiter(100, 200)
	router.POST("/api/v1/auth/register", rl.Middleware(), h.Register)

	t.Run("register jobseeker success", func(t *testing.T) {
		body := `{"email":"test@example.com","username":"testuser","password":"password123","name":"Test User","role":"jobseeker"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		var resp LoginResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if resp.AccessToken == "" {
			t.Fatal("expected access token")
		}
		if resp.User.Email != "test@example.com" {
			t.Fatalf("expected test@example.com, got %s", resp.User.Email)
		}
		if resp.User.Role != "jobseeker" {
			t.Fatalf("expected jobseeker, got %s", resp.User.Role)
		}
	})

	t.Run("register company success", func(t *testing.T) {
		body := `{"email":"company@example.com","username":"testcompany","password":"password123","name":"","role":"company","companyName":"Test Corp"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		var resp LoginResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.User.Name != "Test Corp" {
			t.Fatalf("expected 'Test Corp', got '%s'", resp.User.Name)
		}
	})

	t.Run("register duplicate email", func(t *testing.T) {
		body := `{"email":"test@example.com","username":"another","password":"password123","name":"Another","role":"jobseeker"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("register missing required fields", func(t *testing.T) {
		cases := []struct{ name, body string }{
			{"no email", `{"username":"u","password":"password123","name":"U","role":"jobseeker"}`},
			{"no username", `{"email":"u@test.com","password":"password123","name":"U","role":"jobseeker"}`},
			{"no password", `{"email":"u@test.com","username":"u","name":"U","role":"jobseeker"}`},
			{"invalid email", `{"email":"bad","username":"u","password":"password123","name":"U","role":"jobseeker"}`},
			{"short password", `{"email":"u@test.com","username":"u","password":"short","name":"U","role":"jobseeker"}`},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(tc.body))
				req.Header.Set("Content-Type", "application/json")
				router.ServeHTTP(w, req)
				if w.Code != http.StatusBadRequest {
					t.Fatalf("expected 400 for '%s', got %d: %s", tc.name, w.Code, w.Body.String())
				}
			})
		}
	})
}

func TestLogin(t *testing.T) {
	sqlDB := testutil.SetupTestDB()
	bunDB := db.NewBunDB(sqlDB)

	testutil.CreateUser(sqlDB, "login@example.com", "loginuser", "correct-password", "Login User", "jobseeker")

	router := gin.New()
	h := NewAuthHandler(sqlDB, testutil.TestJWTSecret, bunDB)
	rl := middleware.NewRateLimiter(100, 200)
	router.POST("/api/v1/auth/login", rl.Middleware(), h.Login)

	t.Run("login success", func(t *testing.T) {
		body := `{"email":"login@example.com","password":"correct-password"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("login wrong password", func(t *testing.T) {
		body := `{"email":"login@example.com","password":"wrong-password"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("login nonexistent email", func(t *testing.T) {
		body := `{"email":"nobody@example.com","password":"password123"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("login invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString("{invalid}"))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestRefresh(t *testing.T) {
	sqlDB := testutil.SetupTestDB()
	bunDB := db.NewBunDB(sqlDB)

	userID, err := testutil.CreateUser(sqlDB, "refresh@example.com", "refreshuser", "password123", "Refresh User", "jobseeker")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	refreshID := uuid.New()
	refreshTokenStr := testutil.GenerateRefreshToken(userID.String(), "jobseeker", refreshID.String(), 7*24*time.Hour)
	testutil.InsertRefreshToken(sqlDB, refreshID, userID, refreshTokenStr, time.Now().Add(7*24*time.Hour))

	router := gin.New()
	h := NewAuthHandler(sqlDB, testutil.TestJWTSecret, bunDB)
	rl := middleware.NewRateLimiter(100, 200)
	router.POST("/api/v1/auth/refresh", rl.Middleware(), h.Refresh)

	t.Run("refresh success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/refresh", nil)
		req.AddCookie(&http.Cookie{Name: "refreshToken", Value: refreshTokenStr, Path: "/api/v1/auth"})
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp map[string]string
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["accessToken"] == "" {
			t.Fatal("expected new access token")
		}
	})

	t.Run("refresh no cookie", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/refresh", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("refresh unknown token", func(t *testing.T) {
		badToken := testutil.GenerateRefreshToken(uuid.New().String(), "jobseeker", uuid.New().String(), 7*24*time.Hour)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/refresh", nil)
		req.AddCookie(&http.Cookie{Name: "refreshToken", Value: badToken, Path: "/api/v1/auth"})
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestLogout(t *testing.T) {
	sqlDB := testutil.SetupTestDB()
	bunDB := db.NewBunDB(sqlDB)

	userID, err := testutil.CreateUser(sqlDB, "logout@example.com", "logoutuser", "password123", "Logout User", "jobseeker")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	token := testutil.GenerateToken(userID.String(), "jobseeker", 15*time.Minute)

	router := gin.New()
	h := NewAuthHandler(sqlDB, testutil.TestJWTSecret, bunDB)
	router.POST("/api/v1/auth/logout", middleware.AuthRequired(testutil.TestJWTSecret), h.Logout)

	t.Run("logout success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("logout without auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})
}

func TestCreateSSETicket(t *testing.T) {
	sqlDB := testutil.SetupTestDB()
	bunDB := db.NewBunDB(sqlDB)

	uid, _, _ := testutil.CreateJobseeker(sqlDB, "sseticket@ex.com", "sseticket", "pass123", "SSE User")
	tok := testutil.GenerateToken(uid.String(), "jobseeker", 15*time.Minute)

	sseStore := middleware.NewStreamExchangeStore()

	router := gin.New()
	h := NewAuthHandler(sqlDB, testutil.TestJWTSecret, bunDB)
	h.SetSSEStore(sseStore)
	router.POST("/api/v1/auth/sse-ticket", middleware.AuthRequired(testutil.TestJWTSecret), h.CreateSSETicket)

	t.Run("issues a non-empty exchange ticket", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/sse-ticket", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp SSETicketResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if len(resp.Exchange) < 32 {
			t.Fatalf("expected a long opaque exchange ticket, got %q (len %d)", resp.Exchange, len(resp.Exchange))
		}
		if resp.ExpiresIn != 60 {
			t.Fatalf("expected expiresIn 60, got %d", resp.ExpiresIn)
		}
		if sseStore.Size() != 1 {
			t.Fatalf("expected store to have 1 entry, has %d", sseStore.Size())
		}
	})

	t.Run("rejects request without auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/sse-ticket", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("rejects request with query token (no longer accepted)", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/auth/sse-ticket?token="+tok, nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for ?token= on the ticket endpoint, got %d", w.Code)
		}
	})
}

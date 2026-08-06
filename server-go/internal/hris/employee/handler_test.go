package employee

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"skillpass-server-go/internal/middleware"
	"skillpass-server-go/internal/rbac"
	"skillpass-server-go/internal/testutil"
)

func TestGetMeAndUpdateMe(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "mec@ex.com", "mec", "pass123", "Me Co", true)
	empID, _ := testutil.CreateEmployee(db, cID, "GetMe", "User", testutil.UniqueEmail("getme"))

	// Create a user and link to employee so rbac.RequireCompanyMember can
	// resolve the employeeId/companyId for the authenticated user.
	userID, err := testutil.CreateUser(db, testutil.UniqueEmail("meuser"), "meuser", "pass123", "GetMe User", "company")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if _, err := db.ExecContext(context.Background(),
		`UPDATE employees SET user_id = $1 WHERE id = $2`, userID, empID); err != nil {
		t.Fatalf("link user to employee: %v", err)
	}

	rbacSvc := rbac.NewService(db)
	tok := testutil.GenerateToken(userID.String(), "company", 15*time.Minute)

	h := NewHandler(db, rbacSvc)

	router := gin.New()
	// Matches main.go: hris group uses AuthRequired + RequireCompanyMember.
	hris := router.Group("/api/v1/hris")
	hris.Use(middleware.AuthRequired(testutil.TestJWTSecret), rbac.RequireCompanyMember(rbacSvc))
	hris.GET("/me/employee", h.GetMe)
	hris.PUT("/me/employee", h.UpdateMe)

	t.Run("get me", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/hris/me/employee", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("update me valid NIK", func(t *testing.T) {
		body := `{"nationalId":"1234567890123456"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/v1/hris/me/employee", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("update me ignores HR-managed NIK", func(t *testing.T) {
		// M-4/SEC-06: NIK/NPWP/bank fields are HR-managed and no longer part of
		// the self-service whitelist — the server ignores them rather than
		// rejecting the whole request.
		body := `{"nationalId":"1234567890123456"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/v1/hris/me/employee", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var emp Employee
		if err := json.Unmarshal(w.Body.Bytes(), &emp); err != nil {
			t.Fatalf("parse response: %v", err)
		}
		if emp.NationalID != nil && *emp.NationalID == "1234567890123456" {
			t.Fatal("nationalId must not be settable via /me/employee")
		}
	})

	t.Run("update me ignores HR-managed NPWP", func(t *testing.T) {
		body := `{"npwp":"123456789012345"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/v1/hris/me/employee", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var emp Employee
		if err := json.Unmarshal(w.Body.Bytes(), &emp); err != nil {
			t.Fatalf("parse response: %v", err)
		}
		if emp.NPWP != nil && *emp.NPWP == "123456789012345" {
			t.Fatal("npwp must not be settable via /me/employee")
		}
	})
}

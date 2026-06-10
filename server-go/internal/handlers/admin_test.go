package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"skillpass-server-go/internal/middleware"
	"skillpass-server-go/internal/testutil"
)

func TestListPendingVerifications(t *testing.T) {
	db := testutil.SetupTestDB()

	aID, _ := testutil.CreateAdmin(db, "admin@ex.com", "admin", "adminpass123")
	aTok := testutil.GenerateToken(aID.String(), "admin", 15*time.Minute)

	testutil.CreateCompanyUser(db, "p1@ex.com", "p1", "pass123", "P1 Co", false)
	testutil.CreateCompanyUser(db, "p2@ex.com", "p2", "pass123", "P2 Co", false)
	testutil.CreateCompanyUser(db, "vco@ex.com", "vco", "pass123", "V Co", true)

	router := gin.New()
	h := NewAdminHandler(db)
	g := router.Group("/api/v1/admin")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("admin"))
	g.GET("/verifications/pending", h.ListPendingVerifications)

	t.Run("list pending", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/admin/verifications/pending", nil)
		req.Header.Set("Authorization", "Bearer "+aTok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp []PendingCompany
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) != 2 {
			t.Fatalf("expected 2 pending, got %d", len(resp))
		}
	})

	t.Run("wrong role", func(t *testing.T) {
		wt := testutil.GenerateToken(aID.String(), "jobseeker", 15*time.Minute)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/admin/verifications/pending", nil)
		req.Header.Set("Authorization", "Bearer "+wt)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})
}

func TestHandleVerification(t *testing.T) {
	db := testutil.SetupTestDB()

	aID, _ := testutil.CreateAdmin(db, "a2@ex.com", "a2", "adminpass123")
	aTok := testutil.GenerateToken(aID.String(), "admin", 15*time.Minute)
	_, cID, _ := testutil.CreateCompanyUser(db, "ap@ex.com", "ap", "pass123", "Approve Co", false)

	router := gin.New()
	h := NewAdminHandler(db)
	g := router.Group("/api/v1/admin")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("admin"))
	g.POST("/verifications/:id", h.HandleVerification)

	t.Run("approve", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/admin/verifications/%s", cID.String()), bytes.NewBufferString(`{"action":"approve"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+aTok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp PendingCompany
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.VerificationStatus != "verified" {
			t.Fatalf("expected 'verified', got '%s'", resp.VerificationStatus)
		}
	})

	t.Run("reject", func(t *testing.T) {
		_, c2, _ := testutil.CreateCompanyUser(db, "rj@ex.com", "rj", "pass123", "Reject Co", false)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/admin/verifications/%s", c2.String()), bytes.NewBufferString(`{"action":"reject","reason":"Bad docs"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+aTok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp PendingCompany
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.VerificationStatus != "rejected" {
			t.Fatalf("expected 'rejected', got '%s'", resp.VerificationStatus)
		}
	})

	t.Run("invalid action", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/admin/verifications/%s", cID.String()), bytes.NewBufferString(`{"action":"invalid"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+aTok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/admin/verifications/00000000-0000-0000-0000-000000000000", bytes.NewBufferString(`{"action":"approve"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+aTok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestListPendingVerifications_Empty(t *testing.T) {
	db := testutil.SetupTestDB()

	testutil.CreateCompanyUser(db, "vonly@ex.com", "vonly", "pass123", "Verified Only", true)

	aID, _ := testutil.CreateAdmin(db, "admin-empty@ex.com", "admin-empty", "adminpass123")
	aTok := testutil.GenerateToken(aID.String(), "admin", 15*time.Minute)

	router := gin.New()
	h := NewAdminHandler(db)
	g := router.Group("/api/v1/admin")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("admin"))
	g.GET("/verifications/pending", h.ListPendingVerifications)

	t.Run("no pending companies", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/admin/verifications/pending", nil)
		req.Header.Set("Authorization", "Bearer "+aTok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp []PendingCompany
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) != 0 {
			t.Fatalf("expected 0 pending, got %d", len(resp))
		}
	})
}

func TestListPendingVerifications_LimitOffset(t *testing.T) {
	db := testutil.SetupTestDB()

	aID, _ := testutil.CreateAdmin(db, "admin-lo@ex.com", "admin-lo", "adminpass123")
	aTok := testutil.GenerateToken(aID.String(), "admin", 15*time.Minute)

	for i := 1; i <= 4; i++ {
		email := fmt.Sprintf("lo%d@ex.com", i)
		user := fmt.Sprintf("lo%d", i)
		name := fmt.Sprintf("LO Co %d", i)
		testutil.CreateCompanyUser(db, email, user, "pass123", name, false)
	}

	router := gin.New()
	h := NewAdminHandler(db)
	g := router.Group("/api/v1/admin")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("admin"))
	g.GET("/verifications/pending", h.ListPendingVerifications)

	t.Run("limit 2", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/admin/verifications/pending?limit=2", nil)
		req.Header.Set("Authorization", "Bearer "+aTok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp []PendingCompany
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) != 2 {
			t.Fatalf("expected 2, got %d", len(resp))
		}
	})

	t.Run("offset 2", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/admin/verifications/pending?offset=2", nil)
		req.Header.Set("Authorization", "Bearer "+aTok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp []PendingCompany
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) != 2 {
			t.Fatalf("expected 2, got %d", len(resp))
		}
	})

	t.Run("offset beyond results", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/admin/verifications/pending?offset=10", nil)
		req.Header.Set("Authorization", "Bearer "+aTok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp []PendingCompany
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) != 0 {
			t.Fatalf("expected 0, got %d", len(resp))
		}
	})
}

func TestHandleVerification_RejectWithoutReason(t *testing.T) {
	db := testutil.SetupTestDB()

	aID, _ := testutil.CreateAdmin(db, "a-rwr@ex.com", "a-rwr", "adminpass123")
	aTok := testutil.GenerateToken(aID.String(), "admin", 15*time.Minute)
	_, cID, _ := testutil.CreateCompanyUser(db, "rwr@ex.com", "rwr", "pass123", "RWR Co", false)

	router := gin.New()
	h := NewAdminHandler(db)
	g := router.Group("/api/v1/admin")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("admin"))
	g.POST("/verifications/:id", h.HandleVerification)

	t.Run("reject without reason", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/admin/verifications/%s", cID.String()), bytes.NewBufferString(`{"action":"reject"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+aTok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp PendingCompany
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.VerificationStatus != "rejected" {
			t.Fatalf("expected 'rejected', got '%s'", resp.VerificationStatus)
		}
	})
}

func TestHandleVerification_AlreadyApproved(t *testing.T) {
	db := testutil.SetupTestDB()

	aID, _ := testutil.CreateAdmin(db, "a-aa@ex.com", "a-aa", "adminpass123")
	aTok := testutil.GenerateToken(aID.String(), "admin", 15*time.Minute)
	_, cID, _ := testutil.CreateCompanyUser(db, "aa@ex.com", "aa", "pass123", "AA Co", true)

	router := gin.New()
	h := NewAdminHandler(db)
	g := router.Group("/api/v1/admin")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("admin"))
	g.POST("/verifications/:id", h.HandleVerification)

	t.Run("approve already-approved", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/admin/verifications/%s", cID.String()), bytes.NewBufferString(`{"action":"approve"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+aTok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp PendingCompany
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.VerificationStatus != "verified" {
			t.Fatalf("expected 'verified', got '%s'", resp.VerificationStatus)
		}
	})
}

func TestHandleVerification_AlreadyRejected(t *testing.T) {
	db := testutil.SetupTestDB()

	aID, _ := testutil.CreateAdmin(db, "a-ar@ex.com", "a-ar", "adminpass123")
	aTok := testutil.GenerateToken(aID.String(), "admin", 15*time.Minute)
	_, cID, _ := testutil.CreateCompanyUser(db, "ar@ex.com", "ar", "pass123", "AR Co", false)

	db.Exec("UPDATE companies SET verification_status = 'rejected'::verification_status WHERE id = $1", cID.String())

	router := gin.New()
	h := NewAdminHandler(db)
	g := router.Group("/api/v1/admin")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("admin"))
	g.POST("/verifications/:id", h.HandleVerification)

	t.Run("reject already-rejected", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/admin/verifications/%s", cID.String()), bytes.NewBufferString(`{"action":"reject","reason":"Still bad"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+aTok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp PendingCompany
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.VerificationStatus != "rejected" {
			t.Fatalf("expected 'rejected', got '%s'", resp.VerificationStatus)
		}
	})
}

func TestHandleVerification_NoAuth(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "na@ex.com", "na", "pass123", "NA Co", false)

	router := gin.New()
	h := NewAdminHandler(db)
	g := router.Group("/api/v1/admin")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("admin"))
	g.POST("/verifications/:id", h.HandleVerification)

	t.Run("no auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/admin/verifications/%s", cID.String()), bytes.NewBufferString(`{"action":"approve"}`))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestHandleVerification_WrongRole(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "wr@ex.com", "wr", "pass123", "WR Co", false)
	uID, _, _ := testutil.CreateJobseeker(db, "wruser@ex.com", "wruser", "pass123", "Wrong Role User")
	wt := testutil.GenerateToken(uID.String(), "jobseeker", 15*time.Minute)

	router := gin.New()
	h := NewAdminHandler(db)
	g := router.Group("/api/v1/admin")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("admin"))
	g.POST("/verifications/:id", h.HandleVerification)

	t.Run("wrong role", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/admin/verifications/%s", cID.String()), bytes.NewBufferString(`{"action":"approve"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+wt)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
		}
	})
}

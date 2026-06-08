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

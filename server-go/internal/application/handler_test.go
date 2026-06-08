package application

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

func TestApplicationFlow(t *testing.T) {
	db := testutil.SetupTestDB()

	// Create company with two jobs (one open, one closed)
	cu, cID, _ := testutil.CreateCompanyUser(db, "aco@ex.com", "aco", "pass123", "App Co", true)
	jID, _ := testutil.CreateJob(db, cID, "Software Engineer", "Technology", true)
	cjID, _ := testutil.CreateJob(db, cID, "Closed Position", "Technology", false)

	// Create jobseeker
	uID, _, _ := testutil.CreateJobseeker(db, "app@ex.com", "app", "pass123", "Applicant")
	tok := testutil.GenerateToken(uID.String(), "jobseeker", 15*time.Minute)
	ctok := testutil.GenerateToken(cu.String(), "company", 15*time.Minute)

	svc := NewService(db)
	h := NewHandler(svc)

	router := gin.New()

	// Jobseeker routes
	ag := router.Group("/api/v1/jobs")
	ag.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("jobseeker"))
	ag.POST("/:id/apply", h.Apply)

	lg := router.Group("/api/v1/applications")
	lg.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("jobseeker"))
	lg.GET("/me", h.ListMyApplications)

	// Company routes
	sg := router.Group("/api/v1/applications")
	sg.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"), middleware.RequireVerifiedCompany(db))
	sg.PUT("/:id/status", h.UpdateStatus)

	t.Run("apply success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/jobs/%s/apply", jID.String()), nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("apply duplicate", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/jobs/%s/apply", jID.String()), nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("apply closed job", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/jobs/%s/apply", cjID.String()), nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("apply nonexistent job", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/jobs/00000000-0000-0000-0000-000000000000/apply", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("list my apps", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/applications/me", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp []ApplicationResult
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) != 1 {
			t.Fatalf("expected 1, got %d", len(resp))
		}
	})

	t.Run("update status", func(t *testing.T) {
		// Get app ID from listing
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/applications/me", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		var apps []ApplicationResult
		json.Unmarshal(w.Body.Bytes(), &apps)
		if len(apps) == 0 {
			t.Fatal("no apps")
		}
		appID := apps[0].ID

		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/applications/%s/status", appID), bytes.NewBufferString(`{"status":"reviewed"}`))
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("Authorization", "Bearer "+ctok)
		router.ServeHTTP(w2, req2)
		if w2.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w2.Code, w2.Body.String())
		}
	})
}

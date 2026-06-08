package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"skillpass-server-go/internal/middleware"
	"skillpass-server-go/internal/testutil"
)

func TestSearchCandidates(t *testing.T) {
	db := testutil.SetupTestDB()

	uID, _, _ := testutil.CreateCompanyUser(db, "sc@ex.com", "sc", "pass123", "Search Co", true)
	tok := testutil.GenerateToken(uID.String(), "company", 15*time.Minute)

	_, p1, _ := testutil.CreateJobseeker(db, "c1@ex.com", "c1", "pass123", "Candidate One")
	_, p2, _ := testutil.CreateJobseeker(db, "c2@ex.com", "c2", "pass123", "Candidate Two")

	db.Exec(`INSERT INTO job_experiences (id, profile_id, type, title, organization, start_date, is_current, skills_used) VALUES ($1,$2,'employment','Go Dev','Tech Co','2020-01',true,$3)`,
		"11111111-1111-1111-1111-111111111111", p1.String(), `{Go,PostgreSQL}`)
	db.Exec(`INSERT INTO job_experiences (id, profile_id, type, title, organization, start_date, is_current, skills_used) VALUES ($1,$2,'employment','React Dev','Web Co','2019-06',true,$3)`,
		"22222222-2222-2222-2222-222222222222", p2.String(), `{React,TypeScript}`)
	db.Exec("UPDATE jobseeker_profiles SET headline = 'Senior Go Developer' WHERE id = $1", p1.String())

	router := gin.New()
	h := NewSearchHandler(db)
	g := router.Group("/api/v1/search")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"), middleware.RequireVerifiedCompany(db))
	g.GET("/candidates", h.SearchCandidates)

	t.Run("by text", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/search/candidates?q=candidate", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp []CandidateResult
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) != 2 {
			t.Fatalf("expected 2, got %d", len(resp))
		}
	})

	t.Run("by skills", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/search/candidates?skills=Go", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		var resp []CandidateResult
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) != 1 {
			t.Fatalf("expected 1 with Go, got %d", len(resp))
		}
	})

	t.Run("no results", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/search/candidates?q=nonexistent", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		var resp []CandidateResult
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) != 0 {
			t.Fatalf("expected 0, got %d", len(resp))
		}
	})

	t.Run("no auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/search/candidates?q=test", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})
}

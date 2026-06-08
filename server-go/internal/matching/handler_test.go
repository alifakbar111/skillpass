package matching

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

func TestMatching(t *testing.T) {
	db := testutil.SetupTestDB()

	// Company with job requiring skills
	_, cID, _ := testutil.CreateCompanyUser(db, "mco@ex.com", "mco", "pass123", "Match Co", true)
	jID := "33333333-3333-3333-3333-333333333333"
	db.Exec(`INSERT INTO job_postings (id, company_id, title, description, industry, required_skills, status) VALUES ($1,$2,'Go Dev','Build Go services','Technology',$3,'open')`,
		jID, cID.String(), `{Go,PostgreSQL,Docker}`)

	// Jobseeker with evaluation
	uID, pID, _ := testutil.CreateJobseeker(db, "mee@ex.com", "mee", "pass123", "Matchee")
	testutil.CreateAIEvaluation(db, pID, 90)

	// Another company user for candidate matching
	cu2, _, _ := testutil.CreateCompanyUser(db, "mco2@ex.com", "mco2", "pass123", "Match Co 2", true)

	tok := testutil.GenerateToken(uID.String(), "jobseeker", 15*time.Minute)
	ctok := testutil.GenerateToken(cu2.String(), "company", 15*time.Minute)

	svc := NewService(db)
	h := NewHandler(svc)

	router := gin.New()

	// Jobseeker: match jobs
	jsg := router.Group("/api/v1/jobs")
	jsg.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("jobseeker"))
	jsg.GET("/matches", h.MatchJobs)

	// Company: match candidates
	cog := router.Group("/api/v1/candidates")
	cog.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"), middleware.RequireVerifiedCompany(db))
	cog.GET("/matches", h.MatchCandidates)

	t.Run("match jobs for jobseeker", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/jobs/matches", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp []JobMatch
		json.Unmarshal(w.Body.Bytes(), &resp)
		t.Logf("got %d job matches", len(resp))
	})

	t.Run("match candidates for company", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/candidates/matches?jobId="+jID, nil)
		req.Header.Set("Authorization", "Bearer "+ctok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp []CandidateMatch
		json.Unmarshal(w.Body.Bytes(), &resp)
		t.Logf("got %d candidate matches", len(resp))
	})

	t.Run("match candidates missing jobId", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/candidates/matches", nil)
		req.Header.Set("Authorization", "Bearer "+ctok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("match jobs wrong role", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/jobs/matches", nil)
		req.Header.Set("Authorization", "Bearer "+ctok) // company token
		router.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})
}

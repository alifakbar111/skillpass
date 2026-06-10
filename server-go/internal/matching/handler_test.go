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

	t.Run("match jobs no evaluations", func(t *testing.T) {
		u2, _, _ := testutil.CreateJobseeker(db, "mee2@ex.com", "mee2", "pass123", "Matchee2")
		t2 := testutil.GenerateToken(u2.String(), "jobseeker", 15*time.Minute)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/jobs/matches", nil)
		req.Header.Set("Authorization", "Bearer "+t2)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp []JobMatch
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) != 0 {
			t.Fatalf("expected 0 matches, got %d", len(resp))
		}
	})

	t.Run("match jobs no auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/jobs/matches", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("match candidates no matches", func(t *testing.T) {
		weirdID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
		db.Exec(`INSERT INTO job_postings (id, company_id, title, description, industry, required_skills, status) VALUES ($1,$2,'COBOL Dev','Mainframe','Technology',$3,'open')`,
			weirdID, cID.String(), `{COBOL,Fortran,PunchCard}`)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/candidates/matches?jobId="+weirdID, nil)
		req.Header.Set("Authorization", "Bearer "+ctok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp []CandidateMatch
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) != 0 {
			t.Fatalf("expected 0 candidate matches, got %d", len(resp))
		}
	})

	t.Run("match candidates different company", func(t *testing.T) {
		// A different verified company can also call MatchCandidates
		jID2 := "44444444-4444-4444-4444-444444444444"
		db.Exec(`INSERT INTO job_postings (id, company_id, title, description, industry, required_skills, status) VALUES ($1,$2,'Go Specialist','Go services','Technology',$3,'open')`,
			jID2, cID.String(), `{Go,PostgreSQL}`)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/candidates/matches?jobId="+jID2, nil)
		req.Header.Set("Authorization", "Bearer "+ctok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp []CandidateMatch
		json.Unmarshal(w.Body.Bytes(), &resp)
		t.Logf("different company got %d candidate matches", len(resp))
	})
}

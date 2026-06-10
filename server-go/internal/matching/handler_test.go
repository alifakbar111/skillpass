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

func TestSkillsGap(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "gapco@ex.com", "gapco", "pass123", "Gap Co", true)
	jID := "44444444-4444-4444-4444-444444444444"
	db.Exec(`INSERT INTO job_postings (id, company_id, title, description, industry, required_skills, status) VALUES ($1,$2,'Gap Dev','Build things','Technology',$3,'open')`,
		jID, cID.String(), `{Go,Kubernetes}`)

	// Candidate evaluated with Go (score 90 in the factory's skill_scores)
	uID, pID, _ := testutil.CreateJobseeker(db, "gapjs@ex.com", "gapjs", "pass123", "Gap JS")
	testutil.CreateAIEvaluation(db, pID, 90)
	tok := testutil.GenerateToken(uID.String(), "jobseeker", 15*time.Minute)

	// Candidate without evaluation
	uID2, _, _ := testutil.CreateJobseeker(db, "gapjs2@ex.com", "gapjs2", "pass123", "Gap JS 2")
	tok2 := testutil.GenerateToken(uID2.String(), "jobseeker", 15*time.Minute)

	svc := NewService(db)
	h := NewHandler(svc)

	router := gin.New()
	g := router.Group("/api/v1/jobs")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("jobseeker"))
	g.GET("/:id/skills-gap", h.SkillsGap)

	t.Run("matched and missing skills", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/jobs/"+jID+"/skills-gap", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var gap SkillsGap
		json.Unmarshal(w.Body.Bytes(), &gap)
		if !gap.HasEvaluation {
			t.Fatal("expected hasEvaluation true")
		}
		if len(gap.MatchedSkills) != 1 || gap.MatchedSkills[0] != "Go" {
			t.Fatalf("expected matched [Go], got %v", gap.MatchedSkills)
		}
		if len(gap.MissingSkills) != 1 || gap.MissingSkills[0] != "Kubernetes" {
			t.Fatalf("expected missing [Kubernetes], got %v", gap.MissingSkills)
		}
		if gap.MatchPercent != 50 {
			t.Fatalf("expected 50%%, got %v", gap.MatchPercent)
		}
	})

	t.Run("no evaluation yet", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/jobs/"+jID+"/skills-gap", nil)
		req.Header.Set("Authorization", "Bearer "+tok2)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var gap SkillsGap
		json.Unmarshal(w.Body.Bytes(), &gap)
		if gap.HasEvaluation {
			t.Fatal("expected hasEvaluation false")
		}
		if len(gap.MissingSkills) != 2 {
			t.Fatalf("expected all skills missing, got %v", gap.MissingSkills)
		}
	})

	t.Run("job not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/jobs/00000000-0000-0000-0000-000000000000/skills-gap", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})
}

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
		if len(resp) == 0 {
			t.Fatal("expected at least 1 job match for evaluated candidate (Go) vs job requiring Go")
		}
		if resp[0].Title != "Go Dev" {
			t.Fatalf("expected matched job 'Go Dev', got %q", resp[0].Title)
		}
		if resp[0].CompanyName != "Match Co" {
			t.Fatalf("expected company 'Match Co', got %q", resp[0].CompanyName)
		}
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
		if len(resp) == 0 {
			t.Fatal("expected at least 1 candidate match (evaluated candidate has Go, job requires Go)")
		}
		if resp[0].Name != "Matchee" {
			t.Fatalf("expected candidate 'Matchee', got %q", resp[0].Name)
		}
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

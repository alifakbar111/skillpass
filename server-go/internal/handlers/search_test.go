package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"skillpass-server-go/internal/db"
	"skillpass-server-go/internal/middleware"
	"skillpass-server-go/internal/testutil"
)

func TestSearchCandidates(t *testing.T) {
	sqlDB := testutil.SetupTestDB()
	bunDB := db.NewBunDB(sqlDB)

	uID, _, _ := testutil.CreateCompanyUser(sqlDB, "sc@ex.com", "sc", "pass123", "Search Co", true)
	tok := testutil.GenerateToken(uID.String(), "company", 15*time.Minute)

	_, p1, _ := testutil.CreateJobseeker(sqlDB, "c1@ex.com", "c1", "pass123", "Candidate One")
	_, p2, _ := testutil.CreateJobseeker(sqlDB, "c2@ex.com", "c2", "pass123", "Candidate Two")

	sqlDB.Exec(`INSERT INTO job_experiences (id, profile_id, type, title, organization, start_date, is_current, skills_used) VALUES ($1,$2,'employment','Go Dev','Tech Co','2020-01',true,$3)`,
		"11111111-1111-1111-1111-111111111111", p1.String(), `{Go,PostgreSQL}`)
	sqlDB.Exec(`INSERT INTO job_experiences (id, profile_id, type, title, organization, start_date, is_current, skills_used) VALUES ($1,$2,'employment','React Dev','Web Co','2019-06',true,$3)`,
		"22222222-2222-2222-2222-222222222222", p2.String(), `{React,TypeScript}`)
	sqlDB.Exec("UPDATE jobseeker_profiles SET headline = 'Senior Go Developer' WHERE id = $1", p1.String())

	router := gin.New()
	h := NewSearchHandler(sqlDB, bunDB)
	g := router.Group("/api/v1/search")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"), middleware.RequireVerifiedCompany(bunDB))
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

func TestSearchCandidates_FilterByIndustry(t *testing.T) {
	sqlDB := testutil.SetupTestDB()
	bunDB := db.NewBunDB(sqlDB)

	uID, _, _ := testutil.CreateCompanyUser(sqlDB, "si@ex.com", "si", "pass123", "Search Ind Co", true)
	tok := testutil.GenerateToken(uID.String(), "company", 15*time.Minute)

	_, p1, _ := testutil.CreateJobseeker(sqlDB, "si1@ex.com", "si1", "pass123", "Ind One")
	_, p2, _ := testutil.CreateJobseeker(sqlDB, "si2@ex.com", "si2", "pass123", "Ind Two")

	sqlDB.Exec(`INSERT INTO job_experiences (id, profile_id, type, title, organization, start_date, is_current, industry, skills_used) VALUES ($1,$2,'employment','Go Dev','Tech Co','2020-01',true,$3,$4)`,
		"33333333-3333-3333-3333-333333333333", p1.String(), "Technology", `{Go,PostgreSQL}`)
	sqlDB.Exec(`INSERT INTO job_experiences (id, profile_id, type, title, organization, start_date, is_current, industry, skills_used) VALUES ($1,$2,'employment','Nurse','Hosp A','2019-06',true,$3,$4)`,
		"44444444-4444-4444-4444-444444444444", p2.String(), "Healthcare", `{Nursing}`)

	router := gin.New()
	h := NewSearchHandler(sqlDB, bunDB)
	g := router.Group("/api/v1/search")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"), middleware.RequireVerifiedCompany(bunDB))
	g.GET("/candidates", h.SearchCandidates)

	t.Run("filter by industry", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/search/candidates?industry=Technology", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp []CandidateResult
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) != 1 {
			t.Fatalf("expected 1 with Technology industry, got %d", len(resp))
		}
	})
}

func TestSearchCandidates_Combined(t *testing.T) {
	sqlDB := testutil.SetupTestDB()
	bunDB := db.NewBunDB(sqlDB)

	uID, _, _ := testutil.CreateCompanyUser(sqlDB, "scb@ex.com", "scb", "pass123", "Search Comb Co", true)
	tok := testutil.GenerateToken(uID.String(), "company", 15*time.Minute)

	_, p1, _ := testutil.CreateJobseeker(sqlDB, "cb1@ex.com", "cb1", "pass123", "Senior Go Dev")
	_, p2, _ := testutil.CreateJobseeker(sqlDB, "cb2@ex.com", "cb2", "pass123", "Junior React Dev")

	sqlDB.Exec(`INSERT INTO job_experiences (id, profile_id, type, title, organization, start_date, is_current, industry, skills_used) VALUES ($1,$2,'employment','Go Dev','Tech Co','2020-01',true,$3,$4)`,
		"55555555-5555-5555-5555-555555555555", p1.String(), "Technology", `{Go,PostgreSQL}`)
	sqlDB.Exec(`INSERT INTO job_experiences (id, profile_id, type, title, organization, start_date, is_current, industry, skills_used) VALUES ($1,$2,'employment','React Dev','Web Co','2019-06',true,$3,$4)`,
		"66666666-6666-6666-6666-666666666666", p2.String(), "Technology", `{React,TypeScript}`)
	sqlDB.Exec("UPDATE jobseeker_profiles SET headline = 'Senior Go Developer' WHERE id = $1", p1.String())

	router := gin.New()
	h := NewSearchHandler(sqlDB, bunDB)
	g := router.Group("/api/v1/search")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"), middleware.RequireVerifiedCompany(bunDB))
	g.GET("/candidates", h.SearchCandidates)

	t.Run("combined q + skills + industry", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/search/candidates?q=senior&skills=Go&industry=Technology", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp []CandidateResult
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) != 1 {
			t.Fatalf("expected 1 with all filters, got %d", len(resp))
		}
	})
}

func TestSearchCandidates_Pagination(t *testing.T) {
	sqlDB := testutil.SetupTestDB()
	bunDB := db.NewBunDB(sqlDB)

	uID, _, _ := testutil.CreateCompanyUser(sqlDB, "sp@ex.com", "sp", "pass123", "Search Pag Co", true)
	tok := testutil.GenerateToken(uID.String(), "company", 15*time.Minute)

	for i := 1; i <= 3; i++ {
		email := fmt.Sprintf("pag%d@ex.com", i)
		user := fmt.Sprintf("pag%d", i)
		name := fmt.Sprintf("Paginated %d", i)
		_, pID, _ := testutil.CreateJobseeker(sqlDB, email, user, "pass123", name)
		sqlDB.Exec(`INSERT INTO job_experiences (id, profile_id, type, title, organization, start_date, is_current, skills_used) VALUES ($1,$2,'employment','Dev','Co','2020-01',true,$3)`,
			fmt.Sprintf("70000000-0000-0000-0000-00000000000%d", i), pID.String(), `{Go}`)
	}

	router := gin.New()
	h := NewSearchHandler(sqlDB, bunDB)
	g := router.Group("/api/v1/search")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"), middleware.RequireVerifiedCompany(bunDB))
	g.GET("/candidates", h.SearchCandidates)

	t.Run("limit 2", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/search/candidates?q=paginated&limit=2", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp []CandidateResult
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) > 2 {
			t.Fatalf("expected at most 2, got %d", len(resp))
		}
	})

	t.Run("offset 1", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/search/candidates?q=paginated&offset=1", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp []CandidateResult
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) != 2 {
			t.Fatalf("expected 2 with offset 1, got %d", len(resp))
		}
	})
}

func TestSearchCandidates_InvalidQueryParams(t *testing.T) {
	sqlDB := testutil.SetupTestDB()
	bunDB := db.NewBunDB(sqlDB)

	uID, _, _ := testutil.CreateCompanyUser(sqlDB, "siq@ex.com", "siq", "pass123", "Search Inv Co", true)
	tok := testutil.GenerateToken(uID.String(), "company", 15*time.Minute)

	router := gin.New()
	h := NewSearchHandler(sqlDB, bunDB)
	g := router.Group("/api/v1/search")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"), middleware.RequireVerifiedCompany(bunDB))
	g.GET("/candidates", h.SearchCandidates)

	t.Run("negative limit", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/search/candidates?limit=-1", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("negative offset", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/search/candidates?offset=-1", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("non-numeric limit", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/search/candidates?limit=abc", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestSearchCandidates_UnverifiedCompany(t *testing.T) {
	sqlDB := testutil.SetupTestDB()
	bunDB := db.NewBunDB(sqlDB)

	uID, _, _ := testutil.CreateCompanyUser(sqlDB, "uv@ex.com", "uv", "pass123", "Unverified Co", false)
	tok := testutil.GenerateToken(uID.String(), "company", 15*time.Minute)

	router := gin.New()
	h := NewSearchHandler(sqlDB, bunDB)
	g := router.Group("/api/v1/search")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"), middleware.RequireVerifiedCompany(bunDB))
	g.GET("/candidates", h.SearchCandidates)

	t.Run("unverified company gets 403", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/search/candidates?q=test", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
		}
	})
}

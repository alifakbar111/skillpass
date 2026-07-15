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

	"skillpass-server-go/internal/db"
	"skillpass-server-go/internal/middleware"
	"skillpass-server-go/internal/testutil"
)

func TestListJobs(t *testing.T) {
	sqlDB := testutil.SetupTestDB()
	bunDB := db.NewBunDB(sqlDB)

	_, cID, _ := testutil.CreateCompanyUser(sqlDB, "jc@ex.com", "jc", "pass123", "Job Co", true)
	testutil.CreateJob(sqlDB, cID, "Software Engineer", "Technology", true)
	testutil.CreateJob(sqlDB, cID, "Doctor", "Healthcare", false)

	router := gin.New()
	h := NewJobHandler(bunDB)
	router.GET("/api/v1/jobs", h.ListJobs)

	t.Run("list only open", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/jobs", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp []JobResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) != 1 {
			t.Fatalf("expected 1 open job, got %d", len(resp))
		}
	})

	t.Run("filter by industry", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/jobs?industry=Technology", nil)
		router.ServeHTTP(w, req)
		var resp []JobResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) != 1 {
			t.Fatalf("expected 1, got %d", len(resp))
		}
	})

	t.Run("no match", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/jobs?industry=NoExist", nil)
		router.ServeHTTP(w, req)
		var resp []JobResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) != 0 {
			t.Fatalf("expected 0, got %d", len(resp))
		}
	})

	t.Run("pagination", func(t *testing.T) {
		testutil.CreateJob(sqlDB, cID, "Extra 1", "Tech", true)
		testutil.CreateJob(sqlDB, cID, "Extra 2", "Tech", true)
		testutil.CreateJob(sqlDB, cID, "Extra 3", "Tech", true)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/jobs?limit=2&offset=0", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp []JobResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) != 2 {
			t.Fatalf("expected 2, got %d", len(resp))
		}
	})

	t.Run("filter by experience_level", func(t *testing.T) {
		eID, _ := testutil.CreateJob(sqlDB, cID, "Entry Job", "Tech", true)
		sqlDB.Exec(`UPDATE job_postings SET experience_level = 'entry'::experience_level WHERE id = $1`, eID)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/jobs?experience_level=entry", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp []JobResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) != 1 {
			t.Fatalf("expected 1, got %d", len(resp))
		}
	})

	t.Run("industry + experience_level", func(t *testing.T) {
		mID, _ := testutil.CreateJob(sqlDB, cID, "Mid Tech", "Technology", true)
		testutil.CreateJob(sqlDB, cID, "Mid Health", "Healthcare", true)
		sqlDB.Exec(`UPDATE job_postings SET experience_level = 'mid'::experience_level WHERE id = $1`, mID)
		sqlDB.Exec(`UPDATE job_postings SET experience_level = 'mid'::experience_level WHERE title = 'Mid Health'`)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/jobs?industry=Technology&experience_level=mid", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp []JobResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) != 1 {
			t.Fatalf("expected 1, got %d", len(resp))
		}
	})

	t.Run("closed not listed", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/jobs", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp []JobResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		for _, j := range resp {
			if j.Status == "closed" {
				t.Fatalf("found closed job in list: %s", j.ID)
			}
		}
	})

	t.Run("negative limit defaults", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/jobs?limit=-1", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		var resp []JobResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) == 0 {
			t.Fatal("expected results with defaulted negative limit")
		}
	})

	t.Run("offset beyond results", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/jobs?offset=100", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp []JobResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) != 0 {
			t.Fatalf("expected 0, got %d", len(resp))
		}
	})
}

func TestGetJob(t *testing.T) {
	sqlDB := testutil.SetupTestDB()
	bunDB := db.NewBunDB(sqlDB)

	_, cID, _ := testutil.CreateCompanyUser(sqlDB, "gj@ex.com", "gj", "pass123", "GJ Co", true)
	jID, _ := testutil.CreateJob(sqlDB, cID, "Backend Engineer", "Technology", true)

	router := gin.New()
	h := NewJobHandler(bunDB)
	router.GET("/api/v1/jobs/:id", h.GetJob)

	t.Run("by id", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/jobs/%s", jID.String()), nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/jobs/invalid", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/jobs/00000000-0000-0000-0000-000000000000", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("closed job", func(t *testing.T) {
		closedID, _ := testutil.CreateJob(sqlDB, cID, "Closed Job", "Tech", false)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/jobs/%s", closedID.String()), nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp JobResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Status != "closed" {
			t.Fatalf("expected closed status, got %s", resp.Status)
		}
	})
}

func TestListMyJobs(t *testing.T) {
	sqlDB := testutil.SetupTestDB()
	bunDB := db.NewBunDB(sqlDB)

	uID, cID, _ := testutil.CreateCompanyUser(sqlDB, "mj@ex.com", "mj", "pass123", "MJ Co", true)
	testutil.CreateJob(sqlDB, cID, "Job 1", "Tech", true)
	testutil.CreateJob(sqlDB, cID, "Job 2", "Tech", false)
	tok := testutil.GenerateToken(uID.String(), "company", 15*time.Minute)
	uID2, _, _ := testutil.CreateCompanyUser(sqlDB, "mj-empty@ex.com", "mj-empty", "pass123", "Empty Co", true)
	tokNoJobs := testutil.GenerateToken(uID2.String(), "company", 15*time.Minute)
	jsUID, _, _ := testutil.CreateJobseeker(sqlDB, "js-mj@ex.com", "js-mj", "pass123", "Job Seeker")
	tokJobseeker := testutil.GenerateToken(jsUID.String(), "jobseeker", 15*time.Minute)

	router := gin.New()
	h := NewJobHandler(bunDB)
	g := router.Group("/api/v1/jobs")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"), middleware.RequireVerifiedCompany(bunDB))
	g.GET("/me", h.ListMyJobs)

	t.Run("all statuses", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/jobs/me", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp []JobResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) != 2 {
			t.Fatalf("expected 2 jobs, got %d", len(resp))
		}
	})

	t.Run("no jobs", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/jobs/me", nil)
		req.Header.Set("Authorization", "Bearer "+tokNoJobs)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp []JobResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) != 0 {
			t.Fatalf("expected 0 jobs, got %d", len(resp))
		}
	})

	t.Run("no auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/jobs/me", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("wrong role", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/jobs/me", nil)
		req.Header.Set("Authorization", "Bearer "+tokJobseeker)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})
}

func TestCreateJob(t *testing.T) {
	sqlDB := testutil.SetupTestDB()
	bunDB := db.NewBunDB(sqlDB)

	uID, _, _ := testutil.CreateCompanyUser(sqlDB, "cj@ex.com", "cj", "pass123", "CJ Co", true)
	tok := testutil.GenerateToken(uID.String(), "company", 15*time.Minute)
	uIDUnv, _, _ := testutil.CreateCompanyUser(sqlDB, "cj-unv@ex.com", "cj-unv", "pass123", "Unverified Co", false)
	tokUnverified := testutil.GenerateToken(uIDUnv.String(), "company", 15*time.Minute)

	router := gin.New()
	h := NewJobHandler(bunDB)
	g := router.Group("/api/v1/jobs")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"), middleware.RequireVerifiedCompany(bunDB))
	g.POST("", h.CreateJob)

	t.Run("success", func(t *testing.T) {
		body := `{"title":"New Job","description":"Great","industry":"Technology"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/jobs", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("missing fields", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/jobs", bytes.NewBufferString(`{"title":"Only Title"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("no auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/jobs", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("invalid experience level", func(t *testing.T) {
		body := `{"title":"Bad","description":"desc","industry":"Tech","experienceLevel":"invalid"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/jobs", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("with optional fields", func(t *testing.T) {
		body := `{"title":"Full","description":"desc","industry":"Technology","tags":["go","backend"],"requiredSkills":["Go","PSQL"],"experienceLevel":"senior","location":"Remote","salaryRange":"$100k"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/jobs", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		var resp JobResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.ExperienceLevel == nil || *resp.ExperienceLevel != "senior" {
			t.Fatalf("expected senior experience level")
		}
	})

	t.Run("empty title", func(t *testing.T) {
		body := `{"title":"","description":"desc","industry":"Tech"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/jobs", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("unverified company", func(t *testing.T) {
		body := `{"title":"New","description":"desc","industry":"Tech"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/jobs", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokUnverified)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestUpdateJob(t *testing.T) {
	sqlDB := testutil.SetupTestDB()
	bunDB := db.NewBunDB(sqlDB)

	uID, cID, _ := testutil.CreateCompanyUser(sqlDB, "uj@ex.com", "uj", "pass123", "UJ Co", true)
	jID, _ := testutil.CreateJob(sqlDB, cID, "Original", "Tech", true)
	tok := testutil.GenerateToken(uID.String(), "company", 15*time.Minute)
	uID2, _, _ := testutil.CreateCompanyUser(sqlDB, "uj-other@ex.com", "uj-other", "pass123", "Other Co", true)
	tokOther := testutil.GenerateToken(uID2.String(), "company", 15*time.Minute)

	router := gin.New()
	h := NewJobHandler(bunDB)
	g := router.Group("/api/v1/jobs")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"), middleware.RequireVerifiedCompany(bunDB))
	g.PUT("/:id", h.UpdateJob)

	t.Run("update title", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/jobs/%s", jID.String()), bytes.NewBufferString(`{"title":"Updated"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/v1/jobs/00000000-0000-0000-0000-000000000000", bytes.NewBufferString(`{"title":"X"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/v1/jobs/invalid", bytes.NewBufferString(`{"title":"X"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("update all fields", func(t *testing.T) {
		body := `{"title":"New Title","description":"New Desc","industry":"Healthcare","tags":["new"],"requiredSkills":["X"],"experienceLevel":"lead","location":"Office","salaryRange":"$200k"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/jobs/%s", jID.String()), bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("no fields", func(t *testing.T) {
		body := `{}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/jobs/%s", jID.String()), bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("wrong company", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/jobs/%s", jID.String()), bytes.NewBufferString(`{"title":"X"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokOther)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("status to closed", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/jobs/%s", jID.String()), bytes.NewBufferString(`{"status":"closed"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp JobResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Status != "closed" {
			t.Fatalf("expected closed, got %s", resp.Status)
		}
	})
}

func TestDeleteJob(t *testing.T) {
	sqlDB := testutil.SetupTestDB()
	bunDB := db.NewBunDB(sqlDB)

	uID, cID, _ := testutil.CreateCompanyUser(sqlDB, "dj@ex.com", "dj", "pass123", "DJ Co", true)
	jID, _ := testutil.CreateJob(sqlDB, cID, "To Delete", "Tech", true)
	tok := testutil.GenerateToken(uID.String(), "company", 15*time.Minute)
	jID2, _ := testutil.CreateJob(sqlDB, cID, "To Delete 2", "Tech", true)
	uID2, _, _ := testutil.CreateCompanyUser(sqlDB, "dj-other@ex.com", "dj-other", "pass123", "Other Co", true)
	tokOther := testutil.GenerateToken(uID2.String(), "company", 15*time.Minute)
	jsUID, _, _ := testutil.CreateJobseeker(sqlDB, "js-dj@ex.com", "js-dj", "pass123", "Job Seeker")
	tokJobseeker := testutil.GenerateToken(jsUID.String(), "jobseeker", 15*time.Minute)

	router := gin.New()
	h := NewJobHandler(bunDB)
	g := router.Group("/api/v1/jobs")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"), middleware.RequireVerifiedCompany(bunDB))
	g.DELETE("/:id", h.DeleteJob)

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/jobs/%s", jID.String()), nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/api/v1/jobs/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/api/v1/jobs/invalid", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("no auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/jobs/%s", jID2.String()), nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("wrong company", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/jobs/%s", jID2.String()), nil)
		req.Header.Set("Authorization", "Bearer "+tokOther)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("wrong role", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/jobs/%s", jID2.String()), nil)
		req.Header.Set("Authorization", "Bearer "+tokJobseeker)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})
}

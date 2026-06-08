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

func TestListJobs(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "jc@ex.com", "jc", "pass123", "Job Co", true)
	testutil.CreateJob(db, cID, "Software Engineer", "Technology", true)
	testutil.CreateJob(db, cID, "Doctor", "Healthcare", false)

	router := gin.New()
	h := NewJobHandler(db)
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
}

func TestGetJob(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "gj@ex.com", "gj", "pass123", "GJ Co", true)
	jID, _ := testutil.CreateJob(db, cID, "Backend Engineer", "Technology", true)

	router := gin.New()
	h := NewJobHandler(db)
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
}

func TestListMyJobs(t *testing.T) {
	db := testutil.SetupTestDB()

	uID, cID, _ := testutil.CreateCompanyUser(db, "mj@ex.com", "mj", "pass123", "MJ Co", true)
	testutil.CreateJob(db, cID, "Job 1", "Tech", true)
	testutil.CreateJob(db, cID, "Job 2", "Tech", false)
	tok := testutil.GenerateToken(uID.String(), "company", 15*time.Minute)

	router := gin.New()
	h := NewJobHandler(db)
	g := router.Group("/api/v1/jobs")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"), middleware.RequireVerifiedCompany(db))
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
}

func TestCreateJob(t *testing.T) {
	db := testutil.SetupTestDB()

	uID, _, _ := testutil.CreateCompanyUser(db, "cj@ex.com", "cj", "pass123", "CJ Co", true)
	tok := testutil.GenerateToken(uID.String(), "company", 15*time.Minute)

	router := gin.New()
	h := NewJobHandler(db)
	g := router.Group("/api/v1/jobs")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"), middleware.RequireVerifiedCompany(db))
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
}

func TestUpdateJob(t *testing.T) {
	db := testutil.SetupTestDB()

	uID, cID, _ := testutil.CreateCompanyUser(db, "uj@ex.com", "uj", "pass123", "UJ Co", true)
	jID, _ := testutil.CreateJob(db, cID, "Original", "Tech", true)
	tok := testutil.GenerateToken(uID.String(), "company", 15*time.Minute)

	router := gin.New()
	h := NewJobHandler(db)
	g := router.Group("/api/v1/jobs")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"), middleware.RequireVerifiedCompany(db))
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
}

func TestDeleteJob(t *testing.T) {
	db := testutil.SetupTestDB()

	uID, cID, _ := testutil.CreateCompanyUser(db, "dj@ex.com", "dj", "pass123", "DJ Co", true)
	jID, _ := testutil.CreateJob(db, cID, "To Delete", "Tech", true)
	tok := testutil.GenerateToken(uID.String(), "company", 15*time.Minute)

	router := gin.New()
	h := NewJobHandler(db)
	g := router.Group("/api/v1/jobs")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"), middleware.RequireVerifiedCompany(db))
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
}

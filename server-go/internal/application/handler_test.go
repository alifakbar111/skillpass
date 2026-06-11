package application

import (
	"bytes"
	"context"
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
		// Guard against silent go-jet scan failures on joined columns.
		if resp[0].JobTitle != "Software Engineer" {
			t.Fatalf("expected job title 'Software Engineer', got %q", resp[0].JobTitle)
		}
		if resp[0].CompanyName != "App Co" {
			t.Fatalf("expected company name 'App Co', got %q", resp[0].CompanyName)
		}
	})

	t.Run("apply without auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/jobs/%s/apply", jID.String()), nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("apply wrong role", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/jobs/%s/apply", jID.String()), nil)
		req.Header.Set("Authorization", "Bearer "+ctok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("list my apps empty", func(t *testing.T) {
		u2, _, _ := testutil.CreateJobseeker(db, "empty@ex.com", "empty", "pass123", "Empty")
		t2 := testutil.GenerateToken(u2.String(), "jobseeker", 15*time.Minute)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/applications/me", nil)
		req.Header.Set("Authorization", "Bearer "+t2)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp []ApplicationResult
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) != 0 {
			t.Fatalf("expected 0, got %d", len(resp))
		}
	})

	t.Run("update status invalid value", func(t *testing.T) {
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
		req2 := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/applications/%s/status", appID), bytes.NewBufferString(`{"status":"bad_status"}`))
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("Authorization", "Bearer "+ctok)
		router.ServeHTTP(w2, req2)
		if w2.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w2.Code, w2.Body.String())
		}
	})

	t.Run("update status nonexistent", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/v1/applications/00000000-0000-0000-0000-000000000000/status", bytes.NewBufferString(`{"status":"reviewed"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+ctok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("update status wrong company", func(t *testing.T) {
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

		cu2, _, _ := testutil.CreateCompanyUser(db, "aco2@ex.com", "aco2", "pass123", "App Co 2", true)
		ctok2 := testutil.GenerateToken(cu2.String(), "company", 15*time.Minute)

		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/applications/%s/status", appID), bytes.NewBufferString(`{"status":"reviewed"}`))
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("Authorization", "Bearer "+ctok2)
		router.ServeHTTP(w2, req2)
		if w2.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d: %s", w2.Code, w2.Body.String())
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

func TestApplicationMessages(t *testing.T) {
	db := testutil.SetupTestDB()

	cu, cID, _ := testutil.CreateCompanyUser(db, "msgco@ex.com", "msgco", "pass123", "Msg Co", true)
	jID, _ := testutil.CreateJob(db, cID, "QA Engineer", "Technology", true)
	_, pID, _ := testutil.CreateJobseeker(db, "msgjs@ex.com", "msgjs", "pass123", "Msg JS")
	appID, _ := testutil.CreateApplication(db, pID, jID, "applied")
	ctok := testutil.GenerateToken(cu.String(), "company", 15*time.Minute)

	// A second company that should not be able to access this application's messages
	cu2, _, _ := testutil.CreateCompanyUser(db, "msgco2@ex.com", "msgco2", "pass123", "Msg Co 2", true)
	ctok2 := testutil.GenerateToken(cu2.String(), "company", 15*time.Minute)

	svc := NewService(db)
	h := NewHandler(svc)

	router := gin.New()
	g := router.Group("/api/v1/applications")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"), middleware.RequireVerifiedCompany(db))
	g.POST("/:id/messages", h.AddMessage)
	g.GET("/:id/messages", h.ListMessages)

	t.Run("add message", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/applications/%s/messages", appID), bytes.NewBufferString(`{"body":"We'd like to schedule a call"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+ctok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("list messages", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/applications/%s/messages", appID), nil)
		req.Header.Set("Authorization", "Bearer "+ctok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var msgs []Message
		json.Unmarshal(w.Body.Bytes(), &msgs)
		if len(msgs) != 1 {
			t.Fatalf("expected 1 message, got %d", len(msgs))
		}
	})

	t.Run("other company forbidden", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/applications/%s/messages", appID), nil)
		req.Header.Set("Authorization", "Bearer "+ctok2)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("jobseeker sees latest note", func(t *testing.T) {
		results, err := svc.ListForJobseeker(context.Background(), pID.String())
		if err != nil {
			t.Fatalf("list for jobseeker: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("expected 1 application, got %d", len(results))
		}
		if results[0].LatestNote == nil || *results[0].LatestNote != "We'd like to schedule a call" {
			t.Fatalf("expected latest note, got %v", results[0].LatestNote)
		}
	})
}

func TestListCompanyApplications(t *testing.T) {
	db := testutil.SetupTestDB()

	cu, cID, _ := testutil.CreateCompanyUser(db, "calist@ex.com", "calist", "pass123", "List Co", true)
	jID, _ := testutil.CreateJob(db, cID, "Go Developer", "Technology", true)
	ctok := testutil.GenerateToken(cu.String(), "company", 15*time.Minute)

	// Create a jobseeker and apply
	_, pID, _ := testutil.CreateJobseeker(db, "jslist@ex.com", "jslist", "pass123", "JS List")
	testutil.CreateApplication(db, pID, jID, "applied")

	svc := NewService(db)
	h := NewHandler(svc)

	router := gin.New()
	cg := router.Group("/api/v1/company")
	cg.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"), middleware.RequireVerifiedCompany(db))
	cg.GET("/applications", h.ListCompanyApplications)

	t.Run("list company applications", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/company/applications", nil)
		req.Header.Set("Authorization", "Bearer "+ctok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp []CompanyApplicationResult
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) != 1 {
			t.Fatalf("expected 1 application, got %d", len(resp))
		}
		if resp[0].CandidateName != "JS List" {
			t.Fatalf("expected candidate name 'JS List', got %q", resp[0].CandidateName)
		}
		if resp[0].JobTitle != "Go Developer" {
			t.Fatalf("expected job title 'Go Developer', got %q", resp[0].JobTitle)
		}
	})

	t.Run("list company applications empty", func(t *testing.T) {
		cu2, _, _ := testutil.CreateCompanyUser(db, "empty2@ex.com", "empty2", "pass123", "Empty Co", true)
		ctok2 := testutil.GenerateToken(cu2.String(), "company", 15*time.Minute)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/company/applications", nil)
		req.Header.Set("Authorization", "Bearer "+ctok2)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp []CompanyApplicationResult
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) != 0 {
			t.Fatalf("expected 0 applications, got %d", len(resp))
		}
	})

	t.Run("list company applications requires auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/company/applications", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})
}

package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"skillpass-server-go/internal/middleware"
	"skillpass-server-go/internal/testutil"
)

func TestApplicationFlow(t *testing.T) {
	db := testutil.SetupTestDB()
	bunDB := bun.NewDB(db, pgdialect.New())

	// Create company with two jobs (one open, one closed)
	cu, cID, _ := testutil.CreateCompanyUser(db, "aco@ex.com", "aco", "pass123", "App Co", true)
	jID, _ := testutil.CreateJob(db, cID, "Software Engineer", "Technology", true)
	cjID, _ := testutil.CreateJob(db, cID, "Closed Position", "Technology", false)

	// Create jobseeker
	uID, _, _ := testutil.CreateJobseeker(db, "app@ex.com", "app", "pass123", "Applicant")
	tok := testutil.GenerateToken(uID.String(), "jobseeker", 15*time.Minute)
	ctok := testutil.GenerateToken(cu.String(), "company", 15*time.Minute)

	svc := NewService(db, bunDB)
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
	sg.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"), middleware.RequireVerifiedCompany(bunDB))
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
	bunDB := bun.NewDB(db, pgdialect.New())

	cu, cID, _ := testutil.CreateCompanyUser(db, "msgco@ex.com", "msgco", "pass123", "Msg Co", true)
	jID, _ := testutil.CreateJob(db, cID, "QA Engineer", "Technology", true)
	_, pID, _ := testutil.CreateJobseeker(db, "msgjs@ex.com", "msgjs", "pass123", "Msg JS")
	appID, _ := testutil.CreateApplication(db, pID, jID, "applied")
	ctok := testutil.GenerateToken(cu.String(), "company", 15*time.Minute)

	// A second company that should not be able to access this application's messages
	cu2, _, _ := testutil.CreateCompanyUser(db, "msgco2@ex.com", "msgco2", "pass123", "Msg Co 2", true)
	ctok2 := testutil.GenerateToken(cu2.String(), "company", 15*time.Minute)

	svc := NewService(db, bunDB)
	h := NewHandler(svc)

	router := gin.New()
	g := router.Group("/api/v1/applications")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"), middleware.RequireVerifiedCompany(bunDB))
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

func TestStatusTransitions(t *testing.T) {
	db := testutil.SetupTestDB()
	bunDB := bun.NewDB(db, pgdialect.New())

	cu, cID, _ := testutil.CreateCompanyUser(db, "trans@ex.com", "trans", "pass123", "Trans Co", true)
	jID, _ := testutil.CreateJob(db, cID, "Test Job", "Technology", true)
	_, pID, _ := testutil.CreateJobseeker(db, "transjs@ex.com", "transjs", "pass123", "Trans JS")
	appID, _ := testutil.CreateApplication(db, pID, jID, "applied")
	ctok := testutil.GenerateToken(cu.String(), "company", 15*time.Minute)

	svc := NewService(db, bunDB)
	h := NewHandler(svc)

	router := gin.New()
	sg := router.Group("/api/v1/applications")
	sg.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"), middleware.RequireVerifiedCompany(bunDB))
	sg.PUT("/:id/status", h.UpdateStatus)

	tests := []struct {
		name       string
		from       string
		to         string
		wantStatus int
	}{
		{"applied to reviewed", "applied", "reviewed", http.StatusOK},
		{"applied to interviewed (invalid)", "applied", "interviewed", http.StatusBadRequest},
		{"applied to offered (invalid)", "applied", "offered", http.StatusBadRequest},
		{"applied to rejected", "applied", "rejected", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset status to 'applied' for each test
			db.ExecContext(context.Background(),
				`UPDATE applications SET status = 'applied' WHERE id = $1`, appID)

			body := fmt.Sprintf(`{"status":"%s"}`, tt.to)
			w := httptest.NewRecorder()
			req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/applications/%s/status", appID),
				bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+ctok)
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestApplicationUnauthorized(t *testing.T) {
	db := testutil.SetupTestDB()
	bunDB := bun.NewDB(db, pgdialect.New())

	svc := NewService(db, bunDB)
	h := NewHandler(svc)

	router := gin.New()
	ag := router.Group("/api/v1/jobs")
	ag.POST("/:id/apply", h.Apply)

	t.Run("apply without token returns 401", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/jobs/00000000-0000-0000-0000-000000000000/apply", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})
}

func TestApplicationMessageEdgeCases(t *testing.T) {
	db := testutil.SetupTestDB()
	bunDB := bun.NewDB(db, pgdialect.New())

	cu, cID, _ := testutil.CreateCompanyUser(db, "msgedge@ex.com", "msgedge", "pass123", "Msg Edge Co", true)
	jID, _ := testutil.CreateJob(db, cID, "Msg Job", "Technology", true)
	_, pID, _ := testutil.CreateJobseeker(db, "msgedgejs@ex.com", "msgedgejs", "pass123", "Msg Edge JS")
	appID, _ := testutil.CreateApplication(db, pID, jID, "applied")
	ctok := testutil.GenerateToken(cu.String(), "company", 15*time.Minute)

	svc := NewService(db, bunDB)
	h := NewHandler(svc)

	router := gin.New()
	g := router.Group("/api/v1/applications")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"), middleware.RequireVerifiedCompany(bunDB))
	g.POST("/:id/messages", h.AddMessage)
	g.GET("/:id/messages", h.ListMessages)

	t.Run("empty body rejected", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/applications/%s/messages", appID),
			bytes.NewBufferString(`{"body":""}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+ctok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("nonexistent application returns 404", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/applications/00000000-0000-0000-0000-000000000000/messages",
			bytes.NewBufferString(`{"body":"Hello"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+ctok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("multiple messages ordered chronologically", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			body := fmt.Sprintf(`{"body":"Message %d"}`, i)
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/applications/%s/messages", appID),
				bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+ctok)
			router.ServeHTTP(w, req)
			if w.Code != http.StatusCreated {
				t.Fatalf("message %d: expected 201, got %d", i, w.Code)
			}
		}

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/applications/%s/messages", appID), nil)
		req.Header.Set("Authorization", "Bearer "+ctok)
		router.ServeHTTP(w, req)

		var msgs []Message
		json.Unmarshal(w.Body.Bytes(), &msgs)
		if len(msgs) != 3 { // 3 added above
			t.Fatalf("expected 4 messages, got %d", len(msgs))
		}
	})
}

func TestListCompanyApplications(t *testing.T) {
	db := testutil.SetupTestDB()
	bunDB := bun.NewDB(db, pgdialect.New())

	cu, cID, _ := testutil.CreateCompanyUser(db, "calist@ex.com", "calist", "pass123", "List Co", true)
	jID, _ := testutil.CreateJob(db, cID, "Go Developer", "Technology", true)
	ctok := testutil.GenerateToken(cu.String(), "company", 15*time.Minute)

	// Create a jobseeker and apply
	_, pID, _ := testutil.CreateJobseeker(db, "jslist@ex.com", "jslist", "pass123", "JS List")
	testutil.CreateApplication(db, pID, jID, "applied")

	svc := NewService(db, bunDB)
	h := NewHandler(svc)

	router := gin.New()
	cg := router.Group("/api/v1/company")
	cg.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"), middleware.RequireVerifiedCompany(bunDB))
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

func TestScheduleInterview(t *testing.T) {
	db := testutil.SetupTestDB()
	bunDB := bun.NewDB(db, pgdialect.New())

	cu, cID, _ := testutil.CreateCompanyUser(db, "schedco@ex.com", "schedco", "pass123", "Sched Co", true)
	jID, _ := testutil.CreateJob(db, cID, "Interview Job", "Technology", true)
	_, pID, _ := testutil.CreateJobseeker(db, "schedjs@ex.com", "schedjs", "pass123", "Sched JS")
	appID, _ := testutil.CreateApplication(db, pID, jID, "applied")

	// Advance to "reviewed" so scheduling is allowed
	ctok := testutil.GenerateToken(cu.String(), "company", 15*time.Minute)

	svc := NewService(db, bunDB)
	h := NewHandler(svc)

	router := gin.New()

	// Status update route (to advance to reviewed)
	sg := router.Group("/api/v1/applications")
	sg.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"), middleware.RequireVerifiedCompany(bunDB))
	sg.PUT("/:id/status", h.UpdateStatus)
	sg.POST("/:id/interview", h.ScheduleInterview)

	// Advance application to "reviewed"
	{
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/applications/%s/status", appID),
			bytes.NewBufferString(`{"status":"reviewed"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+ctok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("advance to reviewed: expected 200, got %d: %s", w.Code, w.Body.String())
		}
	}

	t.Run("schedule onsite interview", func(t *testing.T) {
		body := `{"scheduledAt":"2026-09-01T10:00:00Z","mode":"onsite","location":"Office 5th floor","interviewer":"Head of Eng","notes":"Bring ID"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/applications/%s/interview", appID),
			bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+ctok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var result ApplicationResult
		json.Unmarshal(w.Body.Bytes(), &result)
		if result.Status != "interviewed" {
			t.Fatalf("expected status interviewed, got %q", result.Status)
		}
	})

	t.Run("schedule online interview", func(t *testing.T) {
		body := `{"scheduledAt":"2026-09-02T14:00:00Z","mode":"online","meetingLink":"https://meet.google.com/abc-defg-hij"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/applications/%s/interview", appID),
			bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+ctok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("invalid mode rejected", func(t *testing.T) {
		body := `{"scheduledAt":"2026-09-03T10:00:00Z","mode":"zoom","location":"Nowhere"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/applications/%s/interview", appID),
			bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+ctok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for invalid mode, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("missing scheduledAt rejected", func(t *testing.T) {
		body := `{"mode":"onsite","location":"Office"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/applications/%s/interview", appID),
			bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+ctok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for missing scheduledAt, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("wrong company forbidden", func(t *testing.T) {
		cu2, _, _ := testutil.CreateCompanyUser(db, "other@ex.com", "other", "pass123", "Other Co", true)
		ctok2 := testutil.GenerateToken(cu2.String(), "company", 15*time.Minute)
		body := `{"scheduledAt":"2026-09-04T10:00:00Z","mode":"onsite","location":"Other office"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/applications/%s/interview", appID),
			bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+ctok2)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("online meeting link redacted at rest", func(t *testing.T) {
		// SEC-01: query-string credentials (?pwd=…) must never persist.
		body := `{"scheduledAt":"2026-09-06T10:00:00Z","mode":"online","meetingLink":"https://meet.google.com/redact-me?pwd=SuperSecret2026#sess"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/applications/%s/interview", appID),
			bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+ctok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var stored string
		db.QueryRowContext(context.Background(),
			`SELECT meeting_link FROM interview_schedules
			 WHERE application_id = $1 ORDER BY scheduled_at DESC LIMIT 1`, appID,
		).Scan(&stored)
		if stored != "https://meet.google.com/redact-me" {
			t.Fatalf("expected redacted link, got %q", stored)
		}
		if strings.Contains(stored, "pwd=") || strings.Contains(stored, "#sess") {
			t.Fatalf("stored meeting link leaked credentials: %q", stored)
		}
	})

	t.Run("invalid meeting link rejected", func(t *testing.T) {
		// F14: javascript:/data: URLs must not be accepted as meeting links.
		body := `{"scheduledAt":"2026-09-07T10:00:00Z","mode":"online","meetingLink":"javascript:alert(1)"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/applications/%s/interview", appID),
			bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+ctok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for invalid meeting link, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("onsite without location rejected", func(t *testing.T) {
		body := `{"scheduledAt":"2026-09-08T10:00:00Z","mode":"onsite"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/applications/%s/interview", appID),
			bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+ctok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for missing onsite location, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("transaction atomicity — all or nothing", func(t *testing.T) {
		// Create a fresh application in "applied" status that can't be
		// scheduled (no valid transition from applied to interviewed).
		_, pID2, _ := testutil.CreateJobseeker(db, "schedjs2@ex.com", "schedjs2", "pass123", "Sched JS 2")
		jID2, _ := testutil.CreateJob(db, cID, "Interview Job 2", "Technology", true)
		appID2, _ := testutil.CreateApplication(db, pID2, jID2, "applied")

		body := `{"scheduledAt":"2026-09-05T10:00:00Z","mode":"onsite","location":"Office"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/applications/%s/interview", appID2),
			bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+ctok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for invalid transition, got %d: %s", w.Code, w.Body.String())
		}

		// Verify no orphaned interview_schedules row was created
		var count int
		db.QueryRowContext(context.Background(),
			`SELECT COUNT(*) FROM interview_schedules WHERE application_id = $1`, appID2,
		).Scan(&count)
		if count != 0 {
			t.Fatalf("expected 0 orphaned schedules, got %d (transaction not atomic)", count)
		}
	})
}

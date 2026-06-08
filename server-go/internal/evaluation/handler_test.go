package evaluation

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"skillpass-server-go/internal/lib"
	"skillpass-server-go/internal/middleware"
	"skillpass-server-go/internal/testutil"
)

func TestPostEvaluate(t *testing.T) {
	db := testutil.SetupTestDB()

	uID, _, err := testutil.CreateJobseeker(db, "eval@ex.com", "eval", "pass123", "Eval User")
	if err != nil {
		t.Fatalf("create jobseeker: %v", err)
	}
	tok := testutil.GenerateToken(uID.String(), "jobseeker", 15*time.Minute)

	mockLLM := lib.NewMockLLMClient()
	svc := NewService(db, mockLLM)
	h := NewHandler(db, svc)

	router := gin.New()
	g := router.Group("/api/v1/evaluate")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("jobseeker"))
	g.POST("/me", h.PostEvaluate)
	g.GET("/me/results", h.GetLatestEvaluation)

	t.Run("post evaluate", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/evaluate/me", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp EvaluationResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.OverallScore != 75 {
			t.Fatalf("expected 75, got %d", resp.OverallScore)
		}
	})

	t.Run("get latest after post", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/evaluate/me/results", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("get latest no eval", func(t *testing.T) {
		u2, _, _ := testutil.CreateJobseeker(db, "eval2@ex.com", "eval2", "pass123", "Eval2")
		t2 := testutil.GenerateToken(u2.String(), "jobseeker", 15*time.Minute)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/evaluate/me/results", nil)
		req.Header.Set("Authorization", "Bearer "+t2)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("wrong role", func(t *testing.T) {
		cu, _, _ := testutil.CreateCompanyUser(db, "eco@ex.com", "eco", "pass123", "E Co", true)
		ct := testutil.GenerateToken(cu.String(), "company", 15*time.Minute)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/evaluate/me", nil)
		req.Header.Set("Authorization", "Bearer "+ct)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})

	t.Run("no auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/evaluate/me", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})
}

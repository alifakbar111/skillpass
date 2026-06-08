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

func TestGetMyProfile(t *testing.T) {
	db := testutil.SetupTestDB()

	userID, profileID, err := testutil.CreateJobseeker(db, "myprofile@example.com", "myprofile", "password123", "My Profile")
	if err != nil {
		t.Fatalf("create jobseeker: %v", err)
	}
	testutil.CreateExperience(db, profileID, "employment", "Engineer", "Acme")
	token := testutil.GenerateToken(userID.String(), "jobseeker", 15*time.Minute)

	router := gin.New()
	h := NewProfileHandler(db)
	g := router.Group("/api/v1/profiles")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret))
	g.GET("/me", h.GetMyProfile)

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/profiles/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp ProfileResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Username != "myprofile" {
			t.Fatalf("expected 'myprofile', got '%s'", resp.Username)
		}
		if len(resp.Experiences) != 1 {
			t.Fatalf("expected 1 experience, got %d", len(resp.Experiences))
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/profiles/me", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("company user no profile", func(t *testing.T) {
		cu, _, _ := testutil.CreateCompanyUser(db, "comp@ex.com", "comp", "pass123", "Co", true)
		ct := testutil.GenerateToken(cu.String(), "company", 15*time.Minute)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/profiles/me", nil)
		req.Header.Set("Authorization", "Bearer "+ct)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestUpdateMyProfile(t *testing.T) {
	db := testutil.SetupTestDB()

	userID, _, _ := testutil.CreateJobseeker(db, "up@ex.com", "upuser", "pass123", "Up User")
	token := testutil.GenerateToken(userID.String(), "jobseeker", 15*time.Minute)

	router := gin.New()
	h := NewProfileHandler(db)
	g := router.Group("/api/v1/profiles")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret))
	g.PUT("/me", h.UpdateMyProfile)

	t.Run("update headline", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/v1/profiles/me", bytes.NewBufferString(`{"headline":"Engineer","about":"I code"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("update slug reserved", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/v1/profiles/me", bytes.NewBufferString(`{"slug":"admin"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("no fields", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/v1/profiles/me", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestCreateExperience(t *testing.T) {
	db := testutil.SetupTestDB()

	userID, _, _ := testutil.CreateJobseeker(db, "ce@ex.com", "ceuser", "pass123", "CE User")
	token := testutil.GenerateToken(userID.String(), "jobseeker", 15*time.Minute)

	router := gin.New()
	h := NewProfileHandler(db)
	g := router.Group("/api/v1/profiles")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret))
	g.POST("/me/experience", h.CreateExperience)

	t.Run("success", func(t *testing.T) {
		body := `{"type":"employment","title":"Engineer","organization":"Co","startDate":"2020-01"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/profiles/me/experience", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("missing fields", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/profiles/me/experience", bytes.NewBufferString(`{"type":"employment"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("invalid date", func(t *testing.T) {
		body := `{"type":"employment","title":"E","organization":"C","startDate":"2020-13"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/profiles/me/experience", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestUpdateExperience(t *testing.T) {
	db := testutil.SetupTestDB()

	userID, profileID, _ := testutil.CreateJobseeker(db, "ue@ex.com", "ueuser", "pass123", "UE User")
	expID, _ := testutil.CreateExperience(db, profileID, "employment", "Junior", "Small Co")
	token := testutil.GenerateToken(userID.String(), "jobseeker", 15*time.Minute)

	router := gin.New()
	h := NewProfileHandler(db)
	g := router.Group("/api/v1/profiles")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret))
	g.PUT("/me/experience/:id", h.UpdateExperience)

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/profiles/me/experience/%s", expID.String()), bytes.NewBufferString(`{"title":"Senior"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/v1/profiles/me/experience/00000000-0000-0000-0000-000000000000", bytes.NewBufferString(`{"title":"X"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/v1/profiles/me/experience/bad", bytes.NewBufferString(`{"title":"X"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestDeleteExperience(t *testing.T) {
	db := testutil.SetupTestDB()

	userID, profileID, _ := testutil.CreateJobseeker(db, "de@ex.com", "deuser", "pass123", "DE User")
	expID, _ := testutil.CreateExperience(db, profileID, "employment", "Temp", "Temp Co")
	token := testutil.GenerateToken(userID.String(), "jobseeker", 15*time.Minute)

	router := gin.New()
	h := NewProfileHandler(db)
	g := router.Group("/api/v1/profiles")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret))
	g.DELETE("/me/experience/:id", h.DeleteExperience)

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/profiles/me/experience/%s", expID.String()), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/api/v1/profiles/me/experience/00000000-0000-0000-0000-000000000000", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})
}

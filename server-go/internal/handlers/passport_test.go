package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"skillpass-server-go/internal/testutil"
)

func TestGetPublicProfile(t *testing.T) {
	db := testutil.SetupTestDB()

	_, profileID, err := testutil.CreateJobseeker(db, "public@example.com", "publicuser", "password123", "Public User")
	if err != nil {
		t.Fatalf("create jobseeker: %v", err)
	}
	testutil.CreateExperience(db, profileID, "employment", "Software Engineer", "Tech Corp")

	router := gin.New()
	h := NewPassportHandler(db)
	router.GET("/api/v1/profiles/:username", h.GetProfile)

	t.Run("get profile by slug", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/profiles/publicuser", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp PublicProfileResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Name != "Public User" {
			t.Fatalf("expected 'Public User', got '%s'", resp.Name)
		}
		if len(resp.Experiences) != 1 {
			t.Fatalf("expected 1 experience, got %d", len(resp.Experiences))
		}
	})

	t.Run("get profile not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/profiles/nonexistent", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("UUID-style slug not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/profiles/123e4567-e89b-12d3-a456-426614174000", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestPassportViewCount(t *testing.T) {
	db := testutil.SetupTestDB()

	_, _, err := testutil.CreateJobseeker(db, "views@example.com", "viewsuser", "password123", "Views User")
	if err != nil {
		t.Fatalf("create jobseeker: %v", err)
	}

	router := gin.New()
	h := NewPassportHandler(db)
	router.GET("/api/v1/profiles/:username", h.GetProfile)

	get := func() PublicProfileResponse {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/profiles/viewsuser", nil)
		router.ServeHTTP(w, req)
		var resp PublicProfileResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		return resp
	}

	first := get()
	if first.ViewCount != 1 {
		t.Fatalf("expected viewCount 1 after first view, got %d", first.ViewCount)
	}
	second := get()
	if second.ViewCount != 2 {
		t.Fatalf("expected viewCount 2 after second view, got %d", second.ViewCount)
	}
}

func TestGetPublicProfile_NoExperiences(t *testing.T) {
	db := testutil.SetupTestDB()

	_, _, err := testutil.CreateJobseeker(db, "noexp@example.com", "noexpuser", "password123", "No Exp User")
	if err != nil {
		t.Fatalf("create jobseeker: %v", err)
	}

	router := gin.New()
	h := NewPassportHandler(db)
	router.GET("/api/v1/profiles/:username", h.GetProfile)

	t.Run("profile with no experiences", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/profiles/noexpuser", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp PublicProfileResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Name != "No Exp User" {
			t.Fatalf("expected 'No Exp User', got '%s'", resp.Name)
		}
		if len(resp.Experiences) != 0 {
			t.Fatalf("expected 0 experiences, got %d", len(resp.Experiences))
		}
	})
}

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
}

package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"

	"skillpass-server-go/internal/db"
	"skillpass-server-go/internal/testutil"
)

func TestGetPublicProfile(t *testing.T) {
	sqlDB := testutil.SetupTestDB()
	bunDB := db.NewBunDB(sqlDB)

	_, profileID, err := testutil.CreateJobseeker(sqlDB, "public@example.com", "publicuser", "password123", "Public User")
	if err != nil {
		t.Fatalf("create jobseeker: %v", err)
	}
	testutil.CreateExperience(sqlDB, profileID, "employment", "Software Engineer", "Tech Corp")

	router := gin.New()
	h := NewPassportHandler(sqlDB, bunDB)
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
	sqlDB := testutil.SetupTestDB()
	bunDB := db.NewBunDB(sqlDB)

	_, profileID, err := testutil.CreateJobseeker(sqlDB, "views@example.com", "viewsuser", "password123", "Views User")
	if err != nil {
		t.Fatalf("create jobseeker: %v", err)
	}

	// view_count on the public profile must not be written on every GET.
	// It is a DoS amplifier (DB write per unauthenticated request) and is
	// trivially gameable by clients that repeatedly fetch the same profile.
	// The handler returns viewCount=0; an authenticated stats endpoint
	// backed by profile_views is the path forward.

	hook := &viewCountWriteHook{}
	bunDB.AddQueryHook(hook)

	router := gin.New()
	h := NewPassportHandler(sqlDB, bunDB)
	router.GET("/api/v1/profiles/:username", h.GetProfile)

	get := func() PublicProfileResponse {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/profiles/viewsuser", nil)
		router.ServeHTTP(w, req)
		var resp PublicProfileResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		return resp
	}

	if _, err := sqlDB.ExecContext(
		context.Background(),
		"UPDATE jobseeker_profiles SET view_count = 5 WHERE id = $1",
		profileID,
	); err != nil {
		t.Fatalf("seed view_count: %v", err)
	}

	hook.updates = 0
	for i := 0; i < 100; i++ {
		_ = get()
	}
	if hook.updates != 0 {
		t.Fatalf("expected 0 UPDATE statements on jobseeker_profiles after 100 public GETs, got %d (handler must not write view_count on public reads — DoS amplifier)", hook.updates)
	}

	var stored int
	if err := sqlDB.QueryRowContext(
		context.Background(),
		"SELECT view_count FROM jobseeker_profiles WHERE id = $1",
		profileID,
	).Scan(&stored); err != nil {
		t.Fatalf("read view_count: %v", err)
	}
	if stored != 5 {
		t.Fatalf("expected stored view_count to remain 5 after 100 public GETs, got %d", stored)
	}

	resp := get()
	if resp.ViewCount != 0 {
		t.Fatalf("expected response viewCount 0, got %d", resp.ViewCount)
	}
}

type viewCountWriteHook struct {
	updates int
}

func (h *viewCountWriteHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	return ctx
}

func (h *viewCountWriteHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	q := strings.ToUpper(event.Query)
	if strings.HasPrefix(strings.TrimSpace(q), "UPDATE") &&
		strings.Contains(q, "JOBSEEKER_PROFILES") &&
		strings.Contains(q, "VIEW_COUNT") {
		h.updates++
	}
}

func TestGetPublicProfile_NoExperiences(t *testing.T) {
	sqlDB := testutil.SetupTestDB()
	bunDB := db.NewBunDB(sqlDB)

	_, _, err := testutil.CreateJobseeker(sqlDB, "noexp@example.com", "noexpuser", "password123", "No Exp User")
	if err != nil {
		t.Fatalf("create jobseeker: %v", err)
	}

	router := gin.New()
	h := NewPassportHandler(sqlDB, bunDB)
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

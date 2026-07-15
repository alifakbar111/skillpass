package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"skillpass-server-go/internal/db"
	"skillpass-server-go/internal/testutil"
)

func TestGetIndustries(t *testing.T) {
	sqlDB := testutil.SetupTestDB()
	bunDB := db.NewBunDB(sqlDB)

	testutil.CreateIndustry(sqlDB, "Technology", "Software and IT")
	testutil.CreateIndustry(sqlDB, "Healthcare", "Medical services")

	router := gin.New()
	h := NewReferenceHandler(bunDB)
	router.GET("/api/v1/industries", h.GetIndustries)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/industries", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp []interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp) < 2 {
		t.Fatalf("expected at least 2 industries, got %d", len(resp))
	}
}

func TestGetIndustriesEmpty(t *testing.T) {
	sqlDB := testutil.SetupTestDB()
	bunDB := db.NewBunDB(sqlDB)

	router := gin.New()
	h := NewReferenceHandler(bunDB)
	router.GET("/api/v1/industries", h.GetIndustries)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/industries", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp []interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 0 {
		t.Fatalf("expected empty array, got %d items", len(resp))
	}
}

func TestGetTags(t *testing.T) {
	sqlDB := testutil.SetupTestDB()
	bunDB := db.NewBunDB(sqlDB)

	testutil.CreateIndustry(sqlDB, "Technology", "SW")
	var techID string
	sqlDB.QueryRow("SELECT id FROM industry_categories WHERE name = 'Technology'").Scan(&techID)
	testutil.CreateTag(sqlDB, "Go", techID)
	testutil.CreateTag(sqlDB, "React", techID)

	router := gin.New()
	h := NewReferenceHandler(bunDB)
	router.GET("/api/v1/tags", h.GetTags)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/tags", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp []interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) < 2 {
		t.Fatalf("expected at least 2 tags, got %d", len(resp))
	}
}

func TestGetTagsFilterByIndustry(t *testing.T) {
	sqlDB := testutil.SetupTestDB()
	bunDB := db.NewBunDB(sqlDB)

	testutil.CreateIndustry(sqlDB, "Technology", "SW")
	testutil.CreateIndustry(sqlDB, "Healthcare", "Med")
	var techID, healthID string
	sqlDB.QueryRow("SELECT id FROM industry_categories WHERE name = 'Technology'").Scan(&techID)
	sqlDB.QueryRow("SELECT id FROM industry_categories WHERE name = 'Healthcare'").Scan(&healthID)
	testutil.CreateTag(sqlDB, "Go", techID)
	testutil.CreateTag(sqlDB, "Nursing", healthID)

	router := gin.New()
	h := NewReferenceHandler(bunDB)
	router.GET("/api/v1/tags", h.GetTags)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/tags?industry="+techID, nil)
	router.ServeHTTP(w, req)

	var resp []interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 1 {
		t.Fatalf("expected 1 tag filtered by industry, got %d", len(resp))
	}
}

func TestGetTags_InvalidIndustryID(t *testing.T) {
	sqlDB := testutil.SetupTestDB()
	bunDB := db.NewBunDB(sqlDB)

	router := gin.New()
	h := NewReferenceHandler(bunDB)
	router.GET("/api/v1/tags", h.GetTags)

	t.Run("invalid industry ID format", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/tags?industry=not-a-uuid", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
		var resp map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if resp["error"] != "invalid industry ID" {
			t.Fatalf("expected error %q, got %q", "invalid industry ID", resp["error"])
		}
	})
}

func TestGetTags_NonExistentIndustry(t *testing.T) {
	sqlDB := testutil.SetupTestDB()
	bunDB := db.NewBunDB(sqlDB)

	router := gin.New()
	h := NewReferenceHandler(bunDB)
	router.GET("/api/v1/tags", h.GetTags)

	t.Run("non-existent industry returns empty", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/tags?industry=00000000-0000-0000-0000-000000000000", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp []interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp) != 0 {
			t.Fatalf("expected empty array, got %d items", len(resp))
		}
	})
}

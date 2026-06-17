package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"skillpass-server-go/internal/middleware"
	"skillpass-server-go/internal/testutil"
)

func TestBlindModeToggle(t *testing.T) {
	db := testutil.SetupTestDB()

	uID, cID, _ := testutil.CreateCompanyUser(db, "blind@ex.com", "blindco", "pass123", "Blind Corp", true)
	tok := testutil.GenerateToken(uID.String(), "company", 15*time.Minute)

	router := gin.New()
	h := NewCompanyHandler(db)
	g := router.Group("/api/v1/company")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"))
	g.GET("/profile", h.GetProfile)
	g.PUT("/profile", h.UpdateProfile)

	t.Run("defaults to false", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/company/profile", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		var resp CompanyResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.BlindMode {
			t.Fatal("expected blindMode false by default")
		}
	})

	t.Run("enable via update", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/v1/company/profile", bytes.NewBufferString(`{"blindMode":true}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp CompanyResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if !resp.BlindMode {
			t.Fatal("expected blindMode true after update")
		}
	})

	t.Run("helper reflects update", func(t *testing.T) {
		if !CompanyBlindMode(context.Background(), db, cID.String()) {
			t.Fatal("expected CompanyBlindMode true")
		}
	})
}

func TestGetCompanyProfile(t *testing.T) {
	db := testutil.SetupTestDB()

	uID, _, _ := testutil.CreateCompanyUser(db, "comp@ex.com", "comp", "pass123", "Test Corp", true)
	tok := testutil.GenerateToken(uID.String(), "company", 15*time.Minute)

	router := gin.New()
	h := NewCompanyHandler(db)
	g := router.Group("/api/v1/company")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"))
	g.GET("/profile", h.GetProfile)

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/company/profile", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp CompanyResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.CompanyName != "Test Corp" {
			t.Fatalf("expected 'Test Corp', got '%s'", resp.CompanyName)
		}
	})

	t.Run("no auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/company/profile", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("wrong role", func(t *testing.T) {
		wt := testutil.GenerateToken(uID.String(), "jobseeker", 15*time.Minute)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/company/profile", nil)
		req.Header.Set("Authorization", "Bearer "+wt)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})

	t.Run("company not found", func(t *testing.T) {
		uid, err := testutil.CreateUser(db, testutil.UniqueEmail("nocomp"), testutil.UniqueUsername("nocomp"), "pass123", "No Company", "company")
		if err != nil {
			t.Fatal(err)
		}
		nct := testutil.GenerateToken(uid.String(), "company", 15*time.Minute)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/company/profile", nil)
		req.Header.Set("Authorization", "Bearer "+nct)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("malformed userId", func(t *testing.T) {
		mtok := testutil.GenerateToken("not-a-valid-uuid", "company", 15*time.Minute)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/company/profile", nil)
		req.Header.Set("Authorization", "Bearer "+mtok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestUpdateCompanyProfile(t *testing.T) {
	db := testutil.SetupTestDB()

	uID, _, _ := testutil.CreateCompanyUser(db, "uc@ex.com", "uc", "pass123", "Old Name", true)
	tok := testutil.GenerateToken(uID.String(), "company", 15*time.Minute)

	router := gin.New()
	h := NewCompanyHandler(db)
	g := router.Group("/api/v1/company")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"))
	g.PUT("/profile", h.UpdateProfile)

	t.Run("update name", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/v1/company/profile", bytes.NewBufferString(`{"companyName":"New Name Inc"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp CompanyResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.CompanyName != "New Name Inc" {
			t.Fatalf("expected 'New Name Inc', got '%s'", resp.CompanyName)
		}
	})

	t.Run("no fields", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/v1/company/profile", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("invalid website", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/v1/company/profile", bytes.NewBufferString(`{"website":"not-a-url"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("update all fields", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"companyName":"Full Update Inc","website":"https://fullupdate.com","industry":"Healthcare","description":"A full update test"}`
		req := httptest.NewRequest("PUT", "/api/v1/company/profile", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp CompanyResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.CompanyName != "Full Update Inc" {
			t.Fatalf("expected 'Full Update Inc', got '%s'", resp.CompanyName)
		}
		if resp.Website == nil || *resp.Website != "https://fullupdate.com" {
			t.Fatalf("expected 'https://fullupdate.com', got %v", resp.Website)
		}
		if resp.Industry != "Healthcare" {
			t.Fatalf("expected 'Healthcare', got '%s'", resp.Industry)
		}
		if resp.Description == nil || *resp.Description != "A full update test" {
			t.Fatalf("expected 'A full update test', got %v", resp.Description)
		}
	})

	t.Run("company not found", func(t *testing.T) {
		uid, err := testutil.CreateUser(db, testutil.UniqueEmail("nocompup"), testutil.UniqueUsername("nocompup"), "pass123", "No Company Up", "company")
		if err != nil {
			t.Fatal(err)
		}
		nct := testutil.GenerateToken(uid.String(), "company", 15*time.Minute)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/v1/company/profile", bytes.NewBufferString(`{"companyName":"Should Fail"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+nct)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestSubmitVerification(t *testing.T) {
	db := testutil.SetupTestDB()

	uID, _, _ := testutil.CreateCompanyUser(db, "sv@ex.com", "sv", "pass123", "SV Corp", false)
	tok := testutil.GenerateToken(uID.String(), "company", 15*time.Minute)

	router := gin.New()
	h := NewCompanyHandler(db)
	g := router.Group("/api/v1/company")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"))
	g.POST("/verification", h.SubmitVerification)

	t.Run("success", func(t *testing.T) {
		body := `{"businessRegistration":"BR-123","website":"https://ex.com","address":"123 St","contact":"c@ex.com"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/company/verification", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("missing fields", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/company/verification", bytes.NewBufferString(`{"website":"https://ex.com"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("invalid website", func(t *testing.T) {
		body := `{"businessRegistration":"BR-123","website":"not-a-url","address":"123 St","contact":"c@ex.com"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/company/verification", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("non-existent company", func(t *testing.T) {
		uid, err := testutil.CreateUser(db, testutil.UniqueEmail("nover"), testutil.UniqueUsername("nover"), "pass123", "No Ver", "company")
		if err != nil {
			t.Fatal(err)
		}
		nct := testutil.GenerateToken(uid.String(), "company", 15*time.Minute)
		body := `{"businessRegistration":"BR-999","website":"https://nover.com","address":"456 St","contact":"n@ex.com"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/company/verification", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+nct)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("already verified", func(t *testing.T) {
		avU, _, _ := testutil.CreateCompanyUser(db, testutil.UniqueEmail("alreadyv"), testutil.UniqueUsername("alreadyv"), "pass123", "Already V", true)
		avT := testutil.GenerateToken(avU.String(), "company", 15*time.Minute)
		body := `{"businessRegistration":"BR-888","website":"https://alreadyv.com","address":"789 St","contact":"a@ex.com"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/company/verification", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+avT)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestGetVerificationStatus(t *testing.T) {
	db := testutil.SetupTestDB()

	pu, _, _ := testutil.CreateCompanyUser(db, "p@ex.com", "p", "pass123", "P Inc", false)
	vu, _, _ := testutil.CreateCompanyUser(db, "v@ex.com", "v", "pass123", "V Inc", true)
	pt := testutil.GenerateToken(pu.String(), "company", 15*time.Minute)
	vt := testutil.GenerateToken(vu.String(), "company", 15*time.Minute)

	router := gin.New()
	h := NewCompanyHandler(db)
	g := router.Group("/api/v1/company")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"))
	g.GET("/verification-status", h.GetVerificationStatus)

	t.Run("pending", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/company/verification-status", nil)
		req.Header.Set("Authorization", "Bearer "+pt)
		router.ServeHTTP(w, req)
		var m map[string]string
		json.Unmarshal(w.Body.Bytes(), &m)
		if m["verificationStatus"] != "pending" {
			t.Fatalf("expected 'pending', got '%s'", m["verificationStatus"])
		}
	})

	t.Run("verified", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/company/verification-status", nil)
		req.Header.Set("Authorization", "Bearer "+vt)
		router.ServeHTTP(w, req)
		var m map[string]string
		json.Unmarshal(w.Body.Bytes(), &m)
		if m["verificationStatus"] != "verified" {
			t.Fatalf("expected 'verified', got '%s'", m["verificationStatus"])
		}
	})

	t.Run("no company record", func(t *testing.T) {
		uid, err := testutil.CreateUser(db, testutil.UniqueEmail("nostatus"), testutil.UniqueUsername("nostatus"), "pass123", "No Status", "company")
		if err != nil {
			t.Fatal(err)
		}
		nct := testutil.GenerateToken(uid.String(), "company", 15*time.Minute)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/company/verification-status", nil)
		req.Header.Set("Authorization", "Bearer "+nct)
		router.ServeHTTP(w, req)
		var m map[string]string
		json.Unmarshal(w.Body.Bytes(), &m)
		if m["verificationStatus"] != "none" {
			t.Fatalf("expected 'none', got '%s'", m["verificationStatus"])
		}
	})
}

func TestUpdateCompanyAllFields(t *testing.T) {
	db := testutil.SetupTestDB()

	uID, _, _ := testutil.CreateCompanyUser(db, testutil.UniqueEmail("allfields"), testutil.UniqueUsername("allfields"), "pass123", "Original Name", false)
	tok := testutil.GenerateToken(uID.String(), "company", 15*time.Minute)

	router := gin.New()
	h := NewCompanyHandler(db)
	g := router.Group("/api/v1/company")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret), middleware.RequireRole("company"))
	g.PUT("/profile", h.UpdateProfile)

	t.Run("update all fields at once", func(t *testing.T) {
		body := `{"companyName":"Updated Corp","website":"https://updated.com","industry":"Finance","description":"An updated description"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/v1/company/profile", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp CompanyResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.CompanyName != "Updated Corp" {
			t.Fatalf("expected 'Updated Corp', got '%s'", resp.CompanyName)
		}
		if resp.Website == nil || *resp.Website != "https://updated.com" {
			t.Fatalf("expected 'https://updated.com', got %v", resp.Website)
		}
		if resp.Industry != "Finance" {
			t.Fatalf("expected 'Finance', got '%s'", resp.Industry)
		}
		if resp.Description == nil || *resp.Description != "An updated description" {
			t.Fatalf("expected 'An updated description', got %v", resp.Description)
		}
	})
}

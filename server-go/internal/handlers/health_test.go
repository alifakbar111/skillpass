package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetHealth(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/health", nil)
	GetHealth(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if status, ok := resp["status"]; !ok || status != "ok" {
		t.Fatalf("expected status=ok, got %v", resp)
	}
	if _, ok := resp["timestamp"]; !ok {
		t.Fatal("response should contain timestamp")
	}
}

func TestGetHealthOnlyGET(t *testing.T) {
	router := gin.New()
	router.GET("/api/v1/health", GetHealth)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/health", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound && w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 404 or 405, got %d", w.Code)
	}
}

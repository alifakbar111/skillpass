package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		if id, ok := GetUserID(c); ok || id != "" {
			t.Errorf("expected empty/false, got %q/%v", id, ok)
		}
	})
	t.Run("empty string", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("userId", "")
		if id, ok := GetUserID(c); ok || id != "" {
			t.Errorf("expected empty/false, got %q/%v", id, ok)
		}
	})
	t.Run("valid", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("userId", "abc-123")
		if id, ok := GetUserID(c); !ok || id != "abc-123" {
			t.Errorf("expected abc-123/true, got %q/%v", id, ok)
		}
	})
	t.Run("wrong type", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("userId", 42) // not a string
		if id, ok := GetUserID(c); ok || id != "" {
			t.Errorf("expected empty/false for non-string, got %q/%v", id, ok)
		}
	})
}

func TestGetRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		if r, ok := GetRole(c); ok || r != "" {
			t.Errorf("expected empty/false, got %q/%v", r, ok)
		}
	})
	t.Run("valid", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("role", "company")
		if r, ok := GetRole(c); !ok || r != "company" {
			t.Errorf("expected company/true, got %q/%v", r, ok)
		}
	})
}

func TestGetUserID_DoesNotPanicOnUnhandled(t *testing.T) {
	// Regression: the old `userID.(string)` form would panic on missing
	// or wrong-typed context values. Make sure the helper does not.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("GetUserID panicked: %v", r)
		}
	}()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	_, _ = GetUserID(c)
	c.Set("userId", struct{}{})
	_, _ = GetUserID(c)
	_ = httptest.NewRequest(http.MethodGet, "/", nil) // keep http import used
}

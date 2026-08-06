package rbac

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"skillpass-server-go/internal/testutil"
)

func TestRBACHandlers(t *testing.T) {
	db := testutil.SetupTestDB()

	cu, cID, _ := testutil.CreateCompanyUser(db, testutil.UniqueEmail("rbach"), testutil.UniqueUsername("rbach"), "pass123", "RBAC Handler Co", true)
	empID, _ := testutil.CreateEmployee(db, cID, "RBAC", "Handler", testutil.UniqueEmail("rbachh"))
	tok := testutil.GenerateToken(cu.String(), "company", 15*time.Minute)

	svc := NewService(db)
	if err := svc.EnsureCompanyRoles(context.Background(), cID); err != nil {
		t.Fatalf("EnsureCompanyRoles: %v", err)
	}
	h := NewHandler(svc)

	router := gin.New()
	// Simulate what AuthRequired + RequireCompanyMember set in production.
	router.Use(func(c *gin.Context) {
		c.Set("companyId", cID.String())
		c.Set("employeeId", empID.String())
		c.Set("userId", cu.String())
		c.Set("role", "company")
		c.Next()
	})

	g := router.Group("/api/v1/hris")
	g.GET("/roles", h.ListRoles)
	g.GET("/permissions", h.ListPermissions)
	g.GET("/roles/:roleId/permissions", h.GetRolePermissions)
	g.PUT("/roles/:roleId/permissions", h.SetRolePermissions)
	g.POST("/roles", h.CreateRole)
	g.PUT("/roles/:roleId", h.UpdateRole)
	g.DELETE("/roles/:roleId", h.DeleteRole)
	g.GET("/employees/:id/roles", h.GetEmployeeRoles)
	g.GET("/me/permissions", h.GetMyPermissions)

	var createdRoleID string

	t.Run("list roles", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/hris/roles", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var roles []RoleResponse
		json.Unmarshal(w.Body.Bytes(), &roles)
		if len(roles) < 3 {
			t.Fatalf("expected at least 3 system roles, got %d", len(roles))
		}
	})

	t.Run("list permissions", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/hris/permissions", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var perms []PermissionResponse
		json.Unmarshal(w.Body.Bytes(), &perms)
		if len(perms) == 0 {
			t.Fatal("expected non-empty permissions")
		}
	})

	t.Run("create role", func(t *testing.T) {
		body := `{"name":"Test Handler Role","description":"Created via handler test"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/hris/roles", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		var role RoleResponse
		json.Unmarshal(w.Body.Bytes(), &role)
		createdRoleID = role.ID
		if role.Name != "Test Handler Role" {
			t.Fatalf("expected name 'Test Handler Role', got %q", role.Name)
		}
		if role.IsSystem {
			t.Fatal("created role should not be system")
		}
	})

	t.Run("update role", func(t *testing.T) {
		body := `{"name":"Updated Handler Role","description":"Updated via test"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/hris/roles/%s", createdRoleID), bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("set role permissions", func(t *testing.T) {
		perms, _ := svc.ListPermissions(context.Background())
		if len(perms) < 2 {
			t.Skip("not enough permissions")
		}
		body := fmt.Sprintf(`{"permissionIds":["%s","%s"]}`, perms[0].ID, perms[1].ID)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/hris/roles/%s/permissions", createdRoleID), bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("set role permissions rejects invalid IDs", func(t *testing.T) {
		body := `{"permissionIds":["00000000-0000-0000-0000-000000000000"]}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/hris/roles/%s/permissions", createdRoleID), bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		// Invalid permission IDs are a client error → 400 (M-8), not a 500.
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for invalid permission IDs, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("get role permissions", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/hris/roles/%s/permissions", createdRoleID), nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp RolePermissionsResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp.PermissionIDs) != 2 {
			t.Fatalf("expected 2 permission IDs, got %d", len(resp.PermissionIDs))
		}
	})

	t.Run("get employee roles", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/hris/employees/%s/roles", empID), nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("get my permissions", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/hris/me/permissions", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("delete role", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/hris/roles/%s", createdRoleID), nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})
}

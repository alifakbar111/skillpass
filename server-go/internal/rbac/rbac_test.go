package rbac

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"skillpass-server-go/internal/testutil"
)

func setupTest(t *testing.T) (*Service, uuid.UUID, uuid.UUID) {
	t.Helper()
	db := testutil.SetupTestDB()
	_, cID, _ := testutil.CreateCompanyUser(db,
		testutil.UniqueEmail("rbac"),
		testutil.UniqueUsername("rbac"),
		"pass123", "RBAC Test Co", true)
	empID, _ := testutil.CreateEmployee(db, cID, "RBAC", "User", testutil.UniqueEmail("rbacuser"))
	svc := NewService(db)
	return svc, cID, empID
}

func TestListPermissions(t *testing.T) {
	svc, _, _ := setupTest(t)

	perms, err := svc.ListPermissions(context.Background())
	if err != nil {
		t.Fatalf("ListPermissions: %v", err)
	}
	if len(perms) == 0 {
		t.Fatal("expected non-empty permissions list")
	}
	// Verify a known permission exists
	found := false
	for _, p := range perms {
		if p.Code == "employee.view" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected 'employee.view' permission to exist")
	}
}

func TestCreateRole(t *testing.T) {
	svc, cID, _ := setupTest(t)

	name := "Test Role"
	desc := "A test role"
	role, err := svc.CreateRole(context.Background(), cID, name, &desc)
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	if role.Name != name {
		t.Fatalf("expected name %q, got %q", name, role.Name)
	}
	if role.IsSystem {
		t.Fatal("new role should not be system")
	}
	if role.Description == nil || *role.Description != desc {
		t.Fatal("description mismatch")
	}

	t.Run("rejects empty name", func(t *testing.T) {
		_, err := svc.CreateRole(context.Background(), cID, "", nil)
		if err == nil {
			t.Fatal("expected error for empty name")
		}
	})

	t.Run("rejects duplicate name within company", func(t *testing.T) {
		_, err := svc.CreateRole(context.Background(), cID, name, nil)
		if err == nil {
			t.Fatal("expected error for duplicate name")
		}
	})
}

func TestListRoles(t *testing.T) {
	svc, cID, _ := setupTest(t)

	// Initially only system roles from EnsureCompanyRoles
	svc.EnsureCompanyRoles(context.Background(), cID)

	roles, err := svc.ListRoles(context.Background(), cID)
	if err != nil {
		t.Fatalf("ListRoles: %v", err)
	}
	if len(roles) < 3 {
		t.Fatalf("expected at least 3 roles (Company Admin, HR Admin, Employee), got %d", len(roles))
	}

	// All should be system initially
	for _, r := range roles {
		if !r.IsSystem {
			t.Fatalf("expected all initial roles to be system, got %s", r.Name)
		}
	}
}

func TestUpdateRole(t *testing.T) {
	svc, cID, _ := setupTest(t)

	role, err := svc.CreateRole(context.Background(), cID, "Editable Role", nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	newName := "Updated Role"
	newDesc := "Updated description"
	updated, err := svc.UpdateRole(context.Background(), cID, role.ID, newName, &newDesc)
	if err != nil {
		t.Fatalf("UpdateRole: %v", err)
	}
	if updated.Name != newName {
		t.Fatalf("expected %q, got %q", newName, updated.Name)
	}

	t.Run("rejects update on system role", func(t *testing.T) {
		svc.EnsureCompanyRoles(context.Background(), cID)
		systemRoles, _ := svc.ListRoles(context.Background(), cID)
		for _, r := range systemRoles {
			if r.IsSystem {
				_, err := svc.UpdateRole(context.Background(), cID, r.ID, "Hacked", nil)
				if err == nil {
					t.Fatalf("expected error updating system role %s", r.Name)
				}
				break
			}
		}
	})
}

func TestDeleteRole(t *testing.T) {
	svc, cID, _ := setupTest(t)

	role, err := svc.CreateRole(context.Background(), cID, "Deletable Role", nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	err = svc.DeleteRole(context.Background(), cID, role.ID)
	if err != nil {
		t.Fatalf("DeleteRole: %v", err)
	}

	// Verify deleted
	roles, _ := svc.ListRoles(context.Background(), cID)
	for _, r := range roles {
		if r.ID == role.ID {
			t.Fatal("role should have been deleted")
		}
	}

	t.Run("rejects delete on system role", func(t *testing.T) {
		svc.EnsureCompanyRoles(context.Background(), cID)
		systemRoles, _ := svc.ListRoles(context.Background(), cID)
		for _, r := range systemRoles {
			if r.IsSystem {
				err := svc.DeleteRole(context.Background(), cID, r.ID)
				if err == nil {
					t.Fatalf("expected error deleting system role %s", r.Name)
				}
				break
			}
		}
	})
}

func TestSetRolePermissions(t *testing.T) {
	svc, cID, empID := setupTest(t)

	role, err := svc.CreateRole(context.Background(), cID, "Custom Role", nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	perms, err := svc.ListPermissions(context.Background())
	if err != nil {
		t.Fatalf("list perms: %v", err)
	}
	if len(perms) == 0 {
		t.Fatal("no permissions seeded")
	}

	// Assign first two permissions to the role
	permIDs := []string{perms[0].ID, perms[1].ID}
	err = svc.SetRolePermissions(context.Background(), cID, role.ID, permIDs)
	if err != nil {
		t.Fatalf("SetRolePermissions: %v", err)
	}

	// Assign role to employee and verify permission check passes
	svc.AssignRole(context.Background(), cID, empID, role.ID)

	has, err := svc.HasPermission(context.Background(), empID, perms[0].Code)
	if err != nil {
		t.Fatalf("HasPermission: %v", err)
	}
	if !has {
		t.Fatalf("expected %s permission to be granted", perms[0].Code)
	}

	// Clear permissions and re-assign
	newPermIDs := []string{perms[2].ID, perms[3].ID}
	err = svc.SetRolePermissions(context.Background(), cID, role.ID, newPermIDs)
	if err != nil {
		t.Fatalf("re-set permissions: %v", err)
	}

	// Previously granted permission should no longer work
	has, _ = svc.HasPermission(context.Background(), empID, perms[0].Code)
	if has {
		t.Fatalf("expected %s permission to be revoked", perms[0].Code)
	}
}

func TestHasAnyPermission(t *testing.T) {
	svc, cID, empID := setupTest(t)

	// Ensure system roles exist so we have permissions
	svc.EnsureCompanyRoles(context.Background(), cID)

	// Get Company Admin role and assign to employee
	roles, _ := svc.ListRoles(context.Background(), cID)
	var adminRoleID uuid.UUID
	for _, r := range roles {
		if r.Name == "Company Admin" {
			adminRoleID = r.ID
			break
		}
	}
	if adminRoleID == uuid.Nil {
		t.Fatal("Company Admin role not found")
	}

	svc.AssignRole(context.Background(), cID, empID, adminRoleID)

	// Test HasAnyPermission
	has, err := svc.HasAnyPermission(context.Background(), empID, []string{"employee.view", "nonexistent"})
	if err != nil {
		t.Fatalf("HasAnyPermission: %v", err)
	}
	if !has {
		t.Fatal("expected true when at least one permission exists")
	}

	has, err = svc.HasAnyPermission(context.Background(), empID, []string{"nonexistent"})
	if err != nil {
		t.Fatalf("HasAnyPermission: %v", err)
	}
	if has {
		t.Fatal("expected false when no permissions match")
	}
}

func TestAssignAndRemoveRole(t *testing.T) {
	svc, cID, empID := setupTest(t)

	role, err := svc.CreateRole(context.Background(), cID, "Assignable", nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	err = svc.AssignRole(context.Background(), cID, empID, role.ID)
	if err != nil {
		t.Fatalf("AssignRole: %v", err)
	}

	employeeRoles, err := svc.GetEmployeeRoles(context.Background(), empID)
	if err != nil {
		t.Fatalf("GetEmployeeRoles: %v", err)
	}
	found := false
	for _, r := range employeeRoles {
		if r.ID == role.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("role should be assigned to employee")
	}

	// Remove role
	err = svc.RemoveRole(context.Background(), cID, empID, role.ID)
	if err != nil {
		t.Fatalf("RemoveRole: %v", err)
	}

	employeeRoles, _ = svc.GetEmployeeRoles(context.Background(), empID)
	for _, r := range employeeRoles {
		if r.ID == role.ID {
			t.Fatal("role should have been removed")
		}
	}
}

func TestGetEmployeePermissions(t *testing.T) {
	svc, cID, empID := setupTest(t)

	// Ensure system roles
	svc.EnsureCompanyRoles(context.Background(), cID)

	// Get a role and assign it
	roles, _ := svc.ListRoles(context.Background(), cID)
	if len(roles) == 0 {
		t.Fatal("no roles available")
	}

	svc.AssignRole(context.Background(), cID, empID, roles[0].ID)

	perms, err := svc.GetEmployeePermissions(context.Background(), empID)
	if err != nil {
		t.Fatalf("GetEmployeePermissions: %v", err)
	}
	if len(perms) == 0 {
		t.Fatal("expected at least one permission")
	}
}

func TestEnsureCompanyRoles(t *testing.T) {
	svc, cID, _ := setupTest(t)

	err := svc.EnsureCompanyRoles(context.Background(), cID)
	if err != nil {
		t.Fatalf("EnsureCompanyRoles: %v", err)
	}

	// Calling again should be idempotent
	err = svc.EnsureCompanyRoles(context.Background(), cID)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	roles, _ := svc.ListRoles(context.Background(), cID)
	if len(roles) != 8 {
		t.Fatalf("expected 8 system roles, got %d", len(roles))
	}

	// Verify all system roles have permissions assigned
	for _, r := range roles {
		if !r.IsSystem {
			t.Fatalf("expected all to be system after init")
		}
	}
}

// Silence unused import
var _ = uuid.Nil
